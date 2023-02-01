package list

import (
	"fmt"

	"github.com/cli/cli/v2/pkg/cmdutil"

	"github.com/cli/browser"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/github/gh-projects/queries"
	"github.com/spf13/cobra"
)

type listOpts struct {
	limit     int
	web       bool
	userOwner string
	orgOwner  string
	closed    bool
}

type listConfig struct {
	tp        tableprinter.TablePrinter
	client    api.GQLClient
	opts      listOpts
	URLOpener func(string) error
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
		Short: "List the projects for a user or organization",
		Use:   "list",
		Example: `
# list the projects for the current user
gh projects list

# open projects for user "hubot" in the browser
gh projects list --user hubot --web

# list the projects for the github organization including closed projects
gh projects list --org github --closed
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := queries.NewClient()
			if err != nil {
				return err
			}

			URLOpener := func(url string) error {
				return browser.OpenURL(url)
			}
			terminal := term.FromEnv()
			termWidth, _, err := terminal.Size()
			if err != nil {
				return nil
			}

			t := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), termWidth)
			config := listConfig{
				tp:        t,
				client:    client,
				opts:      opts,
				URLOpener: URLOpener,
			}
			return runList(config)
		},
	}

	listCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner")
	listCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	listCmd.Flags().IntVar(&opts.limit, "limit", 0, "Maximum number of projects. Defaults to 100.")
	listCmd.Flags().BoolVarP(&opts.closed, "closed", "c", false, "Show closed projects.")
	listCmd.Flags().BoolVarP(&opts.web, "web", "w", false, "Open projects list in the browser.")

	// owner can be a user or an org
	listCmd.MarkFlagsMutuallyExclusive("user", "org")

	return listCmd
}

func runList(config listConfig) error {
	if config.opts.web {
		url, err := buildURL(config)
		if err != nil {
			return err
		}

		if err := config.URLOpener(url); err != nil {
			return err
		}
		return nil
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

	projects, err := queries.Projects(config.client, login, ownerType, config.opts.first())
	if err != nil {
		return err
	}
	projects = filterProjects(projects, config)

	return printResults(config, projects, login)
}

func buildURL(config listConfig) (string, error) {
	var url string
	if config.opts.userOwner != "" {
		url = fmt.Sprintf("https://github.com/users/%s/projects", config.opts.userOwner)
	} else if config.opts.orgOwner != "" {
		url = fmt.Sprintf("https://github.com/orgs/%s/projects", config.opts.orgOwner)
	} else {
		login, err := queries.ViewerLoginName(config.client)
		if err != nil {
			return "", err
		}
		url = fmt.Sprintf("https://github.com/users/%s/projects", login)
	}

	if config.opts.closed {
		url = fmt.Sprintf("%s?query=is%%3Aclosed", url)
	}

	return url, nil
}

func filterProjects(nodes []queries.Project, config listConfig) []queries.Project {
	projects := make([]queries.Project, 0, len(nodes))
	for _, p := range nodes {
		if !config.opts.closed && p.Closed {
			continue
		}
		projects = append(projects, p)
	}
	return projects
}

func printResults(config listConfig, projects []queries.Project, login string) error {
	if len(projects) == 0 {
		config.tp.AddField(fmt.Sprintf("No projects found for %s", login))
		config.tp.EndRow()
		config.tp.Render()
		return nil
	}

	config.tp.AddField("Title")
	config.tp.AddField("Description")
	config.tp.AddField("URL")
	if config.opts.closed {
		config.tp.AddField("State")
	}
	config.tp.AddField("ID")
	config.tp.EndRow()

	for _, p := range projects {
		config.tp.AddField(p.Title)
		if p.ShortDescription == "" {
			config.tp.AddField(" - ")
		} else {
			config.tp.AddField(p.ShortDescription)
		}
		config.tp.AddField(p.URL)
		if config.opts.closed {
			var state string
			if p.Closed {
				state = "closed"
			} else {
				state = "open"
			}
			config.tp.AddField(state)
		}
		config.tp.AddField(p.ID)
		config.tp.EndRow()
	}

	return config.tp.Render()
}
