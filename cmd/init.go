package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	_ "net/http/pprof"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "cmd",
	Short: "cmd some",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
