package add

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

type addItemOpts struct {
	userOwner string
	orgOwner  string
	viewer    bool
	number    int
	itemURL   string
	// assignees []string
}

type addItemConfig struct {
	tp        tableprinter.TablePrinter
	client    api.GQLClient
	opts      addItemOpts
	projectID string
}

func NewCmdAddItem(f *cmdutil.Factory, runF func(config addItemConfig) error) *cobra.Command {
	opts := addItemOpts{}
	addItemCmd := &cobra.Command{
		Short: "add a pull request or an issue to a project",
		Use:   "add",
		Example: `
# add an item to the current user's project
gh projects item add --me --number 1 --url https://github.com/cli/go-gh/issues/1

# add an item to a user project
gh projects item add --user monalisa --number 1 --url https://github.com/cli/go-gh/issues/1

# add an item to an org project
gh projects item add --org github --number 1 --url https://github.com/cli/go-gh/issues/1
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
			config := addItemConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runAddItem(config)
		},
	}

	addItemCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner.")
	addItemCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	addItemCmd.Flags().StringVar(&opts.itemURL, "url", "", "URL of the issue or pull request to add to the project. Must be of form https://github.com/OWNER/REPO/issues/NUMBER or https://github.com/OWNER/REPO/pull/NUMBER")
	addItemCmd.Flags().BoolVar(&opts.viewer, "me", false, "Login of the current user as the project owner.")
	addItemCmd.Flags().IntVarP(&opts.number, "number", "n", 0, "The project number.")
	addItemCmd.MarkFlagsMutuallyExclusive("user", "org", "me")

	addItemCmd.MarkFlagRequired("number")
	addItemCmd.MarkFlagRequired("url")
	return addItemCmd
}

func runAddItem(config addItemConfig) error {
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

	projectID, err := queries.GetProjectId(config.client, login, ownerType, config.opts.number)
	if err != nil {
		return err
	}
	config.projectID = projectID

	itemID, err := queries.GetIssueOrPullRequestID(config.client, config.opts.itemURL)
	if err != nil {
		return err
	}
	query, variables := buildAddItem(config, itemID)
	err = config.client.Mutate("AddItem", query, variables)
	if err != nil {
		return err
	}

	return printResults(config, query.CreateProjectItem.ProjectV2Item)

}

func buildAddItem(config addItemConfig, itemID string) (*queries.AddProjectItem, map[string]interface{}) {
	return &queries.AddProjectItem{}, map[string]interface{}{
		"input": githubv4.AddProjectV2ItemByIdInput{
			ProjectID: githubv4.ID(config.projectID),
			ContentID: githubv4.ID(itemID),
		},
	}
}

func printResults(config addItemConfig, item queries.ProjectV2Item) error {
	// using table printer here for consistency in case it ends up being needed in the future
	config.tp.AddField("Added item")
	config.tp.EndRow()
	return config.tp.Render()
}
