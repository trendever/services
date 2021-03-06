package views

import (
	"common/log"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"payments/config"
	"payments/models"
	"proto/payment"
	"utils/nats"

	"github.com/gin-gonic/gin"
)

// HandleCallback catches pature HTTP notifications
// @TODO: move it to gateway
func (ps *paymentServer) HandleCallback(c *gin.Context) {
	orderID := c.Query("orderid")
	success, _ := strconv.ParseBool(c.Query("success"))

	// further redirect template
	tmpl := config.Get().HTTP.Redirect

	//find session
	sess, err := ps.repo.GetSessByUID(orderID)
	if err != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf(tmpl, 0, false))
		return
	}

	// we want to redirect client NOW
	// status will be reported afterwards by chat message
	ps.CheckStatusAsync(sess)

	if sess.Payment.Redirect != "" {
		tmpl = sess.Payment.Redirect
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf(tmpl, sess.Payment.LeadID, success))
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

		ps.CheckStatusAsync(sess)

	}()
}

func (ps *paymentServer) PeriodicCheck() {
	for range time.Tick(time.Second * time.Duration(config.Get().PeriodicCheck)) {
		log.Debug("Starting PeriodicCheck")
		ps.CheckStatuses()
		log.Debug("Finished PeriodicCheck")
	}
}

func (ps *paymentServer) CheckStatusAsync(session *models.Session) {
	go ps.shed.process(session)
}

func (ps *paymentServer) checkStatus(session *models.Session) error {

	// Step0: skip already finished sessions
	if session.Finished {
		return nil
	}

	gw, err := ps.gw(session.GatewayType)
	if err != nil {
		return err
	}

	// Step1: check if it's finished
	finished, err := gw.CheckStatus(session)
	if err != nil {
		return err
	}

	if finished {
		var event = payment.Event_PayFailed
		if session.Success {
			event = payment.Event_PaySuccess
		}
		err = nats.StanPublish(natsStream, &payment.PaymentNotification{
			Id:    uint64(session.PaymentID),
			Data:  session.Payment.Encode(),
			Event: event,
			Info:  session.Info(),
		})
		if err != nil {
			return err
		}
	}

	// Step2: save it -- for status updates
	err = ps.repo.SaveSess(session)
	if err != nil {
		return err
	}

	return nil
}

func (ps *paymentServer) CheckStatuses() error {

	// Get session with unfinished state
	toCheck, err := ps.repo.GetUnfinished()
	if err != nil {
		return err
	}

	// check them
	for _, sess := range toCheck {
		var copy = sess
		ps.CheckStatusAsync(&copy)
	}

	return nil
}
