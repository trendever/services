package models

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessages(t *testing.T) {
	var (
		inputs = []Template{
			&EmailTemplate{
				EmailMessage: EmailMessage{
					From:    "String test",
					To:      "without",
					Subject: "variables.",
					Body:    "any",
				},
			},
			&EmailTemplate{
				EmailMessage: EmailMessage{
					From:    "Trendever stuff",
					To:      "{{ object.Email }}",
					Subject: "Hello, dear {{ object.Name }}",
					Body:    "Just die, {{ object.Name }}! Your Trendever",
				},
			},
			&SMSTemplate{
				SMSMessage: SMSMessage{
					To:      "+79991234242",
					Message: "Hello, you made a lead {{ object.Source }}",
				},
			},
			&ChatTemplate{
				Message: "Awesome tests, {{ object.Name }}",
			},
		}

		inputContexts = []interface{}{
			0,
			&User{
				Email: "test@mail.ru",
				Name:  "ehohoh!",
			},
			&Lead{
				Source: "meow meow meow meow",
			},
		}

		excpectedResults = []interface{}{
			&EmailMessage{
				From:    "String test",
				To:      "without",
				Subject: "variables.",
				Body:    "any",
			},
			&EmailMessage{
				From:    "Trendever stuff",
				To:      "test@mail.ru",
				Subject: "Hello, dear ehohoh!",
				Body:    "Just die, ehohoh!! Your Trendever",
			},
			&SMSMessage{
				To:      "+79991234242",
				Message: "Hello, you made a lead meow meow meow meow",
			},
		}
	)

	assert.True(t, len(inputs) == len(inputContexts) && len(inputContexts) == len(excpectedResults), "Incorrect test")

	for i, tmpl := range inputs {
		ctx := inputContexts[i]

		res, err := tmpl.Parse(ctx)
		assert.Nil(t, err)

		assert.EqualValues(t, excpectedResults[i].(interface{}), res)
	}
}

func TestErrorMessages(t *testing.T) {
	var inputs = []Template{ // incorrect templates
		&SMSTemplate{
			SMSMessage: SMSMessage{
				To: "{{ asd dsa dsd ",
			},
		},
		&EmailTemplate{
			EmailMessage: EmailMessage{
				Body: "Just die, }} . {{ ",
			},
		},
	}

	for _, tmpl := range inputs {

		_, err := tmpl.Parse(nil)
		assert.NotNil(t, err, fmt.Sprintf("Template %#v executed OK, but should fail", tmpl))
	}

}
