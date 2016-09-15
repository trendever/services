package views

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"utils/db"
	"utils/log"
	"utils/rpc"

	"payments/api"
	"payments/config"
	"payments/models"
	"payments/payture"
	"proto/payment"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

type paymentServer struct {
	gateway models.Gateway
	repo    models.Repo
	chat    api.ChatNotifier
	shed    *checkerScheduler
}

// Init starts serving
func Init() {

	api.Init()

	var repo = &models.RepoImpl{DB: db.New()}

	server := &paymentServer{
		gateway: payture.GetClient(),
		repo:    repo,
		chat:    api.GetChatNotifier(repo),
	}

	server.shed = createScheduler(server)

	// register API calls
	payment.RegisterPaymentServiceServer(
		rpc.Serve(config.Get().RPC),
		server,
	)

	// register HTTP calls; for notifications
	router := gin.Default()

	router.GET("/callback", server.HandleCallback)
	router.POST("/pay/notify", server.HandleNotification)

	go router.Run(config.Get().HTTP.Listen)
	go server.PeriodicCheck()
}

func (ps *paymentServer) HandleCallback(c *gin.Context) {
	orderID := c.Query("orderid")
	success, _ := strconv.ParseBool(c.Query("success"))

	//find session
	sess, err := ps.repo.GetSessByUID(orderID)
	if err != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf(config.Get().HTTP.Redirect, 0, false))
		return
	}

	// we want to redirect client NOW
	// status will be reported afterwards by chat message
	go func() {
		err := ps.CheckStatus(sess)
		if err != nil {
			log.Error(err)
		}
	}()

	c.Redirect(http.StatusSeeOther, fmt.Sprintf(config.Get().HTTP.Redirect, sess.Payment.LeadID, success))
}

func (ps *paymentServer) HandleNotification(c *gin.Context) {

	orderID := c.PostForm("OrderId")
	if orderID == "" {
		log.Debug("Skipping notification event without OrderId")
		return
	}

	// avoid time attacks
	go func() {

		log.Debug("Got notification event for order=%v; checking", orderID)
		defer log.Debug("Finished notification checking for order=%v;", orderID)

		//find session
		sess, err := ps.repo.GetSessByUID(orderID)
		if err != nil {
			log.Error(err)
			return
		}

		err = ps.CheckStatus(sess)
		if err != nil {
			log.Error(err)
		}

	}()
}

func (ps *paymentServer) PeriodicCheck() {
	for {
		log.Debug("Starting PeriodicCheck")
		ps.CheckStatuses()
		log.Debug("Finished PeriodicCheck")

		<-time.After(time.Second * time.Duration(config.Get().PeriodicCheck))
	}
}

func (ps *paymentServer) CreateOrder(_ context.Context, req *payment.CreateOrderRequest) (*payment.CreateOrderReply, error) {

	// Step1: create order
	pay, err := models.NewPayment(req)
	if err != nil {
		return &payment.CreateOrderReply{Error: payment.Errors_INVALID_DATA, ErrorMessage: err.Error()}, nil
	}

	// Step2: Save pay
	err = ps.repo.CreatePay(pay)
	if err != nil {
		return &payment.CreateOrderReply{Error: payment.Errors_DB_FAILED, ErrorMessage: err.Error()}, nil
	}

	// Step3: Send to chat
	go func() {
		err := ps.chat.SendPayment(pay)
		if err != nil {
			log.Error(err)
		}
	}()

	return &payment.CreateOrderReply{
		Id: uint64(pay.ID),
	}, nil

}

func (ps *paymentServer) BuyOrder(_ context.Context, req *payment.BuyOrderRequest) (*payment.BuyOrderReply, error) {

	// Step0: find pay
	pay, err := ps.repo.GetPayByID(uint(req.PayId))
	if err != nil {
		return &payment.BuyOrderReply{Error: payment.Errors_DB_FAILED, ErrorMessage: err.Error()}, nil
	}

	// Step0.5: check if saved pay parameters are equal to supplied
	if req.LeadId != pay.LeadID || pay.LeadID == 0 || int32(req.Direction) != pay.Direction {
		return &payment.BuyOrderReply{Error: payment.Errors_INVALID_DATA, ErrorMessage: fmt.Sprintf("Access denied: supplied incorrect LeadID (%v)", req.LeadId)}, nil
	}

	// Step0.55: cancelled pays shall not proceed
	if pay.Cancelled {
		return &payment.BuyOrderReply{Error: payment.Errors_PAY_CANCELLED, ErrorMessage: fmt.Sprintf("Payment is cancelled, aborting")}, nil
	}

	// Step0.6: check if TX is already finished
	finished, err := ps.repo.FinishedSessionsForPayID(pay.ID)
	if err != nil {
		return &payment.BuyOrderReply{Error: payment.Errors_DB_FAILED, ErrorMessage: err.Error()}, nil
	}
	if finished > 0 {
		return &payment.BuyOrderReply{Error: payment.Errors_ALREADY_PAYED, ErrorMessage: fmt.Sprintf("payments: This pay is already payed")}, nil
	}

	// Step1: init TX
	sess, err := ps.gateway.Buy(pay, req.Ip)
	if err != nil {
		return &payment.BuyOrderReply{Error: payment.Errors_PAY_FAILED, ErrorMessage: err.Error()}, nil
	}

	// Step2: save session
	err = ps.repo.CreateSess(sess)
	if err != nil {
		return &payment.BuyOrderReply{Error: payment.Errors_DB_FAILED, ErrorMessage: err.Error()}, nil
	}

	// Step3: redirect client
	return &payment.BuyOrderReply{RedirectUrl: ps.gateway.Redirect(sess)}, nil

}

