package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

const SessionTokenBytes = 32 //Session token hash will always be 32 bytes (SHA256 length)
// NewSessionToken takes no inputs, and returns a token string, token hash, and error
func NewSessionToken() (token string, tokenHash [32]byte, err error) {
	raw := make([]byte, SessionTokenBytes)   //Reserves 32 bytes of memory to be filled later by the raw token
	if _, err = rand.Read(raw); err != nil { //Fill allocated slice with random bytes, and if returned error is anything but nil, return failure
		return "", [32]byte{}, err
	}

	token = base64.RawURLEncoding.EncodeToString(raw) //Generate session token
	tokenHash = sha256.Sum256([]byte(token))          //Create token hash to be stored in the user DB
	return token, tokenHash, nil                      //return token, hashed token, and nil error (success)
}
