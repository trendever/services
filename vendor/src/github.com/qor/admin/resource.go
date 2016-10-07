package admin

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
)

// Resource is the most important thing for qor admin, every model is defined as a resource, qor admin will genetate management interface based on its definition
type Resource struct {
	resource.Resource
	Config        *Config
	Metas         []*Meta
	Actions       []*Action
	SearchHandler func(keyword string, context *qor.Context) *gorm.DB

	admin          *Admin
	params         string
	base           *Resource
	scopes         []*Scope
	filters        []*Filter
	searchAttrs    *[]string
	sortableAttrs  *[]string
	indexSections  []*Section
	newSections    []*Section
	editSections   []*Section
	showSections   []*Section
	isSetShowAttrs bool
	cachedMetas    *map[string][]*Meta
}

// Meta register meta for admin resource
func (res *Resource) Meta(meta *Meta) *Meta {
	if res.GetMeta(meta.Name) != nil {
		utils.ExitWithMsg("Duplicated meta %v defined for resource %v", meta.Name, res.Name)
	}
	res.Metas = append(res.Metas, meta)
	meta.baseResource = res
	meta.updateMeta()
	return meta
}

// GetAdmin get admin from resource
func (res Resource) GetAdmin() *Admin {
	return res.admin
}

// GetPrimaryValue get priamry value from request
func (res Resource) GetPrimaryValue(request *http.Request) string {
	if request != nil {
		return request.URL.Query().Get(res.ParamIDName())
	}
	return ""
}

// ParamIDName return param name for primary key like :product_id
func (res Resource) ParamIDName() string {
	return fmt.Sprintf(":%v_id", inflection.Singular(utils.ToParamString(res.Name)))
}

// ToParam used as urls to register routes for resource
func (res *Resource) ToParam() string {
	if res.params == "" {
		if value, ok := res.Value.(interface {
			ToParam() string
		}); ok {
			res.params = value.ToParam()
		} else {
			if res.Config.Singleton == true {
				res.params = utils.ToParamString(res.Name)
			}
			res.params = utils.ToParamString(inflection.Plural(res.Name))
		}
	}
	return res.params
}

// UseTheme use them for resource, will auto load the theme's javascripts, stylesheets for this resource
func (res *Resource) UseTheme(theme interface{}) []ThemeInterface {
	var themeInterface ThemeInterface
	if ti, ok := theme.(ThemeInterface); ok {
		themeInterface = ti
	} else if str, ok := theme.(string); ok {
		themeInterface = Theme{Name: str}
	}

	if themeInterface != nil {
		res.Config.Themes = append(res.Config.Themes, themeInterface)

		// Config Admin Theme
		for _, pth := range themeInterface.GetViewPaths() {
			res.GetAdmin().RegisterViewPath(pth)
		}
		themeInterface.ConfigAdminTheme(res)
	}
	return res.Config.Themes
}

// GetTheme get registered theme with name
func (res *Resource) GetTheme(name string) ThemeInterface {
	for _, theme := range res.Config.Themes {
		if theme.GetName() == name {
			return theme
		}
	}
	return nil
}

// Decode decode context into a value
func (res *Resource) Decode(context *qor.Context, value interface{}) error {
	return resource.Decode(context, value, res)
}

func (res *Resource) convertObjectToJSONMap(context *Context, value interface{}, kind string) interface{} {
	reflectValue := reflect.ValueOf(value)
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}

	switch reflectValue.Kind() {
	case reflect.Slice:
		values := []interface{}{}
		for i := 0; i < reflectValue.Len(); i++ {
			if reflectValue.Index(i).Kind() == reflect.Ptr {
				values = append(values, res.convertObjectToJSONMap(context, reflectValue.Index(i).Interface(), kind))
			} else {
				values = append(values, res.convertObjectToJSONMap(context, reflectValue.Index(i).Addr().Interface(), kind))
			}
		}
		return values
	case reflect.Struct:
		var metas []*Meta
		if kind == "index" {
			metas = res.ConvertSectionToMetas(res.allowedSections(res.IndexAttrs(), context, roles.Update))
		} else if kind == "edit" {
			metas = res.ConvertSectionToMetas(res.allowedSections(res.EditAttrs(), context, roles.Update))
		} else if kind == "show" {
			metas = res.ConvertSectionToMetas(res.allowedSections(res.ShowAttrs(), context, roles.Read))
		}

		values := map[string]interface{}{}
		for _, meta := range metas {
			if meta.HasPermission(roles.Read, context.Context) {
				if meta.Resource != nil && (meta.FieldStruct != nil && meta.FieldStruct.Relationship != nil && (meta.FieldStruct.Relationship.Kind == "has_one" || meta.FieldStruct.Relationship.Kind == "has_many")) {
					values[meta.GetName()] = meta.Resource.convertObjectToJSONMap(context, context.RawValueOf(value, meta), kind)
				} else {
					values[meta.GetName()] = context.FormattedValueOf(value, meta)
				}
			}
		}
		return values
	default:
		return value
	}
}

