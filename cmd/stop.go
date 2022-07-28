package cmd

import (
	"code.hzmantu.com/dts/utils/process"
	"github.com/spf13/cobra"
)

const (
	pidFile = "dts.pid"
)

func init() {
	var cmd = &cobra.Command{
		Use:   "stop",
		Short: "stop source process",
		Long:  "stop source process",
		Run: func(cmd *cobra.Command, args []string) {
			stop()
		},
	}
	// root
	rootCmd.AddCommand(cmd)
}

func stop() {
	pid := process.NewPid(pidFile)
	pid.Kill()
}
