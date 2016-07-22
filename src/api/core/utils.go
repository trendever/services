package core

import (
	. "api/conf"
	gomail "gopkg.in/gomail.v2"
	"math/rand"
	"time"
)

func Random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func SendEmail(subject string, msg string, recipient string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", GetSettings().Mail.Sender)
	m.SetHeader("To", recipient)
	m.SetAddressHeader("Cc", GetSettings().Mail.Sender, "Trendever")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", msg)

	d := gomail.NewPlainDialer(GetSettings().Mail.Host,
		GetSettings().Mail.Port,
		GetSettings().Mail.Username,
		GetSettings().Mail.Password)

	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
