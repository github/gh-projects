package itemadd

import (
	"strconv"

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
	number    int
	itemURL   string
	projectID string
	itemID    string
}

type addItemConfig struct {
	tp     tableprinter.TablePrinter
	client api.GQLClient
	opts   addItemOpts
}

type addProjectItemMutation struct {
	CreateProjectItem struct {
		ProjectV2Item queries.ProjectItem `graphql:"item"`
	} `graphql:"addProjectV2ItemById(input:$input)"`
}

func NewCmdAddItem(f *cmdutil.Factory, runF func(config addItemConfig) error) *cobra.Command {
	opts := addItemOpts{}
	addItemCmd := &cobra.Command{
		Short: "Add a pull request or an issue to a project",
		Use:   "item-add [number]",
		Example: `
# add an item to the current user's project 1
gh projects item-add 1 --user "@me" --url https://github.com/cli/go-gh/issues/1

# add an item to user monalisa's project 1
gh projects item-add 1 --user monalisa --url https://github.com/cli/go-gh/issues/1

# add an item to the org github's project 1
gh projects item-add 1 --org github --url https://github.com/cli/go-gh/issues/1
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := queries.NewClient()
			if err != nil {
				return err
			}

			if len(args) == 1 {
				opts.number, err = strconv.Atoi(args[0])
				if err != nil {
					return err
				}
			}

			terminal := term.FromEnv()
			termWidth, _, err := terminal.Size()
			if err != nil {
				// set a static width in case of error
				termWidth = 80
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

	addItemCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner. Use \"@me\" for the current user.")
	addItemCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	addItemCmd.Flags().StringVar(&opts.itemURL, "url", "", "URL of the issue or pull request to add to the project. Must be of form https://github.com/OWNER/REPO/issues/NUMBER or https://github.com/OWNER/REPO/pull/NUMBER")
	addItemCmd.MarkFlagsMutuallyExclusive("user", "org")

	_ = addItemCmd.MarkFlagRequired("url")

	return addItemCmd
}

func runAddItem(config addItemConfig) error {
	owner, err := queries.NewOwner(config.client, config.opts.userOwner, config.opts.orgOwner)
	if err != nil {
		return err
	}

	project, err := queries.NewProject(config.client, owner, config.opts.number)
	if err != nil {
		return err
	}
	config.opts.projectID = project.ID

	itemID, err := queries.IssueOrPullRequestID(config.client, config.opts.itemURL)
	if err != nil {
		return err
	}

	config.opts.itemID = itemID

	query, variables := addItemArgs(config)
	err = config.client.Mutate("AddItem", query, variables)
	if err != nil {
		return err
	}

	return printResults(config, query.CreateProjectItem.ProjectV2Item)

}

func addItemArgs(config addItemConfig) (*addProjectItemMutation, map[string]interface{}) {
	return &addProjectItemMutation{}, map[string]interface{}{
		"input": githubv4.AddProjectV2ItemByIdInput{
			ProjectID: githubv4.ID(config.opts.projectID),
			ContentID: githubv4.ID(config.opts.itemID),
		},
	}
}

func printResults(config addItemConfig, item queries.ProjectItem) error {
	// using table printer here for consistency in case it ends up being needed in the future
	config.tp.AddField("Added item")
	config.tp.EndRow()
	return config.tp.Render()
}
