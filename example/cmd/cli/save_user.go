package main

import (
	"github.com/ralsnet/grepo"
	"github.com/ralsnet/grepo/example/usecase"
	"github.com/spf13/cobra"
)

var saveUserInput usecase.SaveUserInput

var saveUserCmd = &cobra.Command{
	Use:   "save-user",
	Short: "Save a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		api := ctx.Value("api").(*grepo.API)

		output, err := grepo.UseCase[usecase.SaveUserInput, usecase.SaveUserOutput](api, usecase.SaveUserOperation).Execute(ctx, saveUserInput)
		if err != nil {
			return err
		}
		return PrintJSON(output)
	},
}

func init() {
	saveUserCmd.Flags().StringVar(&saveUserInput.Name, "name", "", "Name of the user to save")
	saveUserCmd.MarkFlagRequired("name")

	saveUserCmd.Flags().StringVar(&saveUserInput.Authority, "authority", "", "Authority of the user to save (options: admin, user)")
	saveUserCmd.MarkFlagRequired("authority")
}
