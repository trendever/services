package models

import "github.com/jinzhu/gorm"

const (
	typeUserLiked      = "liked"
	typeUserSaveTrend  = "savetrend"
	usersProductsTable = "users_products"
)

//UsersProducts is a model which represent link between a user and a product
type UsersProducts struct {
	gorm.Model
	UserID    uint `gorm:"index"`
	User      User
	ProductID uint `gorm:"index"`
	Product   Product
	Type      string `gorm:"index"`
}

//UserProductsRepo is a repository for manipulation with UsersProducts models
type UserProductsRepo interface {
	Like(productID, userID uint) error
	Unlike(productID, userID uint) error
}

//NewUserProductsRepo returns UserProductsRepo instance
func NewUserProductsRepo(db *gorm.DB) UserProductsRepo {
	return &userProductsRepo{
		db: db,
	}
}

type userProductsRepo struct {
	db *gorm.DB
}

func (up *userProductsRepo) Like(productID, userID uint) error {
	rel := &UsersProducts{
		ProductID: productID,
		UserID:    userID,
		Type:      typeUserLiked,
	}
	return up.addRel(rel)
}
func (up *userProductsRepo) Unlike(productID, userID uint) error {
	rel := &UsersProducts{
		ProductID: productID,
		UserID:    userID,
		Type:      typeUserLiked,
	}
	return up.removeRel(rel)
}

func (up *userProductsRepo) addRel(rel *UsersProducts) error {
	exists := &UsersProducts{}
	scope := up.db.Where(rel).Find(exists)
	if scope.RecordNotFound() {
		return up.db.Save(rel).Error
	}
	return scope.Error
}

func (up *userProductsRepo) removeRel(rel *UsersProducts) error {
	return up.db.Where(rel).Delete(&UsersProducts{}).Error
}
