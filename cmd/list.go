package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cli/browser"
	gh "github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/shurcooL/graphql"
	"github.com/spf13/cobra"
)

type listOpts struct {
	limit     int
	login     string
	web       bool
	userOwner bool
	orgOwner  bool
	closed    bool
}

type listConfig struct {
	tp        tableprinter.TablePrinter
	out       io.Writer
	client    querier
	opts      listOpts
	URLOpener func(string) error
}

func (opts *listOpts) first() int {
	if opts.limit == 0 {
		return 100
	}
	return opts.limit
}

func NewListCmd() *cobra.Command {
	opts := listOpts{}
	listCmd := &cobra.Command{
		Short: "list the projects",
		Use:   "list",
		Example: `
# list the projects for the current user
gh projects list

# open projects for user "hubot" in the browser
gh projects list --login hubot --user --web

# list the projects for the github organization including closed projects
gh projects list --login github --org --closed
`,
		Run: func(cmd *cobra.Command, args []string) {
			apiOpts := api.ClientOptions{
				Timeout: 5 * time.Second,
			}

			client, err := gh.GQLClient(&apiOpts)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			URLOpener := func(url string) error {
				return browser.OpenURL(url)
			}
			terminal := term.FromEnv()
			termWidth, _, err := terminal.Size()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			t := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), termWidth)
			config := listConfig{
				tp:        t,
				out:       terminal.Out(),
				client:    client,
				opts:      opts,
				URLOpener: URLOpener,
			}
			runList(config)
		},
	}

	listCmd.Flags().StringVarP(&opts.login, "login", "l", "", "Login of the project owner. Defaults to current user.")
	listCmd.Flags().IntVar(&opts.limit, "limit", 0, "Maximum number of queue entries to get. Defaults to 100.")
	listCmd.Flags().BoolVarP(&opts.closed, "closed", "c", false, "Show closed projects.")
	listCmd.Flags().BoolVarP(&opts.web, "web", "w", false, "Open projects list in the browser.")
	listCmd.Flags().BoolVar(&opts.userOwner, "user", false, "Owner is a user.")
	listCmd.Flags().BoolVar(&opts.orgOwner, "org", false, "Owner is an organization.")
	// owner can be a user or an org
	listCmd.MarkFlagsMutuallyExclusive("user", "org")

	return listCmd
}

func runList(config listConfig) {
	if config.opts.login != "" && !config.opts.userOwner && !config.opts.orgOwner {
		fmt.Println("One of --user or --org is required with --login")
		os.Exit(1)
	}

	if config.opts.web {
		url, err := buildURL(config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := config.URLOpener(url); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	projectsQuery, variables := buildQuery(config)

	err := config.client.Query("ProjectsQuery", projectsQuery, variables)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	projects := filterProjects(projectsQuery.projects().Nodes, config)

	printResults(config, projects, projectsQuery.login())
}

func buildQuery(config listConfig) (query, map[string]interface{}) {
	var projectsQuery query
	variables := map[string]interface{}{
		"first": graphql.Int(config.opts.first()),
	}

	if config.opts.login == "" {
		projectsQuery = &viewerQuery{}
	} else if config.opts.userOwner {
		projectsQuery = &userQuery{}
		variables["login"] = graphql.String(config.opts.login)
	} else if config.opts.orgOwner {
		projectsQuery = &organizationQuery{}
		variables["login"] = graphql.String(config.opts.login)
	}

	return projectsQuery, variables
}

func buildURL(config listConfig) (string, error) {
	var url string
	if config.opts.login == "" {
		// get the current user's login
		err := config.client.Query("Viewer", &queryViewer, map[string]interface{}{})
		if err != nil {
			return "", err
		}
		user := queryViewer.Viewer.Login
		url = fmt.Sprintf("https://github.com/users/%s/projects", user)
	} else if config.opts.userOwner {
		url = fmt.Sprintf("https://github.com/users/%s/projects", config.opts.login)
	} else if config.opts.orgOwner {
		url = fmt.Sprintf("https://github.com/orgs/%s/projects", config.opts.login)
	}

	if config.opts.closed {
		url = fmt.Sprintf("%s?query=is%%3Aclosed", url)
	}

	return url, nil
}

func filterProjects(nodes []projectNode, config listConfig) []projectNode {
	projects := make([]projectNode, 0, len(nodes))
	for _, p := range nodes {
		if !config.opts.closed && p.Closed {
			continue
		}
		projects = append(projects, p)
	}
	return projects
}

func printResults(config listConfig, projects []projectNode, login string) {
	// no projects
	if len(projects) == 0 {
		fmt.Fprintf(
			config.out,
			"No projects found for %s\n",
			login,
		)
		return
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

	if err := config.tp.Render(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
