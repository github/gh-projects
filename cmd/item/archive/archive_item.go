package archive

import (
	"fmt"
	"strconv"

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
	number    int
	undo      bool
	itemID    string
	projectID string
}

type archiveItemConfig struct {
	tp     tableprinter.TablePrinter
	client api.GQLClient
	opts   archiveItemOpts
}

type archiveProjectItemMutation struct {
	ArchiveProjectItem struct {
		ProjectV2Item queries.ProjectItem `graphql:"item"`
	} `graphql:"archiveProjectV2Item(input:$input)"`
}

type unarchiveProjectItemMutation struct {
	UnarchiveProjectItem struct {
		ProjectV2Item queries.ProjectItem `graphql:"item"`
	} `graphql:"unarchiveProjectV2Item(input:$input)"`
}

func NewCmdArchiveItem(f *cmdutil.Factory, runF func(config archiveItemConfig) error) *cobra.Command {
	opts := archiveItemOpts{}
	archiveItemCmd := &cobra.Command{
		Short: "Archive an item in a project",
		Use:   "archive number",
		Example: `
# archive an item in the current user's project 1
gh projects item archive 1 --user "@me" --id ID

# archive an item in the monalisa user project 1
gh projects item archive 1 --user monalisa --id ID

# archive an item in the github org project 1
gh projects item archive 1 --org github --id ID

# unarchive an item
gh projects item archive 1 --user "@me" --id ID --undo
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
				return err
			}

			opts.number, err = strconv.Atoi(args[0])
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

	archiveItemCmd.Flags().StringVar(&opts.userOwner, "user", "", "Login of the user owner. Use \"@me\" for the current user.")
	archiveItemCmd.Flags().StringVar(&opts.orgOwner, "org", "", "Login of the organization owner.")
	archiveItemCmd.Flags().StringVar(&opts.itemID, "id", "", "Global ID of the item to archive from the project.")
	archiveItemCmd.Flags().BoolVar(&opts.undo, "undo", false, "Undo archive (unarchive) of an item.")

	archiveItemCmd.MarkFlagsMutuallyExclusive("user", "org")
	archiveItemCmd.MarkFlagRequired("id")

	return archiveItemCmd
}

func runArchiveItem(config archiveItemConfig) error {
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

	projectID, err := queries.ProjectId(config.client, login, ownerType, config.opts.number)
	if err != nil {
		return err
	}
	config.opts.projectID = projectID

	if config.opts.undo {
		query, variables := unarchiveItemArgs(config, config.opts.itemID)
		err = config.client.Mutate("UnarchiveProjectItem", query, variables)
		if err != nil {
			return err
		}

		return printResults(config, query.UnarchiveProjectItem.ProjectV2Item)
	}
	query, variables := archiveItemArgs(config, config.opts.itemID)
	err = config.client.Mutate("ArchiveProjectItem", query, variables)
	if err != nil {
		return err
	}

	return printResults(config, query.ArchiveProjectItem.ProjectV2Item)
}

func archiveItemArgs(config archiveItemConfig, itemID string) (*archiveProjectItemMutation, map[string]interface{}) {
	return &archiveProjectItemMutation{}, map[string]interface{}{
		"input": githubv4.ArchiveProjectV2ItemInput{
			ProjectID: githubv4.ID(config.opts.projectID),
			ItemID:    githubv4.ID(itemID),
		},
	}
}

func unarchiveItemArgs(config archiveItemConfig, itemID string) (*unarchiveProjectItemMutation, map[string]interface{}) {
	return &unarchiveProjectItemMutation{}, map[string]interface{}{
		"input": githubv4.UnarchiveProjectV2ItemInput{
			ProjectID: githubv4.ID(config.opts.projectID),
			ItemID:    githubv4.ID(itemID),
		},
	}
}

func printResults(config archiveItemConfig, item queries.ProjectItem) error {
	// using table printer here for consistency in case it ends up being needed in the future
	if config.opts.undo {
		config.tp.AddField("Unarchived item")
	} else {
		config.tp.AddField("Archived item")
	}
	config.tp.EndRow()
	return config.tp.Render()
}
