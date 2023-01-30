package create

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

type createItemOpts struct {
	title     string
	body      string
	userOwner string
	orgOwner  string
	viewer    bool
	number    int
}

type createItemConfig struct {
	tp        tableprinter.TablePrinter
	client    api.GQLClient
	opts      createItemOpts
	projectID string
}

type createProjectDraftItemMutation struct {
	CreateProjectDraftItem struct {
		ProjectV2Item queries.ProjectItem `graphql:"projectItem"`
	} `graphql:"addProjectV2DraftIssue(input:$input)"`
}

func NewCmdCreateItem(f *cmdutil.Factory, runF func(config createItemConfig) error) *cobra.Command {
	opts := createItemOpts{}
	createItemCmd := &cobra.Command{
		Short: "Create a draft issue in a project",
		Use:   "create",
		Example: `
# create a draft issue in the current user's project 1 with title "new item" and body "new item body"
gh projects item create --me --number 1 --title "new item" --body "new item body"

# create a draft issue in monalisa user project 1 with title "new item" and body "new item body"
gh projects item create --user monalisa --number 1 --title "new item" --body "new item body"

# create a draft issue in github org project 1 with title "new item" and body "new item body"
gh projects item create --org github --number 1 --title "new item" --body "new item body"
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
			config := createItemConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runCreateItem(config)
		},
	}

	createItemCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner.")
	createItemCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	createItemCmd.Flags().BoolVar(&opts.viewer, "me", false, "Login of the current user as the project user owner.")
	createItemCmd.Flags().IntVarP(&opts.number, "number", "n", 0, "The project number.")
	createItemCmd.Flags().StringVar(&opts.title, "title", "", "Title of the draft issue item.")
	createItemCmd.Flags().StringVar(&opts.body, "body", "", "Body of the draft issue item.")

	createItemCmd.MarkFlagsMutuallyExclusive("user", "org", "me")
	createItemCmd.MarkFlagRequired("number")
	createItemCmd.MarkFlagRequired("title")

	return createItemCmd
}

func runCreateItem(config createItemConfig) error {
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

	query, variables := createDraftIssueArgs(config)

	err = config.client.Mutate("CreateDraftItem", query, variables)
	if err != nil {
		return err
	}

	return printResults(config, query.CreateProjectDraftItem.ProjectV2Item)
}

func createDraftIssueArgs(config createItemConfig) (*createProjectDraftItemMutation, map[string]interface{}) {
	return &createProjectDraftItemMutation{}, map[string]interface{}{
		"input": githubv4.AddProjectV2DraftIssueInput{
			Body:      githubv4.NewString(githubv4.String(config.opts.body)),
			ProjectID: githubv4.ID(config.projectID),
			Title:     githubv4.String(config.opts.title),
		},
	}
}

func printResults(config createItemConfig, item queries.ProjectItem) error {
	// using table printer here for consistency in case it ends up being needed in the future
	config.tp.AddField("Created item")
	config.tp.EndRow()
	return config.tp.Render()
}
