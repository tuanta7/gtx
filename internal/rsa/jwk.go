package rsa

import (
	"crypto/rsa"
	"errors"
	"math/big"
)

// JSONWebKey describes an JWK as defined by RFC 7517/7518.
//
// All numeric parameters are Base64URL-encoded, unsigned, big-endian byte sequences.
// Base64URL encoding MUST use the URL-safe alphabet and omit padding (`=`).
type JSONWebKey struct {
	KeyType         string `json:"kty"`
	Algorithm       string `json:"alg,omitempty"`
	Use             string `json:"use,omitempty"`
	KeyID           string `json:"kid,omitempty"`
	Modulus         *Bytes `json:"n,omitempty"`
	PublicExponent  *Bytes `json:"e,omitempty"`
	PrivateExponent *Bytes `json:"d,omitempty"`
	Prime0          *Bytes `json:"p,omitempty"`
	Prime1          *Bytes `json:"q,omitempty"`
	Dp              *Bytes `json:"dp,omitempty"`
	Dq              *Bytes `json:"dq,omitempty"`
	Qi              *Bytes `json:"qi,omitempty"`
}

func (j JSONWebKey) RSAPrivateKey() (*rsa.PrivateKey, error) {
	publicKey, err := j.RSAPublicKey()
	if err != nil {
		return nil, err
	}

	if j.PrivateExponent == nil || j.Prime0 == nil || j.Prime1 == nil {
		return nil, errors.New("missing private exponent, prime0 or prime1")
	}

	privateKey := &rsa.PrivateKey{
		PublicKey: *publicKey,
		D:         j.PrivateExponent.BigInt(),
		Primes:    []*big.Int{j.Prime0.BigInt(), j.Prime1.BigInt()},
	}

	if j.Dp != nil && j.Dq != nil && j.Qi != nil {
		privateKey.Precomputed = rsa.PrecomputedValues{
			Dp:   j.Dp.BigInt(),
			Dq:   j.Dq.BigInt(),
			Qinv: j.Qi.BigInt(),
		}
	}

	if err = privateKey.Validate(); err != nil {
		return nil, err
	}

	return privateKey, nil
}

func (j JSONWebKey) RSAPublicKey() (*rsa.PublicKey, error) {
	if j.Modulus == nil || j.PublicExponent == nil {
		return nil, errors.New("missing modulus or public exponent")
	}

	return &rsa.PublicKey{
		N: j.Modulus.BigInt(),
		E: j.PublicExponent.Int(),
	}, nil
}
