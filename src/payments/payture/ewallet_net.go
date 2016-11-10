package payture

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"proto/payment"
)

const (
	sessionTypeAdd = "Add"
	sessionTypePay = "Pay"
	vwInitPath     = "/vwapi/Init"
)

type payDef struct {
	cardID  string
	orderID string
	amount  uint64
}

type vwInitResponse struct {
	XMLName   xml.Name `xml:"Init"`
	Success   bool     `xml:"Success,attr"`
	OrderID   string   `xml:"OrderId,attr"`
	Amount    uint64   `xml:"Amount,attr"`
	SessionID string   `xml:"SessionId,attr"`
	ErrCode   string   `xml:"ErrCode,attr"`
}

// Init request
func (ew *Ewallet) vwInit(sessionType, key string, user *payment.UserInfo, pay *payDef) (*vwInitResponse, error) {
	params := map[string]string{
		"VWID": key,
	}

	login, password := ew.creds(user.UserId)

	data := map[string]string{
		"SessionType": sessionType,
		"VWUserLgn":   login,
		"VWUserPsw":   password,
		"PhoneNumber": user.Phone,
		"IP":          user.Ip,
	}

	resp := vwInitResponse{}
	err := xmlRequest(ew.URL+vwInitPath, &resp, data, params)
	return &resp, err
}

// generate login && password by UID
func (ew *Ewallet) creds(userID uint64) (string, string) {

	login := fmt.Sprintf("tndvrid_%v", userID)

	h := sha256.New()
	h.Write([]byte(ew.Secret))
	h.Write([]byte(login))
	h.Write([]byte(ew.Secret))

	password := hex.EncodeToString(h.Sum(nil))

	return login, password
}
