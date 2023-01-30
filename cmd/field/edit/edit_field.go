package edit

import (
	"github.com/cli/cli/v2/pkg/cmdutil"

	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/github/gh-projects/queries"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
)

type editFieldOpts struct {
	name                string
	singleSelectOptions []string
	fieldID             string
}

type editFieldConfig struct {
	tp     tableprinter.TablePrinter
	client api.GQLClient
	opts   editFieldOpts
}

type editProjectV2FieldMutation struct {
	UpdateProjectV2Field struct {
		Field queries.ProjectField `graphql:"projectV2Field"`
	} `graphql:"updateProjectV2Field(input:$input)"`
}

// since this api is still in preview, this struct doesn't yet exist in githubv4
type UpdateProjectV2FieldInput struct {
	FieldID             githubv4.ID     `json:"fieldId"`
	Name                githubv4.String `json:"name,omitempty"`
	SingleSelectOptions []string        `json:"singleSelectOptions,omitempty"`
}

func NewCmdEditField(f *cmdutil.Factory, runF func(config editFieldConfig) error) *cobra.Command {
	opts := editFieldOpts{}
	editFieldCmd := &cobra.Command{
		Short: "Edit a field in a project",
		Use:   "edit",
		Example: `
# edit a field to have the title "new name"
gh projects field edit --id ID --name "new name"

# edit a single select field to have new options
gh projects field edit --id ID --single-select-options "one,two,three"
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
			config := editFieldConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runEditField(config)
		},
	}

	editFieldCmd.Flags().StringVar(&opts.name, "name", "", "New name of the field. OPTIONAL.")
	editFieldCmd.Flags().StringVar(&opts.fieldID, "id", "", "ID of the field to edit.")
	editFieldCmd.Flags().StringSliceVar(&opts.singleSelectOptions, "single-select-options", []string{}, "New options for a field of type SINGLE_SELECT. OPTIONAL.")

	editFieldCmd.MarkFlagRequired("id")

	return editFieldCmd
}

func runEditField(config editFieldConfig) error {
	if config.opts.name == "" && len(config.opts.singleSelectOptions) == 0 {
		config.tp.AddField("No changes to make")
		config.tp.Render()
		return nil
	}

	query, variables := editFieldArgs(config)

	err := config.client.Mutate("EditField", query, variables)
	if err != nil {
		return err
	}

	return printResults(config, query.UpdateProjectV2Field.Field)
}

func editFieldArgs(config editFieldConfig) (*editProjectV2FieldMutation, map[string]interface{}) {
	input := UpdateProjectV2FieldInput{
		FieldID: githubv4.ID(config.opts.fieldID),
	}

	if config.opts.name != "" {
		input.Name = githubv4.String(config.opts.name)
	}

	if len(config.opts.singleSelectOptions) != 0 {
		input.SingleSelectOptions = config.opts.singleSelectOptions
	}

	return &editProjectV2FieldMutation{}, map[string]interface{}{
		"input": input,
	}
}

func printResults(config editFieldConfig, field queries.ProjectField) error {
	// using table printer here for consistency in case it ends up being needed in the future
	config.tp.AddField("Edited field")
	config.tp.EndRow()
	return config.tp.Render()
}
