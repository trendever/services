package filters

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor/utils"
	"time"
	"utils/db"
	"utils/log"
)

func SetDateFilters(res *admin.Resource, field string) {
	res.UseTheme("filter-workaround")

	scope := db.New().NewScope(res.Value)
	tableName := scope.QuotedTableName()
	fieldInfo, ok := scope.FieldByName(field)
	if !ok {
		log.Error(fmt.Errorf("field %v not fount in table %v", field, tableName))
		return
	}
	dbName := fieldInfo.DBName
	label := utils.HumanizeString(fieldInfo.Name)

	for name, act := range map[string]string{"from": ">", "to": "<"} {
		op := act
		res.Filter(&admin.Filter{
			Name:  dbName + "_" + name,
			Label: label + " " + name,
			Handler: func(scope *gorm.DB, arg *admin.FilterArgument) *gorm.DB {
				metaValue := arg.Value.Get("Value")
				if metaValue == nil {
					return scope
				}
				data, ok := metaValue.Value.([]string)
				if !ok || len(data) < 1 {
					return scope
				}
				val, err := time.Parse("2006-01-02 15:04", data[0])
				if err != nil {
					log.Error(fmt.Errorf("failed to parse time in filter argument: %v", err))
					return scope
				}
				return scope.Where(fmt.Sprintf("%v.%v %v ?", tableName, dbName, op), val)
			},
			Type: "date",
		})
	}
}
