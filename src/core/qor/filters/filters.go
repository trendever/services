package filters

import (
	"fmt"
	"core/db"

	"github.com/qor/admin"
	"utils/log"
)

var filters = map[*admin.Resource][]filter{}

type filter struct {
	Meta        *admin.Meta
	Type        string
	FilterName  string
	Placeholder string
}

func (f filter) InputName() string {
	return fmt.Sprintf("filters[%v]", f.FilterName)
}

// Init initializes filters module
func Init(admin *admin.Admin) {

	admin.RegisterFuncMap("filter_list", filterList)
	admin.RegisterFuncMap("filter_values", filterValues)
	admin.RegisterFuncMap("filter_value", filterValue)
}

// filterList generates filter list for displaying on webpage
func filterList(context *admin.Context) []filter {

	return filters[context.Resource]
}

// return only one value. raw
func filterValue(context *admin.Context, flt filter) string {

	return context.Request.Form.Get(flt.InputName())
}

// return filter values (load them from database by primary keys from fields)
func filterValues(context *admin.Context, flt filter) interface{} {

	filterKeys := context.Request.Form[fmt.Sprintf("filters[%v]", flt.FilterName)]

	resource := flt.Meta.Resource

	result := resource.NewSlice()

	err := db.New().
		Model(resource.Value).
		Where(fmt.Sprintf("%v in (?)", resource.PrimaryDBName()), filterKeys).
		Find(result).
		Error

	if err != nil {
		log.Debug("Error while loading value for filter: %v", err)
		return nil
	}

	return result
}

// MetaFilter creates GUI representation of filter (with type operation) for meta with supplied name
//   such a meta just reloads the page with ?filters[sql_field_name_eq]=selected_value
//   supported meta types: {ajaxor_,}select_one, date
func MetaFilter(resource *admin.Resource, metaName, operation string) {

	meta := resource.GetMetaOrNew(metaName)

	if meta == nil {
		panic(fmt.Errorf("Meta with name %v is not found in resource %v", metaName, resource.Name))
	}

	var (
		metaType    = meta.Type
		placeholder = meta.Label
		filterName  string
	)

	// pretty much hardcode, but I have no idea how to make it better at this moment
	switch meta.Type {
	case "select_one", "ajaxor_select_one",
		"select_many", "ajaxor_select_many": // @TODO: make them really searched by multiple values
		metaType = "select_one"
		filterName = fmt.Sprintf("%s_%s_%s", meta.DBName(), meta.Resource.PrimaryDBName(), operation)
	// ok raw ones
	case "date", "datetime":
		filterName = fmt.Sprintf("%s_%s", meta.DBName(), operation)
		metaType = "date"

		// make placeholder more informative
		switch operation {
		case "gt":
			placeholder = placeholder + " From"
		case "lt":
			placeholder = placeholder + " To"
		}
	default:
		panic(fmt.Errorf("Unsupported meta type %v", meta.Type))
	}

	filters[resource] = append(filters[resource], filter{
		Meta:        meta,
		FilterName:  filterName,
		Type:        metaType,
		Placeholder: placeholder,
	})

	// load custom javascripts
	resource.UseTheme("filter-workaround")
	resource.UseTheme("jquery.query-object")
}
