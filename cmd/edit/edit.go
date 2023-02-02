package edit

import (
	"fmt"
	"strconv"

	"github.com/cli/cli/v2/pkg/cmdutil"

	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/github/gh-projects/queries"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
)

type editOpts struct {
	number           int
	userOwner        string
	orgOwner         string
	title            string
	readme           string
	visibility       string
	shortDescription string
	projectID        string
}

type editConfig struct {
	tp     tableprinter.TablePrinter
	client api.GQLClient
	opts   editOpts
}

type updateProjectMutation struct {
	UpdateProjectV2 struct {
		ProjectV2 queries.Project `graphql:"projectV2"`
	} `graphql:"updateProjectV2(input:$input)"`
}

const projectVisibilityPublic = "PUBLIC"
const projectVisibilityPrivate = "PRIVATE"

func NewCmdEdit(f *cmdutil.Factory, runF func(config editConfig) error) *cobra.Command {
	opts := editOpts{}
	editCmd := &cobra.Command{
		Short: "Edit a project",
		Use:   "edit [number]",
		Example: `
# edit project 1 owned by user monalisa to have the new title "New title"
gh projects edit 1 --user monalisa --title "New title"

# edit project 1 owned by org github to have the new title "New title"
gh projects edit 1 --org github --title "New title"

# edit project 1 owned by org github to have visibility public
gh projects edit 1 --org github --visibility PUBLIC
`,
		Args: cobra.MaximumNArgs(1),
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

			if len(args) == 1 {
				opts.number, err = strconv.Atoi(args[0])
				if err != nil {
					return err
				}
			}

			t := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), termWidth)
			config := editConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runEdit(config)
		},
	}

	editCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner. Use \"@me\" for the current user.")
	editCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	editCmd.Flags().StringVar(&opts.visibility, "visibility", "", "Update the visibility of the project public. Must be one of PUBLIC or PRIVATE.")
	editCmd.Flags().StringVar(&opts.title, "title", "", "The edited title of the project.")
	editCmd.Flags().StringVar(&opts.readme, "readme", "", "The edited readme of the project.")
	editCmd.Flags().StringVarP(&opts.shortDescription, "description", "d", "", "The edited short description of the project.")

	editCmd.MarkFlagsMutuallyExclusive("user", "org")

	return editCmd
}

func runEdit(config editConfig) error {
	if config.opts.visibility != "" && config.opts.visibility != projectVisibilityPublic && config.opts.visibility != projectVisibilityPrivate {
		return fmt.Errorf("visibility must match either PUBLIC or PRIVATE")
	}

	if config.opts.title == "" && config.opts.shortDescription == "" && config.opts.readme == "" && config.opts.visibility == "" {
		return fmt.Errorf("no fields to edit")
	}

	owner, err := queries.NewOwner(config.client, config.opts.userOwner, config.opts.orgOwner)
	if err != nil {
		return err
	}

	project, err := queries.NewProject(config.client, owner, config.opts.number)
	if err != nil {
		return err
	}
	config.opts.projectID = project.ID

	query, variables := editArgs(config)
	err = config.client.Mutate("UpdateProjectV2", query, variables)
	if err != nil {
		return err
	}

	return printResults(config, query.UpdateProjectV2.ProjectV2)
}

func editArgs(config editConfig) (*updateProjectMutation, map[string]interface{}) {
	variables := githubv4.UpdateProjectV2Input{ProjectID: githubv4.ID(config.opts.projectID)}
	if config.opts.title != "" {
		variables.Title = githubv4.NewString(githubv4.String(config.opts.title))
	}
	if config.opts.shortDescription != "" {
		variables.ShortDescription = githubv4.NewString(githubv4.String(config.opts.shortDescription))
	}
	if config.opts.readme != "" {
		variables.Readme = githubv4.NewString(githubv4.String(config.opts.readme))
	}
	if config.opts.visibility != "" {
		if config.opts.visibility == projectVisibilityPublic {
			variables.Public = githubv4.NewBoolean(githubv4.Boolean(true))
		} else if config.opts.visibility == projectVisibilityPrivate {
			variables.Public = githubv4.NewBoolean(githubv4.Boolean(false))
		}
	}

	return &updateProjectMutation{}, map[string]interface{}{
		"input": variables,
	}
}

func printResults(config editConfig, project queries.Project) error {
	// using table printer here for consistency in case it ends up being needed in the future
	config.tp.AddField(fmt.Sprintf("Updated project %s", project.URL))
	config.tp.EndRow()
	return config.tp.Render()
}
