package qor

import (
	"core/conf"
	"core/models"
	//"core/qor/filters"
	"core/qor/resources"
	"net/http"
	"utils/db"

	"github.com/gin-gonic/gin"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/sorting"
	"github.com/qor/transition"
	"github.com/qor/validations"

	// views; must side effect to bind servers
	_ "core/views"
	"github.com/jinzhu/gorm"
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
		&models.LeadCancelReason{},
		&models.SMSTemplate{},
		&models.PushTemplate{},
		&models.ChatTemplate{},
		&models.ChatTemplateMessage{},
		&models.EmailTemplate{},
		&transition.StateChangeLog{},
		&models.UsersProducts{},
		&models.PushToken{},
		&models.ShopNote{},
		&models.CoinsOffer{},
		&models.MonetizationPlan{},
	}
)

// Init starts the qor!
func Init(engine *gin.Engine) {
	db := db.New()
	sorting.RegisterCallbacks(db)
	validations.RegisterCallbacks(db)
	db.Callback().Query().Register("qor:load_exterals", loadExternals)
	Admin = admin.New(&qor.Config{
		DB: db,
	})

	Admin.SetSiteName(conf.AdminName)
	Admin.SetAuth(Auth{})

	Admin.AddMenu(&admin.Menu{Name: "Dashboard", Link: "/admin"})

	resources.Init(Admin)

	// attach this qor instance to gin
	mux := http.NewServeMux()

	Admin.MountTo("/qor", mux)
	engine.Any("/qor/*w", gin.WrapH(mux))
}

// invoke `LoadExternals` method
func loadExternals(scope *gorm.Scope) {
	if !scope.HasError() {
		scope.CallMethod("LoadExternals")
	}
}
