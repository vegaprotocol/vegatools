package withdraw

import (
	"encoding/hex"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

var keystoreDir = ".unsafe_withdraw_keystore.tmp"

// Store holds information about Key Store
type Store struct {
	ks *keystore.KeyStore
}

// NewStore creates new instance of Store
func NewStore() *Store {
	return &Store{
		ks: keystore.NewKeyStore(filepath.Join(".", keystoreDir), keystore.StandardScryptN, keystore.StandardScryptP),
	}
}

// SignWithPassphrase sign to the Key Store
func (s *Store) SignWithPassphrase(address, passphrase string, data []byte) ([]byte, error) {
	acc := accounts.Account{
		Address: ethcmn.Address(ethcmn.HexToAddress(address)),
	}
	return s.ks.SignHashWithPassphrase(acc, passphrase, data)
}

// Import gets account address from Key Store
func (s *Store) Import(privKey, passphrase string) (string, error) {
	privBytes, err := hex.DecodeString(privKey)
	if err != nil {
		return "", err
	}
	key, err := ethcrypto.ToECDSA(privBytes)
	if err != nil {
		return "", err
	}
	account, err := s.ks.ImportECDSA(key, passphrase)
	if err != nil {
		return "", err
	}
	return account.Address.Hex(), nil
}
