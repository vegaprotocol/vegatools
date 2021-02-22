package withdraw

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// this is irrelevant, as we create a temporary
	// keystore for the key that we import
	passphrase = "vega"

	withdrawContractName = "withdraw_asset"
)

func UnsafeWithdrawImpersonateValidators(
	privKeys []string,
	receiverAddress string,
	ethereumAssetID string,
	bridgeAddress string,
	amountStr string,
) error {
	// validate amount
	amount, ok := big.NewInt(0).SetString(amountStr, 10)
	if !ok {
		return errors.New("invalid amount, make sure you specified a base 10 price without decimals")
	}

	// the keystore where we will temporay load our private keys
	wstore := NewStore()
	// the final signed payload
	sigs := "0x"
	msgs := map[string]struct{}{}

	var finalmsg string
	expiry := time.Now().Unix()

	nonce := big.NewInt(expiry + 42)

	for _, priv := range privKeys {
		address, err := wstore.Import(priv, passphrase)
		if err != nil {
			return fmt.Errorf("unable to import key: %w, (you may want to delete the unsafe_withdraw_keystore folder)", err)
		}
		fmt.Printf("generating signature for address: %v\n", address)

		msg, sig, err := signWithdrawal(
			amount, expiry, nonce, wstore, passphrase, address, bridgeAddress, ethereumAssetID, receiverAddress)
		if err != nil {
			return fmt.Errorf("unable to create signature: %w", err)
		}

		fmt.Printf("0x%v\n", hex.EncodeToString(sig))

		msgs[string(msg)] = struct{}{}
		finalmsg = "0x" + hex.EncodeToString(msg)
		sigs = sigs + hex.EncodeToString(sig)
	}

	fmt.Printf("bridge address: %v\n", bridgeAddress)
	fmt.Printf("asset source: %v\n", ethereumAssetID)
	fmt.Printf("receiver: %v\n", receiverAddress)
	fmt.Printf("msgs are unique: %v\n", len(msgs) == 1)
	fmt.Printf("msg: %v\n", finalmsg)
	fmt.Printf("sig bundle: %v\n", sigs)
	fmt.Printf("amount: %v\n", amount)
	fmt.Printf("expiry: %v\n", expiry)
	fmt.Printf("nonce: %v\n", nonce)

	return nil
}

func signWithdrawal(
	amount *big.Int,
	expiry int64,
	withdrawRef *big.Int,
	store *Store,
	passphrase string,
	account string,
	bridgeAddress string,
	assetAddress string,
	receiverAddress string,
) (msg []byte, sig []byte, err error) {
	typAddr, err := abi.NewType("address", "", nil)
	if err != nil {
		return nil, nil, err
	}
	typString, err := abi.NewType("string", "", nil)
	if err != nil {
		return nil, nil, err
	}
	typU256, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, nil, err
	}
	typBytes, err := abi.NewType("bytes", "", nil)
	if err != nil {
		return nil, nil, err
	}

	args := abi.Arguments([]abi.Argument{
		{
			Name: "address",
			Type: typAddr,
		},
		{
			Name: "uint256",
			Type: typU256,
		},
		{
			Name: "uint256",
			Type: typU256,
		},
		{
			Name: "address",
			Type: typAddr,
		},
		{
			Name: "nonce",
			Type: typU256,
		},
		{
			Name: "func_name",
			Type: typString,
		},
	})

	hexAssetAddress := ethcmn.HexToAddress(assetAddress)
	hexReceiverAddress := ethcmn.HexToAddress(receiverAddress)

	// we use the withdrawRef as a nonce
	// they are unique as generated as an increment from the banking
	// layer
	buf, err := args.Pack([]interface{}{hexAssetAddress, amount, big.NewInt(expiry), hexReceiverAddress, withdrawRef, withdrawContractName}...)
	if err != nil {
		return nil, nil, err
	}

	hexBridgeAddress := ethcmn.HexToAddress(bridgeAddress)
	args2 := abi.Arguments([]abi.Argument{
		{
			Name: "bytes",
			Type: typBytes,
		},
		{
			Name: "address",
			Type: typAddr,
		},
	})

	msg, err = args2.Pack(buf, hexBridgeAddress)
	if err != nil {
		return nil, nil, err
	}

	// hash our message before signing it
	hash := crypto.Keccak256(msg)

	// now sign the message using our wallet private key
	sig, err = store.SignWithPassphrase(account, passphrase, hash)
	if err != nil {
		return nil, nil, err
	}

	return msg, sig, nil
}
