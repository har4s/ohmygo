package dao

import (
	"errors"
	"strings"
	"time"

	"github.com/har4s/ohmygo/dbx"
	"github.com/har4s/ohmygo/model"
)

const DefaultMaxFailRetries = 5

// New creates a new Dao instance with the provided db builder
// (for both async and sync db operations).
func New(db dbx.Builder) *Dao {
	return NewMultiDB(db, db)
}

// New creates a new Dao instance with the provided dedicated
// async and sync db builders.
func NewMultiDB(concurrentDB, nonconcurrentDB dbx.Builder) *Dao {
	return &Dao{
		concurrentDB:    concurrentDB,
		nonconcurrentDB: nonconcurrentDB,
	}
}

type Dao struct {
	// in a transaction both refer to the same *dbx.TX instance
	concurrentDB    dbx.Builder
	nonconcurrentDB dbx.Builder

	BeforeCreateFunc func(eventDao *Dao, m model.Model) error
	AfterCreateFunc  func(eventDao *Dao, m model.Model)
	BeforeUpdateFunc func(eventDao *Dao, m model.Model) error
	AfterUpdateFunc  func(eventDao *Dao, m model.Model)
	BeforeDeleteFunc func(eventDao *Dao, m model.Model) error
	AfterDeleteFunc  func(eventDao *Dao, m model.Model)
}

// DB returns the default dao db builder (*dbx.DB or *dbx.TX).
//
// Currently the default db builder is dao.concurrentDB but that may change in the future.
func (dao *Dao) DB() dbx.Builder {
	return dao.ConcurrentDB()
}

// ConcurrentDB returns the dao concurrent (aka. multiple open connections)
// db builder (*dbx.DB or *dbx.TX).
//
// In a transaction the concurrentDB and nonconcurrentDB refer to the same *dbx.TX instance.
func (dao *Dao) ConcurrentDB() dbx.Builder {
	return dao.concurrentDB
}

// NonconcurrentDB returns the dao nonconcurrent (aka. single open connection)
// db builder (*dbx.DB or *dbx.TX).
//
// In a transaction the concurrentDB and nonconcurrentDB refer to the same *dbx.TX instance.
func (dao *Dao) NonconcurrentDB() dbx.Builder {
	return dao.nonconcurrentDB
}

// ModelQuery creates a new query with preset Select and From fields
// based on the provided model argument.
func (dao *Dao) ModelQuery(m model.Model) *dbx.SelectQuery {
	tableName := m.TableName()
	return dao.DB().Select("{{" + tableName + "}}.*").From(tableName)
}

// FindById finds a single db record with the specified id and
// scans the result into m.
func (dao *Dao) FindById(m model.Model, id string) error {
	return dao.ModelQuery(m).Where(dbx.HashExp{"id": id}).Limit(1).One(m)
}

type afterCallGroup struct {
	Action   string
	EventDao *Dao
	Model    model.Model
}

// RunInTransaction wraps fn into a transaction.
//
// It is safe to nest RunInTransaction calls as long as you use the txDao.
func (dao *Dao) RunInTransaction(fn func(txDao *Dao) error) error {
	switch txOrDB := dao.NonconcurrentDB().(type) {
	case *dbx.Tx:
		// nested transactions are not supported by default
		// so execute the function within the current transaction
		// ---
		// create a new dao with the same hooks to avoid semaphore deadlock when nesting
		txDao := New(txOrDB)
		txDao.BeforeCreateFunc = dao.BeforeCreateFunc
		txDao.BeforeUpdateFunc = dao.BeforeUpdateFunc
		txDao.BeforeDeleteFunc = dao.BeforeDeleteFunc
		txDao.AfterCreateFunc = dao.AfterCreateFunc
		txDao.AfterUpdateFunc = dao.AfterUpdateFunc
		txDao.AfterDeleteFunc = dao.AfterDeleteFunc

		return fn(txDao)
	case *dbx.DB:
		afterCalls := []afterCallGroup{}

		txError := txOrDB.Transactional(func(tx *dbx.Tx) error {
			txDao := New(tx)

			if dao.BeforeCreateFunc != nil {
				txDao.BeforeCreateFunc = func(eventDao *Dao, m model.Model) error {
					return dao.BeforeCreateFunc(eventDao, m)
				}
			}
			if dao.BeforeUpdateFunc != nil {
				txDao.BeforeUpdateFunc = func(eventDao *Dao, m model.Model) error {
					return dao.BeforeUpdateFunc(eventDao, m)
				}
			}
			if dao.BeforeDeleteFunc != nil {
				txDao.BeforeDeleteFunc = func(eventDao *Dao, m model.Model) error {
					return dao.BeforeDeleteFunc(eventDao, m)
				}
			}

			if dao.AfterCreateFunc != nil {
				txDao.AfterCreateFunc = func(eventDao *Dao, m model.Model) {
					afterCalls = append(afterCalls, afterCallGroup{"create", eventDao, m})
				}
			}
			if dao.AfterUpdateFunc != nil {
				txDao.AfterUpdateFunc = func(eventDao *Dao, m model.Model) {
					afterCalls = append(afterCalls, afterCallGroup{"update", eventDao, m})
				}
			}
			if dao.AfterDeleteFunc != nil {
				txDao.AfterDeleteFunc = func(eventDao *Dao, m model.Model) {
					afterCalls = append(afterCalls, afterCallGroup{"delete", eventDao, m})
				}
			}

			return fn(txDao)
		})

		if txError == nil {
			// execute after event calls on successful transaction
			// (note: using the non-transaction dao to allow following queries in the after hooks)
			for _, call := range afterCalls {
				switch call.Action {
				case "create":
					dao.AfterCreateFunc(dao, call.Model)
				case "update":
					dao.AfterUpdateFunc(dao, call.Model)
				case "delete":
					dao.AfterDeleteFunc(dao, call.Model)
				}
			}
		}

		return txError
	}

	return errors.New("failed to start transaction (unknown dao.NonconcurrentDB() instance)")
}

