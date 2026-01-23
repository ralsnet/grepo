package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/ralsnet/grepo"
	"github.com/ralsnet/grepo/refl"
	"github.com/spf13/cobra"
)

type apikey struct{}

func WithAPIContext(ctx context.Context, api *grepo.API) context.Context {
	return context.WithValue(ctx, apikey{}, api)
}

type SetupFunc func(cmd *cobra.Command, uc grepo.Descriptor)

func New(api *grepo.API, name string, setups ...SetupFunc) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   name,
		Short: api.Description(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ctx = WithAPIContext(ctx, api)
			cmd.SetContext(ctx)
			return nil
		},
	}

	for _, uc := range api.UseCases() {
		cmd := newUseCaseCommand(uc, setups...)
		rootCmd.AddCommand(cmd)
	}
	rootCmd.AddCommand(specCmd(api))

	return rootCmd
}

func newUseCaseCommand(uc grepo.Descriptor, setups ...SetupFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   uc.Operation(),
		Short: uc.Description(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			api := ctx.Value(apikey{}).(*grepo.API)

			input, err := getInput(cmd, args, uc)
			if err != nil {
				return err
			}

			output, err := api.ExecuteAny(ctx, uc.Operation(), input)
			if err != nil {
				return err
			}

			b, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return err
			}
			_, err = os.Stdout.Write(b)
			if err != nil {
				return err
			}

			return nil
		},
	}
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("%s\n\n", uc.Operation()))
	if desc := uc.Description(); desc != "" {
		b.WriteString(fmt.Sprintf("%s\n\n", desc))
	}

	input := uc.Input()
	inputSpec := refl.TypeOf(input)
	inputJSON, _ := json.MarshalIndent(inputSpec, "", "  ")
	b.WriteString("Input schema:\n")
	b.WriteString(string(inputJSON))

	output := uc.Output()
	outputSpec := refl.TypeOf(output)
	outputJSON, _ := json.MarshalIndent(outputSpec, "", "  ")
	b.WriteString("\n\n")
	b.WriteString("Output schema:\n")
	b.WriteString(string(outputJSON))

	cmd.Long = b.String()

	cmd.Flags().String("input", "", "Path to JSON file containing input data")
	cmd.Flags().Bool("stdin", false, "Read input data from standard input")
	cmd.ArgAliases = []string{"input-data"}

	for _, setup := range setups {
		setup(cmd, uc)
	}

	return cmd
}

func specCmd(api *grepo.API) *cobra.Command {
	return &cobra.Command{
		Use:   "spec",
		Short: "Show the API specification",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := json.MarshalIndent(api, "", "  ")
			if err != nil {
				return err
			}
			_, err = os.Stdout.Write(b)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func getInput(cmd *cobra.Command, args []string, uc grepo.Descriptor) (any, error) {
	var b []byte

	if flagInput, err := cmd.Flags().GetString("input"); err == nil && flagInput != "" {
		b, err = os.ReadFile(flagInput)
	} else if len(args) > 0 {
		b = []byte(args[0])
	} else if flagStdin, err := cmd.Flags().GetBool("stdin"); err == nil && flagStdin {
		stdin := cmd.InOrStdin()
		b, err = io.ReadAll(stdin)
	} else {
		b = []byte("{}")
	}

	p := reflect.New(reflect.TypeOf(uc.Input())).Interface()
	if err := json.Unmarshal(b, p); err != nil {
		return nil, err
	}

	return reflect.ValueOf(p).Elem().Interface(), nil
}
