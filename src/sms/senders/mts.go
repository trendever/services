package senders

import (
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"github.com/bsm/ratelimit"
	"io/ioutil"
	"net/http"
	"net/url"
	"sms/models"
	"sms/server"
	"strconv"
	"time"
)

const mtsSOAPEndpoint = "https://www.mcommunicator.ru"

const (
	mtsSOAPActionSendMessage      = "/m2m/m2m_api.asmx/SendMessage"
	mtsSOAPActionGetMessageStatus = "/m2m/m2m_api.asmx/GetMessageStatus"
)

type MTS struct {
	login    string
	password string
	naming   string
	rates    *ratelimit.RateLimiter
}

type soapMessageStatus struct {
	Status string `xml:"Body>GetMessageStatusResponse>GetMessageStatusResult>DeliveryInfo>DeliveryStatus"`
}

func NewMTSClient(login, password, naming string, rate int) server.Sender {
	return &MTS{
		login:    login,
		password: fmt.Sprintf("%x", md5.Sum([]byte(password))),
		naming:   naming,
		rates:    ratelimit.New(rate, time.Second),
	}
}

func (s *MTS) SendSMS(sms *models.SmsDB) (err error) {
	for {
		if s.rates.Limit() {
			<-time.After(time.Second / 10)
			continue
		}
		sms.SmsID, err = s.send(sms.Message, sms.Phone)
		if err != nil {
			sms.SmsError = err.Error()
			return err
		}
		<-time.After(time.Second)
		sms.SmsStatus, err = s.status(sms.SmsID)
		return err
	}
}

func (s *MTS) request(path string, params url.Values) ([]byte, error) {
	u, err := url.Parse(mtsSOAPEndpoint)
	if err != nil {
		return nil, err
	}
	u.Path = path
	u.RawQuery = params.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	m, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return m, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error [%v]: %v", resp.StatusCode, string(m))
	}

	return m, err
}

func (s *MTS) send(message, phone string) (id int64, err error) {

	params := url.Values{
		"msid":     {phone[1:]},
		"message":  {message},
		"naming":   {s.naming},
		"login":    {s.login},
		"password": {s.password},
	}

	m, err := s.request(mtsSOAPActionSendMessage, params)

	if err != nil {
		return
	}

	err = xml.Unmarshal(m, &id)
	if err != nil {
		return
	}

	return

}

/*
Status example
<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
    <soap:Body>
        <GetMessageStatusResponse xmlns="http://mcommunicator.ru/M2M">
            <GetMessageStatusResult>
                <DeliveryInfo>
                    <Msid>77779992211</Msid>
                    <DeliveryStatus>Delivered</DeliveryStatus>
                    <DeliveryDate>2016-06-09T20:44:00</DeliveryDate>
                    <UserDeliveryDate>2016-06-09T20:44:28</UserDeliveryDate>
                    <PartCount>1</PartCount>
                </DeliveryInfo>
            </GetMessageStatusResult>
        </GetMessageStatusResponse>
    </soap:Body>
</soap:Envelope>
*/
func (s *MTS) status(id int64) (string, error) {
	params := url.Values{
		"messageID": {strconv.FormatInt(id, 10)},
		"login":     {s.login},
		"password":  {s.password},
	}

	m, err := s.request(mtsSOAPActionGetMessageStatus, params)

	if err != nil {
		return "", err
	}

	status := &soapMessageStatus{}
	err = xml.Unmarshal(m, status)
	if err != nil {
		return "", err
	}

	return status.Status, nil
}
