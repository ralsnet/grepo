package main

import (
	"github.com/ralsnet/grepo"
	"github.com/ralsnet/grepo/example/usecase"
	"github.com/spf13/cobra"
)

var findUsersInput usecase.FindUsersInput

var findUsersCmd = &cobra.Command{
	Use:   "find-users",
	Short: "Find users by criteria",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		api := ctx.Value("api").(*grepo.API)

		output, err := grepo.UseCase[usecase.FindUsersInput, usecase.FindUsersOutput](api, usecase.FindUsersOperation).Execute(ctx, findUsersInput)
		if err != nil {
			return err
		}
		return PrintJSON(output)
	},
}

func init() {
	findUsersCmd.Flags().StringVar(&findUsersInput.Name, "name", "", "Filter users whose names contain the specified substring")
}
