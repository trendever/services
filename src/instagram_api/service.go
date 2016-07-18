package instagram_api

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"

	"github.com/pborman/uuid"
)

const (
	API_URL         = "https://i.instagram.com/api/v1"
	USER_AGENT      = "Instagram 8.2.0 Android (23/6.0.1; 515dpi; 1440x2416; huawei/google; Nexus 6P; angler; angler; en_US)"
	IG_SIG_KEY      = "55e91155636eaa89ba5ed619eb4645a4daf1103f2161dbfe6fd94d5ea7716095"
	SIG_KEY_VERSION = "4"
)

func generateUUID(t bool) string {
	u := uuid.New()
	if !t {
		return strings.Replace(u, "-", "", -1)
	}
	return u
}

func generateSignature(data []byte) string {
	h := hmac.New(sha256.New, []byte(IG_SIG_KEY))
	h.Write(data)
	hash := hex.EncodeToString(h.Sum(nil))
	return "ig_sig_key_version=" + SIG_KEY_VERSION + "&signed_body=" + hash + "." + url.QueryEscape(string(data))
}

func generateDeviceId(salt string) string {
	hash := md5.New()
	hash.Write([]byte(salt))
	hash.Write([]byte{1, 2, 3})
	return "android-" + hex.EncodeToString(hash.Sum(nil))[:16]
}
