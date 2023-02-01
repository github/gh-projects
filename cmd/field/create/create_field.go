package create

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

type createFieldOpts struct {
	name                string
	dataType            string
	userOwner           string
	singleSelectOptions []string
	orgOwner            string
	number              int
	projectID           string
}

type createFieldConfig struct {
	tp     tableprinter.TablePrinter
	client api.GQLClient
	opts   createFieldOpts
}

type createProjectV2FieldMutation struct {
	CreateProjectV2Field struct {
		Field queries.ProjectField `graphql:"projectV2Field"`
	} `graphql:"createProjectV2Field(input:$input)"`
}

// since this api is still in preview, this struct doesn't yet exist in githubv4
type CreateProjectV2FieldInput struct {
	ProjectID           githubv4.ID     `json:"projectId"`
	DataType            githubv4.String `json:"dataType"`
	Name                githubv4.String `json:"name"`
	SingleSelectOptions []string        `json:"singleSelectOptions,omitempty"`
}

func NewCmdCreateField(f *cmdutil.Factory, runF func(config createFieldConfig) error) *cobra.Command {
	opts := createFieldOpts{}
	createFieldCmd := &cobra.Command{
		Short: "Create a field in a project",
		Use:   "create number",
		Example: `
# create a field in the current user's project 1 with title "new item" and dataType "text"
gh projects field create 1 --user "@me" --name "new field" --data-type "text"

# create a field in monalisa user project 1 with title "new item" and dataType "text"
gh projects field create 1 --user monalisa --name "new field" --data-type "text"

# create a field in the github org project 1 with title "new item" and dataType "text"
gh projects field create 1 --org github --name "new field" --data-type "text"

# create a field with single select options
gh projects field create 1 --user monalisa --name "new field" --data-type "SINGLE_SELECT" --single-select-options "one,two,three"
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
			config := createFieldConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runCreateField(config)
		},
	}

	createFieldCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner. Use \"@me\" for the current user.")
	createFieldCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	createFieldCmd.Flags().StringVar(&opts.name, "name", "", "Name of the new field.")
	createFieldCmd.Flags().StringVar(&opts.dataType, "data-type", "", "DataType of the new field. Must be one of TEXT, SINGLE_SELECT, DATE, NUMBER.")
	createFieldCmd.Flags().StringSliceVar(&opts.singleSelectOptions, "single-select-options", []string{}, "At least one option is required when data type is SINGLE_SELECT.")

	createFieldCmd.MarkFlagsMutuallyExclusive("user", "org")
	createFieldCmd.MarkFlagRequired("name")
	createFieldCmd.MarkFlagRequired("data-type")

	return createFieldCmd
}

func runCreateField(config createFieldConfig) error {
	if config.opts.userOwner == "" && config.opts.orgOwner == "" {
		return fmt.Errorf("one of --user or --org is required")
	}

	if config.opts.dataType == "SINGLE_SELECT" && len(config.opts.singleSelectOptions) == 0 {
		return fmt.Errorf("at least one single select options is required with data type is SINGLE_SELECT")
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
	config.opts.projectID = projectID

	query, variables := createFieldArgs(config)

	err = config.client.Mutate("CreateField", query, variables)
	if err != nil {
		return err
	}

	return printResults(config, query.CreateProjectV2Field.Field)
}

func createFieldArgs(config createFieldConfig) (*createProjectV2FieldMutation, map[string]interface{}) {
	input := CreateProjectV2FieldInput{
		ProjectID: githubv4.ID(config.opts.projectID),
		DataType:  githubv4.String(config.opts.dataType),
		Name:      githubv4.String(config.opts.name),
	}

	if len(config.opts.singleSelectOptions) != 0 {
		input.SingleSelectOptions = config.opts.singleSelectOptions
	}

	return &createProjectV2FieldMutation{}, map[string]interface{}{
		"input": input,
	}
}

func printResults(config createFieldConfig, field queries.ProjectField) error {
	// using table printer here for consistency in case it ends up being needed in the future
	config.tp.AddField("Created field")
	config.tp.EndRow()
	return config.tp.Render()
}
