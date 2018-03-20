package views

import (
	"common/db"
	"common/log"
	"fmt"
	"utils/coins"
	"utils/nats"
	"utils/rpc"

	"payments/config"
	"payments/gateway"
	"payments/models"
	"proto/payment"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"proto/trendcoin"
)

const natsStream = "payments.event"

type paymentServer struct {
	gateways map[string]gateway.Gateway
	repo     models.Repo
	shed     *checkerScheduler
}

// Init starts serving
func Init() {

	coins.SetGRPCCli(trendcoin.NewTrendcoinServiceClient(rpc.Connect(config.Get().CoinsServer)))
	log.Fatal(gateway.LoadAll())

	var repo = &models.RepoImpl{DB: db.New()}

	server := &paymentServer{
		gateways: gateway.Gateways,
		repo:     repo,
	}

	server.shed = createScheduler(server)

	s := rpc.MakeServer(config.Get().RPC)
	// register API calls
	payment.RegisterPaymentServiceServer(
		s.Server,
		server,
	)
	s.StartServe()

	// register HTTP calls; for notifications
	router := gin.Default()

	router.GET("/callback", server.HandleCallback)
	router.POST("/pay/notify", server.HandleNotification)

	go router.Run(config.Get().HTTP.Listen)
	go server.PeriodicCheck()
}

// get gateway by name
func (ps *paymentServer) gw(name string) (gateway.PaymentGateway, error) {
	gw, found := ps.gateways[name]
	if !found {
		log.Debug("Gateway %v not found, available ones: %v", name, ps.gateways)
		return nil, fmt.Errorf("Unknown gateway (%v)", name)
	}

	gwCasted, ok := gw.(gateway.PaymentGateway)
	if !ok {
		return nil, fmt.Errorf("This gw is not payment gateway")
	}

	return gwCasted, nil
}

// crd gateway by name
func (ps *paymentServer) crd(name string) (gateway.CardGateway, error) {
	gw, found := ps.gateways[name]
	if !found {
		log.Debug("Gateway %v not found, available ones: %v", name, ps.gateways)
		return nil, fmt.Errorf("Unknown gateway (%v)", name)
	}

	gwCasted, ok := gw.(gateway.CardGateway)
	if !ok {
		return nil, fmt.Errorf("This gw is not payment gateway")
	}

	return gwCasted, nil
}

func (ps *paymentServer) CreateOrder(_ context.Context, req *payment.CreateOrderRequest) (*payment.CreateOrderReply, error) {
	log.Debug("create order request:\n%v", log.IndentEncode(req))

	if req.Data == nil {
		// wut the fuck
		return nil, fmt.Errorf("payments: req.data is nil")
	}

	// Step0: check if Lead already has open order
	canOpen, err := ps.repo.CanCreateOrder(req.Data.LeadId)
	if err != nil {
		return &payment.CreateOrderReply{Error: payment.Errors_DB_FAILED, ErrorMessage: err.Error()}, nil
	} else if !canOpen {
		return &payment.CreateOrderReply{
			Error:        payment.Errors_ANOTHER_OPEN_ORDER,
			ErrorMessage: fmt.Sprintf("Error! LeadId=%v has another open order", req.Data.LeadId),
		}, nil
	}

	// Step1: create order
	pay, err := models.NewPayment(req)
	if err != nil {
		return &payment.CreateOrderReply{Error: payment.Errors_INVALID_DATA, ErrorMessage: err.Error()}, nil
	}

	// Step2: Save pay
	if pay.CommissionFee != 0 {
		err := coins.CheckWriteOff(
			pay.CommissionSource, pay.CommissionFee, "payment commission", false,
			func() error {
				return ps.repo.CreatePay(pay)
			},
		)
		switch err {
		case nil:

		case coins.CallbackFailed:
			return &payment.CreateOrderReply{Error: payment.Errors_DB_FAILED, ErrorMessage: "failed to save pay"}, nil

		case coins.ServiceError:
			return &payment.CreateOrderReply{Error: payment.Errors_COINS_DOWN, ErrorMessage: err.Error()}, nil

		case coins.InsufficientFunds:
			return &payment.CreateOrderReply{Error: payment.Errors_CANT_PAY_FEE, ErrorMessage: "insufficient funds to pay commission fee"}, nil

		case coins.RefundError:
			return &payment.CreateOrderReply{Error: payment.Errors_REFUND_ERROR, ErrorMessage: "failed to refund commission fee after db error"}, nil

		default:
			return &payment.CreateOrderReply{Error: payment.Errors_UNKNOWN_ERROR, ErrorMessage: fmt.Sprintf("unexpected error happend on commission write-off: %v", err)}, nil
		}

	} else { // pay without commission
		err = ps.repo.CreatePay(pay)
		if err != nil {
			return &payment.CreateOrderReply{Error: payment.Errors_DB_FAILED, ErrorMessage: err.Error()}, nil
		}
	}

	err = nats.StanPublish(natsStream, &payment.PaymentNotification{
		Id:    uint64(pay.ID),
		Data:  pay.Encode(),
		Info:  pay.Info(),
		Event: payment.Event_Created,
	})
	if err != nil {
		return &payment.CreateOrderReply{Error: payment.Errors_NATS_FAILED, ErrorMessage: err.Error()}, nil
	}

	return &payment.CreateOrderReply{
		Id: uint64(pay.ID),
	}, nil

}

