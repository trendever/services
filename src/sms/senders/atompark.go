package senders

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"sms/conf"
	"sms/models"
)

const (
	apiURL      = "https://api.atompark.com/sms/3.0/"
	apiVersion  = "3.0"
	smsLifetime = "0" // Время жизни сообщения, 0 - максимальное время
	datetime    = ""
	smsType     = "2" // 2 - отправка смс через прямые подклюения к операторам (разрешены только альфа имена)
)

//Atompark is Sender interface implementation for Atompark service
type Atompark struct {
	test       string
	keyPublic  string
	keyPrivate string
	sender     string
}

//NewAtompark returns NewAtompark instance
func NewAtompark() *Atompark {

	settings := conf.GetSettings()

	a := &Atompark{
		test:       settings.Atompark.Test,
		keyPublic:  settings.Atompark.KeyPublic,
		keyPrivate: settings.Atompark.KeyPrivate,
		sender:     settings.Atompark.Sender,
	}

	return a
}

//SendSMS sends sms
func (ap *Atompark) SendSMS(sms *models.SmsDB) error {

	// args for calculating checksum of request
	args := map[string]string{
		"sender":       ap.sender,
		"asender":      ap.sender,
		"key":          ap.keyPublic,
		"test":         ap.test,
		"text":         sms.Message,
		"phone":        sms.Phone,
		"datetime":     datetime,
		"sms_lifetime": smsLifetime,
		"version":      apiVersion,
		"action":       "sendSMS",
		"type":         smsType,
	}

	// keys sorting
	keys := make([]string, len(args))
	i := 0
	for k := range args {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	// concat params
	var concatString string
	for _, k := range keys {
		concatString += args[k]
	}
	concatString += ap.keyPrivate

	// get md5 hash
	hash := md5.New()
	hash.Write([]byte(concatString))
	sum := hex.EncodeToString(hash.Sum(nil))

	// POST params
	form := url.Values{}
	form.Add("key", ap.keyPublic)
	form.Add("sum", sum)
	form.Add("sender", ap.sender)
	form.Add("asender", ap.sender)
	form.Add("text", sms.Message)
	form.Add("phone", sms.Phone)
	form.Add("datetime", datetime)
	form.Add("sms_lifetime", smsLifetime)
	form.Add("test", ap.test)
	form.Add("type", smsType)

	// request
	resp, err := ap.request("POST", apiURL+"sendSMS?", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// read json from Body
	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// unmarshal json
	var objectJSON *models.SmsJSON
	if err = json.Unmarshal(jsonBody, &objectJSON); err != nil {
		return err
	}

	// check error form atompark
	if objectJSON.Error != "" {
		// if error, fill sms db model
		sms.SmsError = objectJSON.Error
		sms.SmsStatus = "error"
	} else {
		// if success, parse result
		var result *models.ResultSuccess

		// unmarshal json
		if err = json.Unmarshal(objectJSON.Result, &result); err != nil {
			return err
		}

		// fill sms db model
		sms.SmsID = result.ID
		sms.SmsStatus = "sent"
	}

	return nil
}

// Main request
func (ap *Atompark) request(method, endpoint string, body io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