func (ps *paymentServer) CancelOrder(_ context.Context, req *payment.CancelOrderRequest) (*payment.CancelOrderReply, error) {

	// Step0: find pay
	pay, err := ps.repo.GetPayByID(uint(req.PayId))
	if err != nil {
		return &payment.CancelOrderReply{Error: payment.Errors_DB_FAILED, ErrorMessage: err.Error()}, nil
	}

	// Step0.5: check if saved pay parameters are equal to supplied

	// allow both sides cancel
	// we have only 2 direction and this may seem useless; but it is planned to have other people in conversation
	// so I will disable checks only for new payment sides
	switch req.Direction {
	case payment.Direction_CLIENT_PAYS, payment.Direction_CLIENT_RECV:
		if pay.Direction != int32(payment.Direction_CLIENT_PAYS) && pay.Direction != int32(payment.Direction_CLIENT_RECV) {
			return &payment.CancelOrderReply{
				Error:        payment.Errors_INVALID_DATA,
				ErrorMessage: fmt.Sprintf("Access denied: you are not the side of payment who can cancel it (%v)", req.LeadId),
			}, nil
		}
	default:
		if int32(req.Direction) != pay.Direction {
			return &payment.CancelOrderReply{
				Error:        payment.Errors_INVALID_DATA,
				ErrorMessage: fmt.Sprintf("Access denied: you are not the side of payment who can cancel it (%v)", req.LeadId),
			}, nil
		}
	}

	// checks to make sure leadID is not mangled
	if req.LeadId != pay.LeadID || pay.LeadID == 0 {
		return &payment.CancelOrderReply{
			Error:        payment.Errors_INVALID_DATA,
			ErrorMessage: fmt.Sprintf("Access denied: supplied incorrect LeadID (%v)", req.LeadId),
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
			ErrorMessage: fmt.Sprintf("payments: This pay is already payed; why do you want to cancel it?"),
		}, nil
	}

	// Step0.8: notify chat; check before sending to chat to avoid inconsistiency if chat is down
	err = ps.chat.SendCancelOrder(pay)
	if err != nil {
		return &payment.CancelOrderReply{
			Error:        payment.Errors_CHAT_DOWN,
			ErrorMessage: fmt.Sprintf("payments: chat service is unreachable"),
		}, nil
	}

	// Step0.7: do the cancel
	pay.Cancelled = true
	err = ps.repo.SavePay(pay)
	if err != nil {
		return &payment.CancelOrderReply{
			Error:        payment.Errors_DB_FAILED,
			ErrorMessage: fmt.Sprintf("payments: could not save modified pay"),
		}, nil
	}

	return &payment.CancelOrderReply{Cancelled: true}, nil
}

func (ps *paymentServer) CheckStatusAsync(session *models.Session) {
	go ps.shed.process(session)
}

func (ps *paymentServer) CheckStatus(session *models.Session) error {

	// Step0: skip already finished sessions
	if session.Finished && session.ChatNotified {
		return nil
	}

	// Step1: check if it's finished
	finished, err := ps.gateway.CheckStatus(session)
	if err != nil {
		return err
	}

	// Step2: save it -- for status updates
	err = ps.repo.SaveSess(session)
	if err != nil {
		return err
	}

	// Step3: check if it's finished
	if !finished {
		return nil
	}

	// Step4: notify chat
	err = ps.chat.SendSession(session)
	if err != nil {
		log.Error(err)
		return err
	}

	// Step5: remember about it
	session.ChatNotified = true
	err = ps.repo.SaveSess(session)
	if err != nil {
		return err
	}

	return nil
}

func (ps *paymentServer) CheckStatuses() error {

	// Get session with unfinished state
	toCheck, err := ps.repo.GetUnfinished(ps.gateway.GatewayType())
	if err != nil {
		return err
	}

	// check them
	for _, sess := range toCheck {
		err := ps.CheckStatus(&sess)
		if err != nil {
			// this errors are not fatal; let's just log them
			log.Error(err)
		}
	}

	return nil
}
