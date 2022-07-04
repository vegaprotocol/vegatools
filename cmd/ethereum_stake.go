package cmd

import (
	"fmt"
	"math/big"

	vgethereum "code.vegaprotocol.io/shared/libs/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

var (
	ethereumStakeOpts struct {
		ownerPrivateKey string
		bridgeAddress   string
		assetAddress    string
		vegaPubKey      string
		amount          int64
	}

	ethereumStakeCmd = &cobra.Command{
		Use:   "stake",
		Short: "Stake allows to stake an asset to given Vega public key.",
		RunE:  ethereumStake,
	}
)

func init() {
	ethereumCmd.AddCommand(ethereumStakeCmd)

	ethereumStakeCmd.Flags().StringVar(&ethereumStakeOpts.ownerPrivateKey, "owner-private-key", "", "private key of the bridge contract owner")
	ethereumStakeCmd.Flags().StringVar(&ethereumStakeOpts.bridgeAddress, "bridge-addr", "", "smart contract address of the bridge")
	ethereumStakeCmd.Flags().StringVar(&ethereumStakeOpts.assetAddress, "asset-addr", "", "address of the asset to be staked")
	ethereumStakeCmd.Flags().StringVar(&ethereumStakeOpts.vegaPubKey, "pub-key", "", "Vega public key to where the asset will be staked")
	ethereumStakeCmd.Flags().Int64Var(&ethereumStakeOpts.amount, "amount", 0, "amount to be staked")
	ethereumStakeCmd.MarkFlagRequired("owner-private-key")
	ethereumStakeCmd.MarkFlagRequired("bridge-addr")
	ethereumStakeCmd.MarkFlagRequired("asset-addr")
	ethereumStakeCmd.MarkFlagRequired("pub-key")
	ethereumStakeCmd.MarkFlagRequired("amount")
}

func ethereumStake(cmd *cobra.Command, args []string) error {
	client, err := vgethereum.NewClient(cmd.Context(), ethereumOpts.address, ethereumOpts.chainID)
	if err != nil {
		return fmt.Errorf("falied to create Ethereum client: %w", err)
	}

	bridgeAddr := common.HexToAddress(ethereumStakeOpts.bridgeAddress)

	bridgeSession, err := client.NewStakingBridgeSession(
		cmd.Context(),
		ethereumStakeOpts.ownerPrivateKey,
		bridgeAddr,
		&defaultEthSyncTimeout,
	)
	if err != nil {
		return fmt.Errorf("failed to create staking bridge session for %s: %w", ethereumStakeOpts.bridgeAddress, err)
	}

	tokenSession, err := client.NewBaseTokenSession(
		cmd.Context(),
		ethereumStakeOpts.ownerPrivateKey,
		common.HexToAddress(ethereumStakeOpts.assetAddress),
		&defaultEthSyncTimeout,
	)
	if err != nil {
		return fmt.Errorf("failed to create token session for %s: %w", ethereumStakeOpts.assetAddress, err)
	}

	amount := big.NewInt(ethereumStakeOpts.amount)

	if _, err := tokenSession.ApproveSync(bridgeAddr, amount); err != nil {
		return fmt.Errorf("failef to approve asset amount to bridge: %w", err)
	}

	vegaPubKeyArr, err := vgethereum.HexStringToByte32Array(ethereumStakeOpts.vegaPubKey)
	if err != nil {
		return fmt.Errorf("failed to convert Vega pub key string to byte array: %w", err)
	}

	tx, err := bridgeSession.Stake(
		big.NewInt(ethereumStakeOpts.amount),
		vegaPubKeyArr,
	)
	if err != nil {
		return fmt.Errorf("failed to stake asset: %w", err)
	}

	return printEthereumTx(tx)
}
