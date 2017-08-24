package product_code

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	alphabet  = "abcdefghijklmnopqrstuvwxyz"
	numbers   = "0123456789"
	base64URL = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	// largest possible len of post ulr part(for uint64 id type)
	instagramMaxLen = 11
)

var (
	// yeah, this is first code for product with ID==0
	codeStart, _   = revCode("te2121")
	reversedBase64 map[rune]uint64
)

func init() {
	reversedBase64 = map[rune]uint64{}
	for i, c := range base64URL {
		reversedBase64[c] = uint64(i)
	}
}

// convert instagram decimal id to url part
// 1436562162406503138_4118841035 -> BPvsazsAbLi
func ID2URL(id string) (out string, err error) {
	parts := strings.Split(id, "_")
	num, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return "", errors.New("invalid id")
	}
	var shift int
	buf := make([]byte, instagramMaxLen)
	for num > 0 {
		shift++
		buf[instagramMaxLen-shift] = base64URL[num%64]
		num /= 64
	}
	return string(buf[instagramMaxLen-shift:]), nil
}

// instagram url part -> id
// BYE7q4cBWpT -> 1586655400303094355
func URL2ID(str string) (ret string, err error) {
	if len(str) > instagramMaxLen {
		return "", errors.New("invalid code")
	}
	var id uint64
	for _, c := range str {
		id *= 64
		i, ok := reversedBase64[c]
		if !ok {
			return "", errors.New("invalid code")
		}
		id += i
	}
	return strconv.FormatUint(id, 10), nil
}

var postURLRegexp = regexp.MustCompile("^https?://(?:www.)?instagram.com/p/([\\w|-]+)/.*")

// return post id from instagram url
// https://www.instagram.com/p/BYE7q4cBWpT/?taken-by=nasa -> 1586655400303094355
func ParsePostURL(url string) (post_id string, err error) {
	if url == "" {
		return "", nil
	}
	sub := postURLRegexp.FindStringSubmatch(url)
	if sub == nil {
		return "", errors.New("invalid url")
	}
	return URL2ID(sub[1])
}

// return instagram post url for given id
// 1587331875411361939_12345 => https://www.instagram.com/p/BYHVe4-jICT/
func MakePostURL(id string) (url string, err error) {
	code, err := ID2URL(id)
	if err != nil {
		return "", err
	}
	return "https://www.instagram.com/p/" + code + "/", nil
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
