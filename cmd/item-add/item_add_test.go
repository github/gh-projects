package itemadd

import (
	"bytes"
	"testing"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestRunAddItem_User(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get user ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query UserLogin.*",
			"variables": map[string]interface{}{
				"login": "monalisa",
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"id": "an ID",
				},
			},
		})

	// get project ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query UserProject.*",
			"variables": map[string]interface{}{
				"login":       "monalisa",
				"number":      1,
				"firstItems":  0,
				"afterItems":  nil,
				"firstFields": 0,
				"afterFields": nil,
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

	client, err := api.NewGraphQLClient(api.ClientOptions{AuthToken: "token"})
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
	// get org ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query OrgLogin.*",
			"variables": map[string]interface{}{
				"login": "github",
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"organization": map[string]interface{}{
					"id": "an ID",
				},
			},
		})

	// get project ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query OrgProject.*",
			"variables": map[string]interface{}{
				"login":       "github",
				"number":      1,
				"firstItems":  0,
				"afterItems":  nil,
				"firstFields": 0,
				"afterFields": nil,
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

	client, err := api.NewGraphQLClient(api.ClientOptions{AuthToken: "token"})
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
	// get viewer ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query ViewerLogin.*",
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"id": "an ID",
				},
			},
		})

	// get project ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query ViewerProject.*",
			"variables": map[string]interface{}{
				"number":      1,
				"firstItems":  0,
				"afterItems":  nil,
				"firstFields": 0,
				"afterFields": nil,
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
				"url": "https://github.com/cli/go-gh/v2/pull/1",
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

	client, err := api.NewGraphQLClient(api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := addItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: addItemOpts{
			userOwner: "@me",
			number:    1,
			itemURL:   "https://github.com/cli/go-gh/v2/pull/1",
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
