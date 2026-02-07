package rsa

import (
	"crypto/rsa"
	"encoding/json"
	"errors"

	"github.com/tuanta7/keys/internal/config"
)

type Key struct {
	Value     any
	Algorithm string
	KeyID     string
	Use       string
}

func (k Key) MarshalJSON() (b []byte, err error) {
	var jwk *JSONWebKey

	switch t := k.Value.(type) {
	case *rsa.PublicKey:
		jwk, err = rsaPublicKeyToJWK(t)
	case *rsa.PrivateKey:
		jwk, err = rsaPrivateKeyToJWK(t)
	default:
		return nil, errors.New("unsupported key type")
	}

	if err != nil {
		return nil, err
	}

	jwk.Use = k.Use
	jwk.Algorithm = k.Algorithm
	jwk.KeyID = k.KeyID

	return json.Marshal(jwk)
}

func (k *Key) UnmarshalJSON(b []byte) error {
	var jwk JSONWebKey
	err := json.Unmarshal(b, &jwk)
	if err != nil {
		return err
	}

	var key any

	switch jwk.KeyType {
	case config.KeyTypeRSA:
		if jwk.PrivateExponent != nil {
			key, err = jwk.RSAPrivateKey()
		} else {
			key, err = jwk.RSAPublicKey()
		}
	default:
		return errors.New("unsupported key type")
	}

	k.Value = key
	k.Algorithm = jwk.Algorithm
	k.KeyID = jwk.KeyID
	k.Use = jwk.Use

	return nil
}
