package phone

import (
	"errors"
	"github.com/ttacon/libphonenumber"
)

func CheckNumber(phoneNumber, country string) (string, error) {
	if country == "" {
		country = "RU"
	}

	number, err := libphonenumber.Parse(phoneNumber, country)
	if err != nil {
		return "", err
	}

	if !libphonenumber.IsValidNumber(number) {
		return "", errors.New("Phone number isn't valid")
	}

	return libphonenumber.Format(number, libphonenumber.E164), nil
}
