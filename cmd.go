package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	// DefaultConfigPath is the default path to the config file
	DefaultConfigPath = "grape.json"
)

var RunNotice = `🍇 now watching for changes ✨`
var StopNotice = `🍇 stopped watching for changes, cleaning up... ✨`

var onCmd = &cobra.Command{
	Use:   "on",
	Short: "use [on] to configure grape on the go without a config file.",
	Run: func(cmd *cobra.Command, args []string) {
		targets, _ := cmd.Flags().GetStringSlice("ext")
		exclude, _ := cmd.Flags().GetStringSlice("exclude")
		run, _ := cmd.Flags().GetString("run")
		if run == "" {
			fmt.Println(failText("run command is required"))
			cmd.Help()
			return
		}

		config, err := FromFlags(run, targets, exclude)
		if err != nil {
			fmt.Println(failText(err.Error()))
			cmd.Help()
			return
		}

		if err := Run(config, "default"); err != nil {
			fmt.Println(failText(err.Error()))
			cmd.Help()
			return
		}
	},
}

var grapeCmd = &cobra.Command{
	Use:  "grape",
	Long: `🍇 grape is a tiny tool for watching files and running commands when they change during development.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "use [run] to run grape with a config file and switch between namespaces.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println(failText("namespace argument is missing"))
			cmd.Help()
			return
		}

		config, err := FromJson(cmd.Flag("config").Value.String())
		if err != nil {
			fmt.Println(failText(err.Error()))
			cmd.Help()
			return
		}

		namespace := args[0]
		if err := Run(config, namespace); err != nil {
			fmt.Println(failText(err.Error()))
			cmd.Help()
		}
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "use [init] to create a config file in the current directory.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := NewDefault(); err != nil {
			fmt.Println(failText(err.Error()))
			cmd.Help()
		}
	},
}

func RootCmd() *cobra.Command {
	grapeCmd.AddCommand(runCmd)
	grapeCmd.AddCommand(onCmd)
	grapeCmd.AddCommand(initCmd)
	onCmd.Flags().StringSliceP("ext", "e", []string{"*.go"}, "comma separated list of file extensions to watch [ default: .go ]")
	onCmd.Flags().StringSliceP("exclude", "x", []string{"vendor"}, "comma separated list of directories to exclude from watching")
	onCmd.Flags().StringP("run", "r", "", "command to run when a file is changed")
	runCmd.Flags().StringP("config", "c", DefaultConfigPath, "path to config file")
	return grapeCmd
}
