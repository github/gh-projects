package copy

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

type copyOpts struct {
	includeDraftIssues bool
	number             int
	ownerID            string
	projectID          string
	sourceOrgOwner     string
	sourceUserOwner    string
	targetOrgOwner     string
	targetUserOwner    string
	title              string
	format             string
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

func NewCmdCopy(f *cmdutil.Factory, runF func(config copyConfig) error) *cobra.Command {
	opts := copyOpts{}
	copyCmd := &cobra.Command{
		Short: "Copy a project",
		Use:   "copy [number]",
		Example: `
# copy project 1 owned by user monalisa to org github with title "a new project"
gh projects copy 1 --source-user monalisa --title "a new project" --target-org github

# copy project 1 owned by the org github to current user with title "a new project"
gh projects copy 1 --source-org github --title "a new project" --target-me

# copy project 1 owned by the org github to user monalisa with title "a new project" and include draft issues
gh projects copy 1 --source-org github --title "a new project" --target-user monalisa --drafts

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

			config := copyConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runCopy(config)
		},
	}

	copyCmd.Flags().StringVar(&opts.sourceUserOwner, "source-user", "", "Login of the source user owner. Use \"@me\" for the current user.")
	copyCmd.Flags().StringVar(&opts.sourceOrgOwner, "source-org", "", "Login of the source organization owner.")
	copyCmd.Flags().StringVar(&opts.targetUserOwner, "target-user", "", "Login of the target organization owner. Use \"@me\" for the current user.")
	copyCmd.Flags().StringVar(&opts.targetOrgOwner, "target-org", "", "Login of the target organization owner.")
	copyCmd.Flags().StringVar(&opts.title, "title", "", "Title of the new project copy. Titles do not need to be unique.")
	copyCmd.Flags().BoolVar(&opts.includeDraftIssues, "drafts", false, "Include draft issues in new copy.")
	copyCmd.Flags().StringVar(&opts.format, "format", "", "Output format, must be 'json'.")

	_ = copyCmd.MarkFlagRequired("title")
	copyCmd.MarkFlagsMutuallyExclusive("source-user", "source-org")
	copyCmd.MarkFlagsMutuallyExclusive("target-user", "target-org")

	return copyCmd
}

func runCopy(config copyConfig) error {
	if config.opts.format != "" && config.opts.format != "json" {
		return fmt.Errorf("format must be 'json'")
	}

	sourceOwner, err := queries.NewOwner(config.client, config.opts.sourceUserOwner, config.opts.sourceOrgOwner)
	if err != nil {
		return err
	}

	targetOwner, err := queries.NewOwner(config.client, config.opts.targetUserOwner, config.opts.targetOrgOwner)
	if err != nil {
		return err
	}

	project, err := queries.NewProject(config.client, sourceOwner, config.opts.number, false)
	if err != nil {
		return err
	}

	config.opts.projectID = project.ID
	config.opts.ownerID = targetOwner.ID

	query, variables := copyArgs(config)

	err = config.client.Mutate("CopyProjectV2", query, variables)
	if err != nil {
		return err
	}

	if config.opts.format == "json" {
		return printJSON(config, query.CopyProjectV2.ProjectV2)
	}

	return printResults(config, query.CopyProjectV2.ProjectV2)
}

func copyArgs(config copyConfig) (*copyProjectMutation, map[string]interface{}) {
	return &copyProjectMutation{}, map[string]interface{}{
		"input": githubv4.CopyProjectV2Input{
			OwnerID:            githubv4.ID(config.opts.ownerID),
			ProjectID:          githubv4.ID(config.opts.projectID),
			Title:              githubv4.String(config.opts.title),
			IncludeDraftIssues: githubv4.NewBoolean(githubv4.Boolean(config.opts.includeDraftIssues)),
		},
		"firstItems":  githubv4.Int(0),
		"afterItems":  (*githubv4.String)(nil),
		"firstFields": githubv4.Int(0),
		"afterFields": (*githubv4.String)(nil),
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

func printJSON(config copyConfig, project queries.Project) error {
	b, err := format.JSONProject(project)
	if err != nil {
		return err
	}
	config.tp.AddField(string(b))
	return config.tp.Render()
}
