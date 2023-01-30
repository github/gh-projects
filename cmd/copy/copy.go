package copy

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

type copyOpts struct {
	title              string
	sourceUserOwner    string
	targetUserOwner    string
	sourceOrgOwner     string
	targetOrgOwner     string
	sourceViewer       bool
	targetViewer       bool
	includeDraftIssues bool
	number             int
	ownerID            string
	projectID          string
}

type copyConfig struct {
	tp     tableprinter.TablePrinter
	client api.GQLClient
	opts   copyOpts
}

type copyProjectMutation struct {
	CopyProjectV2 struct {
		ProjectV2 queries.Project `graphql:"projectV2"`
	} `graphql:"copyProjectV2(input:$input)"`
}

// since this api is still in preview, this struct doesn't yet exist in githubv4
type CopyProjectV2Input struct {
	OwnerID            githubv4.ID      `json:"ownerId"`
	ProjectID          githubv4.ID      `json:"projectId"`
	Title              githubv4.String  `json:"title"`
	IncludeDraftIssues githubv4.Boolean `json:"includeDraftIssues"`
}

func NewCmdCopy(f *cmdutil.Factory, runF func(config copyConfig) error) *cobra.Command {
	opts := copyOpts{}
	copyCmd := &cobra.Command{
		Short: "copy a project",
		Use:   "copy",
		Example: `
# copy a project in interative mode
gh projects copy

# copy a project owned by user monalisa
gh projects copy --user monalisa --number 1 --title "a new project"

# copy a project owned by the org github
gh projects copy --org github --number 1 --title "a new project"

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
			config := copyConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runCopy(config)
		},
	}

	copyCmd.Flags().StringVar(&opts.title, "title", "", "Title of the project copy. Titles do not need to be unique.")
	copyCmd.Flags().StringVar(&opts.sourceUserOwner, "source-user", "", "Login of the source user owner.")
	copyCmd.Flags().StringVar(&opts.sourceOrgOwner, "source-org", "", "Login of the source organization owner.")
	copyCmd.Flags().BoolVar(&opts.sourceViewer, "source-me", false, "Login of the current user as the source project owner.")
	copyCmd.Flags().StringVar(&opts.targetOrgOwner, "target-org", "", "Login of the target organization owner.")
	copyCmd.Flags().StringVar(&opts.targetUserOwner, "target-user", "", "Login of the target organization owner.")
	copyCmd.Flags().BoolVar(&opts.targetViewer, "target-me", false, "Login of the current user as the target project owner.")
	copyCmd.Flags().BoolVar(&opts.includeDraftIssues, "drafts", false, "Include draft issues in new copy.")
	copyCmd.Flags().IntVarP(&opts.number, "number", "n", 0, "Number of the source project.")
	copyCmd.MarkFlagsMutuallyExclusive("source-user", "source-org", "source-me")
	copyCmd.MarkFlagsMutuallyExclusive("target-user", "target-org", "target-me")
	return copyCmd
}

func runCopy(config copyConfig) error {
	// TODO interactive survey if no arguments are provided
	if !config.opts.sourceViewer && config.opts.sourceUserOwner == "" && config.opts.sourceOrgOwner == "" {
		return fmt.Errorf("one of --source-user, --source-org or --source-me is required")
	}

	if !config.opts.targetViewer && config.opts.targetUserOwner == "" && config.opts.targetOrgOwner == "" {
		return fmt.Errorf("one of --target-user, --target-org or --target-me is required")
	}

	// source project
	var sourceLogin string
	var sourceOwnerType queries.OwnerType
	if config.opts.sourceUserOwner != "" {
		sourceLogin = config.opts.sourceUserOwner
		sourceOwnerType = queries.UserOwner
	} else if config.opts.sourceOrgOwner != "" {
		sourceLogin = config.opts.sourceOrgOwner
		sourceOwnerType = queries.OrgOwner
	} else {
		sourceOwnerType = queries.ViewerOwner
	}

	projectID, err := queries.ProjectId(config.client, sourceLogin, sourceOwnerType, config.opts.number)
	if err != nil {
		return err
	}
	config.opts.projectID = projectID

	// target owner
	var targetLogin string
	var targetOwnerType queries.OwnerType
	if config.opts.targetUserOwner != "" {
		targetLogin = config.opts.targetUserOwner
		targetOwnerType = queries.UserOwner
	} else if config.opts.targetOrgOwner != "" {
		targetLogin = config.opts.targetOrgOwner
		targetOwnerType = queries.OrgOwner
	} else {
		targetOwnerType = queries.ViewerOwner
	}

	ownerId, err := queries.OwnerID(config.client, targetLogin, targetOwnerType)
	if err != nil {
		return err
	}
	config.opts.ownerID = ownerId

	query, variables := buildCopyQuery(config)

	err = config.client.Mutate("CopyProjectV2", query, variables)
	if err != nil {
		return err
	}

	return printResults(config, query.CopyProjectV2.ProjectV2)
}

func buildCopyQuery(config copyConfig) (*copyProjectMutation, map[string]interface{}) {
	return &copyProjectMutation{}, map[string]interface{}{
		"input": CopyProjectV2Input{
			OwnerID:            githubv4.ID(config.opts.ownerID),
			ProjectID:          githubv4.ID(config.opts.projectID),
			Title:              githubv4.String(config.opts.title),
			IncludeDraftIssues: githubv4.Boolean(config.opts.includeDraftIssues),
		},
	}
}

func printResults(config copyConfig, project queries.Project) error {
	// using table printer here for consistency in case it ends up being needed in the future
	config.tp.AddField(fmt.Sprintf("Created project copy '%s'", project.Title))
	config.tp.EndRow()
	config.tp.AddField(project.URL)
	config.tp.EndRow()
	return config.tp.Render()
}
