package main

import (
	"fmt"
	"os"

	"github.com/ralsnet/grepo/cli"
	"github.com/ralsnet/grepo/example/internal"
)

func main() {
	api := internal.InitializeAPI()
	if err := cli.New(api, "clitest").Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
