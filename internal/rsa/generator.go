package rsa

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tuanta7/tig/internal/config"
)

type RSAKeyGenerator struct {
	OutputDir string
	Format    string
}

func (g *RSAKeyGenerator) WriteKeyPair(privateKey *rsa.PrivateKey) error {
	if err := g.writePrivateKey(privateKey); err != nil {
		return err
	}

	return g.writePublicKey(&privateKey.PublicKey)
}

func (g *RSAKeyGenerator) writePrivateKey(privateKey *rsa.PrivateKey) error {
	filePath := filepath.Join(g.OutputDir, g.fileName("id_rsa"))
	data, err := g.marshalPrivateKey(privateKey)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0600)
}

func (g *RSAKeyGenerator) marshalPrivateKey(privateKey *rsa.PrivateKey) ([]byte, error) {
	switch g.Format {
	case config.KeyFormatDER:
		return x509.MarshalPKCS1PrivateKey(privateKey), nil
	case config.KeyFormatPEM:
		return pem.EncodeToMemory(&pem.Block{
			Type:  config.KeyTypeRSAPrivateKey,
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}), nil
	case config.KeyFormatJWK:
		return json.MarshalIndent(Key{Value: privateKey}, "", "\t")
	default:
		return nil, fmt.Errorf("unsupported format: %s", g.Format)
	}
}

func (g *RSAKeyGenerator) writePublicKey(publicKey *rsa.PublicKey) error {
	filePath := filepath.Join(g.OutputDir, g.fileName("id_rsa.pub"))
	data, err := g.marshalPublicKey(publicKey)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func (g *RSAKeyGenerator) marshalPublicKey(publicKey *rsa.PublicKey) ([]byte, error) {
	switch g.Format {
	case config.KeyFormatDER:
		return x509.MarshalPKCS1PublicKey(publicKey), nil
	case config.KeyFormatPEM:
		return pem.EncodeToMemory(&pem.Block{
			Type:  config.KeyTypeRSAPublicKey,
			Bytes: x509.MarshalPKCS1PublicKey(publicKey),
		}), nil
	case config.KeyFormatJWK:
		return json.MarshalIndent(Key{Value: publicKey}, "", "\t")
	default:
		return nil, fmt.Errorf("unsupported format: %s", g.Format)
	}
}

func (g *RSAKeyGenerator) fileName(def string) string {
	switch g.Format {
	case config.KeyFormatPEM:
		return def
	case config.KeyFormatDER:
		return def + ".der"
	case config.KeyFormatJWK:
		return def + ".json"
	}

	return def
}
