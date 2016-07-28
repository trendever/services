package views

import (
	"payments/config"
	"payments/db"
	"payments/models"
	"payments/payture"
	"proto/payment"
	"utils/rpc"

	"golang.org/x/net/context"
)

type paymentServer struct {
	payture *payture.Client
	repo    models.Repo
}

// Init starts serving
func Init() {

	// register API calls
	payment.RegisterPaymentServiceServer(
		rpc.Serve(config.Get().RPC),
		&paymentServer{
			payture.GetSandboxClient(),
			&models.RepoImpl{DB: db.New()},
		},
	)

	// register HTTP calls
	// @TODO

}

func (ps *paymentServer) CreateOrder(_ context.Context, req *payment.CreateOrderRequest) (*payment.CreateOrderReply, error) {

	// Step1: create order
	pay, err := models.NewPayment(req)
	if err != nil {
		return &payment.CreateOrderReply{Error: payment.Errors_INVALID_DATA}, err
	}

	// Step2: Save pay
	err = ps.repo.CreatePay(pay)
	if err != nil {
		return &payment.CreateOrderReply{Error: payment.Errors_DB_FAILED}, err
	}

	// Step3: Send to chat
	// @TODO

	return &payment.CreateOrderReply{
		Id: uint64(pay.ID),
	}, nil

}

func (ps *paymentServer) BuyOrder(_ context.Context, req *payment.BuyOrderRequest) (*payment.BuyOrderReply, error) {

	// Step0: find pay
	pay, err := ps.repo.GetPayByID(uint(req.PayId))
	if err != nil {
		return &payment.BuyOrderReply{Error: payment.Errors_DB_FAILED}, err
	}

	// Step1: init TX
	sess, err := ps.payture.Buy(pay, req.Ip)
	if err != nil {
		return &payment.BuyOrderReply{Error: payment.Errors_PAY_FAILED}, err
	}

	// Step2: save session
	err = ps.repo.CreateSess(sess)
	if err != nil {
		return &payment.BuyOrderReply{Error: payment.Errors_DB_FAILED}, err
	}

	// Step3: redirect client
	return &payment.BuyOrderReply{RedirectUrl: ps.payture.Redirect(sess)}, nil

}
