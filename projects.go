package main

import (
	"log"

	"github.com/cli/cli/v2/pkg/cmd/factory"
	cmdClose "github.com/github/gh-projects/cmd/close"
	cmdCreate "github.com/github/gh-projects/cmd/create"
	cmdEdit "github.com/github/gh-projects/cmd/edit"
	cmdItemsArchive "github.com/github/gh-projects/cmd/items/archive"
	cmdItemsCreate "github.com/github/gh-projects/cmd/items/create"
	cmdItemsDelete "github.com/github/gh-projects/cmd/items/delete"
	cmdItemsList "github.com/github/gh-projects/cmd/items/list"
	cmdList "github.com/github/gh-projects/cmd/list"
	"github.com/spf13/cobra"
)

// analogous to cli/pkg/cmd/pr.go in cli/cli
func main() {
	var rootCmd = &cobra.Command{
		Use: "projects",
	}

	var itemsCmd = &cobra.Command{
		Use: "items",
	}

	cmdFactory := factory.New("0.1.0") // will be replaced by buildVersion := build.Version

	rootCmd.AddCommand(cmdList.NewCmdList(cmdFactory, nil))
	rootCmd.AddCommand(cmdCreate.NewCmdCreate(cmdFactory, nil))
	rootCmd.AddCommand(cmdClose.NewCmdClose(cmdFactory, nil))
	rootCmd.AddCommand(cmdEdit.NewCmdEdit(cmdFactory, nil))

	// items subcommand
	rootCmd.AddCommand(itemsCmd)
	itemsCmd.AddCommand(cmdItemsList.NewCmdList(cmdFactory, nil))
	itemsCmd.AddCommand(cmdItemsCreate.NewCmdCreateItem(cmdFactory, nil))
	itemsCmd.AddCommand(cmdItemsArchive.NewCmdArchiveItem(cmdFactory, nil))
	itemsCmd.AddCommand(cmdItemsDelete.NewCmdDeleteItem(cmdFactory, nil))

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
