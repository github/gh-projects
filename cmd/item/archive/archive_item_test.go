package archive

import (
	"bytes"
	"testing"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestRunArchive_User(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get project ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query UserProject.*",
			"variables": map[string]interface{}{
				"login":  "monalisa",
				"number": 1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id": "an ID",
					},
				},
			},
		})

	// archive item
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation ArchiveProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":"item ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"archiveProjectV2Item": map[string]interface{}{
					"item": map[string]interface{}{
						"id": "item ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := archiveItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: archiveItemOpts{
			userOwner: "monalisa",
			number:    1,
			itemID:    "item ID",
		},
		client: client,
	}

	err = runArchiveItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Archived item\n",
		buf.String())
}

func TestRunArchive_Org(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get project ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query OrgProject.*",
			"variables": map[string]interface{}{
				"login":  "github",
				"number": 1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"organization": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id": "an ID",
					},
				},
			},
		})

	// archive item
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation ArchiveProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":"item ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"archiveProjectV2Item": map[string]interface{}{
					"item": map[string]interface{}{
						"id": "item ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := archiveItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: archiveItemOpts{
			orgOwner: "github",
			number:   1,
			itemID:   "item ID",
		},
		client: client,
	}

	err = runArchiveItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Archived item\n",
		buf.String())
}

func TestRunArchive_Me(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get project ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query ViewerProject.*",
			"variables": map[string]interface{}{
				"number": 1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id": "an ID",
					},
				},
			},
		})

	// archive item
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation ArchiveProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":"item ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"archiveProjectV2Item": map[string]interface{}{
					"item": map[string]interface{}{
						"id": "item ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := archiveItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: archiveItemOpts{
			viewer: true,
			number: 1,
			itemID: "item ID",
		},
		client: client,
	}

	err = runArchiveItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Archived item\n",
		buf.String())
}

func TestRunArchive_User_Undo(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get project ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query UserProject.*",
			"variables": map[string]interface{}{
				"login":  "monalisa",
				"number": 1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id": "an ID",
					},
				},
			},
		})

	// archive item
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation UnarchiveProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":"item ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"unarchiveProjectV2Item": map[string]interface{}{
					"item": map[string]interface{}{
						"id": "item ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := archiveItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: archiveItemOpts{
			userOwner: "monalisa",
			number:    1,
			itemID:    "item ID",
			undo:      true,
		},
		client: client,
	}

	err = runArchiveItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Unarchived item\n",
		buf.String())
}

func TestRunArchive_Org_Undo(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get project ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query OrgProject.*",
			"variables": map[string]interface{}{
				"login":  "github",
				"number": 1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"organization": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id": "an ID",
					},
				},
			},
		})

	// archive item
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation UnarchiveProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":"item ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"unarchiveProjectV2Item": map[string]interface{}{
					"item": map[string]interface{}{
						"id": "item ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := archiveItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: archiveItemOpts{
			orgOwner: "github",
			number:   1,
			itemID:   "item ID",
			undo:     true,
		},
		client: client,
	}

	err = runArchiveItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Unarchived item\n",
		buf.String())
}

func TestRunArchive_Me_Undo(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get project ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query ViewerProject.*",
			"variables": map[string]interface{}{
				"number": 1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id": "an ID",
					},
				},
			},
		})

	// archive item
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation UnarchiveProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":"item ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"unarchiveProjectV2Item": map[string]interface{}{
					"item": map[string]interface{}{
						"id": "item ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := archiveItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: archiveItemOpts{
			viewer: true,
			number: 1,
			itemID: "item ID",
			undo:   true,
		},
		client: client,
	}

	err = runArchiveItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Unarchived item\n",
		buf.String())
}

func TestRunArchiveItem_NoOrgOrUserSpecified(t *testing.T) {
	buf := bytes.Buffer{}
	config := archiveItemConfig{
		tp:   tableprinter.New(&buf, false, 0),
		opts: archiveItemOpts{},
	}

	err := runArchiveItem(config)
	assert.EqualError(t, err, "one of --user, --org or --me is required")
}
