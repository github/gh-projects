package delete

import (
	"fmt"
	"strconv"

	"github.com/cli/cli/v2/pkg/cmdutil"

	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/github/gh-projects/queries"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
)

type deleteItemOpts struct {
	userOwner string
	orgOwner  string
	number    int
	itemID    string
	projectID string
}

type deleteItemConfig struct {
	tp     tableprinter.TablePrinter
	client api.GQLClient
	opts   deleteItemOpts
}

type deleteProjectItemMutation struct {
	DeleteProjectItem struct {
		DeletedItemId githubv4.ID `graphql:"deletedItemId"`
	} `graphql:"deleteProjectV2Item(input:$input)"`
}

func NewCmdDeleteItem(f *cmdutil.Factory, runF func(config deleteItemConfig) error) *cobra.Command {
	opts := deleteItemOpts{}
	deleteItemCmd := &cobra.Command{
		Short: "Delete an item from a project",
		Use:   "delete number",
		Example: `
# delete an item in the current user's project 1
gh projects item delete 1 --user "@me" --id ID

# delete an item in the monalisa user project 1
gh projects item delete 1 --user monalisa --id ID

# delete an item in the github org project 1
gh projects item delete 1 --org github --id ID
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := queries.NewClient()
			if err != nil {
				return err
			}

			terminal := term.FromEnv()
			termWidth, _, err := terminal.Size()
			if err != nil {
				return err
			}

			opts.number, err = strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			t := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), termWidth)
			config := deleteItemConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runDeleteItem(config)
		},
	}

	deleteItemCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner. Use \"@me\" for the current user.")
	deleteItemCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	deleteItemCmd.Flags().StringVar(&opts.itemID, "id", "", "Global ID of the item to delete from the project.")

	deleteItemCmd.MarkFlagsMutuallyExclusive("user", "org")
	deleteItemCmd.MarkFlagRequired("id")

	return deleteItemCmd
}

func runDeleteItem(config deleteItemConfig) error {
	if config.opts.userOwner == "" && config.opts.orgOwner == "" {
		return fmt.Errorf("one of --user or --org is required")
	}

	var login string
	var ownerType queries.OwnerType
	if config.opts.userOwner == "@me" {
		login = "me"
		ownerType = queries.ViewerOwner
	} else if config.opts.userOwner != "" {
		login = config.opts.userOwner
		ownerType = queries.UserOwner
	} else if config.opts.orgOwner != "" {
		login = config.opts.orgOwner
		ownerType = queries.OrgOwner
	}
	projectID, err := queries.ProjectID(config.client, login, ownerType, config.opts.number)
	if err != nil {
		return err
	}
	config.opts.projectID = projectID

	query, variables := deleteItemArgs(config, config.opts.itemID)
	err = config.client.Mutate("DeleteProjectItem", query, variables)
	if err != nil {
		return err
	}

	return printResults(config)

}

func deleteItemArgs(config deleteItemConfig, itemID string) (*deleteProjectItemMutation, map[string]interface{}) {
	return &deleteProjectItemMutation{}, map[string]interface{}{
		"input": githubv4.DeleteProjectV2ItemInput{
			ProjectID: githubv4.ID(config.opts.projectID),
			ItemID:    githubv4.ID(itemID),
		},
	}
}

func printResults(config deleteItemConfig) error {
	// using table printer here for consistency in case it ends up being needed in the future
	config.tp.AddField("Deleted item")
	config.tp.EndRow()
	return config.tp.Render()
}
