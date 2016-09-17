package admin

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
)

// Filter register filter for qor resource
func (res *Resource) Filter(filter *Filter) {
	filter.Resource = res

	if filter.Label == "" {
		filter.Label = utils.HumanizeString(filter.Name)
	}

	if filter.Config != nil {
		filter.Config.ConfigureQORAdminFilter(filter)
	}

	if filter.Handler == nil {
		// generate default handler
		filter.Handler = func(db *gorm.DB, filterArgument *FilterArgument) *gorm.DB {
			if metaValue := filterArgument.Value.Get("Value"); metaValue != nil {
				return defaultFieldFilter(res, []string{filter.Name}, utils.ToString(metaValue.Value), db)
			}
			return db
		}
	}

	if filter.Type != "" {
		res.filters = append(res.filters, filter)
	} else {
		utils.ExitWithMsg("Invalid filter definition %v for resource %v", filter.Name, res.Name)
	}
}

func (res *Resource) GetFilters() []*Filter {
	return res.filters
}

// Filter filter definiation
type Filter struct {
	Name     string
	Label    string
	Type     string
	Resource *Resource
	Handler  func(*gorm.DB, *FilterArgument) *gorm.DB
	Config   FilterConfigInterface
}

// FilterConfigInterface filter config interface
type FilterConfigInterface interface {
	ConfigureQORAdminFilter(*Filter)
}

// FilterArgument filter argument that used in handler
type FilterArgument struct {
	Value    *resource.MetaValues
	Resource *Resource
	Context  *qor.Context
}
