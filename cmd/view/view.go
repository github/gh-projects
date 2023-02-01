package view

import (
	"fmt"
	"strconv"

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

	config.tp.AddField("Title")
	config.tp.AddField("Description")
	config.tp.AddField("Visibility")
	config.tp.AddField("URL")
	config.tp.AddField("ID")
	config.tp.AddField("Item count")
	config.tp.EndRow()

	config.tp.AddField(project.Title)
	if project.ShortDescription == "" {
		config.tp.AddField(" - ")
	} else {
		config.tp.AddField(project.ShortDescription)
	}
	if project.Public {
		config.tp.AddField("Public")
	} else {
		config.tp.AddField("Private")
	}
	config.tp.AddField(project.URL)
	config.tp.AddField(project.ID)
	config.tp.AddField(fmt.Sprintf("%d", project.Items.TotalCount))
	config.tp.EndRow()
	// empty space
	config.tp.AddField("")
	config.tp.EndRow()

	config.tp.AddField("Readme")
	config.tp.EndRow()
	if project.Readme == "" {
		config.tp.AddField(" - ")
	} else {
		config.tp.AddField(project.Readme)
	}
	config.tp.EndRow()
	// empty space
	config.tp.AddField("")
	config.tp.EndRow()

	config.tp.AddField("Field Name")
	config.tp.AddField("Field Type")
	config.tp.EndRow()
	for _, f := range project.Fields.Nodes {
		config.tp.AddField(f.Name())
		config.tp.AddField(f.Type())
		config.tp.EndRow()
	}
	return config.tp.Render()
}