func (res *Resource) allAttrs() []string {
	var attrs []string
	scope := &gorm.Scope{Value: res.Value}

Fields:
	for _, field := range scope.GetModelStruct().StructFields {
		for _, meta := range res.Metas {
			if field.Name == meta.FieldName {
				attrs = append(attrs, meta.Name)
				continue Fields
			}
		}

		if field.IsForeignKey {
			continue
		}

		for _, value := range []string{"CreatedAt", "UpdatedAt", "DeletedAt"} {
			if value == field.Name {
				continue Fields
			}
		}

		if (field.IsNormal || field.Relationship != nil) && !field.IsIgnored {
			attrs = append(attrs, field.Name)
			continue
		}

		fieldType := field.Struct.Type
		for fieldType.Kind() == reflect.Ptr || fieldType.Kind() == reflect.Slice {
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct {
			attrs = append(attrs, field.Name)
		}
	}

MetaIncluded:
	for _, meta := range res.Metas {
		for _, attr := range attrs {
			if attr == meta.FieldName || attr == meta.Name {
				continue MetaIncluded
			}
		}
		attrs = append(attrs, meta.Name)
	}

	return attrs
}

func (res *Resource) getAttrs(attrs []string) []string {
	if len(attrs) == 0 {
		return res.allAttrs()
	}

	var onlyExcludeAttrs = true
	for _, attr := range attrs {
		if !strings.HasPrefix(attr, "-") {
			onlyExcludeAttrs = false
			break
		}
	}

	if onlyExcludeAttrs {
		return append(res.allAttrs(), attrs...)
	}
	return attrs
}

// IndexAttrs set attributes will be shown in the index page
//     // show given attributes in the index page
//     order.IndexAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")
//     // show all attributes except `State` in the index page
//     order.IndexAttrs("-State")
func (res *Resource) IndexAttrs(values ...interface{}) []*Section {
	res.setSections(&res.indexSections, values...)
	res.SearchAttrs()
	return res.indexSections
}

// NewAttrs set attributes will be shown in the new page
//     // show given attributes in the new page
//     order.NewAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")
//     // show all attributes except `State` in the new page
//     order.NewAttrs("-State")
//  You could also use `Section` to structure form to make it tidy and clean
//     product.NewAttrs(
//       &admin.Section{
//       	Title: "Basic Information",
//       	Rows: [][]string{
//       		{"Name"},
//       		{"Code", "Price"},
//       	}},
//       &admin.Section{
//       	Title: "Organization",
//       	Rows: [][]string{
//       		{"Category", "Collections", "MadeCountry"},
//       	}},
//       "Description",
//       "ColorVariations",
//     }
func (res *Resource) NewAttrs(values ...interface{}) []*Section {
	res.setSections(&res.newSections, values...)
	return res.newSections
}

// EditAttrs set attributes will be shown in the edit page
//     // show given attributes in the new page
//     order.EditAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")
//     // show all attributes except `State` in the edit page
//     order.EditAttrs("-State")
//  You could also use `Section` to structure form to make it tidy and clean
//     product.EditAttrs(
//       &admin.Section{
//       	Title: "Basic Information",
//       	Rows: [][]string{
//       		{"Name"},
//       		{"Code", "Price"},
//       	}},
//       &admin.Section{
//       	Title: "Organization",
//       	Rows: [][]string{
//       		{"Category", "Collections", "MadeCountry"},
//       	}},
//       "Description",
//       "ColorVariations",
//     }
func (res *Resource) EditAttrs(values ...interface{}) []*Section {
	res.setSections(&res.editSections, values...)
	return res.editSections
}

// ShowAttrs set attributes will be shown in the show page
//     // show given attributes in the show page
//     order.ShowAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")
//     // show all attributes except `State` in the show page
//     order.ShowAttrs("-State")
//  You could also use `Section` to structure form to make it tidy and clean
//     product.ShowAttrs(
//       &admin.Section{
//       	Title: "Basic Information",
//       	Rows: [][]string{
//       		{"Name"},
//       		{"Code", "Price"},
//       	}},
//       &admin.Section{
//       	Title: "Organization",
//       	Rows: [][]string{
//       		{"Category", "Collections", "MadeCountry"},
//       	}},
//       "Description",
//       "ColorVariations",
//     }
func (res *Resource) ShowAttrs(values ...interface{}) []*Section {
	if len(values) > 0 {
		if values[len(values)-1] == false {
			values = values[:len(values)-1]
		} else {
			res.isSetShowAttrs = true
		}
	}
	res.setSections(&res.showSections, values...)
	return res.showSections
}

// SortableAttrs set sortable attributes, sortable attributes could be click to order in qor table
func (res *Resource) SortableAttrs(columns ...string) []string {
	if len(columns) != 0 || res.sortableAttrs == nil {
		if len(columns) == 0 {
			columns = res.ConvertSectionToStrings(res.indexSections)
		}
		res.sortableAttrs = &[]string{}
		scope := res.GetAdmin().Config.DB.NewScope(res.Value)
		for _, column := range columns {
			if field, ok := scope.FieldByName(column); ok && field.DBName != "" {
				attrs := append(*res.sortableAttrs, column)
				res.sortableAttrs = &attrs
			}
		}
	}
	return *res.sortableAttrs
}

// SearchAttrs set search attributes, when search resources, will use those columns to search
//     // Search products with its name, code, category's name, brand's name
//	   product.SearchAttrs("Name", "Code", "Category.Name", "Brand.Name")
func (res *Resource) SearchAttrs(columns ...string) []string {
	if len(columns) != 0 || res.searchAttrs == nil {
		if len(columns) == 0 {
			columns = res.ConvertSectionToStrings(res.indexSections)
		}

		if len(columns) > 0 {
			res.searchAttrs = &columns
			res.SearchHandler = func(keyword string, context *qor.Context) *gorm.DB {
				return defaultFieldFilter(res, columns, keyword, context.GetDB(), context)
			}
		}
	}

	return columns
}

func (res *Resource) getCachedMetas(cacheKey string, fc func() []resource.Metaor) []*Meta {
	if res.cachedMetas == nil {
		res.cachedMetas = &map[string][]*Meta{}
	}

	if values, ok := (*res.cachedMetas)[cacheKey]; ok {
		return values
	}

	values := fc()
	var metas []*Meta
	for _, value := range values {
		metas = append(metas, value.(*Meta))
	}
	(*res.cachedMetas)[cacheKey] = metas
	return metas
}

// GetMetas get metas with give attrs
func (res *Resource) GetMetas(attrs []string) []resource.Metaor {
	if len(attrs) == 0 {
		attrs = res.allAttrs()
	}
	var showSections, ignoredAttrs []string
	for _, attr := range attrs {
		if strings.HasPrefix(attr, "-") {
			ignoredAttrs = append(ignoredAttrs, strings.TrimLeft(attr, "-"))
		} else {
			showSections = append(showSections, attr)
		}
	}

	primaryKey := res.PrimaryFieldName()

	metas := []resource.Metaor{}

Attrs:
	for _, attr := range showSections {
		for _, a := range ignoredAttrs {
			if attr == a {
				continue Attrs
			}
		}

		var meta *Meta
		for _, m := range res.Metas {
			if m.GetName() == attr {
				meta = m
				break
			}
		}

		if meta == nil {
			meta = &Meta{Name: attr, baseResource: res}
			if attr == primaryKey {
				meta.Type = "hidden"
			}
			meta.updateMeta()
		}

		metas = append(metas, meta)
	}

	return metas
}

// GetMeta get meta with name
func (res *Resource) GetMeta(name string) *Meta {
	for _, meta := range res.Metas {
		if meta.Name == name || meta.GetFieldName() == name {
			return meta
		}
	}
	return nil
}

// GetMetaOrNew get meta or initalize a new one
func (res *Resource) GetMetaOrNew(name string) *Meta {
	if meta := res.GetMeta(name); meta != nil {
		return meta
	}

	if field, ok := res.GetAdmin().Config.DB.NewScope(res.Value).FieldByName(name); ok {
		meta := &Meta{Name: name, baseResource: res}
		if field.IsPrimaryKey {
			meta.Type = "hidden"
		}
		meta.updateMeta()
		res.Metas = append(res.Metas, meta)
		return meta
	}

	return nil
}

func (res *Resource) allowedSections(sections []*Section, context *Context, roles ...roles.PermissionMode) []*Section {
	var newSections []*Section
	for _, section := range sections {
		newSection := Section{Resource: section.Resource, Title: section.Title}
		var editableRows [][]string
		for _, row := range section.Rows {
			var editableColumns []string
			for _, column := range row {
				for _, role := range roles {
					meta := res.GetMetaOrNew(column)
					if meta != nil && meta.HasPermission(role, context.Context) {
						editableColumns = append(editableColumns, column)
						break
					}
				}
			}
			if len(editableColumns) > 0 {
				editableRows = append(editableRows, editableColumns)
			}
		}
		newSection.Rows = editableRows
		newSections = append(newSections, &newSection)
	}
	return newSections
}

func (res *Resource) configure() {
	modelType := res.GetAdmin().Config.DB.NewScope(res.Value).GetModelStruct().ModelType
	for i := 0; i < modelType.NumField(); i++ {
		if fieldStruct := modelType.Field(i); fieldStruct.Anonymous {
			if injector, ok := reflect.New(fieldStruct.Type).Interface().(resource.ConfigureResourceInterface); ok {
				injector.ConfigureQorResource(res)
			}
		}
	}

	if injector, ok := res.Value.(resource.ConfigureResourceInterface); ok {
		injector.ConfigureQorResource(res)
	}
}

func defaultFieldFilter(res *Resource, columns []string, keyword string, db *gorm.DB, context *qor.Context) *gorm.DB {
	var (
		joinConditionsMap  = map[string][]string{}
		conditions         []string
		keywords           []interface{}
		generateConditions func(column string, scope *gorm.Scope)
	)

	generateConditions = func(column string, scope *gorm.Scope) {
		currentScope, nextScope := scope, scope
		if strings.Contains(column, ".") {
			for _, field := range strings.Split(column, ".") {
				column = field
				currentScope = nextScope
				if field, ok := currentScope.FieldByName(field); ok {
					if relationship := field.Relationship; relationship != nil {
						nextScope = currentScope.New(reflect.New(field.Field.Type()).Interface())
						key := fmt.Sprintf("LEFT JOIN %v ON", nextScope.QuotedTableName())

						for index := range relationship.ForeignDBNames {
							if relationship.Kind == "has_one" || relationship.Kind == "has_many" {
								joinConditionsMap[key] = append(joinConditionsMap[key],
									fmt.Sprintf("%v.%v = %v.%v",
										nextScope.QuotedTableName(), scope.Quote(relationship.ForeignDBNames[index]),
										currentScope.QuotedTableName(), scope.Quote(relationship.AssociationForeignDBNames[index]),
									))
							} else if relationship.Kind == "belongs_to" {
								joinConditionsMap[key] = append(joinConditionsMap[key],
									fmt.Sprintf("%v.%v = %v.%v",
										currentScope.QuotedTableName(), scope.Quote(relationship.ForeignDBNames[index]),
										nextScope.QuotedTableName(), scope.Quote(relationship.AssociationForeignDBNames[index]),
									))
							}
						}
					}
				}
			}
		}

		tableName := currentScope.QuotedTableName()
		if field, ok := currentScope.FieldByName(column); ok {
			if field.IsNormal {
				switch field.Field.Kind() {
				case reflect.String:
					conditions = append(conditions, fmt.Sprintf("upper(%v.%v) like upper(?)", tableName, scope.Quote(field.DBName)))
					keywords = append(keywords, "%"+keyword+"%")
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					if _, err := strconv.Atoi(keyword); err == nil {
						conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
						keywords = append(keywords, keyword)
					}
				case reflect.Float32, reflect.Float64:
					if _, err := strconv.ParseFloat(keyword, 64); err == nil {
						conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
						keywords = append(keywords, keyword)
					}
				case reflect.Bool:
					if value, err := strconv.ParseBool(keyword); err == nil {
						conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
						keywords = append(keywords, value)
					}
				case reflect.Struct:
					// time ?
					if _, ok := field.Field.Interface().(time.Time); ok {
						if parsedTime, err := utils.ParseTime(keyword, context); err == nil {
							conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
							keywords = append(keywords, parsedTime)
						}
					}
				case reflect.Ptr:
					// time ?
					if _, ok := field.Field.Interface().(*time.Time); ok {
						if parsedTime, err := utils.ParseTime(keyword, context); err == nil {
							conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
							keywords = append(keywords, parsedTime)
						}
					}
				default:
					conditions = append(conditions, fmt.Sprintf("%v.%v = ?", tableName, scope.Quote(field.DBName)))
					keywords = append(keywords, keyword)
				}
			} else if relationship := field.Relationship; relationship != nil {
				switch relationship.Kind {
				case "select_one", "select_many":
					for _, foreignFieldName := range relationship.ForeignFieldNames {
						generateConditions(strings.Join([]string{field.Name, foreignFieldName}, "."), currentScope)
					}
				case "belongs_to":
					for _, foreignFieldName := range relationship.ForeignFieldNames {
						generateConditions(foreignFieldName, currentScope)
					}
				case "many_to_many":
					// TODO
					panic("not supported")
				}
			}
		}
	}

	scope := db.NewScope(res.Value)
	for _, column := range columns {
		generateConditions(column, scope)
	}

	// join conditions
	if len(joinConditionsMap) > 0 {
		var joinConditions []string
		for key, values := range joinConditionsMap {
			joinConditions = append(joinConditions, fmt.Sprintf("%v %v", key, strings.Join(values, " AND ")))
		}
		db = db.Joins(strings.Join(joinConditions, " "))
	}

	// search conditions
	if len(conditions) > 0 {
		return db.Where(strings.Join(conditions, " OR "), keywords...)
	}

	return db
}
