package perftest

import (
	"context"
	"fmt"
	"math/big"

	"code.vegaprotocol.io/shared/libs/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	stakingBridgeAddress    = common.HexToAddress("0x5E93B6db35EeD20C62E4a9f2a37d15e2D1359eA1")
	vegaTokenAddress        = common.HexToAddress("0x67175Da1D5e966e40D11c4B2519392B2058373de")
	contractOwnerPrivateKey = "a37f4c2a678aefb5037bf415a826df1540b330b7e471aa54184877ba901b9ef0"
)

type token interface {
	Mint(to common.Address, amount *big.Int) (*types.Transaction, error)
	MintSync(to common.Address, amount *big.Int) (*types.Transaction, error)
	BalanceOf(account common.Address) (*big.Int, error)
	ApproveSync(spender common.Address, value *big.Int) (*types.Transaction, error)
	Address() common.Address
}

func approveAndStakeToken(token token, bridge *ethereum.StakingBridgeSession, amount *big.Int, vegaPubKey string) error {
	if _, err := token.ApproveSync(bridge.Address(), amount); err != nil {
		return fmt.Errorf("failed to approve token: %w", err)
	}

	vegaPubKeyByte32, err := ethereum.HexStringToByte32Array(vegaPubKey)
	if err != nil {
		return err
	}

	if _, err := bridge.Stake(amount, vegaPubKeyByte32); err != nil {
		return fmt.Errorf("failed to stake asset: %w", err)
	}
	return nil
}
func sendVegaTokens(vegaPubKey, ganacheURL string) error {
	ctx := context.Background()

	url := "ws://" + ganacheURL

	// Create a connection to ganache
	client, err := ethereum.NewClient(ctx, url)
	if err != nil {
		return err
	}

	stakingBridge, err := client.NewStakingBridgeSession(ctx, contractOwnerPrivateKey, stakingBridgeAddress, nil)
	if err != nil {
		return err
	}

	vegaToken, err := client.NewBaseTokenSession(ctx, contractOwnerPrivateKey, vegaTokenAddress, nil)
	if err != nil {
		return err
	}

	if err := approveAndStakeToken(vegaToken, stakingBridge, big.NewInt(1000000000000000000), vegaPubKey); err != nil {
		return err
	}
	return nil
}
