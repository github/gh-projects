package itemedit

import (
	"errors"
	"strings"

	"github.com/cli/cli/v2/pkg/cmdutil"

	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/github/gh-projects/queries"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
)

type editItemOpts struct {
	title  string
	body   string
	itemID string
	// assignees []string
}

type editItemConfig struct {
	tp     tableprinter.TablePrinter
	client api.GQLClient
	opts   editItemOpts
}

type EditProjectDraftIssue struct {
	UpdateProjectV2DraftIssue struct {
		DraftIssue queries.DraftIssue `graphql:"draftIssue"`
	} `graphql:"updateProjectV2DraftIssue(input:$input)"`
}

func NewCmdEditItem(f *cmdutil.Factory, runF func(config editItemConfig) error) *cobra.Command {
	opts := editItemOpts{}
	editItemCmd := &cobra.Command{
		Short: "Edit a draft issue in a project",
		Use:   "item-edit",
		Example: `
# edit a draft issue
gh projects item-edit --id ID --title "a new title" --body "a new body"
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
			config := editItemConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runEditItem(config)
		},
	}

	editItemCmd.Flags().StringVar(&opts.title, "title", "", "Title of the draft issue item to edit.")
	editItemCmd.Flags().StringVar(&opts.body, "body", "", "Body of the draft issue item to edit.")
	editItemCmd.Flags().StringVar(&opts.itemID, "id", "", "ID of the draft issue item to edit. Must be the ID of the draft issue content which is prefixed with `DI_`")

	return editItemCmd
}

func runEditItem(config editItemConfig) error {
	if config.opts.title == "" && config.opts.body == "" {
		config.tp.AddField("No changes to make")
		config.tp.Render()
		return nil
	}

	if !strings.HasPrefix(config.opts.itemID, "DI_") {
		return errors.New("ID must be the ID of the draft issue content which is prefixed with `DI_`")
	}

	query, variables := buildEditDraftIssue(config)

	err := config.client.Mutate("EditDraftIssueItem", query, variables)
	if err != nil {
		return err
	}

	return printResults(config, query.UpdateProjectV2DraftIssue.DraftIssue)
}

func buildEditDraftIssue(config editItemConfig) (*EditProjectDraftIssue, map[string]interface{}) {
	return &EditProjectDraftIssue{}, map[string]interface{}{
		"input": githubv4.UpdateProjectV2DraftIssueInput{
			Body:         githubv4.NewString(githubv4.String(config.opts.body)),
			DraftIssueID: githubv4.ID(config.opts.itemID),
			Title:        githubv4.NewString(githubv4.String(config.opts.title)),
			// optionally include assignees
		},
	}
}

func printResults(config editItemConfig, item queries.DraftIssue) error {
	// using table printer here for consistency in case it ends up being needed in the future
	config.tp.AddField("Title")
	config.tp.AddField("Body")
	config.tp.EndRow()
	config.tp.AddField(item.Title)
	config.tp.AddField(item.Body)
	config.tp.EndRow()
	return config.tp.Render()
}
