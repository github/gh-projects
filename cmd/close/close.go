package close

import (
	"fmt"
	"strconv"

	"github.com/cli/cli/v2/pkg/cmdutil"

	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/github/gh-projects/format"
	"github.com/github/gh-projects/queries"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
)

type closeOpts struct {
	number    int
	userOwner string
	orgOwner  string
	reopen    bool
	projectID string
	format    string
}

type closeConfig struct {
	tp     tableprinter.TablePrinter
	client api.GQLClient
	opts   closeOpts
}

// the close command relies on the updateProjectV2 mutation
type updateProjectMutation struct {
	UpdateProjectV2 struct {
		ProjectV2 queries.Project `graphql:"projectV2"`
	} `graphql:"updateProjectV2(input:$input)"`
}

func NewCmdClose(f *cmdutil.Factory, runF func(config closeConfig) error) *cobra.Command {
	opts := closeOpts{}
	closeCmd := &cobra.Command{
		Short: "Close a project",
		Use:   "close [number]",
		Example: `
# close project 1 owned by user monalisa
gh projects close 1 --user monalisa

# close project 1 owned by org github
gh projects close 1 --org github

# reopen closed project 1 owned by org github
gh projects close 1 --org github --undo

# add --format=json to output in JSON format
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

			config := closeConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runClose(config)
		},
	}

	closeCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner. Use \"@me\" for the current user.")
	closeCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	closeCmd.Flags().BoolVar(&opts.reopen, "undo", false, "Reopen a closed project.")
	closeCmd.Flags().StringVar(&opts.format, "format", "", "Output format, must be 'json'.")
	closeCmd.MarkFlagsMutuallyExclusive("user", "org")

	return closeCmd
}

func runClose(config closeConfig) error {
	if config.opts.format != "" && config.opts.format != "json" {
		return fmt.Errorf("format must be 'json'")
	}

	owner, err := queries.NewOwner(config.client, config.opts.userOwner, config.opts.orgOwner)
	if err != nil {
		return err
	}

	project, err := queries.NewProject(config.client, owner, config.opts.number, false)
	if err != nil {
		return err
	}
	config.opts.projectID = project.ID

	query, variables := closeArgs(config)

	err = config.client.Mutate("CloseProjectV2", query, variables)
	if err != nil {
		return err
	}

	if config.opts.format == "json" {
		return printJSON(config, *project)
	}

	return printResults(config, query.UpdateProjectV2.ProjectV2)
}

func closeArgs(config closeConfig) (*updateProjectMutation, map[string]interface{}) {
	closed := !config.opts.reopen
	return &updateProjectMutation{}, map[string]interface{}{
		"input": githubv4.UpdateProjectV2Input{
			ProjectID: githubv4.ID(config.opts.projectID),
			Closed:    githubv4.NewBoolean(githubv4.Boolean(closed)),
		},
		"firstItems":  githubv4.Int(queries.LimitMax),
		"afterItems":  (*githubv4.String)(nil),
		"firstFields": githubv4.Int(queries.LimitMax),
		"afterFields": (*githubv4.String)(nil),
	}
}

func printResults(config closeConfig, project queries.Project) error {
	// using table printer here for consistency in case it ends up being needed in the future
	var action string
	if config.opts.reopen {
		action = "Reopened"
	} else {
		action = "Closed"
	}
	config.tp.AddField(fmt.Sprintf("%s project %s", action, project.URL))
	config.tp.EndRow()
	return config.tp.Render()
}

func printJSON(config closeConfig, project queries.Project) error {
	b, err := format.JSONProject(project)
	if err != nil {
		return err
	}
	config.tp.AddField(string(b))
	return config.tp.Render()
}
