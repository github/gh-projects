package delete

import (
	"fmt"

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
	viewer    bool
	number    int
	itemID    string
}

type deleteItemConfig struct {
	tp        tableprinter.TablePrinter
	client    api.GQLClient
	opts      deleteItemOpts
	projectID string
}

type deleteProjectItemMutation struct {
	DeleteProjectItem struct {
		DeletedItemId githubv4.ID `graphql:"deletedItemId"`
	} `graphql:"deleteProjectV2Item(input:$input)"`
}

func NewCmdDeleteItem(f *cmdutil.Factory, runF func(config deleteItemConfig) error) *cobra.Command {
	opts := deleteItemOpts{}
	deleteItemCmd := &cobra.Command{
		Short: "delete an item from a project",
		Use:   "delete",
		Example: `
# delete an item in the current user's project
gh projects item delete --me --number 1 --id ID

# delete an item in a user project
gh projects item delete --user monalisa --number 1 --id ID

# delete an item in an org project
gh projects item delete --org github --number 1 --id ID
`,
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

			t := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), termWidth)
			config := deleteItemConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runDeleteItem(config)
		},
	}

	deleteItemCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner.")
	deleteItemCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	deleteItemCmd.Flags().StringVar(&opts.itemID, "id", "", "Global ID of the item to delete from the project.")
	deleteItemCmd.Flags().BoolVar(&opts.viewer, "me", false, "Login of the current user as the project owner.")
	deleteItemCmd.Flags().IntVarP(&opts.number, "number", "n", 0, "The project number.")

	deleteItemCmd.MarkFlagsMutuallyExclusive("user", "org", "me")
	deleteItemCmd.MarkFlagRequired("number")
	deleteItemCmd.MarkFlagRequired("id")
	return deleteItemCmd
}

func runDeleteItem(config deleteItemConfig) error {
	// TODO interactive survey if no arguments are provided
	if !config.opts.viewer && config.opts.userOwner == "" && config.opts.orgOwner == "" {
		return fmt.Errorf("one of --user, --org or --me is required")
	}

	var login string
	var ownerType queries.OwnerType
	if config.opts.userOwner != "" {
		login = config.opts.userOwner
		ownerType = queries.UserOwner
	} else if config.opts.orgOwner != "" {
		login = config.opts.orgOwner
		ownerType = queries.OrgOwner
	} else {
		ownerType = queries.ViewerOwner
	}

	projectID, err := queries.ProjectId(config.client, login, ownerType, config.opts.number)
	if err != nil {
		return err
	}
	config.projectID = projectID

	query, variables := buildDeleteItem(config, config.opts.itemID)
	err = config.client.Mutate("DeleteProjectItem", query, variables)
	if err != nil {
		return err
	}

	return printResults(config)

}

func buildDeleteItem(config deleteItemConfig, itemID string) (*deleteProjectItemMutation, map[string]interface{}) {
	return &deleteProjectItemMutation{}, map[string]interface{}{
		"input": githubv4.DeleteProjectV2ItemInput{
			ProjectID: githubv4.ID(config.projectID),
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
