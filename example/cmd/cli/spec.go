package main

import (
	"github.com/ralsnet/grepo"
	"github.com/spf13/cobra"
)

var specCmd = &cobra.Command{
	Use:   "spec",
	Short: "Show the API specification",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		api := ctx.Value("api").(*grepo.API)
		return PrintJSON(api)
	},
}
