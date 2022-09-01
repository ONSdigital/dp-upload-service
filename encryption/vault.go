package encryption

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

//go:generate moq -out mock/vault.go -pkg mock_encryption . VaultClienter

const (
	vaultKey = "key"
)

var (
	ErrVaultWrite           = errors.New("failed to write to vault")
	ErrVaultRead            = errors.New("failed to read from vault")
	ErrKeyGeneration        = errors.New("failed to generate encryption key")
	ErrInvalidEncryptionKey = errors.New("encryption key invalid")
)

type Vault struct {
	keyGenerator GenerateKey
	client       VaultClienter
	path         string
}

type VaultClienter interface {
	ReadKey(path, key string) (string, error)
	WriteKey(path, key, value string) error
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

func NewVault(keyGenerator GenerateKey, client VaultClienter, path string) *Vault {
	return &Vault{keyGenerator, client, path}
}

func (v *Vault) GenerateEncryptionKey(ctx context.Context, filepath string) ([]byte, error) {
	encryptionKey, err := v.keyGenerator()
	if err != nil {
		log.Error(ctx, "failed to generate encryption key", err)
		return nil, ErrKeyGeneration
	}
	if err := v.client.WriteKey(v.vaultPath(filepath), vaultKey, hex.EncodeToString(encryptionKey)); err != nil {
		log.Error(ctx, "failed to write encryption encryptionKey to vault", err, log.Data{"vault-path": v.vaultPath(filepath), "vault-encryptionKey": vaultKey})
		return nil, ErrVaultWrite
	}

	return encryptionKey, nil
}

func (v *Vault) EncryptionKey(ctx context.Context, filepath string) ([]byte, error) {
	strKey, err := v.client.ReadKey(v.vaultPath(filepath), vaultKey)
	if err != nil {
		log.Error(ctx, "failed to read encryption encryptionkey from vault", err, log.Data{"vault-path": v.vaultPath(filepath), "vault-encryptionkey": vaultKey})
		return nil, ErrVaultRead
	}

	encryptionKey, err := hex.DecodeString(strKey)
	if err != nil {
		log.Error(ctx, "encryption key contains non-hexadecimal characters", err, log.Data{"vault-path": v.vaultPath(filepath), "vault-encryptionkey": vaultKey})
		return nil, ErrInvalidEncryptionKey
	}
	return encryptionKey, nil
}

func (v *Vault) vaultPath(filepath string) string {
	return fmt.Sprintf("%s/%s", v.path, filepath)
}
