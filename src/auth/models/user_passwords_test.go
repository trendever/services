package models

import "testing"

func TestGenerateRandomPass(t *testing.T) {
	var data = []struct {
		param    int
		expected int
	}{
		{param: 1, expected: 1},
		{param: 2, expected: 2},
		{param: 3, expected: 3},
		{param: 6, expected: 6},
	}

	for _, test := range data {
		actual := len(generateRandomPass(test.param))
		if test.expected != actual {
			t.Errorf("Expected len(GenerateRandomPass(%q)) to be %v, but got %v", test.param, test.expected, actual)
		}
	}
}