// Delete deletes the provided model.
func (dao *Dao) Delete(m model.Model) error {
	if !m.HasId() {
		return errors.New("ID is not set")
	}

	return dao.failRetry(func(retryDao *Dao) error {
		if retryDao.BeforeDeleteFunc != nil {
			if err := retryDao.BeforeDeleteFunc(retryDao, m); err != nil {
				return err
			}
		}

		if err := retryDao.NonconcurrentDB().Model(m).Delete(); err != nil {
			return err
		}

		if retryDao.AfterDeleteFunc != nil {
			retryDao.AfterDeleteFunc(retryDao, m)
		}

		return nil
	}, DefaultMaxFailRetries)
}

// Save upserts (update or create if primary key is not set) the provided model.
func (dao *Dao) Save(m model.Model) error {
	if m.IsNew() {
		return dao.failRetry(func(retryDao *Dao) error {
			return retryDao.create(m)
		}, DefaultMaxFailRetries)
	}

	return dao.failRetry(func(retryDao *Dao) error {
		return retryDao.update(m)
	}, DefaultMaxFailRetries)
}

func (dao *Dao) update(m model.Model) error {
	if !m.HasId() {
		return errors.New("ID is not set")
	}

	if m.GetCreated().IsZero() {
		m.RefreshCreated()
	}

	m.RefreshUpdated()

	if dao.BeforeUpdateFunc != nil {
		if err := dao.BeforeUpdateFunc(dao, m); err != nil {
			return err
		}
	}

	if v, ok := any(m).(model.ColumnValueMapper); ok {
		dataMap := v.ColumnValueMap()

		_, err := dao.NonconcurrentDB().Update(
			m.TableName(),
			dataMap,
			dbx.HashExp{"id": m.GetId()},
		).Execute()

		if err != nil {
			return err
		}
	} else {
		if err := dao.NonconcurrentDB().Model(m).Update(); err != nil {
			return err
		}
	}

	if dao.AfterUpdateFunc != nil {
		dao.AfterUpdateFunc(dao, m)
	}

	return nil
}

func (dao *Dao) create(m model.Model) error {
	if !m.HasId() {
		// auto generate id
		m.RefreshId()
	}

	// mark the model as "new" since the model now always has an ID
	m.MarkAsNew()

	if m.GetCreated().IsZero() {
		m.RefreshCreated()
	}

	if m.GetUpdated().IsZero() {
		m.RefreshUpdated()
	}

	if dao.BeforeCreateFunc != nil {
		if err := dao.BeforeCreateFunc(dao, m); err != nil {
			return err
		}
	}

	if v, ok := any(m).(model.ColumnValueMapper); ok {
		dataMap := v.ColumnValueMap()
		if _, ok := dataMap["id"]; !ok {
			dataMap["id"] = m.GetId()
		}

		_, err := dao.NonconcurrentDB().Insert(m.TableName(), dataMap).Execute()
		if err != nil {
			return err
		}
	} else {
		if err := dao.NonconcurrentDB().Model(m).Insert(); err != nil {
			return err
		}
	}

	// clears the "new" model flag
	m.MarkAsNotNew()

	if dao.AfterCreateFunc != nil {
		dao.AfterCreateFunc(dao, m)
	}

	return nil
}

func (dao *Dao) failRetry(op func(retryDao *Dao) error, maxRetries int) error {
	retryDao := dao
	attempts := 1

Retry:
	if attempts == 2 {
		// assign new Dao without the before hooks to avoid triggering
		// the already fired before callbacks multiple times
		retryDao = NewMultiDB(dao.concurrentDB, dao.nonconcurrentDB)
		retryDao.AfterCreateFunc = dao.AfterCreateFunc
		retryDao.AfterUpdateFunc = dao.AfterUpdateFunc
		retryDao.AfterDeleteFunc = dao.AfterDeleteFunc
	}

	// execute
	err := op(retryDao)

	if err != nil &&
		attempts < maxRetries &&
		// note: we are checking the err message so that we can handle both the cgo and noncgo errors
		strings.Contains(err.Error(), "database is locked") {
		// wait and retry
		time.Sleep(time.Duration(200*attempts) * time.Millisecond)
		attempts++
		goto Retry
	}

	return err
}
