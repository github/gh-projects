package view

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/cli/cli/v2/pkg/cmdutil"

	"github.com/cli/browser"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/github/gh-projects/queries"
	"github.com/spf13/cobra"
)

type viewOpts struct {
	web       bool
	userOwner string
	orgOwner  string
	number    int
}

type viewConfig struct {
	tp        tableprinter.TablePrinter
	client    api.GQLClient
	opts      viewOpts
	URLOpener func(string) error
}

func NewCmdView(f *cmdutil.Factory, runF func(config viewConfig) error) *cobra.Command {
	opts := viewOpts{}
	viewCmd := &cobra.Command{
		Short: "View a project",
		Use:   "view number",
		Example: `
# view project 1 for the current user
gh projects view 1

# open project 1 for user "monalisa" in the browser
gh projects view 1 --user monalisa --web

# view project 1 for the github organization including closed projects
gh projects view 1 --org github --closed
`,
		Args: cobra.ExactArgs(1),
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

			opts.number, err = strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			t := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), termWidth)
			config := viewConfig{
				tp:        t,
				client:    client,
				opts:      opts,
				URLOpener: URLOpener,
			}
			return runView(config)
		},
	}

	viewCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner. Use \"@me\" for the current user.")
	viewCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	viewCmd.Flags().BoolVarP(&opts.web, "web", "w", false, "Open project in the browser.")

	// owner can be a user or an org
	viewCmd.MarkFlagsMutuallyExclusive("user", "org")

	return viewCmd
}

func runView(config viewConfig) error {
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

	project, err := queries.ProjectView(config.client, login, ownerType, config.opts.number)
	if err != nil {
		return err
	}

	return printResults(config, project, login)
}

func buildURL(config viewConfig) (string, error) {
	var url string
	if config.opts.userOwner == "@me" {
		login, err := queries.ViewerLoginName(config.client)
		if err != nil {
			return "", err
		}
		url = fmt.Sprintf("https://github.com/users/%s/projects/%d", login, config.opts.number)
	} else if config.opts.userOwner != "" {
		url = fmt.Sprintf("https://github.com/users/%s/projects/%d", config.opts.userOwner, config.opts.number)
	} else if config.opts.orgOwner != "" {
		url = fmt.Sprintf("https://github.com/orgs/%s/projects/%d", config.opts.orgOwner, config.opts.number)
	}

	return url, nil
}

func printResults(config viewConfig, project queries.Project, login string) error {

	var sb strings.Builder
	sb.WriteString("# Title\n")
	sb.WriteString(project.Title)
	sb.WriteString("\n")

	sb.WriteString("## Description\n")
	if project.ShortDescription == "" {
		sb.WriteString(" -- ")
	} else {
		sb.WriteString(project.ShortDescription)
	}
	sb.WriteString("\n")

	sb.WriteString("## Visibility\n")
	if project.Public {
		sb.WriteString("Public")
	} else {
		sb.WriteString("Private")
	}
	sb.WriteString("\n")

	sb.WriteString("## URL\n")
	sb.WriteString(project.URL)
	sb.WriteString("\n")

	sb.WriteString("## ID\n")
	sb.WriteString(project.ID)
	sb.WriteString("\n")

	sb.WriteString("## Item count\n")
	sb.WriteString(fmt.Sprintf("%d", project.Items.TotalCount))
	sb.WriteString("\n")

	sb.WriteString("## Readme\n")
	if project.Readme == "" {
		sb.WriteString(" -- ")
	} else {
		sb.WriteString(project.Readme)
	}
	sb.WriteString("\n")

	sb.WriteString("## Field Name (Field Type)\n")
	for _, f := range project.Fields.Nodes {
		sb.WriteString(fmt.Sprintf("%s (%s)\n\n", f.Name(), f.Type()))
	}

	// TODO: respect the glamour env var if set
	out, err := glamour.Render(sb.String(), "dark")
	if err != nil {
		return err
	}
	fmt.Println(out)

	return config.tp.Render()
}
