package views

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"core/fixtures"
	"core/models"
	"proto/core"

	"github.com/golang/mock/gomock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

var _ = assert.EqualValues

func TestCards(t *testing.T) {

	mock := gomock.NewController(t)
	defer mock.Finish()

	obj := fixtures.NewMockCardRepository(mock)
	server := shopCardServer{obj}

	type shopDef struct {
		id       uint
		supplier uint
		sellers  []uint
		cards    []models.ShopCard
	}

	var (
		card1 = models.ShopCard{
			Model:  gorm.Model{ID: 1},
			Name:   "Card1",
			Number: "4255 6677 1231 5555",

			ShopID: 1,
		}
		card2 = models.ShopCard{
			Model:  gorm.Model{ID: 2},
			Name:   "Card2",
			Number: "5412 1234 9999 7120",

			ShopID: 2,
		}
		card3 = models.ShopCard{
			Model:  gorm.Model{ID: 3},
			Name:   "Card1",
			Number: "4365 3543 3456 1234",

			ShopID: 2,
		}
		card4 = models.ShopCard{
			Model:  gorm.Model{ID: 4},
			Name:   "Card1",
			Number: "3333 4444 5555 6666",

			ShopID: 2,
		}
	)

	for _, c := range []models.ShopCard{card1, card2, card3, card4} {
		var card = c

		obj.EXPECT().
			GetCardByID(card.ID).Return(&card, nil).
			MinTimes(0)
	}

	users := []*models.User{
		{
			Model: gorm.Model{ID: 1},
		},
		{
			Model: gorm.Model{ID: 2},
		},
		{
			Model: gorm.Model{ID: 4},
		},
		{
			Model:       gorm.Model{ID: 5},
			SuperSeller: true,
		},
		{
			Model: gorm.Model{ID: 6},
		},
		{
			Model: gorm.Model{ID: 15},
		},
	}

	for _, u := range users {
		var user = u
		obj.EXPECT().
			GetUserByID(user.ID).Return(user, nil).MinTimes(0)
	}

	obj.EXPECT().
		GetUserByID(3).Return(nil, errors.New("User not found")).MinTimes(0)

	for _, s := range []shopDef{
		{
			id:       1,
			supplier: 1,
			sellers:  []uint{5, 6},
			cards:    []models.ShopCard{card1},
		},
		{
			id:       2,
			supplier: 5,
			sellers:  []uint{6, 15},
			cards:    []models.ShopCard{card2, card3, card4},
		},
		{
			id:       4,
			supplier: 1,
			sellers:  []uint{16, 15},
			cards:    []models.ShopCard{},
		},
	} {

		var shop = s

		obj.EXPECT().
			GetShopSupplierID(shop.id).Return(shop.supplier, nil).
			MinTimes(0)
		obj.EXPECT().
			GetShopSellers(shop.id).Return(shop.sellers, nil).
			MinTimes(0)
		obj.EXPECT().
			GetCardsForShop(shop.id).Return(shop.cards, nil).
			MinTimes(0)
	}

	obj.EXPECT().
		GetCardsForShop(uint(42)).Return(nil, errors.New("Shop not found")).
		MinTimes(0)

	//==
	// Create tests
	//==
	{

		type createTest struct {
			shopID   uint
			req      core.CreateCardRequest
			wantSucc bool
		}

		for _, test := range []createTest{
			// success creation (user is supplier)
			{
				shopID: 1, wantSucc: true,
				req: core.CreateCardRequest{
					Card: &core.ShopCard{
						ShopId: 1,
						UserId: 1,
						Name:   "ehoho",
						Number: "ecoco",
					},
				},
			},

			// success creation (user is seller)
			{
				shopID: 2, wantSucc: true,
				req: core.CreateCardRequest{
					Card: &core.ShopCard{
						UserId: 6,
						ShopId: 2,
						Name:   "ehoho",
						Number: "ecoco",
					},
				},
			},

			// success creation (user is superseller)
			{
				shopID: 2, wantSucc: true,
				req: core.CreateCardRequest{
					Card: &core.ShopCard{
						UserId: 5,
						ShopId: 2,
						Name:   "ehoho",
						Number: "ecoco",
					},
				},
			},

			// unsuccessfull creation (user is nobody)
			{
				shopID: 2, wantSucc: false,
				req: core.CreateCardRequest{
					Card: &core.ShopCard{
						UserId: 1,
						ShopId: 2,
						Name:   "ehoho",
						Number: "ecoco",
					},
				},
			},
		} {

			var created = false
			var genID = rand.Intn(2000)

			if test.wantSucc {
				// one call for each shop if want success
				obj.EXPECT().
					CreateCard(gomock.Any()).Do(func(c *models.ShopCard) {
					assert.Equal(t, test.shopID, c.ShopID)
					created = true
					c.ID = uint(genID)
				})
			}

			res, err := server.CreateCard(context.Background(), &test.req)

			assert.True(t, (err == nil) == test.wantSucc)
			assert.EqualValues(t, test.wantSucc, created)
			if test.wantSucc {
				assert.EqualValues(t, genID, res.Id)
			}
		}

	}

	//==
	// Get tests
	//==
	{
		type getTest struct {
			req       core.GetCardsRequest
			wantCards []uint
			wantSucc  bool
		}

		for _, test := range []getTest{
			// unknown shop
			{
				wantCards: []uint{1},
				req: core.GetCardsRequest{
					ShopId: 42,
					UserId: 5,
				},
			},
			// just ok
			{
				wantCards: []uint{2, 3, 4},
				wantSucc:  true,
				req: core.GetCardsRequest{
					ShopId: 2,
					UserId: 5,
				},
			},
			// just ok -- no cards
			{
				wantCards: []uint{},
				wantSucc:  true,
				req: core.GetCardsRequest{
					ShopId: 4,
					UserId: 15,
				},
			},
		} {

			res, err := server.GetCards(context.Background(), &test.req)

			assert.True(t, (err == nil) == test.wantSucc)
			if test.wantSucc {
				if assert.Equal(t, len(res.Cards), len(test.wantCards)) {
					for _, c := range res.Cards {
						assert.Contains(t, test.wantCards, uint(c.Id))
					}
				}
			}

		}

	}

	//==
	// Delete tests
	//==
	{
		type deleteTest struct {
			req      core.DeleteCardRequest
			wantSucc bool
		}

		for i, test := range []deleteTest{
			{
				req: core.DeleteCardRequest{
					//ShopId: 1,
					UserId: 1,
					Id:     1,
				},
				wantSucc: true,
			},

			{
				req: core.DeleteCardRequest{
					//ShopId: 3,
					UserId: 15,
					Id:     1,
				},
				wantSucc: false,
			},
		} {
			var deleted = false

			if test.wantSucc {
				// one call for each shop if want success
				obj.EXPECT().
					DeleteCardByID(gomock.Any()).Do(func(id uint) {
					assert.EqualValues(t, test.req.Id, id)
					deleted = true
				})
			}

			_, err := server.DeleteCard(context.Background(), &test.req)

			assert.True(t, (err == nil) == test.wantSucc, fmt.Sprintf("Test #%v", i))
			assert.EqualValues(t, test.wantSucc, deleted)

		}

	}
}
