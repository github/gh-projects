package add

import (
	"bytes"
	"testing"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestRunAddItem_User(t *testing.T) {
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

	// get item ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query GetIssueOrPullRequest.*",
			"variables": map[string]interface{}{
				"url": "https://github.com/cli/go-gh/issues/1",
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"resource": map[string]interface{}{
					"id":         "item ID",
					"__typename": "Issue",
				},
			},
		})

	// create item
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation AddItem.*","variables":{"input":{"projectId":"an ID","contentId":"item ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"addProjectV2ItemById": map[string]interface{}{
					"item": map[string]interface{}{
						"id": "item ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := addItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: addItemOpts{
			userOwner: "monalisa",
			number:    1,
			itemURL:   "https://github.com/cli/go-gh/issues/1",
		},
		client: client,
	}

	err = runAddItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Added item\n",
		buf.String())
}

func TestRunAddItem_Org(t *testing.T) {
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

	// get item ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query GetIssueOrPullRequest.*",
			"variables": map[string]interface{}{
				"url": "https://github.com/cli/go-gh/issues/1",
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"resource": map[string]interface{}{
					"id":         "item ID",
					"__typename": "Issue",
				},
			},
		})

	// create item
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation AddItem.*","variables":{"input":{"projectId":"an ID","contentId":"item ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"addProjectV2ItemById": map[string]interface{}{
					"item": map[string]interface{}{
						"id": "item ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := addItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: addItemOpts{
			orgOwner: "github",
			number:   1,
			itemURL:  "https://github.com/cli/go-gh/issues/1",
		},
		client: client,
	}

	err = runAddItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Added item\n",
		buf.String())
}

func TestRunAddItem_Me(t *testing.T) {
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

	// get item ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query GetIssueOrPullRequest.*",
			"variables": map[string]interface{}{
				"url": "https://github.com/cli/go-gh/pull/1",
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"resource": map[string]interface{}{
					"id":         "item ID",
					"__typename": "PullRequest",
				},
			},
		})

	// create item
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation AddItem.*","variables":{"input":{"projectId":"an ID","contentId":"item ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"addProjectV2ItemById": map[string]interface{}{
					"item": map[string]interface{}{
						"id": "item ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := addItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: addItemOpts{
			viewer:  true,
			number:  1,
			itemURL: "https://github.com/cli/go-gh/pull/1",
		},
		client: client,
	}

	err = runAddItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Added item\n",
		buf.String())
}

func TestRunAddItem_NoOrgOrUserSpecified(t *testing.T) {
	buf := bytes.Buffer{}
	config := addItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: addItemOpts{
			itemURL: "a URL",
		},
	}

	err := runAddItem(config)
	assert.EqualError(t, err, "one of --user, --org or --me is required")
}
