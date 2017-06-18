package product_code

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	alphabet  = "abcdefghijklmnopqrstuvwxyz"
	numbers   = "0123456789"
	base64URL = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
)

var (
	// yeah, this is first code for product with ID==0
	codeStart, _ = revCode("te2121")
)

// convert instagram decimal id to url part
// 1436562162406503138_4118841035 -> BPvsazsAbLi
func ID2URL(id string) (out string, err error) {
	parts := strings.Split(id, "_")
	if len(parts) == 0 {
		return "", errors.New("invalid id")
	}
	num, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return "", errors.New("invalid id")
	}
	const maxLen = 11
	var shift int
	buf := make([]byte, maxLen)
	for num > 0 {
		shift++
		buf[maxLen-shift] = base64URL[num%64]
		num /= 64
	}
	return string(buf[maxLen-shift:]), nil
}

func int2ascii(num int64) (out string) {
	radix := int64(len(alphabet))
	for {
		out = string(alphabet[num%radix]) + out

		if num < radix {
			return
		}

		num = num / radix

	}
}

func ascii2int(num string) (out int64, err error) {
	for len(num) > 0 {
		el := num[0]
		num = num[1:]

		index := strings.IndexByte(alphabet, el)
		if index < 0 {
			return 0, fmt.Errorf("Unknown code character: %v", el)
		}

		out = int64(len(alphabet))*out + int64(index)
	}
	return
}

// Convert an id to string like te0001
// t - fixed prefix
// e - alphabetically incremented number
// after reaching z, a new symbol is added: {z, aa, ab, ..., zz, aaa, aab}
// codes start with te0001
// 0000 is never used
func GenCode(id int64) string {
	return genCode(id + codeStart)
}

func genCode(id int64) string {
	prefixInt, suffixInt := id/9999, 1+id%9999

	return fmt.Sprintf("t%s%04d", int2ascii(prefixInt), suffixInt)
}

// Decode code to ID
func RevCode(code string) (int64, error) {
	res, err := revCode(code)
	return res - codeStart, err
}

func revCode(code string) (int64, error) {

	code = code[1:]

	suffixStart := strings.IndexAny(code, numbers)
	if suffixStart < 0 {
		return 0, fmt.Errorf("Incorrect product code %v", code)
	}

	prefix, suffix := code[:suffixStart], code[suffixStart:]

	suffixInt, err := strconv.ParseInt(suffix, 10, 64)
	if err != nil {
		return 0, err
	}

	prefixInt, err := ascii2int(prefix)
	if err != nil {
		return 0, err
	}

	return prefixInt*9999 + suffixInt - 1, nil
}
