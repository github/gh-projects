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
	draft     bool
	number    int
	itemURL   string
	// assignees []string
}

type createItemConfig struct {
	tp        tableprinter.TablePrinter
	client    api.GQLClient
	opts      createItemOpts
	projectID string
}

func NewCmdCreateItem(f *cmdutil.Factory, runF func(config createItemConfig) error) *cobra.Command {
	opts := createItemOpts{}
	createItemCmd := &cobra.Command{
		Short: "create an item in a project",
		Use:   "create",
		Example: `
# create a draft item in the current user's project
gh projects items create --draft --me --number 1 --title "a new item"

# create a draft item in a user project
gh projects items create --draft --user monalisa --number 1 --title "a new item"

# create a draft item in an org project
gh projects items create --draft --org github --number 1 --title "a new item"

# create an item in the current user's project
gh projects items create  --me --number 1 --url https://github.com/cli/go-gh/issues/1

# create an item in a user project
gh projects items create --user monalisa --number 1 --url https://github.com/cli/go-gh/issues/1

# create an item in an org project
gh projects items create --org github --number 1 --url https://github.com/cli/go-gh/issues/1
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

	createItemCmd.Flags().StringVar(&opts.title, "title", "t", "Title of the draft issue item.")
	createItemCmd.Flags().StringVar(&opts.body, "body", "b", "Body of the draft issue item.")
	createItemCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner.")
	createItemCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	createItemCmd.Flags().StringVar(&opts.itemURL, "url", "", "URL of the issue or pull request to add to the project. Must be of form https://github.com/OWNER/REPO/issues/NUMBER or https://github.com/OWNER/REPO/pull/NUMBER")
	createItemCmd.Flags().BoolVar(&opts.viewer, "me", false, "User the login of the current use as the organization owner.")
	createItemCmd.Flags().BoolVar(&opts.draft, "draft", false, "Create a draft issue.")
	createItemCmd.Flags().IntVarP(&opts.number, "number", "n", 0, "The project number.")
	createItemCmd.MarkFlagsMutuallyExclusive("user", "org", "me")
	createItemCmd.MarkFlagsMutuallyExclusive("draft", "url")

	createItemCmd.MarkFlagRequired("number")
	return createItemCmd
}

func runCreateItem(config createItemConfig) error {
	// TODO interactive survey if no arguments are provided
	if !config.opts.viewer && config.opts.userOwner == "" && config.opts.orgOwner == "" {
		return fmt.Errorf("one of --user, --org or --me is required")
	}

	if config.opts.draft && config.opts.title == "" {
		return fmt.Errorf("--title is required with draft issues")
	}

	if !config.opts.draft && config.opts.itemURL == "" {
		return fmt.Errorf("one of --url or --draft is required")
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

	if config.opts.draft {
		query, variables := buildCreateDraftIssue(config)

		err = config.client.Mutate("CreateDraftItem", query, variables)
		if err != nil {
			return err
		}

		return printResults(config, query.CreateProjectDraftItem.ProjectV2Item)
	}

	itemID, err := queries.GetIssueOrPullRequestID(config.client, config.opts.itemURL)
	if err != nil {
		return err
	}
	query, variables := buildCreateItem(config, itemID)
	err = config.client.Mutate("CreateDraftItem", query, variables)
	if err != nil {
		return err
	}

	return printResults(config, query.CreateProjectItem.ProjectV2Item)

}

func buildCreateItem(config createItemConfig, itemID string) (*queries.CreateProjectItem, map[string]interface{}) {
	return &queries.CreateProjectItem{}, map[string]interface{}{
		"input": githubv4.AddProjectV2ItemByIdInput{
			ProjectID: githubv4.ID(config.projectID),
			ContentID: githubv4.ID(itemID),
		},
	}
}

func buildCreateDraftIssue(config createItemConfig) (*queries.CreateProjectDraftItem, map[string]interface{}) {
	return &queries.CreateProjectDraftItem{}, map[string]interface{}{
		"input": githubv4.AddProjectV2DraftIssueInput{
			Body:      githubv4.NewString(githubv4.String(config.opts.body)),
			ProjectID: githubv4.ID(config.projectID),
			Title:     githubv4.String(config.opts.title),
			// optionally include assignees
		},
	}
}

func printResults(config createItemConfig, item queries.ProjectV2Item) error {
	// using table printer here for consistency in case it ends up being needed in the future
	config.tp.AddField("Created item")
	config.tp.EndRow()
	return config.tp.Render()
}
