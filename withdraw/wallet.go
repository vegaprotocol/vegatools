package withdraw

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

type Store struct {
	ks *keystore.KeyStore
}

func NewStore() *Store {
	return &Store{
		ks: keystore.NewKeyStore("./unsafe_withdraw_keystore", keystore.StandardScryptN, keystore.StandardScryptP),
	}
}

func (s *Store) SignWithPassphrase(address, passphrase string, data []byte) ([]byte, error) {
	acc := accounts.Account{
		Address: ethcmn.Address(ethcmn.HexToAddress(address)),
	}
	return s.ks.SignHashWithPassphrase(acc, passphrase, data)
}

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
