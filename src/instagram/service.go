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
	Version       = "10.8.0"
	UserAgent     = "Instagram " + Version + " Android (23/6.0.1; 515dpi; 1440x2416; huawei/google; Nexus 6P; angler; angler; en_US)"
	SigKey        = "68a04945eb02970e2e8d15266fc256f7295da123e123f44b88f09d594a5902df"
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
