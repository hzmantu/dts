package cmd

import (
	"code.hzmantu.com/dts/structs"
	"code.hzmantu.com/dts/task"
	"code.hzmantu.com/dts/utils/process"
	"github.com/spf13/cobra"
	"net/http"
)

var (
	configFile string
)

func init() {
	var cmd = &cobra.Command{
		Use:   "reader",
		Short: "reader source data to queue",
		Long:  "reader source data to queue",
		Run: func(cmd *cobra.Command, args []string) {
			reader()
		},
	}
	// args
	cmd.PersistentFlags().StringVar(&configFile, "config", "task.yaml", "config yaml")
	// root
	rootCmd.AddCommand(cmd)
}

func reader() {
	pid := process.NewPid(pidFile)
	pid.SaveFile()
	defer func() {
		pid.RemoveFile()
	}()
	// reade config
	config := structs.GetConfig(configFile)
	// Indicator view
	if config.StatAddr != "" {
		go func() {
			_ = http.ListenAndServe(config.StatAddr, nil)
		}()
	}
	// start manager
	task.NewManager(config).Start()
}