func (ps *paymentServer) AddCard(_ context.Context, req *payment.AddCardRequest) (*payment.AddCardReply, error) {

	Gateway, err := ps.crd(req.Gateway)
	if err != nil {
		return &payment.AddCardReply{Error: payment.Errors_PAY_FAILED, ErrorMessage: err.Error()}, nil
	}

	// Step1: init TX
	redirectURL, err := Gateway.Add(req.User)
	if err != nil {
		return &payment.AddCardReply{Error: payment.Errors_PAY_FAILED, ErrorMessage: err.Error()}, nil
	}

	// Step2: redirect client
	return &payment.AddCardReply{RedirectUrl: redirectURL}, nil
}

func (ps *paymentServer) GetCards(_ context.Context, req *payment.GetCardsRequest) (*payment.GetCardsReply, error) {
	Gateway, err := ps.crd(req.Gateway)
	if err != nil {
		return &payment.GetCardsReply{Error: payment.Errors_PAY_FAILED, ErrorMessage: err.Error()}, nil
	}

	// Step1: init TX
	cards, err := Gateway.GetCards(req.User)
	if err != nil {
		return &payment.GetCardsReply{Error: payment.Errors_PAY_FAILED, ErrorMessage: err.Error()}, nil
	}

	// Step2: redirect client
	return &payment.GetCardsReply{Cards: cards}, nil
}

func (ps *paymentServer) DelCard(_ context.Context, req *payment.DelCardRequest) (*payment.DelCardReply, error) {
	Gateway, err := ps.crd(req.Gateway)
	if err != nil {
		return &payment.DelCardReply{Error: payment.Errors_PAY_FAILED, ErrorMessage: err.Error()}, nil
	}

	// Step1: init TX
	err = Gateway.DelCard(req.CardId, req.User)
	if err != nil {
		return &payment.DelCardReply{Error: payment.Errors_PAY_FAILED, ErrorMessage: err.Error()}, nil
	}

	// Step2: redirect client
	return &payment.DelCardReply{}, nil
}

func (ps *paymentServer) BuyOrder(_ context.Context, req *payment.BuyOrderRequest) (*payment.BuyOrderReply, error) {
	log.Debug("buy order request:\n%v", log.IndentEncode(req))

	// Step0: find pay
	Payment, err := ps.repo.GetPayByID(req.PayId)
	if err != nil {
		return &payment.BuyOrderReply{Error: payment.Errors_DB_FAILED, ErrorMessage: err.Error()}, nil
	}

	// Step0.55: cancelled pays shall not proceed
	if Payment.Cancelled {
		return &payment.BuyOrderReply{Error: payment.Errors_PAY_CANCELLED, ErrorMessage: "Payment is cancelled, aborting"}, nil
	}

	// Step0.6: check if TX is already finished
	finished, err := ps.repo.FinishedSessionsForPayID(Payment.ID)
	if err != nil {
		return &payment.BuyOrderReply{Error: payment.Errors_DB_FAILED, ErrorMessage: err.Error()}, nil
	}
	if finished > 0 {
		return &payment.BuyOrderReply{Error: payment.Errors_ALREADY_PAYED, ErrorMessage: "payments: This pay is already payed"}, nil
	}

	Gateway, err := ps.gw(Payment.GatewayType)
	if err != nil {
		return &payment.BuyOrderReply{Error: payment.Errors_PAY_FAILED, ErrorMessage: err.Error()}, nil
	}

	// Step1: init TX
	sess, err := Gateway.Buy(Payment, req.User, false)
	if err != nil {
		return &payment.BuyOrderReply{Error: payment.Errors_PAY_FAILED, ErrorMessage: err.Error()}, nil
	}

	// Step2: save session
	err = ps.repo.CreateSess(sess)
	if err != nil {
		return &payment.BuyOrderReply{Error: payment.Errors_DB_FAILED, ErrorMessage: err.Error()}, nil
	}

	// Step3: redirect client
	return &payment.BuyOrderReply{RedirectUrl: Gateway.Redirect(sess)}, nil

}

