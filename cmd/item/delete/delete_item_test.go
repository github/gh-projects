package delete

import (
	"bytes"
	"testing"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestRunDelete_User(t *testing.T) {
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

	// delete item
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation DeleteProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":"item ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"deleteProjectV2Item": map[string]interface{}{
					"deletedItemId": "item ID",
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := deleteItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: deleteItemOpts{
			userOwner: "monalisa",
			number:    1,
			itemID:    "item ID",
		},
		client: client,
	}

	err = runDeleteItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Deleted item\n",
		buf.String())
}

func TestRunDelete_Org(t *testing.T) {
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

	// delete item
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation DeleteProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":"item ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"deleteProjectV2Item": map[string]interface{}{
					"deletedItemId": "item ID",
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := deleteItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: deleteItemOpts{
			orgOwner: "github",
			number:   1,
			itemID:   "item ID",
		},
		client: client,
	}

	err = runDeleteItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Deleted item\n",
		buf.String())
}

func TestRunDelete_Me(t *testing.T) {
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

	// delete item
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation DeleteProjectItem.*","variables":{"input":{"projectId":"an ID","itemId":"item ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"deleteProjectV2Item": map[string]interface{}{
					"deletedItemId": "item ID",
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := deleteItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: deleteItemOpts{
			userOwner: "@me",
			number:    1,
			itemID:    "item ID",
		},
		client: client,
	}

	err = runDeleteItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Deleted item\n",
		buf.String())
}

func TestRunDeleteItem_NoOrgOrUserSpecified(t *testing.T) {
	buf := bytes.Buffer{}
	config := deleteItemConfig{
		tp:   tableprinter.New(&buf, false, 0),
		opts: deleteItemOpts{},
	}

	err := runDeleteItem(config)
	assert.EqualError(t, err, "one of --user or --org is required")
}
