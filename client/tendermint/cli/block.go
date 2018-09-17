package cli

import (
	"fmt"
	"github.com/irisnet/irishub/client"
	"github.com/irisnet/irishub/client/context"
	"github.com/spf13/cobra"
	tmliteProxy "github.com/tendermint/tendermint/lite/proxy"
	"strconv"
)

//BlockCommand returns the verified block data for a given heights
func BlockCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block [height]",
		Short: "Get verified data for a the block at given height",
		Args:  cobra.MaximumNArgs(1),
		RunE:  printBlock,
	}
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().Bool(client.FlagTrustNode, false, "Trust connected full node (don't verify proofs for responses)")
	cmd.Flags().String(client.FlagChainID, "", "Chain ID of Tendermint node")
	return cmd
}

func getBlock(cliCtx context.CLIContext, height *int64) ([]byte, error) {
	// get the node
	node, err := cliCtx.GetNode()
	if err != nil {
		return nil, err
	}

	// header -> BlockchainInfo
	// header, tx -> Block
	// results -> BlockResults
	res, err := node.Block(height)
	if err != nil {
		return nil, err
	}

	if !cliCtx.TrustNode {
		check, err := cliCtx.Certify(res.Block.Height)
		if err != nil {
			return nil, err
		}

		err = tmliteProxy.ValidateBlockMeta(res.BlockMeta, check)
		if err != nil {
			return nil, err
		}

		err = tmliteProxy.ValidateBlock(res.Block, check)
		if err != nil {
			return nil, err
		}
	}

	// TODO move maarshalling into cmd/rest functions
	// output, err := tmcodec.MarshalJSON(res)
	output, err := cdc.MarshalJSONIndent(res,"", "  ")
	if err != nil {
		return nil, err
	}
	return output, nil
}

// get the current blockchain height
func GetChainHeight(cliCtx context.CLIContext) (int64, error) {
	node, err := cliCtx.GetNode()
	if err != nil {
		return -1, err
	}
	status, err := node.Status()
	if err != nil {
		return -1, err
	}
	height := status.SyncInfo.LatestBlockHeight
	return height, nil
}

// CMD

func printBlock(cmd *cobra.Command, args []string) error {
	var height *int64
	// optional height
	if len(args) > 0 {
		h, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		if h > 0 {
			tmp := int64(h)
			height = &tmp
		}
	}

	output, err := getBlock(context.NewCLIContext(), height)
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
