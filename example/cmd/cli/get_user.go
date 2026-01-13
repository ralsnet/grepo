package main

import (
	"github.com/ralsnet/grepo"
	"github.com/ralsnet/grepo/example/usecase"
	"github.com/spf13/cobra"
)

var getUserInput usecase.GetUserInput

var getUserCmd = &cobra.Command{
	Use:   "get-user",
	Short: "Get a user by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		api := ctx.Value("api").(*grepo.API)

		output, err := grepo.UseCase[usecase.GetUserInput, usecase.GetUserOutput](api, usecase.GetUserOperation).Execute(ctx, getUserInput)
		if err != nil {
			return err
		}
		return PrintJSON(output)
	},
}

func init() {
	getUserCmd.Flags().StringVar(&getUserInput.ID, "id", "", "ID of the user to retrieve")
	getUserCmd.MarkFlagRequired("id")
}
