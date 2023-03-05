package fieldlistvalues

import (
	"fmt"
	"strconv"

	"github.com/cli/cli/v2/pkg/cmdutil"

	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/github/gh-projects/queries"
	"github.com/spf13/cobra"
)

type listOpts struct {
	limit     int // TODO
	userOwner string
	orgOwner  string
	number    int
	itemID    string
	projectID string
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
		Short: "List the field values in a project.",
		Use:   "field-list-values [number]",
		Example: `
# list the field values in the current user's project number 1
gh projects field-list-values 1 --id ID --user "@me"

# list the field values in user monalisa's project number 1
gh projects field-list-values 1 --id ID --user monalisa

# list the first 30 fields in org github's project number 1
gh projects field-list-values 1 --id ID --org github --limit 30
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

	// owner can be a user or an org
	listCmd.MarkFlagsMutuallyExclusive("user", "org")

	return listCmd
}

func runList(config listConfig) error {
	owner, err := queries.NewOwner(config.client, config.opts.userOwner, config.opts.orgOwner)
	if err != nil {
		return err
	}

	fields, err := queries.ProjectFieldWithValues(config.client, owner, config.opts.number, config.opts.first())
	if err != nil {
		return err
	}
	var field queries.ProjectFieldWithValue
	for _, f := range fields {
		if f.ID() == config.opts.itemID {
			field = f
		}
	}

	return printResults(config, field, owner.Login)
}

func printResults(config listConfig, field queries.ProjectFieldWithValue, login string) error {
	if field.ID() == "" {
		config.tp.AddField(fmt.Sprintf("Project %d for login %s has no fields with given ID", config.opts.number, login))
		config.tp.EndRow()
		return config.tp.Render()
	}

	if field.TypeName == "ProjectV2IterationField" {
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
	}

	return config.tp.Render()
}

func reverseSlice[T comparable](s []T) []T {
	var r []T
	for i := len(s) - 1; i >= 0; i-- {
		r = append(r, s[i])
	}
	return r
}
