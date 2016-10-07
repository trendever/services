package admin

import (
	"errors"
	"html/template"
	"reflect"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
)

// SelectManyConfig meta configuration used for select many
type SelectManyConfig struct {
	Collection               interface{} // []string, [][]string, func(interface{}, *qor.Context) [][]string, func(interface{}, *admin.Context) [][]string
	SelectionTemplate        string
	SelectMode               string // select, select_async, bottom_sheet
	Select2ResultTemplate    template.JS
	Select2SelectionTemplate template.JS
	RemoteDataResource       *Resource
	SelectOneConfig
}

// GetTemplate get template for selection template
func (selectManyConfig SelectManyConfig) GetTemplate(context *Context, metaType string) ([]byte, error) {
	if metaType == "form" && selectManyConfig.SelectionTemplate != "" {
		return context.Asset(selectManyConfig.SelectionTemplate)
	}
	return nil, errors.New("not implemented")
}

// ConfigureQorMeta configure select many meta
func (selectManyConfig *SelectManyConfig) ConfigureQorMeta(metaor resource.Metaor) {
	if meta, ok := metaor.(*Meta); ok {
		selectManyConfig.SelectOneConfig.Collection = selectManyConfig.Collection
		selectManyConfig.SelectOneConfig.SelectMode = selectManyConfig.SelectMode
		selectManyConfig.SelectOneConfig.RemoteDataResource = selectManyConfig.RemoteDataResource

		selectManyConfig.SelectOneConfig.ConfigureQorMeta(meta)

		selectManyConfig.RemoteDataResource = selectManyConfig.SelectOneConfig.RemoteDataResource
		meta.Type = "select_many"

		// Set FormattedValuer
		if meta.FormattedValuer == nil {
			meta.SetFormattedValuer(func(record interface{}, context *qor.Context) interface{} {
				reflectValues := reflect.Indirect(reflect.ValueOf(meta.GetValuer()(record, context)))
				var results []string
				for i := 0; i < reflectValues.Len(); i++ {
					results = append(results, utils.Stringify(reflectValues.Index(i).Interface()))
				}
				return results
			})
		}
	}
}
