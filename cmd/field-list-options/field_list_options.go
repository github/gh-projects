package fieldlistvalues

import (
	"fmt"
	"github.com/github/gh-projects/format"
	"strconv"

	"github.com/cli/cli/v2/pkg/cmdutil"

	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/github/gh-projects/queries"
	"github.com/spf13/cobra"
)

type listOpts struct {
	limit     int
	userOwner string
	orgOwner  string
	number    int
	itemID    string
	format    string
}

type listConfig struct {
	tp     tableprinter.TablePrinter
	client api.GQLClient
	opts   listOpts
}

func (opts *listOpts) first() int {
	if opts.limit == 0 {
		return 100
	}
	return opts.limit
}

func NewCmdListValues(f *cmdutil.Factory, runF func(config listConfig) error) *cobra.Command {
	opts := listOpts{}
	listCmd := &cobra.Command{
		Short: "List the field options.",
		Use:   "field-list-options [number]",
		Example: `
# list the field options in the current user's project number 1
gh projects field-list-options 1 --id ID --user "@me"

# list the field values in user monalisa's project number 1
gh projects field-list-options 1 --id ID --user monalisa

# list the first 30 fields in org github's project number 1
gh projects field-list-options 1 --id ID --org github --limit 30

# add --format=json to output in JSON format
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
				return nil
			}

			opts.number, err = strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			t := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), termWidth)
			config := listConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runList(config)
		},
	}

	listCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner. Use \"@me\" for the current user.")
	listCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	listCmd.Flags().StringVar(&opts.itemID, "id", "", "ID of the field to list from the project.")
	listCmd.Flags().IntVar(&opts.limit, "limit", 0, "Maximum number of fields to get. Defaults to 100.")
	listCmd.Flags().StringVar(&opts.format, "format", "", "Output format, must be 'json'.")

	// owner can be a user or an org
	listCmd.MarkFlagsMutuallyExclusive("user", "org")

	return listCmd
}

func runList(config listConfig) error {
	if config.opts.format != "" && config.opts.format != "json" {
		return fmt.Errorf("format must be 'json'")
	}

	owner, err := queries.NewOwner(config.client, config.opts.userOwner, config.opts.orgOwner)
	if err != nil {
		return err
	}

	fields, err := queries.ProjectFieldWithValues(config.client, owner, config.opts.number, config.opts.first())
	if err != nil {
		return err
	}
	var field queries.ProjectFieldWithOptions
	for _, f := range fields {
		if f.ID() == config.opts.itemID {
			field = f
		}
	}

	return printResults(config, field, owner.Login)
}

func printResults(config listConfig, field queries.ProjectFieldWithOptions, login string) error {
	if field.ID() == "" {
		config.tp.AddField(fmt.Sprintf("Project %d for login %s has no fields with given ID.", config.opts.number, login))
		config.tp.EndRow()
		return config.tp.Render()
	}

	if field.TypeName != "ProjectV2IterationField" && field.TypeName != "ProjectV2SingleSelectField" {
		config.tp.AddField(fmt.Sprintf("Field \"%s\" does not have options.", field.Name()))
		config.tp.EndRow()
		return config.tp.Render()
	}

	if field.TypeName == "ProjectV2IterationField" {
		return printIterationField(config, field)
	}

	if field.TypeName == "ProjectV2SingleSelectField" {
		return printSingleSelectField(config, field)
	}

	return nil
}

func printIterationField(config listConfig, field queries.ProjectFieldWithOptions) error {
	if config.opts.format == "json" {
		return printIterationFieldJSON(config, field)
	}

	config.tp.AddField("ID")
	config.tp.AddField("Title")
	config.tp.AddField("Start Date")
	config.tp.AddField("Duration")
	config.tp.AddField("Completed")
	config.tp.EndRow()

	for _, i := range reverseSlice(field.IterationField.Configuration.CompletedIterations) {
		config.tp.AddField(i.Id)
		config.tp.AddField(i.Title)
		config.tp.AddField(i.StartDate.String())
		config.tp.AddField(strconv.Itoa(i.Duration))
		config.tp.AddField(strconv.FormatBool(true))
		config.tp.EndRow()
	}
	for _, i := range field.IterationField.Configuration.Iterations {
		config.tp.AddField(i.Id)
		config.tp.AddField(i.Title)
		config.tp.AddField(i.StartDate.String())
		config.tp.AddField(strconv.Itoa(i.Duration))
		config.tp.EndRow()
	}
	return config.tp.Render()
}

func printIterationFieldJSON(config listConfig, field queries.ProjectFieldWithOptions) error {
	b, err := format.JSONProjectFieldsIterableOptions(
		reverseSlice(field.IterationField.Configuration.CompletedIterations),
		field.IterationField.Configuration.Iterations,
	)
	if err != nil {
		return err
	}
	config.tp.AddField(string(b))
	return config.tp.Render()
}

func printSingleSelectField(config listConfig, field queries.ProjectFieldWithOptions) error {
	if config.opts.format == "json" {
		return printSingleSelectFieldJSON(config, field)
	}

	config.tp.AddField("ID")
	config.tp.AddField("Name")
	config.tp.EndRow()

	for _, o := range field.SingleSelectField.Options {
		config.tp.AddField(o.ID)
		config.tp.AddField(o.Name)
		config.tp.EndRow()
	}
	return config.tp.Render()
}

func printSingleSelectFieldJSON(config listConfig, field queries.ProjectFieldWithOptions) error {
	b, err := format.JSONProjectFieldsSingleSelectOptions(field.SingleSelectField.Options)
	if err != nil {
		return err
	}
	config.tp.AddField(string(b))
	return config.tp.Render()
}

func reverseSlice[T comparable](s []T) []T {
	var r []T
	for i := len(s) - 1; i >= 0; i-- {
		r = append(r, s[i])
	}
	return r
}
