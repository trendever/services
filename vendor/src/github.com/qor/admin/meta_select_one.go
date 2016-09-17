package admin

import (
	"errors"
	"fmt"
	"html/template"
	"path"
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
)

// SelectOneConfig meta configuration used for select one
type SelectOneConfig struct {
	Collection               interface{} // []string, [][]string, func(interface{}, *qor.Context) [][]string, func(interface{}, *admin.Context) [][]string
	AllowBlank               bool
	SelectionTemplate        string
	SelectMode               string // select, select_async, bottom_sheet
	Select2ResultTemplate    template.JS
	Select2SelectionTemplate template.JS
	RemoteDataResource       *Resource
	metaConfig
	getCollection func(interface{}, *Context) [][]string
}

// GetTemplate get template for selection template
func (selectOneConfig SelectOneConfig) GetTemplate(context *Context, metaType string) ([]byte, error) {
	if metaType == "form" && selectOneConfig.SelectionTemplate != "" {
		return context.Asset(selectOneConfig.SelectionTemplate)
	}
	return nil, errors.New("not implemented")
}

// GetCollection get collections from select one meta
func (selectOneConfig SelectOneConfig) GetCollection(value interface{}, context *Context) [][]string {
	if selectOneConfig.getCollection != nil {
		return selectOneConfig.getCollection(value, context)
	}
	return [][]string{}
}

// ConfigureQorMeta configure select one meta
func (selectOneConfig *SelectOneConfig) ConfigureQorMeta(metaor resource.Metaor) {
	if meta, ok := metaor.(*Meta); ok {
		// Set FormattedValuer
		if meta.FormattedValuer == nil {
			meta.SetFormattedValuer(func(record interface{}, context *qor.Context) interface{} {
				return utils.Stringify(meta.GetValuer()(record, context))
			})
		}

		selectOneConfig.prepareDataSource(meta.FieldStruct, meta.baseResource, "!remote_data_selector")

		meta.Type = "select_one"
	}
}

func (selectOneConfig *SelectOneConfig) ConfigureQORAdminFilter(filter *Filter) {
	var structField *gorm.StructField
	if field, ok := filter.Resource.GetAdmin().Config.DB.NewScope(filter.Resource.Value).FieldByName(filter.Name); ok {
		structField = field.StructField
	}

	selectOneConfig.prepareDataSource(structField, filter.Resource, "!remote_data_filter")

	filter.Type = "select_one"
}

func (selectOneConfig *SelectOneConfig) FilterValue(filter *Filter, context *Context) interface{} {
	var (
		prefix  = fmt.Sprintf("filters[%v].", filter.Name)
		keyword string
	)

	if metaValues, err := resource.ConvertFormToMetaValues(context.Request, []resource.Metaor{}, prefix); err == nil {
		if metaValue := metaValues.Get("Value"); metaValue != nil {
			keyword = utils.ToString(metaValue.Value)
		}
	}

	if keyword != "" && selectOneConfig.RemoteDataResource != nil {
		result := selectOneConfig.RemoteDataResource.NewStruct()
		clone := context.Clone()
		clone.ResourceID = keyword
		selectOneConfig.RemoteDataResource.CallFindOne(result, nil, clone)
		return result
	}

	return keyword
}

func (selectOneConfig *SelectOneConfig) prepareDataSource(field *gorm.StructField, res *Resource, routePrefix string) {
	// Set GetCollection
	if selectOneConfig.Collection != nil {
		selectOneConfig.SelectMode = "select"

		if values, ok := selectOneConfig.Collection.([]string); ok {
			selectOneConfig.getCollection = func(interface{}, *Context) (results [][]string) {
				for _, value := range values {
					results = append(results, []string{value, value})
				}
				return
			}
		} else if maps, ok := selectOneConfig.Collection.([][]string); ok {
			selectOneConfig.getCollection = func(interface{}, *Context) [][]string {
				return maps
			}
		} else if fc, ok := selectOneConfig.Collection.(func(interface{}, *qor.Context) [][]string); ok {
			selectOneConfig.getCollection = func(record interface{}, context *Context) [][]string {
				return fc(record, context.Context)
			}
		} else if fc, ok := selectOneConfig.Collection.(func(interface{}, *Context) [][]string); ok {
			selectOneConfig.getCollection = fc
		} else {
			utils.ExitWithMsg("Unsupported Collection format")
		}
	}

	// Set GetCollection if normal select mode
	if selectOneConfig.getCollection == nil {
		if selectOneConfig.RemoteDataResource == nil {
			fieldType := field.Struct.Type
			for fieldType.Kind() == reflect.Ptr || fieldType.Kind() == reflect.Slice {
				fieldType = fieldType.Elem()
			}
			selectOneConfig.RemoteDataResource = res.GetAdmin().NewResource(reflect.New(fieldType).Interface())
		}

		if selectOneConfig.SelectMode == "" {
			selectOneConfig.SelectMode = "select_async"
		}

		selectOneConfig.getCollection = func(_ interface{}, context *Context) (results [][]string) {
			cloneContext := context.clone()
			cloneContext.setResource(selectOneConfig.RemoteDataResource)
			searcher := &Searcher{Context: cloneContext}
			searchResults, _ := searcher.FindMany()

			reflectValues := reflect.Indirect(reflect.ValueOf(searchResults))
			for i := 0; i < reflectValues.Len(); i++ {
				value := reflectValues.Index(i).Interface()
				scope := context.GetDB().NewScope(value)
				results = append(results, []string{fmt.Sprint(scope.PrimaryKeyValue()), utils.Stringify(value)})
			}
			return
		}
	}

	if selectOneConfig.SelectMode == "select_async" || selectOneConfig.SelectMode == "bottom_sheet" {
		if remoteDataResource := selectOneConfig.RemoteDataResource; remoteDataResource != nil {
			if remoteDataResource.params == "" {
				remoteDataResource.params = path.Join(routePrefix, res.ToParam(), field.Name)
				remoteDataSearcherController := &controller{Admin: res.GetAdmin()}
				res.GetAdmin().registerResourceToRouter(remoteDataSearcherController, remoteDataResource, "create", "update", "read", "delete")
			}
		} else {
			utils.ExitWithMsg("RemoteDataResource not configured")
		}
	}
}
