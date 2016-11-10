package payture

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"utils/log"
)

func request(endpoint string, params map[string]string) ([]byte, error) {

	// add parameters
	urlValues := url.Values{}
	for k, v := range params {
		urlValues.Add(k, v)
	}

	log.Debug("Req parameters (%v): %#v", endpoint, urlValues)

	response, err := http.PostForm(endpoint, urlValues)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	return body, err
}

func encodeData(data map[string]string) string {
	var dataParams []string
	for k, v := range data {
		dataParams = append(dataParams, fmt.Sprintf("%v=%v", k, v))
	}

	return strings.Join(dataParams, ";")
}

func xmlRequest(url string, decodeTo interface{}, data, params map[string]string) error {

	if len(data) > 0 {
		params["Data"] = encodeData(data)
	}

	body, err := request(url, params)

	if err != nil {
		return err
	}

	log.Debug("Payture resp body (for %v): %v", url, string(body))

	err = xml.Unmarshal(body, decodeTo)

	log.Debug("Payture unmarshal: %+v", decodeTo)

	return err
}

func (c *InPay) xmlRequest(method string, decodeTo interface{}, data, extraParams map[string]string) error {
	params := map[string]string{
		"Key": c.Key,
	}

	for k, v := range extraParams {
		params[k] = v
	}

	url := fmt.Sprintf("%v/apim/%v", c.URL, method)

	return xmlRequest(url, decodeTo, data, params)
}
