package delete

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

type deleteOpts struct {
	userOwner string
	orgOwner  string
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
		Use:   "delete number",
		Example: `
# delete the current user's project 1
gh projects delete 1 --user "@me" --id ID

# delete the monalisa user project 1
gh projects delete 1 --user monalisa --id ID

# delete the github org project 1
gh projects delete 1 --org github --id ID
`,
		Args: cobra.ExactArgs(1),
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

			opts.number, err = strconv.Atoi(args[0])
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

	deleteCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner. Use \"@me\" for the current user.")
	deleteCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")

	deleteCmd.MarkFlagsMutuallyExclusive("user", "org")

	return deleteCmd
}

func runDelete(config deleteConfig) error {
	if config.opts.userOwner == "" && config.opts.orgOwner == "" {
		return fmt.Errorf("one of --user or --org is required")
	}

	var login string
	var ownerType queries.OwnerType
	if config.opts.userOwner == "@me" {
		login = "me"
		ownerType = queries.ViewerOwner
	} else if config.opts.userOwner != "" {
		login = config.opts.userOwner
		ownerType = queries.UserOwner
	} else if config.opts.orgOwner != "" {
		login = config.opts.orgOwner
		ownerType = queries.OrgOwner
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
