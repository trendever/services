package product_code

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	testCases = map[string]int64{
		"tf2209": 10087,
		"tf2021": 9899,
		"te2209": 88,

		// sequential test
		"tf9999": 17877,
		"tg0001": 17878,
		"tg0002": 17879,

		// sequential prefix test
		"tz9999":  217857,
		"tba0001": 217858,
		"tba0002": 217859,

		// sequential prefix test
		"tzzzz9999":  4569260907,
		"tbaaaa0001": 4569260908,
		"tbaaaa0002": 4569260909,
	}
)

func TestCodeGen(t *testing.T) {

	for code, id := range testCases {
		// test encode
		assert.Equal(t, code, GenCode(id))

		// test decode
		id_got, err := RevCode(code)
		assert.Nil(t, err)
		assert.Equal(t, id, id_got)
	}
}
