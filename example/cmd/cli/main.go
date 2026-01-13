package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ralsnet/grepo/example/internal"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "example-cli",
	Short: "An example CLI application",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	rootCmd.AddCommand(specCmd)
	rootCmd.AddCommand(findUsersCmd)
	rootCmd.AddCommand(getUserCmd)
	rootCmd.AddCommand(saveUserCmd)

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		api := internal.InitializeAPI()

		ctx := cmd.Context()
		ctx = context.WithValue(ctx, "api", api)

		cmd.SetContext(ctx)

		return nil
	}
}

func main() {
	if err := Execute(); err != nil {
		panic(err)
	}
}
