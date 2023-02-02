package list

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
	limit     int
	userOwner string
	orgOwner  string
	number    int
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
		Use:   "list number",
		Example: `
# list the fields in project number 1 for the current user
gh projects field list 1 --user "@me"

# list the fields in project number 1 for user monalisa
gh projects field list 1 --user monalisa

# list the first 30 fields in project number 1 for org github
gh projects field list 1 --org github --limit 30
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

	fields, err := queries.ProjectFields(config.client, owner, config.opts.number, config.opts.first())
	if err != nil {
		return err
	}

	return printResults(config, fields, owner.Login)
}

func printResults(config listConfig, fields []queries.ProjectField, login string) error {
	if len(fields) == 0 {
		config.tp.AddField(fmt.Sprintf("Project %d for login %s has no fields", config.opts.number, login))
		config.tp.EndRow()
		config.tp.Render()
		return nil
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
