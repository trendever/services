package resources

import (
	"github.com/qor/admin"
)

// qorAdder contains slice of callbacks that should be
//  launched on qor init
var qorAdder []qorAdderFunc

type qorAdderFunc func(*admin.Admin)

// Init itializes qor resources for qor/admin
func Init(adm *admin.Admin) {
	for _, qorAdder := range qorAdder {
		qorAdder(adm)
	}
}

func addOnQorInitCallback(cb qorAdderFunc) {
	qorAdder = append(qorAdder, cb)
}
