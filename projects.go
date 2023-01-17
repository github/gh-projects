package main

import (
	"log"

	"github.com/cli/cli/v2/pkg/cmd/factory"
	cmdClose "github.com/github/gh-projects/cmd/close"
	cmdCreate "github.com/github/gh-projects/cmd/create"
	cmdList "github.com/github/gh-projects/cmd/list"
	"github.com/spf13/cobra"
)

// analogous to cli/pkg/cmd/pr.go in cli/cli
func main() {
	var rootCmd = &cobra.Command{
		Use: "projects",
	}

	cmdFactory := factory.New("0.1.0") // will be replaced by buildVersion := build.Version

	rootCmd.AddCommand(cmdList.NewCmdList(cmdFactory, nil))
	rootCmd.AddCommand(cmdCreate.NewCmdCreate(cmdFactory, nil))
	rootCmd.AddCommand(cmdClose.NewCmdClose(cmdFactory, nil))

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
