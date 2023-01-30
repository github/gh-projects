package list

import (
	"fmt"

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
	viewer    bool
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
		Short: "list the fields in a project",
		Use:   "list",
		Example: `
# list the fields in project number 1 for the current user
gh projects field list --number 1

# list the fields in project number 1 for user monalisa
gh projects field list --number 1 --user monalisa
`,
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

			t := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), termWidth)
			config := listConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runList(config)
		},
	}

	listCmd.Flags().IntVar(&opts.limit, "limit", 0, "Maximum number of fields to get. Defaults to 100.")
	listCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner.")
	listCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	listCmd.Flags().IntVarP(&opts.number, "number", "n", 0, "The project number.")
	listCmd.Flags().BoolVar(&opts.viewer, "me", false, "Login of the current user as the project owner.")

	// owner can be a user or an org
	listCmd.MarkFlagsMutuallyExclusive("user", "org", "me")

	listCmd.MarkFlagRequired("number")

	return listCmd
}

func runList(config listConfig) error {
	// TODO interactive survey if no arguments are provided
	if !config.opts.viewer && config.opts.userOwner == "" && config.opts.orgOwner == "" {
		return fmt.Errorf("one of --user, --org or --me is required")
	}

	var login string
	var ownerType queries.OwnerType
	if config.opts.userOwner != "" {
		login = config.opts.userOwner
		ownerType = queries.UserOwner
	} else if config.opts.orgOwner != "" {
		login = config.opts.orgOwner
		ownerType = queries.OrgOwner
	} else {
		login = "me"
		ownerType = queries.ViewerOwner
	}

	fields, err := queries.ProjectFields(config.client, login, ownerType, config.opts.number, config.opts.first())
	if err != nil {
		return err
	}

	return printResults(config, fields, login)
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
	config.tp.EndRow()

	for _, f := range fields {
		config.tp.AddField(f.Name())
		config.tp.AddField(f.Type())
		config.tp.EndRow()
	}

	return config.tp.Render()
}
