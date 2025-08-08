package jwt

import (
	"crypto"
	"crypto/hmac"
)

type Signer interface {
	GetAlg() string
}

type SigningMethod struct {
	Name string
	Hash crypto.Hash
}

func (m *SigningMethod) GetAlg() string {
	return m.Name
}

var (
	SignMethodHS256 *SigningMethod
	SignMethodHS384 *SigningMethod
	SignMethodHS512 *SigningMethod
	signingMethods  = map[string]*SigningMethod{}
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
	if m, ok := signingMethods[name]; ok {
		return m
	}
	return nil
}

func (m *SigningMethod) Sign(signingString string, key []byte) []byte {
	hasher := hmac.New(m.Hash.New, key)
	hasher.Write([]byte(signingString))
	return hasher.Sum(nil)
}
