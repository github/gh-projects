package update

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

type updateOpts struct {
	number    int
	userOwner string
	orgOwner  string
	viewer    bool
	title     string
	readme    string
	// using a pointer to a boolean here to know if the user set the flag for public or not
	// otherwise we would make public projects private if the user didn't set the flag
	public           *bool
	shortDescription string
}

type updateConfig struct {
	tp        tableprinter.TablePrinter
	client    api.GQLClient
	opts      updateOpts
	projectId string
}

func NewCmdUpdate(f *cmdutil.Factory, runF func(config updateConfig) error) *cobra.Command {
	opts := updateOpts{}
	updateCmd := &cobra.Command{
		Short: "update a project",
		Use:   "update",
		Example: `
# update a project in interative mode
gh projects update

# update a project owned by user monalisa
gh projects update --user monalisa --number 1 --title "New title"

# update a project owned by org github
gh projects update --org github --number 1 --title "New title"
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
			config := updateConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runUpdate(config)
		},
	}

	updateCmd.Flags().IntVarP(&opts.number, "number", "n", 0, "Number of the project.")
	updateCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner.")
	updateCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	updateCmd.Flags().BoolVar(&opts.viewer, "me", false, "Use the login of the current use as the organization owner.")
	updateCmd.Flags().BoolVar(opts.public, "public", false, "Change the visibility to public.")
	updateCmd.Flags().StringVar(&opts.title, "title", "", "The updated title of the project.")
	updateCmd.Flags().StringVar(&opts.readme, "readme", "", "The updated readme of the project.")
	updateCmd.Flags().StringVarP(&opts.shortDescription, "description", "d", "", "The updated short description of the project.")
	updateCmd.MarkFlagsMutuallyExclusive("user", "org", "me")
	return updateCmd
}

func runUpdate(config updateConfig) error {
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
	query, variables := buildUpdateQuery(config)

	err = config.client.Mutate("UpdateProjectV2", query, variables)
	if err != nil {
		return err
	}

	return printResults(config, query.UpdateProjectV2.ProjectV2)
}

func buildUpdateQuery(config updateConfig) (*queries.UpdateProjectMutation, map[string]interface{}) {
	variables := githubv4.UpdateProjectV2Input{ProjectID: githubv4.ID(config.projectId)}
	if config.opts.title != "" {
		variables.Title = githubv4.NewString(githubv4.String(config.opts.title))
	}
	if config.opts.shortDescription != "" {
		variables.ShortDescription = githubv4.NewString(githubv4.String(config.opts.shortDescription))
	}
	if config.opts.readme != "" {
		variables.Readme = githubv4.NewString(githubv4.String(config.opts.readme))
	}
	if config.opts.public != nil {
		variables.Public = githubv4.NewBoolean(githubv4.Boolean(*config.opts.public))
	}

	return &queries.UpdateProjectMutation{}, map[string]interface{}{
		"input": variables,
	}
}

func printResults(config updateConfig, project queries.ProjectV2) error {
	// using table printer here for consistency in case it ends up being needed in the future
	config.tp.AddField(fmt.Sprintf("Updated project %s", project.Url))
	config.tp.EndRow()
	return config.tp.Render()
}
