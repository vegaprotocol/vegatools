package cmd

import (
	"fmt"
	"math/big"

	vgethereum "code.vegaprotocol.io/shared/libs/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

var (
	ethereumDepositOpts struct {
		ownerPrivateKey string
		bridgeAddress   string
		assetAddress    string
		vegaPubKey      string
		amount          int64
	}

	ethereumDepositCmd = &cobra.Command{
		Use:   "deposit",
		Short: "Deposit allows to deposit an asset to given Vega public key.",
		RunE:  ethereumDeposit,
	}
)

func init() {
	ethereumCmd.AddCommand(ethereumDepositCmd)

	ethereumDepositCmd.Flags().StringVar(&ethereumDepositOpts.ownerPrivateKey, "owner-private-key", "", "private key of the bridge contract owner")
	ethereumDepositCmd.Flags().StringVar(&ethereumDepositOpts.bridgeAddress, "bridge-addr", "", "smart contract address of the bridge")
	ethereumDepositCmd.Flags().StringVar(&ethereumDepositOpts.assetAddress, "asset-addr", "", "address of the asset to be deposited")
	ethereumDepositCmd.Flags().StringVar(&ethereumDepositOpts.vegaPubKey, "pub-key", "", "Vega public key to where the asset will be deposited")
	ethereumDepositCmd.Flags().Int64Var(&ethereumDepositOpts.amount, "amount", 0, "amount to be deposited")
	ethereumDepositCmd.MarkFlagRequired("owner-private-key")
	ethereumDepositCmd.MarkFlagRequired("bridge-addr")
	ethereumDepositCmd.MarkFlagRequired("asset-addr")
	ethereumDepositCmd.MarkFlagRequired("pub-key")
	ethereumDepositCmd.MarkFlagRequired("amount")
}

func ethereumDeposit(cmd *cobra.Command, args []string) error {
	client, err := vgethereum.NewClient(cmd.Context(), ethereumOpts.address, ethereumOpts.chainID)
	if err != nil {
		return fmt.Errorf("falied to create Ethereum client: %w", err)
	}

	bridgeAddr := common.HexToAddress(ethereumDepositOpts.bridgeAddress)

	bridgeSession, err := client.NewERC20BridgeSession(
		cmd.Context(),
		ethereumDepositOpts.ownerPrivateKey,
		bridgeAddr,
		&defaultEthSyncTimeout,
	)
	if err != nil {
		return fmt.Errorf("failed to create erc20 bridge session for %s: %w", ethereumDepositOpts.bridgeAddress, err)
	}

	tokenSession, err := client.NewBaseTokenSession(
		cmd.Context(),
		ethereumDepositOpts.ownerPrivateKey,
		common.HexToAddress(ethereumDepositOpts.assetAddress),
		&defaultEthSyncTimeout,
	)
	if err != nil {
		return fmt.Errorf("failed to create token session for %s: %w", ethereumDepositOpts.assetAddress, err)
	}

	amount := big.NewInt(ethereumDepositOpts.amount)

	if _, err := tokenSession.ApproveSync(bridgeAddr, amount); err != nil {
		return fmt.Errorf("failef to approve asset amount to bridge: %w", err)
	}

	vegaPubKeyArr, err := vgethereum.HexStringToByte32Array(ethereumDepositOpts.vegaPubKey)
	if err != nil {
		return fmt.Errorf("failed to convert Vega pub key string to byte array: %w", err)
	}

	tx, err := bridgeSession.DepositAssetSync(
		common.HexToAddress(ethereumDepositOpts.assetAddress),
		amount,
		vegaPubKeyArr,
	)
	if err != nil {
		return fmt.Errorf("failed to deposit asset: %w", err)
	}

	return printEthereumTx(tx)
}
