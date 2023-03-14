package fielddelete

import (
	"fmt"

	"github.com/cli/cli/v2/pkg/cmdutil"

	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/github/gh-projects/format"
	"github.com/github/gh-projects/queries"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
)

type deleteFieldOpts struct {
	fieldID string
	format  string
}

type deleteFieldConfig struct {
	tp     tableprinter.TablePrinter
	client api.GQLClient
	opts   deleteFieldOpts
}

type deleteProjectV2FieldMutation struct {
	DeleteProjectV2Field struct {
		Field queries.ProjectField `graphql:"projectV2Field"`
	} `graphql:"deleteProjectV2Field(input:$input)"`
}

// TODO: update this to use githubv4.DeleteProjectV2FieldInput once it is available there
type DeleteProjectV2FieldInput struct {
	FieldID githubv4.ID `json:"fieldId"`
}

func NewCmdDeleteField(f *cmdutil.Factory, runF func(config deleteFieldConfig) error) *cobra.Command {
	opts := deleteFieldOpts{}
	deleteFieldCmd := &cobra.Command{
		Short: "Delete a field in a project by ID",
		Use:   "field-delete",
		Example: `
# delete a field by ID
gh projects field-delete --id ID
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := queries.NewClient()
			if err != nil {
				return err
			}

			terminal := term.FromEnv()
			termWidth, _, err := terminal.Size()
			if err != nil {
				// set a static width in case of error
				termWidth = 80
			}
			t := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), termWidth)

			config := deleteFieldConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runDeleteField(config)
		},
	}

	deleteFieldCmd.Flags().StringVar(&opts.fieldID, "id", "", "ID of the field to delete.")
	deleteFieldCmd.Flags().StringVar(&opts.format, "format", "", "Output format, must be 'json'.")

	_ = deleteFieldCmd.MarkFlagRequired("id")

	return deleteFieldCmd
}

func runDeleteField(config deleteFieldConfig) error {
	if config.opts.format != "" && config.opts.format != "json" {
		return fmt.Errorf("format must be 'json'")
	}

	query, variables := deleteFieldArgs(config)

	err := config.client.Mutate("DeleteField", query, variables)
	if err != nil {
		return err
	}

	if config.opts.format == "json" {
		return printJSON(config, query.DeleteProjectV2Field.Field)
	}

	return printResults(config, query.DeleteProjectV2Field.Field)
}

func deleteFieldArgs(config deleteFieldConfig) (*deleteProjectV2FieldMutation, map[string]interface{}) {
	return &deleteProjectV2FieldMutation{}, map[string]interface{}{
		"input": DeleteProjectV2FieldInput{
			FieldID: githubv4.ID(config.opts.fieldID),
		},
	}
}

func printResults(config deleteFieldConfig, field queries.ProjectField) error {
	// using table printer here for consistency in case it ends up being needed in the future
	config.tp.AddField("Deleted field")
	config.tp.EndRow()
	return config.tp.Render()
}

func printJSON(config deleteFieldConfig, field queries.ProjectField) error {
	b, err := format.JSONProjectField(field)
	if err != nil {
		return err
	}
	config.tp.AddField(string(b))
	return config.tp.Render()
}
