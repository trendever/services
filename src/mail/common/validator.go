package common

import (
	"github.com/asaskevich/govalidator"
	"regexp"
	"strings"
)

const (
	EmailWithName = "^(.+)<(.+)>$"
)

var rxEmailWithName = regexp.MustCompile(EmailWithName)

//IsEmailWithName checks if value is email or email with name: Name <email@domain.com>
func IsEmailWithName(value string) bool {
	if subs := rxEmailWithName.FindStringSubmatch(value); len(subs) > 0 {
		return govalidator.IsEmail(subs[2])
	}

	return govalidator.IsEmail(value)
}

func init() {
	govalidator.TagMap["email_with_name"] = govalidator.Validator(IsEmailWithName)

	govalidator.TagMap["emails"] = govalidator.Validator(func(value string) bool {
		valid := true
		for _, email := range strings.Split(value, ",") {
			if !IsEmailWithName(email) {
				valid = false
				break
			}
		}
		return valid
	})
}
