package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func ComputeHMACSHA256(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	signature := hex.EncodeToString(h.Sum(nil))
	fmt.Printf("Agent signing data: '%s'\nComputed hash: %s\n", data, signature)
	return signature
}
