package models

import (
	"fmt"
	"utils/db"
)

// variables that shall be saved between service runs
type VarData struct {
	Name  string `gorm:"primary_key"`
	Value string
}

func GetVar(name string, output interface{}) error {
	return db.New().Model(&VarData{Name: name}).Select("value").Row().Scan(output)
}

func SetVar(name string, value interface{}) error {
	data := VarData{
		Name:  name,
		Value: fmt.Sprintf("%v", value),
	}
	res := db.New().Save(&data)
	if res.Error == nil && res.RowsAffected == 0 {
		res = res.Create(&data)
	}
	return res.Error
}
