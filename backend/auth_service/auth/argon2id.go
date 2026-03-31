package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Argon2idParams struct {
	Memory      uint32 //Memory used in kibibytes (KiB)
	Iterations  uint32 //Number of passes over the memory
	Parallelism uint8  // Number of threads used in the algo
	SaltLength  uint32 //Length of random salt
	KeyLength   uint32 //Length of generated key/password hash
}

var DefaultArgon2idParameters = Argon2idParams{
	Memory:      64 * 1024, // 64 Mibibytes
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

func HashPasswordArgon2id(password string, param Argon2idParams) (string, error) {
	if password == "" {
		return "", errors.New("password is empty")
	}

	salt := make([]byte, param.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, param.Iterations, param.Memory, param.Parallelism, param.KeyLength)

	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=19$m=%d, t=%d, p=%d$%s$%s",
		param.Memory, param.Iterations, param.Parallelism, saltB64, hashB64)

	return encoded, nil
}

func VerifyPasswordArgon2id(encodedHash, password string) (bool, error) {
	if encodedHash == "" {
		return false, errors.New("encoded hash is empty")
	}

	param, salt, hash, err := decodeArgon2idHash(encodedHash)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey([]byte(password), salt, param.Iterations, param.Memory,
		param.Parallelism, param.KeyLength)

	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

func decodeArgon2idHash(encoded string) (Argon2idParams, []byte, []byte, error) {
	//Expected format: $argon2id$v=19$m=65536, t=3, p=2$<salt>$<hash>
	parts := strings.Split(encoded, "$") //Split encoded pass using $ as delimiter
	if len(parts) != 6 || parts[1] != "argon2id" {
		return Argon2idParams{}, nil, nil, errors.New("invalid argon2id hash format")
	}

	if parts[2] != "v=19" {
		return Argon2idParams{}, nil, nil, errors.New("incompatible argon2 version")
	}

	var params Argon2idParams
	_, err := fmt.Sscanf(parts[3], "m=%d, t=%d, p=%d", &params.Memory, &params.Iterations, &params.Parallelism)
	if err != nil {
		return Argon2idParams{}, nil, nil, errors.New("invalid argon2 parameters")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Argon2idParams{}, nil, nil, errors.New("invalid salt encoding")
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Argon2idParams{}, nil, nil, errors.New("invalid hash encoding")
	}

	params.SaltLength = uint32(len(salt))
	params.KeyLength = uint32(len(hash))

	return params, salt, hash, nil
}
