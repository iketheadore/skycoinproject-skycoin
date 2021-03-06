package crypto

import (
	"errors"
	"fmt"

	"github.com/SkycoinProject/skycoin/src/cipher/encrypt"
)

// Cryptor wraps the Encrypt and Decrypt method
type Cryptor interface {
	Encrypt(data, password []byte) ([]byte, error)
	Decrypt(data, password []byte) ([]byte, error)
}

// CryptoType represents the type of crypto name
type CryptoType string

// CryptoTypeFromString converts string to CryptoType
func CryptoTypeFromString(s string) (CryptoType, error) {
	switch CryptoType(s) {
	case CryptoTypeSha256Xor:
		return CryptoTypeSha256Xor, nil
	case CryptoTypeScryptChacha20poly1305:
		return CryptoTypeScryptChacha20poly1305, nil
	case CryptoTypeScryptChacha20poly1305Insecure:
		return CryptoTypeScryptChacha20poly1305Insecure, nil
	default:
		return "", errors.New("unknown crypto type")
	}
}

// Crypto types
const (
	// CryptoTypeSha256Xor uses the SHA256-XOR encryption method (unsafe - no key derivation)
	CryptoTypeSha256Xor = CryptoType("sha256-xor")
	// CryptoTypeScryptChacha20poly1305 uses chacha20poly1305 + scrypt key derivation (use this)
	CryptoTypeScryptChacha20poly1305 = CryptoType("scrypt-chacha20poly1305")
	// CryptoTypeScryptChacha20poly1305Insecure uses chacha20poly1305 + scrypt key derivation with a weak work factor (unsafe)
	CryptoTypeScryptChacha20poly1305Insecure = CryptoType("scrypt-chacha20poly1305-insecure")

	// DefaultCryptoType is the default CryptoType used
	DefaultCryptoType = CryptoTypeScryptChacha20poly1305
)

// cryptoTable records all supported wallet crypto methods
// If want to support new crypto methods, register here.
var cryptoTable = map[CryptoType]Cryptor{
	CryptoTypeSha256Xor:              encrypt.DefaultSha256Xor,
	CryptoTypeScryptChacha20poly1305: encrypt.DefaultScryptChacha20poly1305,
	CryptoTypeScryptChacha20poly1305Insecure: encrypt.ScryptChacha20poly1305{
		N:      1 << 15,
		R:      encrypt.ScryptR,
		P:      encrypt.ScryptP,
		KeyLen: encrypt.ScryptKeyLen,
	},
}

// GetCrypto gets crypto of given type
func GetCrypto(cryptoType CryptoType) (Cryptor, error) {
	c, ok := cryptoTable[cryptoType]
	if !ok {
		return nil, fmt.Errorf("can not find crypto %v in crypto table", cryptoType)
	}

	return c, nil
}

// Types returns all supported crypto types
func Types() []CryptoType {
	return []CryptoType{
		CryptoTypeSha256Xor,
		CryptoTypeScryptChacha20poly1305,
		CryptoTypeScryptChacha20poly1305Insecure,
		DefaultCryptoType,
	}
}

// TypesInsecure returns the CryptoTypeScryptChacha21poly1305Insecure, it
// would be used in testing to speed up the tests.
func TypesInsecure() []CryptoType {
	return []CryptoType{
		CryptoTypeSha256Xor,
		CryptoTypeScryptChacha20poly1305Insecure,
	}
}
