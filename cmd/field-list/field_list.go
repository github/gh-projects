package fieldlist

import (
	"fmt"
	"strconv"

	"github.com/cli/cli/v2/pkg/cmdutil"

	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/github/gh-projects/format"
	"github.com/github/gh-projects/queries"
	"github.com/spf13/cobra"
)

type listOpts struct {
	limit     int
	userOwner string
	orgOwner  string
	number    int
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

func NewCmdList(f *cmdutil.Factory, runF func(config listConfig) error) *cobra.Command {
	opts := listOpts{}
	listCmd := &cobra.Command{
		Short: "List the fields in a project",
		Use:   "field-list number",
		Example: `
# list the fields in the current user's project number 1
gh projects field-list 1 --user "@me"

# list the fields in user monalisa's project number 1
gh projects field-list 1 --user monalisa

# list the first 30 fields in org github's project number 1
gh projects field-list 1 --org github --limit 30

# add --format=json to output in JSON format
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := queries.NewClient()
			if err != nil {
				return err
			}

			opts.number, err = strconv.Atoi(args[0])
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

	project, err := queries.ProjectFields(config.client, owner, config.opts.number, config.opts.first())
	if err != nil {
		return err
	}

	if config.opts.format == "json" {
		return printJSON(config, project.Fields.Nodes)
	}

	return printResults(config, project.Fields.Nodes, owner.Login)
}

func printResults(config listConfig, fields []queries.ProjectField, login string) error {
	if len(fields) == 0 {
		config.tp.AddField(fmt.Sprintf("Project %d for login %s has no fields", config.opts.number, login))
		config.tp.EndRow()
		return config.tp.Render()
	}

	config.tp.AddField("Name")
	config.tp.AddField("DataType")
	config.tp.AddField("ID")
	config.tp.EndRow()

	for _, f := range fields {
		config.tp.AddField(f.Name())
		config.tp.AddField(f.Type())
		config.tp.AddField(f.ID())
		config.tp.EndRow()
	}

	return config.tp.Render()
}

func printJSON(config listConfig, fields []queries.ProjectField) error {
	b, err := format.JSONProjectFields(fields)
	if err != nil {
		return err
	}
	config.tp.AddField(string(b))

	return config.tp.Render()
}
