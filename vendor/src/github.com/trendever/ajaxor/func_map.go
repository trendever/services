package ajaxor

import (
	"github.com/qor/admin"
	"github.com/qor/qor/utils"
	"path"
	"reflect"
)

// URLOverrider is interface for overriding links
type URLOverrider interface {

	// return Value should be used for generating links instead of interface implementer
	GetURLValue() interface{}
}

// URLForOverride is template func used to override links
//  for example, we have child resource, that point to one parent
//  and not unaccessible by direct link
// For a convinient usage, we will implement overrider on it
//  so it will return (stub) object of it's parent (and thus link)
func URLForOverride(context *admin.Context, value interface{}) string {

	if overrider, ok := value.(URLOverrider); ok {
		value = overrider.GetURLValue()
	}

	// resource has no show page defined
	if res := getResourceByValue(context, value); res == nil || len(res.ShowAttrs()) == 0 {
		return ""
	}

	return context.URLFor(value)
}

// getResourceByValue returns resource name by raw value
func getResourceNameByValue(value interface{}) string {
	// assume it's ResourceNamer -- get resource name
	if inter, ok := value.(admin.ResourceNamer); value != nil && ok {
		return inter.ResourceName()
	}

	// last resort: raw struct name
	return reflect.Indirect(reflect.ValueOf(value)).Type().String()
}

func getResourceByValue(ctx *admin.Context, value interface{}) *admin.Resource {
	return ctx.Admin.GetResource(getResourceNameByValue(value))
}

// ResourceName generates resourceName; that uses our meta model
//  model should implement admin.ResourceNamer interface
func ResourceName(meta *admin.Meta) string {
	// follow ptr && slice
	elemType := meta.FieldStruct.Struct.Type
	for elemType.Kind() == reflect.Slice || elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	// get empty struct
	value := reflect.New(elemType).Interface()

	return getResourceNameByValue(value)
}

// generate resource url
// different from qor URLFor (it appends weird prefix if it's child resource)
func ajaxorURL(context *admin.Context, res *admin.Resource, value interface{}) string {

	var (
		// main admin prefix
		prefix = res.GetAdmin().GetRouter().Prefix

		// ID of entry
		primaryKey = utils.Stringify(context.GetDB().NewScope(value).PrimaryKeyValue())
	)

	return path.Join(prefix, res.ToParam(), primaryKey)
}
