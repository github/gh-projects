package itemlist

import (
	"encoding/json"
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
		Short: "List the items in a project",
		Use:   "item-list [number]",
		Example: `
# list the items in the current users's project number 1
gh projects item-list 1 --user "@me"

# list the items in user monalisa's project number 1
gh projects item-list 1 --user monalisa

# list the items in org github's project number 1
gh projects item-list 1 --org github
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := queries.NewClient()
			if err != nil {
				return err
			}

			if len(args) == 1 {
				opts.number, err = strconv.Atoi(args[0])
				if err != nil {
					return err
				}
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
	listCmd.Flags().StringVar(&opts.format, "format", "", "Output format, must be one of 'json' or 'csv'.")
	listCmd.Flags().IntVar(&opts.limit, "limit", 0, "Maximum number of items to get. Defaults to 100.")

	// owner can be a user or an org
	listCmd.MarkFlagsMutuallyExclusive("user", "org")

	return listCmd
}

func runList(config listConfig) error {
	if config.opts.format != "" && config.opts.format != "csv" && config.opts.format != "json" {
		return fmt.Errorf("format must be one of 'json' or 'csv'")
	}

	owner, err := queries.NewOwner(config.client, config.opts.userOwner, config.opts.orgOwner)
	if err != nil {
		return err
	}

	project, err := queries.ProjectItems(config.client, owner, config.opts.number, config.opts.first())
	if err != nil {
		return err
	}

	if config.opts.format == "json" {
		return jsonPrint(config, project)
	}

	return printResults(config, project.Items.Nodes, owner.Login)
}

func printResults(config listConfig, items []queries.ProjectItem, login string) error {
	if len(items) == 0 {
		config.tp.AddField(fmt.Sprintf("Project %d for login %s has no items", config.opts.number, login))
		config.tp.EndRow()
		return config.tp.Render()
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

func serialize(project queries.Project) []map[string]any {
	fields := make(map[string]string)

	for _, f := range project.Fields.Nodes {
		fields[f.ID()] = f.Name()
	}
	itemsSlice := make([]map[string]any, 0)
	for _, i := range project.Items.Nodes {
		o := make(map[string]any)
		for _, v := range i.FieldValues.Nodes {
			// name and value based on type
			switch v.Type {
			case "ProjectV2ItemFieldDateValue":
				o[fields[v.ProjectV2ItemFieldDateValue.Field.ID()]] = v.ProjectV2ItemFieldDateValue.Date
			case "ProjectV2ItemFieldIterationValue":
				o[fields[v.ProjectV2ItemFieldIterationValue.Field.ID()]] = v.ProjectV2ItemFieldIterationValue.StartDate // what about duration
			case "ProjectV2ItemFieldNumberValue":
				o[fields[v.ProjectV2ItemFieldNumberValue.Field.ID()]] = fmt.Sprintf("%f", v.ProjectV2ItemFieldNumberValue.Number)
			case "ProjectV2ItemFieldSingleSelectValue":
				o[fields[v.ProjectV2ItemFieldSingleSelectValue.Field.ID()]] = v.ProjectV2ItemFieldSingleSelectValue.Name
			case "ProjectV2ItemFieldTextValue":
				o[fields[v.ProjectV2ItemFieldTextValue.Field.ID()]] = v.ProjectV2ItemFieldTextValue.Text
			case "ProjectV2ItemFieldMilestoneValue":
				o[fields[v.ProjectV2ItemFieldMilestoneValue.Field.ID()]] = struct {
					Description string
					DueOn       string
				}{
					Description: v.ProjectV2ItemFieldMilestoneValue.Milestone.Description,
					DueOn:       v.ProjectV2ItemFieldMilestoneValue.Milestone.DueOn,
				}
			case "ProjectV2ItemFieldLabelValue":
				name := make([]string, 0)
				for _, p := range v.ProjectV2ItemFieldLabelValue.Labels.Nodes {
					name = append(name, p.Name)
				}
				o[fields[v.ProjectV2ItemFieldLabelValue.Field.ID()]] = name

			case "ProjectV2ItemFieldPullRequestValue":
				urls := make([]string, 0)
				for _, p := range v.ProjectV2ItemFieldPullRequestValue.PullRequests.Nodes {
					urls = append(urls, p.Url)
				}
				o[fields[v.ProjectV2ItemFieldPullRequestValue.Field.ID()]] = urls
			case "ProjectV2ItemFieldRepositoryValue":
				o[fields[v.ProjectV2ItemFieldRepositoryValue.Field.ID()]] = v.ProjectV2ItemFieldRepositoryValue.Repository.Url
			case "ProjectV2ItemFieldUserValue":
				logins := make([]string, 0)
				for _, p := range v.ProjectV2ItemFieldUserValue.Users.Nodes {
					logins = append(logins, p.Login)
				}
				o[fields[v.ProjectV2ItemFieldUserValue.Field.ID()]] = logins
			case "ProjectV2ItemFieldReviewerValue":
				names := make([]string, 0)
				for _, p := range v.ProjectV2ItemFieldReviewerValue.Reviewers.Nodes {
					if p.Type == "Team" {
						names = append(names, p.Team.Name)
					} else if p.Type == "User" {
						names = append(names, p.User.Login)
					}
				}
				o[fields[v.ProjectV2ItemFieldReviewerValue.Field.ID()]] = names

			}
		}
		itemsSlice = append(itemsSlice, o)
	}
	return itemsSlice
}

func jsonPrint(config listConfig, project queries.Project) error {
	items := serialize(project)
	b, err := json.Marshal(items)
	if err != nil {
		return err
	}
	config.tp.AddField(string(b))
	config.tp.EndRow()
	return config.tp.Render()

}
