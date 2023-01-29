package archive

import (
	"fmt"

	"github.com/cli/cli/v2/pkg/cmdutil"

	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/github/gh-projects/queries"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
)

type archiveItemOpts struct {
	userOwner string
	orgOwner  string
	viewer    bool
	number    int
	// itemURL   string
	itemID string
}

type archiveItemConfig struct {
	tp        tableprinter.TablePrinter
	client    api.GQLClient
	opts      archiveItemOpts
	projectID string
}

func NewCmdArchiveItem(f *cmdutil.Factory, runF func(config archiveItemConfig) error) *cobra.Command {
	opts := archiveItemOpts{}
	archiveItemCmd := &cobra.Command{
		Short: "archive an item from a project",
		Use:   "archive",
		Example: `
# archive an item in the current user's project
gh projects item archive --me --number 1 --id ID

# archive an item in a user project
gh projects item archive --user monalisa --number 1 --id ID

# archive an item in an org project
gh projects item archive --org github --number 1 --id ID
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := queries.NewClient()
			if err != nil {
				return err
			}

			terminal := term.FromEnv()
			termWidth, _, err := terminal.Size()
			if err != nil {
				return err
			}

			t := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), termWidth)
			config := archiveItemConfig{
				tp:     t,
				client: client,
				opts:   opts,
			}
			return runArchiveItem(config)
		},
	}

	archiveItemCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner.")
	archiveItemCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	archiveItemCmd.Flags().StringVar(&opts.itemID, "id", "", "Global ID of the item to archive from the project.")
	archiveItemCmd.Flags().BoolVar(&opts.viewer, "me", false, "Login of the current user as the project owner.")
	archiveItemCmd.Flags().IntVarP(&opts.number, "number", "n", 0, "The project number.")
	archiveItemCmd.MarkFlagsMutuallyExclusive("user", "org", "me")

	archiveItemCmd.MarkFlagRequired("number")
	archiveItemCmd.MarkFlagRequired("id")
	return archiveItemCmd
}

func runArchiveItem(config archiveItemConfig) error {
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
		ownerType = queries.ViewerOwner
	}

	projectID, err := queries.GetProjectId(config.client, login, ownerType, config.opts.number)
	if err != nil {
		return err
	}
	config.projectID = projectID

	query, variables := buildArchiveItem(config, config.opts.itemID)
	err = config.client.Mutate("ArchiveProjectItem", query, variables)
	if err != nil {
		return err
	}

	return printResults(config, query.ArchiveProjectItem.ProjectV2Item)
}

func buildArchiveItem(config archiveItemConfig, itemID string) (*queries.ArchiveProjectItem, map[string]interface{}) {
	return &queries.ArchiveProjectItem{}, map[string]interface{}{
		"input": githubv4.ArchiveProjectV2ItemInput{
			ProjectID: githubv4.ID(config.projectID),
			ItemID:    githubv4.ID(itemID),
		},
	}
}

func printResults(config archiveItemConfig, item queries.ProjectV2Item) error {
	// using table printer here for consistency in case it ends up being needed in the future
	config.tp.AddField("Archived item")
	config.tp.EndRow()
	return config.tp.Render()
}
