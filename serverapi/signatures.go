package serverapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"github.com/unixpickle/optishop-server/optishop/db"
)

// SignatureKey is the user metadata field which stores a
// secret key used to sign data on behalf of a user.
const SignatureKey = "signatureKey"

// SignData mixes a piece of data with a secret that
// prevents a user from modifying the data while still
// producing the same signature.
func SignData(sigKey string, data []byte) string {
	mac := hmac.New(sha256.New, []byte(sigKey))
	return base64.StdEncoding.EncodeToString(mac.Sum(data))
}

// SignStore generates a unique signature for a store
// search result.
func SignStore(sigKey, storeName string, storeData []byte) string {
	nameSig := SignData(sigKey, []byte(storeName))
	dataSig := SignData(sigKey, storeData)
	return SignData(sigKey, []byte(nameSig+dataSig))
}

// SignInventoryItem generates a unique signature for an
// inventory search result.
func SignInventoryItem(sigKey string, storeID db.StoreID, data []byte) string {
	idSig := SignData(sigKey, []byte(storeID))
	dataSig := SignData(sigKey, data)
	return SignData(sigKey, []byte(idSig+dataSig))
}
