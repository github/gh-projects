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
	title     string
	userOwner string
	orgOwner  string
	ownerID   string
}

type createConfig struct {
	tp     tableprinter.TablePrinter
	client api.GQLClient
	opts   createOpts
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
gh projects create --user '@me' --title "a new project"
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
	createCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner. Use \"@me\" for the current user.")
	createCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")

	createCmd.MarkFlagRequired("title")
	createCmd.MarkFlagsMutuallyExclusive("user", "org")

	return createCmd
}

func runCreate(config createConfig) error {
	owner, err := queries.NewOwner(config.client, config.opts.userOwner, config.opts.orgOwner)
	if err != nil {
		return err
	}

	config.opts.ownerID = owner.ID
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
			OwnerID: githubv4.ID(config.opts.ownerID),
			Title:   githubv4.String(config.opts.title),
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
