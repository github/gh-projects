package list

import (
	"fmt"

	"github.com/cli/cli/v2/pkg/cmdutil"

	"github.com/cli/browser"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/github/gh-projects/queries"
	"github.com/shurcooL/graphql"
	"github.com/spf13/cobra"
)

type listOpts struct {
	limit     int
	web       bool
	userOwner string
	orgOwner  string
	viewer    bool
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
		Short: "list the projects",
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

	listCmd.Flags().IntVar(&opts.limit, "limit", 0, "Maximum number of queue entries to get. Defaults to 100.")
	listCmd.Flags().BoolVarP(&opts.closed, "closed", "c", false, "Show closed projects.")
	listCmd.Flags().BoolVarP(&opts.web, "web", "w", false, "Open projects list in the browser.")
	listCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner.")
	listCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	listCmd.Flags().BoolVar(&opts.viewer, "me", false, "Login of the current user as the project owner.")

	// owner can be a user or an org
	listCmd.MarkFlagsMutuallyExclusive("user", "org", "me")

	return listCmd
}

func runList(config listConfig) error {
	// TODO interactive survey if no arguments are provided
	if !config.opts.viewer && config.opts.userOwner == "" && config.opts.orgOwner == "" {
		return fmt.Errorf("one of --user, --org or --me is required")
	}

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

	projectsQuery, variables := buildQuery(config)

	err := config.client.Query("ProjectsQuery", projectsQuery, variables)
	if err != nil {
		return err
	}

	projects := filterProjects(projectsQuery.Projects().Nodes, config)

	return printResults(config, projects, projectsQuery.Login())
}

func buildQuery(config listConfig) (queries.ProjectsQuery, map[string]interface{}) {
	var projectsQuery queries.ProjectsQuery
	variables := map[string]interface{}{
		"first": graphql.Int(config.opts.first()),
	}

	if config.opts.viewer {
		projectsQuery = &queries.ProjectsViewerQuery{}
	} else if config.opts.userOwner != "" {
		projectsQuery = &queries.ProjectsUserQuery{}
		variables["login"] = graphql.String(config.opts.userOwner)
	} else if config.opts.orgOwner != "" {
		projectsQuery = &queries.ProjectsOrganizationQuery{}
		variables["login"] = graphql.String(config.opts.orgOwner)
	}

	return projectsQuery, variables
}

func buildURL(config listConfig) (string, error) {
	var url string
	if config.opts.viewer {
		viewer := &queries.ProjectViewerLogin{}
		// get the current user's login
		err := config.client.Query("Viewer", viewer, map[string]interface{}{})
		if err != nil {
			return "", err
		}
		login := viewer.Viewer.Login
		url = fmt.Sprintf("https://github.com/users/%s/projects", login)
	} else if config.opts.userOwner != "" {
		url = fmt.Sprintf("https://github.com/users/%s/projects", config.opts.userOwner)
	} else if config.opts.orgOwner != "" {
		url = fmt.Sprintf("https://github.com/orgs/%s/projects", config.opts.orgOwner)
	}

	if config.opts.closed {
		url = fmt.Sprintf("%s?query=is%%3Aclosed", url)
	}

	return url, nil
}

func filterProjects(nodes []queries.ProjectNode, config listConfig) []queries.ProjectNode {
	projects := make([]queries.ProjectNode, 0, len(nodes))
	for _, p := range nodes {
		if !config.opts.closed && p.Closed {
			continue
		}
		projects = append(projects, p)
	}
	return projects
}

func printResults(config listConfig, projects []queries.ProjectNode, login string) error {
	// no projects
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
		config.tp.EndRow()
	}

	return config.tp.Render()
}
