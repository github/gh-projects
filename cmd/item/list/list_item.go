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
		Short: "List the items in a project",
		Use:   "list number",
		Example: `
# list the items in project number 1 for the current user
gh projects item list 1 --user "@me"

# list the items in project number 1 for user monalisa
gh projects item list 1 --user monalisa

# list the items in project number 1 for org github
gh projects item list 1 --org github
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
	listCmd.Flags().IntVar(&opts.limit, "limit", 0, "Maximum number of items to get. Defaults to 100.")

	// owner can be a user or an org
	listCmd.MarkFlagsMutuallyExclusive("user", "org")

	return listCmd
}

func runList(config listConfig) error {
	if config.opts.userOwner == "" && config.opts.orgOwner == "" {
		return fmt.Errorf("one of --user or --org is required")
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

	items, err := queries.ProjectItems(config.client, login, ownerType, config.opts.number, config.opts.first())
	if err != nil {
		return err
	}

	return printResults(config, items, login)
}

func printResults(config listConfig, items []queries.ProjectItem, login string) error {
	if len(items) == 0 {
		config.tp.AddField(fmt.Sprintf("Project %d for login %s has no items", config.opts.number, login))
		config.tp.EndRow()
		config.tp.Render()
		return nil
	}

	config.tp.AddField("Type")
	config.tp.AddField("Title")
	config.tp.AddField("Number")
	config.tp.AddField("Repository")
	config.tp.AddField("ID")
	config.tp.EndRow()

	for _, i := range items {
		config.tp.AddField(i.Type())
		config.tp.AddField(i.Title())
		if i.Number() == 0 {
			config.tp.AddField(" - ")
		} else {
			config.tp.AddField(fmt.Sprintf("%d", i.Number()))
		}
		if i.Repo() == "" {
			config.tp.AddField(" - ")
		} else {
			config.tp.AddField(i.Repo())
		}
		config.tp.AddField(i.ID())
		config.tp.EndRow()
	}

	return config.tp.Render()
}
