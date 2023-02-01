package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cli/cli/v2/pkg/cmd/factory"
	cmdClose "github.com/github/gh-projects/cmd/close"
	cmdCopy "github.com/github/gh-projects/cmd/copy"
	cmdCreate "github.com/github/gh-projects/cmd/create"
	cmdDelete "github.com/github/gh-projects/cmd/delete"
	cmdEdit "github.com/github/gh-projects/cmd/edit"
	cmdFieldCreate "github.com/github/gh-projects/cmd/field/create"
	cmdFieldDelete "github.com/github/gh-projects/cmd/field/delete"
	cmdFieldEdit "github.com/github/gh-projects/cmd/field/edit"
	cmdFieldList "github.com/github/gh-projects/cmd/field/list"
	cmdItemAdd "github.com/github/gh-projects/cmd/item/add"
	cmdItemArchive "github.com/github/gh-projects/cmd/item/archive"
	cmdItemCreate "github.com/github/gh-projects/cmd/item/create"
	cmdItemDelete "github.com/github/gh-projects/cmd/item/delete"
	cmdItemEdit "github.com/github/gh-projects/cmd/item/edit"
	cmdItemList "github.com/github/gh-projects/cmd/item/list"
	cmdList "github.com/github/gh-projects/cmd/list"
	"github.com/spf13/cobra"
)

// analogous to cli/pkg/cmd/pr.go in cli/cli
func main() {
	var rootCmd = &cobra.Command{
		Use:           "projects",
		Short:         "Work with GitHub Projects.",
		Long:          "Work with GitHub Projects. Note that the token you are using must have 'project' scope, which is not set by default. You can verify your token scope by running 'gh auth status' and add the project scope by running 'gh auth refresh -s project'.",
		SilenceErrors: true,
	}

	var itemCmd = &cobra.Command{
		Use:           "item",
		Short:         "Commands for items",
		SilenceErrors: true,
	}

	var fieldCmd = &cobra.Command{
		Use:           "field",
		Short:         "Commands for fields",
		SilenceErrors: true,
	}

	cmdFactory := factory.New("0.1.0") // will be replaced by buildVersion := build.Version

	rootCmd.AddCommand(cmdList.NewCmdList(cmdFactory, nil))
	rootCmd.AddCommand(cmdCreate.NewCmdCreate(cmdFactory, nil))
	rootCmd.AddCommand(cmdCopy.NewCmdCopy(cmdFactory, nil))
	rootCmd.AddCommand(cmdClose.NewCmdClose(cmdFactory, nil))
	rootCmd.AddCommand(cmdDelete.NewCmdDelete(cmdFactory, nil))
	rootCmd.AddCommand(cmdEdit.NewCmdEdit(cmdFactory, nil))

	// item subcommand
	rootCmd.AddCommand(itemCmd)
	itemCmd.AddCommand(cmdItemList.NewCmdList(cmdFactory, nil))
	itemCmd.AddCommand(cmdItemCreate.NewCmdCreateItem(cmdFactory, nil))
	itemCmd.AddCommand(cmdItemAdd.NewCmdAddItem(cmdFactory, nil))
	itemCmd.AddCommand(cmdItemEdit.NewCmdEditItem(cmdFactory, nil))
	itemCmd.AddCommand(cmdItemArchive.NewCmdArchiveItem(cmdFactory, nil))
	itemCmd.AddCommand(cmdItemDelete.NewCmdDeleteItem(cmdFactory, nil))

	// field subcommand
	rootCmd.AddCommand(fieldCmd)
	fieldCmd.AddCommand(cmdFieldList.NewCmdList(cmdFactory, nil))
	fieldCmd.AddCommand(cmdFieldCreate.NewCmdCreateField(cmdFactory, nil))
	fieldCmd.AddCommand(cmdFieldEdit.NewCmdEditField(cmdFactory, nil))
	fieldCmd.AddCommand(cmdFieldDelete.NewCmdDeleteField(cmdFactory, nil))

	if err := rootCmd.Execute(); err != nil {
		if strings.HasPrefix(err.Error(), "Message: Your token has not been granted the required scopes to execute this query") {
			fmt.Println("Your token has not been granted the required scopes to execute this query.\nRun 'gh auth refresh -s project' to add the 'project' scope.\nRun 'gh auth status' to see your current token scopes.")
			os.Exit(1)
		}
		fmt.Println(err)
		os.Exit(1)
	}
}
