package qor

import (
	"core/conf"
	"core/db"
	"core/models"
	"core/qor/filters"
	"core/qor/resources"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/qor/activity"
	"github.com/qor/admin"
	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/database"
	"github.com/qor/qor"
	"github.com/qor/sorting"
	"github.com/qor/transition"
	"github.com/qor/validations"
	"github.com/trendever/ajaxor"

	// views; must side effect to bind servers
	_ "core/views"
)

var (
	// Admin is qor/admin instance
	Admin *admin.Admin

	// Models contains list of used db models
	Models = []interface{}{
		&models.User{},
		&models.Shop{},
		&models.ShopCard{},
		&models.Tag{},
		&models.Product{},
		&models.ProductItem{},
		&models.ImageCandidate{},
		&models.TagGroup{},
		&models.Lead{},
		&models.SMSTemplate{},
		&models.EmailTemplate{},
		&transition.StateChangeLog{},
		&activity.QorActivity{},
		&models.UsersProducts{},
	}
)

// Init starts the qor!
func Init(engine *gin.Engine) {

	Admin = admin.New(&qor.Config{
		DB: db.New(),
	})

	Admin.SetSiteName(conf.AdminName)
	Admin.SetAuth(Auth{})

	Admin.AddMenu(&admin.Menu{Name: "Dashboard", Link: "/admin"})

	//TODO: this is dirty workaround, needs to be fixed
	Admin.I18n = i18n.New(
		database.New(db.New()),
	)

	resources.Init(Admin)
	ajaxor.Init(Admin)
	filters.Init(Admin)

	sorting.RegisterCallbacks(db.New())
	validations.RegisterCallbacks(db.New())

	// attach this qor instance to gin
	mux := http.NewServeMux()

	Admin.MountTo("/qor", mux)
	engine.Any("/qor/*w", gin.WrapH(mux))
}
