package dao

import (
	"github.com/har4s/ohmygo/dbx"
	"github.com/har4s/ohmygo/model"
	"github.com/har4s/ohmygo/tools/types"
)

// ParamQuery returns a new Param select query.
func (dao *Dao) ParamQuery() *dbx.SelectQuery {
	return dao.ModelQuery(&model.Param{})
}

// FindParamByKey finds the first Param model with the provided key.
func (dao *Dao) FindParamByKey(key string) (*model.Param, error) {
	param := &model.Param{}

	err := dao.ParamQuery().
		AndWhere(dbx.HashExp{"key": key}).
		Limit(1).
		One(param)

	if err != nil {
		return nil, err
	}

	return param, nil
}

// SaveParam creates or updates a Param model by the provided key-value pair.
// The value argument will be encoded as json string.
func (dao *Dao) SaveParam(key string, value any) error {
	param, _ := dao.FindParamByKey(key)
	if param == nil {
		param = &model.Param{Key: key}
	}

	normalizedValue := value

	encodedValue := types.JsonRaw{}
	if err := encodedValue.Scan(normalizedValue); err != nil {
		return err
	}

	param.Value = encodedValue

	return dao.Save(param)
}

// DeleteParam deletes the provided Param model.
func (dao *Dao) DeleteParam(param *model.Param) error {
	return dao.Delete(param)
}
