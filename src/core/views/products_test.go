package views

import (
	"core/db"
	"core/models"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"proto/core"
	"testing"
	"time"
	"utils/log"
)

func TestMain(t *testing.T) {
	log.Init(true, "Tests")
	db.Init()
}

//@CHECK: is this really necessary?
func lastID() uint {
	product := models.Product{}

	err := db.
		New().
		Order("id desc").
		First(&product).
		Error

	if err != nil {
		return 0
	}
	return product.ID + 1
}

// Insert test product, recieve it and delete
func TestRetrieveProduct(t *testing.T) {

	DB := db.New()

	defer func() {
		DB.Delete(models.Product{}, "code like ?", "Tst%")
		// gorm uses soft-delete, but we don't want to save these entries
		DB.Unscoped().Delete(models.Product{}, "title like ?", "Tst%")
	}()

	prod := models.Product{
		Title: "Test",
		Code:  "Tst01",
		Model: gorm.Model{ID: lastID()},
	}

	err := DB.Create(&prod).Error
	if err != nil {
		log.Error(err)
		return
	}

	// Test #1: by id
	result, err := productServer{}.GetProduct(context.Background(), &core.GetProductRequest{
		SearchBy: &core.GetProductRequest_Id{Id: int64(prod.ID)},
	})

	if !assert.Nil(t, err) || !assert.Equal(t, 1, len(result.Result), "Got incorrect number of results") {
		return
	}

	assert.EqualValues(t, prod.ID, result.Result[0].Id, "Id is different")

	// Test #2: offset && limit

	// adding 20 entries in case DB is empty

	last := lastID()
	for i := 0; i < 20; i++ {
		pr := models.Product{
			Title: fmt.Sprintf("Tst%02v", i),
			//make sure they are sorted
			Model: gorm.Model{ID: last + uint(i)},
		}
		pr.CreatedAt = time.Now().Add(time.Second * time.Duration(i))

		err := DB.Create(&pr).Error
		if !assert.Nil(t, err, "Entry creation error") {
			return
		}

	}

	tst := [][]interface{}{
		// limit - offset - el.id - el.title
		{10, 10, 5, "Tst14"},
		{5, 5, 0, "Tst04"},
	}

	for i, suite := range tst {

		limit := suite[0].(int)
		offset := suite[1].(int)

		result, err := productServer{}.SearchProducts(context.Background(), &core.SearchProductRequest{
			Limit:  int64(limit),
			Offset: int64(offset),
		})

		if !assert.Nil(t, err) ||
			!assert.EqualValues(t, limit, len(result.Result), fmt.Sprintf("Incorrect len #%v", i)) {
			return
		}

		id := suite[2].(int)
		title := suite[3].(string)
		assert.EqualValues(t, title, result.Result[id].Title, "Title is different")
	}
}
