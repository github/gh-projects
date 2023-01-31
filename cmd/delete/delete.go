package delete

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

type deleteOpts struct {
	userOwner string
	orgOwner  string
	viewer    bool
	number    int
}

type deleteConfig struct {
	tp     tableprinter.TablePrinter
	client api.GQLClient
	opts   deleteOpts
}

type deleteProjectMutation struct {
	DeleteProject struct {
		Project queries.Project `graphql:"projectV2"`
	} `graphql:"deleteProjectV2(input:$input)"`
}

// since this api is still in preview, this struct doesn't yet exist in githubv4
type DeleteProjectV2Input struct {
	ProjectID githubv4.ID `json:"projectId"`
}

func NewCmdDelete(f *cmdutil.Factory, runF func(config deleteConfig) error) *cobra.Command {
	opts := deleteOpts{}
	deleteCmd := &cobra.Command{
		Short: "Delete a project",
		Use:   "delete",
		Example: `
# delete the current user's project 1
gh projects delete --me --number 1 --id ID

# delete the monalisa user project 1
gh projects delete --user monalisa --number 1 --id ID

# delete the github org project 1
gh projects delete --org github --number 1 --id ID
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
			config := deleteConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runDelete(config)
		},
	}

	deleteCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner.")
	deleteCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	deleteCmd.Flags().BoolVar(&opts.viewer, "me", false, "Login of the current user as the project user owner.")
	deleteCmd.Flags().IntVarP(&opts.number, "number", "n", 0, "The project number.")

	deleteCmd.MarkFlagsMutuallyExclusive("user", "org", "me")
	deleteCmd.MarkFlagRequired("number")

	return deleteCmd
}

func runDelete(config deleteConfig) error {
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

	query, variables := deleteItemArgs(config, projectID)
	err = config.client.Mutate("DeleteProject", query, variables)
	if err != nil {
		return err
	}

	return printResults(config)

}

func deleteItemArgs(config deleteConfig, projectID string) (*deleteProjectMutation, map[string]interface{}) {
	return &deleteProjectMutation{}, map[string]interface{}{
		"input": DeleteProjectV2Input{
			ProjectID: githubv4.ID(projectID),
		},
	}
}

func printResults(config deleteConfig) error {
	// using table printer here for consistency in case it ends up being needed in the future
	config.tp.AddField("Deleted project")
	config.tp.EndRow()
	return config.tp.Render()
}
