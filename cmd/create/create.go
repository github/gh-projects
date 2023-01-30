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

type createOpts struct {
	title        string
	userOwner    string
	orgOwner     string
	viewer       bool
	teamID       string
	repositoryID string
}

type createConfig struct {
	tp      tableprinter.TablePrinter
	client  api.GQLClient
	opts    createOpts
	ownerId string
}

type createProjectMutation struct {
	CreateProjectV2 struct {
		ProjectV2 queries.Project `graphql:"projectV2"`
	} `graphql:"createProjectV2(input:$input)"`
}

func NewCmdCreate(f *cmdutil.Factory, runF func(config createConfig) error) *cobra.Command {
	opts := createOpts{}
	createCmd := &cobra.Command{
		Short: "Create a project",
		Use:   "create",
		Example: `
# create a new project owned by user monalisa with title "a new project"
gh projects create --user monalisa --title "a new project"

# create a new project owned by the org github with title "a new project"
gh projects create --org github --title "a new project"

# create a new project owned by the current user with title "a new project"
gh projects create --me --title "a new project"
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
			config := createConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runCreate(config)
		},
	}

	createCmd.Flags().StringVar(&opts.title, "title", "", "Title of the project. Titles do not need to be unique.")
	createCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner.")
	createCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	createCmd.Flags().BoolVar(&opts.viewer, "me", false, "Login of the current user as the project user owner.")
	createCmd.Flags().StringVar(&opts.teamID, "team-id", "", "ID of the team to associate with the project.")
	createCmd.Flags().StringVar(&opts.repositoryID, "repo-id", "", "ID of the repository to associate with the project.")

	createCmd.MarkFlagRequired("title")
	createCmd.MarkFlagsMutuallyExclusive("user", "org", "me")

	return createCmd
}

func runCreate(config createConfig) error {
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

	ownerId, err := queries.OwnerID(config.client, login, ownerType)
	if err != nil {
		return err
	}
	config.ownerId = ownerId
	query, variables := createArgs(config)

	err = config.client.Mutate("CreateProjectV2", query, variables)
	if err != nil {
		return err
	}

	return printResults(config, query.CreateProjectV2.ProjectV2)
}

func createArgs(config createConfig) (*createProjectMutation, map[string]interface{}) {
	return &createProjectMutation{}, map[string]interface{}{
		"input": githubv4.CreateProjectV2Input{
			OwnerID:      githubv4.ID(config.ownerId),
			Title:        githubv4.String(config.opts.title),
			TeamID:       githubv4.NewID(githubv4.ID(config.opts.teamID)),
			RepositoryID: githubv4.NewID(githubv4.ID(config.opts.repositoryID)),
		},
	}
}

func printResults(config createConfig, project queries.Project) error {
	// using table printer here for consistency in case it ends up being needed in the future
	config.tp.AddField(fmt.Sprintf("Created project '%s'", project.Title))
	config.tp.EndRow()
	config.tp.AddField(project.URL)
	config.tp.EndRow()
	return config.tp.Render()
}
