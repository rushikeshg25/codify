package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
)

// codegroundSecret is the master key that per-codeground signing keys are
// derived from. It must match CODEGROUND_TOKEN_SECRET in the API server, which
// performs the identical derivation when it mints tokens.
var codegroundSecret []byte

func init() {
	secret := os.Getenv("CODEGROUND_TOKEN_SECRET")
	if secret == "" {
		log.Fatal("CODEGROUND_TOKEN_SECRET must be set: refusing to create codegrounds that cannot authenticate callers")
	}
	codegroundSecret = []byte(secret)
}

/*
codegroundKey derives the signing key handed to a single codeground pod.

Deriving per codeground rather than sharing one key means a pod running hostile
user code can only verify tokens minted for itself: it cannot forge access to
anyone else's codeground, and it never holds the key that signs session
cookies.

This must stay byte-for-byte identical to CodegroundKey in the API server.
*/
func codegroundKey(codegroundID string) string {
	mac := hmac.New(sha256.New, codegroundSecret)
	mac.Write([]byte("codeground:" + codegroundID))
	return hex.EncodeToString(mac.Sum(nil))
}
