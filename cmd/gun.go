package cmd

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"strconv"
	"time"
)

func gunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gun [src-chain-id] [dst-chain-id] [amount] [source] [dst-chain-addr] [msgs-count]",
		Aliases: []string{"g"},
		Short:   "transfer tokens from a source chain to a destination chain in one command",
		Long:    "This sends tokens from a relayers configured wallet on chain src to a dst addr on dst",
		Args:    cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, dst := args[0], args[1]
			c, err := config.Chains.Gets(src, dst)
			if err != nil {
				return err
			}

			pth, err := cmd.Flags().GetString(flagPath)
			if err != nil {
				return err
			}

			gas, err := cmd.Flags().GetUint64(flagGas)
			if err != nil {
				return err
			}

			if _, err = setPathsFromArgs(c[src], c[dst], pth); err != nil {
				return err
			}

			amount, err := sdk.ParseCoin(args[2])
			if err != nil {
				return err
			}

			source, err := strconv.ParseBool(args[3])
			if err != nil {
				return err
			}

			msgsCount, err := strconv.Atoi(args[5])
			if err != nil {
				return err
			}

			dstAddr, err := sdk.AccAddressFromBech32(args[4])
			if err != nil {
				return err
			}

			c[src].NewGas = gas
			c[dst].NewGas = gas

			return c[src].Gun(c[dst], amount, dstAddr, source, msgsCount)
		},
	}
	cmd = pathFlag(cmd)
	return gasFlag(cmd)
}

func slowGunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gun_update_client [src-chain-id] [dst-chain-id] [timeout]",
		Aliases: []string{"g"},
		Short:   "update client",
		Long:    "",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, dst := args[0], args[1]
			c, err := config.Chains.Gets(src, dst)
			if err != nil {
				return err
			}

			pth, err := cmd.Flags().GetString(flagPath)
			if err != nil {
				return err
			}

			if _, err = setPathsFromArgs(c[src], c[dst], pth); err != nil {
				return err
			}

			timeout, err := time.ParseDuration(args[2])
			if err != nil {
				return err
			}

			return c[src].SlowGun(c[dst], timeout)
		},
	}
	cmd = pathFlag(cmd)
	return gasFlag(cmd)
}
