package dao

import (
	"encoding/json"
	"errors"

	"github.com/har4s/ohmygo/model"
	"github.com/har4s/ohmygo/model/settings"
)

// FindSettings returns and decode the serialized app settings param value.
//
// Returns an error if it fails to decode the stored serialized param value.
func (d *Dao) FindSettings() (*settings.Settings, error) {
	param, err := d.FindParamByKey(model.ParamAppSettings)
	if err != nil {
		return nil, err
	}

	result := settings.New()

	if err := json.Unmarshal(param.Value, result); err != nil {
		return nil, errors.New("failed to load the stored app settings")
	}

	return result, nil
}

// SaveSettings persists the specified settings configuration.
func (dao *Dao) SaveSettings(newSettings *settings.Settings) error {
	return dao.SaveParam(model.ParamAppSettings, newSettings)
}
