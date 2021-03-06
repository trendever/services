package payture

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"proto/payment"
)

const (
	sessionTypeAdd  = "Add"
	sessionTypePay  = "Pay"
	vwInitPath      = "/vwapi/Init"
	vwPayPath       = "/vwapi/Pay"
	vwAddPath       = "/vwapi/Add"
	vwCardsPath     = "/vwapi/GetList"
	vwDelCardPath   = "/vwapi/Remove"
	vwPayStatusPath = "/vwapi/PayStatus"
)

type payDef struct {
	cardID  string
	orderID string
	amount  uint64
}

type vwInitResponse struct {
	XMLName   xml.Name `xml:"Init"`
	Success   bool     `xml:"Success,attr"`
	ErrCode   string   `xml:"ErrCode,attr"`
	SessionID string   `xml:"SessionId,attr"`
	OrderID   string   `xml:"OrderId,attr"`
}

type vwPayResponse struct {
	XMLName xml.Name `xml:"Pay"`

	VWUserLgn       string `xml:"VWUserLgn,attr"`
	OrderID         string `xml:"OrderId,attr"`
	MerchantOrderID string `xml:"MerchantOrderId,attr"`
	Amount          uint64 `xml:"Amount,attr"`

	Success bool   `xml:"Success,attr"`
	ErrCode string `xml:"ErrCode,attr"`

	Additional []struct {
		Key   string `xml:"Key,attr"`
		Value string `xml:"Value,attr"`
	} `xml:"AddInfo"`
}

type vwCardsResponse struct {
	XMLName xml.Name `xml:"GetList"`
	Success bool     `xml:"Success,attr"`
	ErrCode string   `xml:"ErrCode,attr"`
	Items   []struct {
		XMLName    xml.Name `xml:"Item"`
		CardName   string   `xml:"CardName,attr"`
		CardID     string   `xml:"CardId,attr"`
		CardHolder string   `xml:"CardHolder,attr"`
		Status     string   `xml:"Status,attr"`
		NoCVV      bool     `xml:"NoCVV,attr"`
		Expired    bool     `xml:"Expired,attr"`
	} `xml:"Item"`
}

type vwDelCardResponse struct {
	XMLName xml.Name `xml:"Remove"`
	Success bool     `xml:"Success,attr"`
	ErrCode string   `xml:"ErrCode,attr"`
}

type vwPayStatusResponse struct {
	XMLName xml.Name `xml:"PayStatus"`
	Success bool     `xml:"Success,attr"`
	OrderID string   `xml:"OrderId,attr"`
	Amount  uint64   `xml:"Amount,attr"`
	State   string   `xml:"Status,attr"`
	ErrCode string   `xml:"ErrCode,attr"`
}

// Init request
func (ew *Ewallet) vwInit(sessionType, key string, user *payment.UserInfo, pay *payDef) (*vwInitResponse, error) {
	params := map[string]string{
		"VWID": key,
	}

	login, password, err := ew.creds(user.UserId)
	if err != nil {
		return nil, err
	}

	data := map[string]string{
		"SessionType": sessionType,
		"VWUserLgn":   login,
		"VWUserPsw":   password,
		"PhoneNumber": user.Phone,
		"IP":          user.Ip,
	}

	if pay != nil {
		data["OrderId"] = pay.orderID
		data["CardId"] = pay.cardID
		data["Amount"] = fmt.Sprintf("%v", pay.amount)
	}

	resp := vwInitResponse{}
	err = xmlRequest(ew.URL+vwInitPath, &resp, data, params)
	return &resp, err
}

func (ew *Ewallet) vwPay(user *payment.UserInfo, pay *payDef) (*vwInitResponse, error) {
	params := map[string]string{
		"VWID": ew.KeyAdd,
	}

	login, password, err := ew.creds(user.UserId)
	if err != nil {
		return nil, err
	}

	data := map[string]string{
		"SessionType": sessionTypePay,
		"VWUserLgn":   login,
		"VWUserPsw":   password,
		"PhoneNumber": user.Phone,
		"OrderId":     pay.orderID,
		"Amount":      fmt.Sprintf("%v", pay.amount),
		"IP":          user.Ip,
	}

	if pay.cardID != "" {
		data["CardId"] = pay.cardID
	}

	resp := vwInitResponse{}
	err = xmlRequest(ew.URL+vwInitPath, &resp, data, params)
	return &resp, err
}

func (ew *Ewallet) vwPayAsync(user *payment.UserInfo, pay *payDef) (*vwPayResponse, error) {
	params := map[string]string{
		"VWID": ew.KeyPay,
	}

	login, password, err := ew.creds(user.UserId)
	if err != nil {
		return nil, err
	}

	data := map[string]string{
		"SessionType": sessionTypePay,
		"VWUserLgn":   login,
		"VWUserPsw":   password,
		"PhoneNumber": user.Phone,
		"OrderId":     pay.orderID,
		"Amount":      fmt.Sprintf("%v", pay.amount),
		"IP":          user.Ip,
	}

	if pay.cardID != "" {
		data["CardId"] = pay.cardID
	}

	resp := vwPayResponse{}
	err = xmlRequest(ew.URL+vwPayPath, &resp, data, params)
	return &resp, err
}

func (ew *Ewallet) vwCards(user *payment.UserInfo) (*vwCardsResponse, error) {
	params := map[string]string{
		"VWID": ew.KeyAdd,
	}

	login, password, err := ew.creds(user.UserId)
	if err != nil {
		return nil, err
	}

	data := map[string]string{
		"VWUserLgn": login,
		"VWUserPsw": password,
	}

	resp := vwCardsResponse{}
	err = xmlRequest(ew.URL+vwCardsPath, &resp, data, params)
	return &resp, err
}

func (ew *Ewallet) vwDelCard(cardID string, user *payment.UserInfo) (*vwDelCardResponse, error) {
	params := map[string]string{
		"VWID": ew.KeyAdd,
	}

	login, password, err := ew.creds(user.UserId)
	if err != nil {
		return nil, err
	}

	data := map[string]string{
		"VWUserLgn": login,
		"VWUserPsw": password,
		"CardId":    cardID,
	}

	resp := vwDelCardResponse{}
	err = xmlRequest(ew.URL+vwDelCardPath, &resp, data, params)
	return &resp, err
}

func (ew *Ewallet) vwPayStatus(orderID string) (*vwPayStatusResponse, error) {
	params := map[string]string{
		"VWID": ew.KeyAdd,
	}

	data := map[string]string{
		"OrderId": orderID,
	}

	resp := vwPayStatusResponse{}
	err := xmlRequest(ew.URL+vwPayStatusPath, &resp, data, params)
	return &resp, err
}

// generate login && password by UID
func (ew *Ewallet) creds(userID uint64) (string, string, error) {

	if userID == 0 {
		return "", "", fmt.Errorf("payments: got zero userID for ewallet stuff")
	}

	login := fmt.Sprintf("tndvrid_%v", userID)

	h := sha256.New()
	h.Write([]byte(ew.Secret))
	h.Write([]byte(login))
	h.Write([]byte(ew.Secret))

	password := hex.EncodeToString(h.Sum(nil))

	return login, password, nil
}
