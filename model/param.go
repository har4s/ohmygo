package model

import (
	"github.com/har4s/ohmygo/tools/types"
)

const (
	ParamAppSettings = "settings"
)

type Param struct {
	BaseModel

	Key   string        `db:"key" json:"key"`
	Value types.JsonRaw `db:"value" json:"value"`
}

func (m *Param) TableName() string {
	return "params"
}
