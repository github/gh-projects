package close

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

type closeOpts struct {
	number    int
	userOwner string
	orgOwner  string
	viewer    bool
	reopen    bool
}

type closeConfig struct {
	tp        tableprinter.TablePrinter
	client    api.GQLClient
	opts      closeOpts
	projectId string
}

func NewCmdClose(f *cmdutil.Factory, runF func(config closeConfig) error) *cobra.Command {
	opts := closeOpts{}
	closeCmd := &cobra.Command{
		Short: "close a project",
		Use:   "close",
		Example: `
# close a project in interative mode
gh projects close

# close a project owned by user monalisa
gh projects close --user monalisa --number 1

# close a project owned by org github
gh projects close --org github --number 1

# reopen a closed project owned by org github
gh projects close --org github --number 1 --reopen

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
			config := closeConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runClose(config)
		},
	}

	closeCmd.Flags().IntVarP(&opts.number, "number", "n", 0, "Number of the project.")
	closeCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner.")
	closeCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	closeCmd.Flags().BoolVar(&opts.viewer, "me", false, "Use the login of the current use as the organization owner.")
	closeCmd.Flags().BoolVar(&opts.reopen, "reopen", false, "Reopen a closed project.")
	closeCmd.MarkFlagsMutuallyExclusive("user", "org", "me")
	return closeCmd
}

func runClose(config closeConfig) error {
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
		// login intentionally empty here
	}

	projectId, err := queries.GetProjectId(config.client, login, ownerType, config.opts.number)
	if err != nil {
		return err
	}
	config.projectId = projectId
	query, variables := buildCloseQuery(config)

	err = config.client.Mutate("CloseProjectV2", query, variables)
	if err != nil {
		return err
	}

	return printResults(config, query.UpdateProjectV2.ProjectV2)
}

func buildCloseQuery(config closeConfig) (*queries.UpdateProjectMutation, map[string]interface{}) {
	closed := !config.opts.reopen
	return &queries.UpdateProjectMutation{}, map[string]interface{}{
		"input": githubv4.UpdateProjectV2Input{
			ProjectID: githubv4.ID(config.projectId),
			Closed:    githubv4.NewBoolean(githubv4.Boolean(closed)),
		},
	}
}

func printResults(config closeConfig, project queries.ProjectV2) error {
	// using table printer here for consistency in case it ends up being needed in the future
	var action string
	if config.opts.reopen {
		action = "Reopened"
	} else {
		action = "Closed"
	}
	config.tp.AddField(fmt.Sprintf("%s project %s", action, project.Url))
	config.tp.EndRow()
	return config.tp.Render()
}
