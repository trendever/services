package instagram

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"

	"github.com/pborman/uuid"
)

// Global constants
const (
	URL           = "https://i.instagram.com/api/v1"
	Version       = "9.3.0"
	UserAgent     = "Instagram " + Version + " Android (23/6.0.1; 515dpi; 1440x2416; huawei/google; Nexus 6P; angler; angler; en_US)"
	SigKey        = "26e29e57f4ea61a0ebb4ee0ec483e5efe7ca39093adcfa3689dadbfba139546b"
	SigKeyVersion = "4"
)

func generateUUID(t bool) string {
	u := uuid.New()
	if !t {
		return strings.Replace(u, "-", "", -1)
	}
	return u
}

func generateSignature(data []byte) string {
	h := hmac.New(sha256.New, []byte(SigKey))
	h.Write(data)
	hash := hex.EncodeToString(h.Sum(nil))
	return "ig_sig_key_version=" + SigKeyVersion + "&signed_body=" + hash + "." + url.QueryEscape(string(data))
}

func generateDeviceID(salt string) string {
	hash := md5.New()
	hash.Write([]byte(salt))
	hash.Write([]byte{1, 2, 3})
	return "android-" + hex.EncodeToString(hash.Sum(nil))[:16]
}