func (ps *paymentServer) BuyAsync(ctx context.Context, req *payment.BuyAsyncRequest) (*payment.BuyAsyncReply, error) {
	log.Debug("buy async request:\n%v", log.IndentEncode(req))

	create, err := ps.CreateOrder(ctx, &payment.CreateOrderRequest{
		Data: req.Data,
	})

	if err != nil {
		return nil, err
	}

	if create.Error != payment.Errors_OK {
		return &payment.BuyAsyncReply{
			Error:        create.Error,
			ErrorMessage: create.ErrorMessage,
		}, nil
	}

	// Step0: find pay
	Payment, err := ps.repo.GetPayByID(create.Id)
	if err != nil {
		return &payment.BuyAsyncReply{Error: payment.Errors_DB_FAILED, ErrorMessage: err.Error()}, nil
	}

	Gateway, err := ps.gw(Payment.GatewayType)
	if err != nil {
		return &payment.BuyAsyncReply{Error: payment.Errors_PAY_FAILED, ErrorMessage: err.Error()}, nil
	}

	// now -- async part

	go func() {
		// Step1: init TX
		sess, err := Gateway.Buy(Payment, req.User, true)
		if err != nil {
			log.Errorf("Warning! Async Buy returnted err: %v", err)
			log.Error(nats.StanPublish(natsStream, &payment.PaymentNotification{
				Id:    uint64(Payment.ID),
				Data:  Payment.Encode(),
				Event: payment.Event_PayFailed,
				Info:  Payment.Info(),
			}))
		}

		// Step2: save session
		err = ps.repo.CreateSess(sess)
		if err != nil {
			// @TODO Though db errors should be rare, it would be real fuckup. What else could be done?
			log.Errorf("Warning! Save session returnted err, this should not happen: %v", err)
			if !sess.Success && sess.Finished {
				log.Error(nats.StanPublish(natsStream, &payment.PaymentNotification{
					Id:    uint64(sess.PaymentID),
					Data:  Payment.Encode(),
					Event: payment.Event_PayFailed,
					Info:  sess.Info(),
				}))
				return
			}
		}

		// @CHECK But... Is it really success already?
		if sess.Success && sess.Finished {
			log.Error(nats.StanPublish(natsStream, &payment.PaymentNotification{
				Id:    uint64(sess.PaymentID),
				Data:  Payment.Encode(),
				Event: payment.Event_PaySuccess,
				Info:  sess.Info(),
			}))
		}
	}()

	return &payment.BuyAsyncReply{}, nil
}

func (ps *paymentServer) CancelOrder(_ context.Context, req *payment.CancelOrderRequest) (*payment.CancelOrderReply, error) {
	log.Debug("cancel order request:\n%v", log.IndentEncode(req))

	// Step0: find pay
	pay, err := ps.repo.GetPayByID(req.PayId)
	if err != nil {
		return &payment.CancelOrderReply{Error: payment.Errors_DB_FAILED, ErrorMessage: err.Error()}, nil
	}

	if pay.Cancelled {
		return &payment.CancelOrderReply{
			Error:        payment.Errors_ALREADY_CANCELLED,
			ErrorMessage: "payments: This pay is already cancelled; why do you want to cancel it again?",
		}, nil
	}

	// Step0.6: check if TX is already finished
	finished, err := ps.repo.FinishedSessionsForPayID(pay.ID)
	if err != nil {
		return &payment.CancelOrderReply{Error: payment.Errors_DB_FAILED, ErrorMessage: err.Error()}, nil
	}
	if finished > 0 {
		return &payment.CancelOrderReply{
			Error:        payment.Errors_ALREADY_PAYED,
			ErrorMessage: "payments: This pay is already payed; why do you want to cancel it?",
		}, nil
	}

	// Step0.9: refund commission
	if pay.CommissionFee != 0 {
		text := fmt.Sprintf("refund commission of pay %v", pay.ID)
		err := coins.PostTransactions(&trendcoin.TransactionData{
			Destination:    pay.CommissionSource,
			Amount:         pay.CommissionFee,
			AllowEmptySide: true,
			Reason:         text,
			IdempotencyKey: text,
		})
		if err != nil {
			return &payment.CancelOrderReply{Error: payment.Errors_REFUND_ERROR, ErrorMessage: "failed to refund commission fee"}, nil
		}
	}

	// Step0.10: do the cancel
	pay.Cancelled = true
	err = ps.repo.SavePay(pay)
	if err != nil {
		return &payment.CancelOrderReply{
			Error:        payment.Errors_DB_FAILED,
			ErrorMessage: "payments: could not save modified pay",
		}, nil
	}

	// @TODO: shit; shit; shit. If nat is dead -- the thing will FUCK UP
	err = nats.StanPublish(natsStream, &payment.PaymentNotification{
		Id:            uint64(pay.ID),
		Data:          pay.Encode(),
		Event:         payment.Event_Cancelled,
		InvokerUserId: req.UserId,
		Info:          pay.Info(),
	})
	if err != nil {
		return &payment.CancelOrderReply{Error: payment.Errors_NATS_FAILED, ErrorMessage: err.Error()}, nil
	}

	return &payment.CancelOrderReply{Cancelled: true}, nil
}

func (ps *paymentServer) GetOrder(_ context.Context, req *payment.GetOrderRequest) (*payment.GetOrderReply, error) {

	// Step0: find pay
	pay, err := ps.repo.GetPayByID(req.Id)
	if err != nil {
		return nil, err
	}

	return &payment.GetOrderReply{
		Order: pay.Encode(),
	}, nil
}

func (ps *paymentServer) UpdateServiceData(_ context.Context, req *payment.UpdateServiceDataRequest) (*payment.UpdateServiceDataReply, error) {

	err := ps.repo.UpdateServiceData(req.Id, req.NewData)
	return &payment.UpdateServiceDataReply{}, err
}
