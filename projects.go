package main

import (
	"log"

	"github.com/cli/cli/v2/pkg/cmd/factory"
	cmdClose "github.com/github/gh-projects/cmd/close"
	cmdCreate "github.com/github/gh-projects/cmd/create"
	cmdEdit "github.com/github/gh-projects/cmd/edit"
	cmdItemArchive "github.com/github/gh-projects/cmd/item/archive"
	cmdItemCreate "github.com/github/gh-projects/cmd/item/create"
	cmdItemDelete "github.com/github/gh-projects/cmd/item/delete"
	cmdItemList "github.com/github/gh-projects/cmd/item/list"
	cmdList "github.com/github/gh-projects/cmd/list"
	"github.com/spf13/cobra"
)

// analogous to cli/pkg/cmd/pr.go in cli/cli
func main() {
	var rootCmd = &cobra.Command{
		Use: "projects",
	}

	var itemCmd = &cobra.Command{
		Use:   "item",
		Short: "Commands for items",
	}

	cmdFactory := factory.New("0.1.0") // will be replaced by buildVersion := build.Version

	rootCmd.AddCommand(cmdList.NewCmdList(cmdFactory, nil))
	rootCmd.AddCommand(cmdCreate.NewCmdCreate(cmdFactory, nil))
	rootCmd.AddCommand(cmdClose.NewCmdClose(cmdFactory, nil))
	rootCmd.AddCommand(cmdEdit.NewCmdEdit(cmdFactory, nil))

	// item subcommand
	rootCmd.AddCommand(itemCmd)
	itemCmd.AddCommand(cmdItemList.NewCmdList(cmdFactory, nil))
	itemCmd.AddCommand(cmdItemCreate.NewCmdCreateItem(cmdFactory, nil))
	itemCmd.AddCommand(cmdItemArchive.NewCmdArchiveItem(cmdFactory, nil))
	itemCmd.AddCommand(cmdItemDelete.NewCmdDeleteItem(cmdFactory, nil))

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
