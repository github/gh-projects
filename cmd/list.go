package cmd

import (
	"fmt"
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

			runList(client, opts)
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

func runList(client api.GQLClient, opts listOpts) {
	if opts.login != "" && !opts.userOwner && !opts.orgOwner {
		fmt.Println("One of --user or --org is required with --login")
		os.Exit(1)
	}

	if opts.web {
		if err := openInBrowser(opts, client); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	projectsQuery, variables := buildQuery(opts)

	err := client.Query("ProjectsQuery", projectsQuery, variables)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	projects := make([]projectNode, 0, len(projectsQuery.projects().Nodes))
	for _, p := range projectsQuery.projects().Nodes {
		if !opts.closed && p.Closed {
			continue
		}
		projects = append(projects, p)
	}

	printResults(opts, projects, projectsQuery.login())
}

func buildQuery(opts listOpts) (query, map[string]interface{}) {
	var projectsQuery query
	variables := map[string]interface{}{
		"first": graphql.Int(opts.first()),
	}

	if opts.login == "" {
		projectsQuery = &viewerQuery{}
	} else if opts.userOwner {
		projectsQuery = &userQuery{}
		variables["login"] = graphql.String(opts.login)
	} else if opts.orgOwner {
		projectsQuery = &organizationQuery{}
		variables["login"] = graphql.String(opts.login)
	}

	return projectsQuery, variables
}

func openInBrowser(opts listOpts, client api.GQLClient) error {
	var url string
	if opts.login == "" {
		err := client.Query("Viewer", &queryViewer, map[string]interface{}{})
		if err != nil {
			return err
		}
		user := queryViewer.Viewer.Login
		url = fmt.Sprintf("https://github.com/users/%s/projects", user)
	} else if opts.userOwner {
		url = fmt.Sprintf("https://github.com/users/%s/projects", opts.login)
	} else if opts.orgOwner {
		url = fmt.Sprintf("https://github.com/orgs/%s/projects", opts.login)
	}

	if opts.closed {
		url = fmt.Sprintf("%s?query=is%%3Aclosed", url)
	}

	browser.OpenURL(url)
	return nil
}

func printResults(opts listOpts, projects []projectNode, login string) {
	terminal := term.FromEnv()
	// no projects
	if len(projects) == 0 {
		fmt.Fprintf(terminal.Out(),
			"No projects found for %s\n",
			login,
		)
		return
	}

	termWidth, _, err := terminal.Size()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	t := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), termWidth)

	t.AddField("Title")
	t.AddField("Description")
	t.AddField("URL")
	if opts.closed {
		t.AddField("State")
	}
	t.EndRow()

	for _, p := range projects {
		t.AddField(p.Title)
		if p.ShortDescription == "" {
			t.AddField(" - ")
		} else {
			t.AddField(p.ShortDescription)
		}
		t.AddField(p.URL)
		if opts.closed {
			var state string
			if p.Closed {
				state = "closed"
			} else {
				state = "open"
			}
			t.AddField(state)
		}
		t.EndRow()
	}

	if err := t.Render(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
