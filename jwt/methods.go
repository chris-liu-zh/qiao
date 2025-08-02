package jwt

import (
	"crypto"
	"crypto/hmac"
	"sync"
)

type Signer interface {
	GetAlg() string
}

type SigningMethod struct {
	Name string
	Hash crypto.Hash
}

func (s *SigningMethod) GetAlg() string {
	return s.Name
}

var (
	SignMethodHS256   *SigningMethod
	SignMethodHS384   *SigningMethod
	SignMethodHS512   *SigningMethod
	signingMethodLock = new(sync.RWMutex)
	signingMethods    = map[string]*SigningMethod{}
)

func init() {
	SignMethodHS256 = &SigningMethod{Name: "HS256", Hash: crypto.SHA256}
	SignMethodHS384 = &SigningMethod{Name: "HS384", Hash: crypto.SHA384}
	SignMethodHS512 = &SigningMethod{Name: "HS512", Hash: crypto.SHA512}
	signingMethods = map[string]*SigningMethod{
		"HS256": SignMethodHS256,
		"HS384": SignMethodHS384,
		"HS512": SignMethodHS512,
	}
}

func GetSigningMethod(name string) *SigningMethod {
	return signingMethods[name]
}

func (m *SigningMethod) Sign(signingString string, key []byte) []byte {
	hasher := hmac.New(m.Hash.New, key)
	hasher.Write([]byte(signingString))
	return hasher.Sum(nil)
}
