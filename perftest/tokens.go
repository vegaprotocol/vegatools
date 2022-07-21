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
	stakingBridgeAddress    = common.HexToAddress("0x9135f5afd6F055e731bca2348429482eE614CFfA")
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
func sendVegaTokens(vegaPubKey string) error {
	ctx := context.Background()

	// Create a connection to ganache
	client, err := ethereum.NewClient(ctx, "http://localhost:8545", 1440)
	if err != nil {
		return err
	}

	stakingBridge, err := client.NewStakingBridgeSession(ctx, contractOwnerPrivateKey, stakingBridgeAddress, nil)
	if err != nil {
		return err
	}

	/*	erc20Bridge, err := client.NewERC20BridgeSession(ctx, contractOwnerPrivateKey, erc20BridgeAddress, nil)
		if err != nil {
		  fmt.Errorf("Failed to create ethereum bridge: %w", err)
		}*/

	vegaToken, err := client.NewBaseTokenSession(ctx, contractOwnerPrivateKey, vegaTokenAddress, nil)
	if err != nil {
		return err
	}

	/*	tUSDCToken, err := client.NewBaseTokenSession(ctx, contractOwnerPrivateKey, tUSDCTokenAddress, nil)
		if err != nil {
		  fmt.Errorf("Failed to create tUSDC token: %w", err)
		}*/

	// mint some tokens
	/*	if err := mintTokenAndShowBalances(client, vegaToken, contractOwnerAddress, big.NewInt(1000000000000000000)); err != nil {
	  fmt.Errorf("Failed to mint and show balances for vegaToken: %w", err)
	}*/

	/*	if err := mintTokenAndShowBalances(client, tUSDCToken, contractOwnerAddress, big.NewInt(1000000000000000000)); err != nil {
	  fmt.Errorf("Failed to mint and show balances for tUSDCToken: %w", err)
	}*/

	/*	if err := approveAndDepositToken(tUSDCToken, erc20Bridge, big.NewInt(1000000000000000000), vegaPubKey); err != nil {
	  fmt.Errorf("Failed to approve and deposit token on ethereum bridge: %w", err)
	}*/

	if err := approveAndStakeToken(vegaToken, stakingBridge, big.NewInt(1000000000000000000), vegaPubKey); err != nil {
		return err
	}
	return nil
}