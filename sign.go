package main

import (
	"crypto/sha256"
	"encoding/base64"
)

// SignatureKey is the user metadata field which stores a
// secret key used to sign data on behalf of a user.
const SignatureKey = "signatureKey"

// SignData mixes a piece of data with a secret that
// prevents a user from modifying the data while still
// producing the same signature.
func SignData(sigKey string, data []byte) string {
	combined := append(append([]byte(sigKey), data...), []byte(sigKey)...)
	hashed := sha256.Sum256(combined)
	return base64.StdEncoding.EncodeToString(hashed[:])
}

// SignStore generates a unique signature for a store
// search result.
func SignStore(sigKey, storeName string, storeData []byte) string {
	nameSig := SignData(sigKey, []byte(storeName))
	dataSig := SignData(sigKey, storeData)
	return SignData(sigKey, []byte(nameSig+dataSig))
}