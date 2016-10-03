package resources

import (
	"github.com/qor/admin"
)

// qorAdder contains slice of callbacks that should be
//  launched on qor init
var qorAdder []qorAdderFunc

type resource struct {
	value  interface{}
	config *admin.Config
	res    *admin.Resource
	init   func(*admin.Resource)
}

var resources []*resource

type qorAdderFunc func(*admin.Admin)

// Init itializes qor resources for qor/admin
func Init(adm *admin.Admin) {
	for _, res := range resources {
		res.res = adm.AddResource(res.value, res.config)
	}
	for _, res := range resources {
		if res.init != nil {
			res.init(res.res)
		}
	}
}

// init func will be called after all resources will be created
func addResource(value interface{}, config *admin.Config, init func(*admin.Resource)) {
	resources = append(resources, &resource{
		value:  value,
		config: config,
		init:   init,
	})
}
