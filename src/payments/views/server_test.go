package views

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"payments/fixtures"
	"payments/models"
	"proto/payment"

	"github.com/golang/mock/gomock"
	"github.com/jinzhu/gorm"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type createTest struct {
	desc     string
	wantSucc bool
	request  payment.CreateOrderRequest
}

type buyTest struct {
	desc     string
	wantSucc bool
	request  payment.BuyOrderRequest
	ip       string

	pay     models.Payment
	payErr  error
	sessErr error

	finishedSess    int
	finishedSessErr error
}

var (
	repoMock *fixtures.MockRepo
	gwMock   *fixtures.MockGateway
)

func TestCreateOrder(t *testing.T) {

	mock := gomock.NewController(t)
	defer mock.Finish()

	repoMock = fixtures.NewMockRepo(mock)
	gwMock = fixtures.NewMockGateway(mock)

	server := &paymentServer{
		gateway: gwMock,
		repo:    repoMock,
	}

	for _, test := range []createTest{
		// =======================/
		{
			desc: "Simple successfull order", wantSucc: true,
			request: payment.CreateOrderRequest{
				Amount:         42,
				Currency:       0,
				LeadId:         1,
				ShopCardNumber: "4242424242424242",
			},
		},
		// =======================/
		{
			desc: "Bad cards", wantSucc: false,
			request: payment.CreateOrderRequest{
				Amount:         150,
				Currency:       0,
				LeadId:         1,
				ShopCardNumber: "4242424242112",
			},
		},
		// =======================/
		{
			desc: "Bad cards", wantSucc: false,
			request: payment.CreateOrderRequest{
				Amount:         40,
				Currency:       0,
				LeadId:         1,
				ShopCardNumber: "4012888888881880",
			},
		},
	} {
		doCreate(t, server, &test)
	}

	for _, test := range []buyTest{
		// =======================/
		{
			desc: "Simple successfull order", wantSucc: true,
			request: payment.BuyOrderRequest{
				PayId:  1,
				LeadId: 42,
			},
			pay: models.Payment{
				Model:  gorm.Model{ID: 1},
				LeadID: 42,
			},
		},
		// =======================/
		{
			desc: "Unknown pay err", wantSucc: false,
			request: payment.BuyOrderRequest{
				PayId:  4222,
				LeadId: 534234,
			},
			payErr: errors.New("wut"),
		},
		// =======================/
		{
			desc: "Gateway processing err", wantSucc: false,
			request: payment.BuyOrderRequest{
				PayId:  4322,
				LeadId: 5,
			},
			pay: models.Payment{
				Model:  gorm.Model{ID: 4322},
				LeadID: 5,
			},
			sessErr: errors.New("no sess such wow"),
		},
		// =======================/
		{
			desc: "Incorrect leadID", wantSucc: false,
			request: payment.BuyOrderRequest{
				PayId:  1,
				LeadId: 42,
			},
			pay: models.Payment{
				Model:  gorm.Model{ID: 1},
				LeadID: 43,
			},
		},
		// =======================/
		{
			desc: "Already payed", wantSucc: false,
			request: payment.BuyOrderRequest{
				PayId:  112,
				LeadId: 421,
			},
			pay: models.Payment{
				Model:  gorm.Model{ID: 112},
				LeadID: 421,
			},
			finishedSess: 1,
		},
		// =======================/

	} {
		var copy = test
		doBuy(t, server, &copy)
	}
}

func doCreate(t *testing.T, s *paymentServer, test *createTest) {

	var createdID uint

	repoMock.EXPECT().
		CreatePay(gomock.Any()).Do(func(p *models.Payment) error {

		// check pay fields
		if test.wantSucc {
			if payment.Currency_name[int32(test.request.Currency)] == "RUB" {
				assert.EqualValues(t, test.request.Amount*100, p.Amount, test.desc)
			}
			assert.EqualValues(t, test.request.ShopCardNumber, p.ShopCardNumber, test.desc)
			assert.EqualValues(t, test.request.LeadId, p.LeadID, test.desc)
		}

		p.ID = uint(rand.Intn(1000))
		createdID = p.ID

		return nil
	}).MaxTimes(1)

	res, err := s.CreateOrder(context.Background(), &test.request)
	assert.Equal(t, err == nil, test.wantSucc, test.desc)
	assert.Equal(t, res.Error == 0, test.wantSucc, test.desc)
	assert.EqualValues(t, res.Id, createdID, test.desc)
	assert.Equal(t, createdID == 0, !test.wantSucc, test.desc)
}

func doBuy(t *testing.T, s *paymentServer, test *buyTest) {

	repoMock.EXPECT().
		GetPayByID(gomock.Any()).Return(&test.pay, test.payErr).MaxTimes(1)

	var redirURL = fmt.Sprintf("https://%v/", uuid.New())

	var sess = models.Session{
		PaymentID:  test.pay.ID,
		ExternalID: uuid.New(),
		UniqueID:   uuid.New(),
		Amount:     test.pay.Amount,
		IP:         test.ip,
	}

	repoMock.EXPECT().
		FinishedSessionsForPayID(test.pay.ID).Return(test.finishedSess, test.finishedSessErr).MaxTimes(1)

	if test.payErr == nil {

		if test.sessErr == nil {
			gwMock.EXPECT().
				Buy(gomock.Any(), gomock.Any()).Return(&sess, nil).MaxTimes(1)
		} else {
			gwMock.EXPECT().
				Buy(gomock.Any(), gomock.Any()).Return(nil, test.sessErr).MaxTimes(1)
		}

		var createdID uint
		repoMock.EXPECT().
			CreateSess(gomock.Any()).Do(func(sess *models.Session) error {
			if !assert.NotNil(t, sess, test.desc) {
				return errors.New("Nil sess")
			}
			sess.ID = uint(rand.Intn(1000))
			createdID = sess.ID
			return nil
		}).MaxTimes(1)

		gwMock.EXPECT().
			Redirect(&sess).Return(redirURL).MaxTimes(1)
	}

	res, err := s.BuyOrder(context.Background(), &test.request)

	assert.Equal(t, test.wantSucc, err == nil, test.desc)
	assert.Equal(t, test.wantSucc, res.Error == 0, test.desc)

	if test.wantSucc {
		assert.EqualValues(t, res.RedirectUrl, redirURL, test.desc)
	}
}
