package common

import "testing"

func TestIsEmailWithName(t *testing.T) {
	var data = []struct {
		param    string
		expected bool
	}{
		{"John Doe <foo@bar.com>", true},
		{"John Doe foo@bar.com", false},
		{"foo@bar.com", true},
	}

	for _, test := range data {
		actual := IsEmailWithName(test.param)
		if test.expected != actual {
			t.Errorf("Expected IsEmailWithName(%q) to be %v,  got %v", test.param, test.expected, actual)
		}
	}
}
