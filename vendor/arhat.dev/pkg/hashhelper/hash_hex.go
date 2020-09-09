package hashhelper

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
)

func Sha256SumHex(data []byte) string {
	h := sha256.New()
	_, _ = h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func Sha512SumHex(data []byte) string {
	h := sha512.New()
	_, _ = h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func MD5SumHex(data []byte) string {
	h := md5.New()
	_, _ = h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
