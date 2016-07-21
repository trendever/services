package models

import "testing"

func TestValidate(t *testing.T) {
	//from, subject, message, to, ok
	data := [][]string{
		{"test@mail.ru", "test", "test", "test@mail.ru", "ok"},
		{"test@mail.ru", "test", "test", "test@mail.ru,test2@gmail.com", "ok"},
		//Not ok, because "to" contains spaces
		{"test@mail.ru", "test", "test", "test@mail.ru, test2@gmail.com", "!ok"},
		{"", "", "", "", "!ok"},
	}

	for _, v := range data {
		m := &Mail{
			From:    v[0],
			Subject: v[1],
			Message: v[2],
			To:      v[3],
		}

		if ok, err := m.Validate(); !ok && v[4] == "ok" {
			t.Error(err)
		}
	}
}
