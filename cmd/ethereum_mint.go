package cmd

import (
	"fmt"
	"math/big"

	vgethereum "code.vegaprotocol.io/shared/libs/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

var (
	ethereumMintOpts struct {
		ownerPrivateKey string
		tokenAddress    string
		toAddress       string
		amount          int64
	}

	ethereumMintCmd = &cobra.Command{
		Use:   "mint",
		Short: "Mint allows a token to be minted by a Base Faucet Token contract.",
		RunE:  ethereumMintToken,
	}
)

func init() {
	ethereumCmd.AddCommand(ethereumMintCmd)

	ethereumMintCmd.Flags().StringVar(&ethereumMintOpts.ownerPrivateKey, "owner-private-key", "", "private key of the token contract owner")
	ethereumMintCmd.Flags().StringVar(&ethereumMintOpts.tokenAddress, "token-addr", "", "smart contract address of the token")
	ethereumMintCmd.Flags().StringVar(&ethereumMintOpts.toAddress, "to-addr", "", "address of where the token will be minted to")
	ethereumMintCmd.Flags().Int64Var(&ethereumMintOpts.amount, "amount", 0, "amount to be minted")
	ethereumMintCmd.MarkFlagRequired("owner-private-key")
	ethereumMintCmd.MarkFlagRequired("token-address")
	ethereumMintCmd.MarkFlagRequired("to-address")
	ethereumMintCmd.MarkFlagRequired("amount")
}

func ethereumMintToken(cmd *cobra.Command, args []string) error {
	client, err := vgethereum.NewClient(cmd.Context(), ethereumOpts.address, ethereumOpts.chainID)
	if err != nil {
		return fmt.Errorf("falied to create Ethereum client: %w", err)
	}

	tokenSession, err := client.NewBaseTokenSession(
		cmd.Context(),
		ethereumMintOpts.ownerPrivateKey,
		common.HexToAddress(ethereumMintOpts.tokenAddress),
		&defaultEthSyncTimeout,
	)
	if err != nil {
		return fmt.Errorf("failed to create base token session for %s: %w", ethereumMintOpts.tokenAddress, err)
	}

	tx, err := tokenSession.MintSync(
		common.HexToAddress(ethereumMintOpts.toAddress),
		big.NewInt(ethereumMintOpts.amount),
	)
	if err != nil {
		return fmt.Errorf("failed to mint token: %w", err)
	}

	return printEthereumTx(tx)
}
