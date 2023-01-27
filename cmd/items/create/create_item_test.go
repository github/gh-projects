package create

import (
	"bytes"
	"testing"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestRunCreateItem_Draft_User(t *testing.T) {
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

	// create item
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation CreateDraftItem.*","variables":{"input":{"projectId":"an ID","title":"a title","body":""}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"addProjectV2DraftIssue": map[string]interface{}{
					"projectItem": map[string]interface{}{
						"id": "item ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := createItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createItemOpts{
			title:     "a title",
			userOwner: "monalisa",
			number:    1,
			draft:     true,
		},
		client: client,
	}

	err = runCreateItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created item\n",
		buf.String())
}

func TestRunCreateItem_Draft_Org(t *testing.T) {
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

	// create item
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation CreateDraftItem.*","variables":{"input":{"projectId":"an ID","title":"a title","body":""}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"addProjectV2DraftIssue": map[string]interface{}{
					"projectItem": map[string]interface{}{
						"id": "item ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := createItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createItemOpts{
			title:    "a title",
			orgOwner: "github",
			number:   1,
			draft:    true,
		},
		client: client,
	}

	err = runCreateItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created item\n",
		buf.String())
}

func TestRunCreateItem_Draft_Me(t *testing.T) {
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

	// create item
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation CreateDraftItem.*","variables":{"input":{"projectId":"an ID","title":"a title","body":"a body"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"addProjectV2DraftIssue": map[string]interface{}{
					"projectItem": map[string]interface{}{
						"id": "item ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := createItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createItemOpts{
			title:  "a title",
			viewer: true,
			number: 1,
			draft:  true,
			body:   "a body",
		},
		client: client,
	}

	err = runCreateItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created item\n",
		buf.String())
}

func TestRunCreateItem_User(t *testing.T) {
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
		BodyString(`{"query":"mutation CreateDraftItem.*","variables":{"input":{"projectId":"an ID","contentId":"item ID"}}}`).
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
	config := createItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createItemOpts{
			userOwner: "monalisa",
			number:    1,
			itemURL:   "https://github.com/cli/go-gh/issues/1",
		},
		client: client,
	}

	err = runCreateItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created item\n",
		buf.String())
}

func TestRunCreateItem_Org(t *testing.T) {
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
		BodyString(`{"query":"mutation CreateDraftItem.*","variables":{"input":{"projectId":"an ID","contentId":"item ID"}}}`).
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
	config := createItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createItemOpts{
			orgOwner: "github",
			number:   1,
			itemURL:  "https://github.com/cli/go-gh/issues/1",
		},
		client: client,
	}

	err = runCreateItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created item\n",
		buf.String())
}

func TestRunCreateItem_Me(t *testing.T) {
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
		BodyString(`{"query":"mutation CreateDraftItem.*","variables":{"input":{"projectId":"an ID","contentId":"item ID"}}}`).
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
	config := createItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createItemOpts{
			viewer:  true,
			number:  1,
			itemURL: "https://github.com/cli/go-gh/pull/1",
		},
		client: client,
	}

	err = runCreateItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created item\n",
		buf.String())
}

func TestRunCreateItem_NoOrgOrUserSpecified(t *testing.T) {
	buf := bytes.Buffer{}
	config := createItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createItemOpts{
			title: "a title",
		},
	}

	err := runCreateItem(config)
	assert.EqualError(t, err, "one of --user, --org or --me is required")
}

func TestRunCreateItem_NoDraftOrURL(t *testing.T) {
	buf := bytes.Buffer{}
	config := createItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createItemOpts{
			viewer: true,
		},
	}

	err := runCreateItem(config)
	assert.EqualError(t, err, "one of --url or --draft is required")
}

func TestRunCreateItem_Draft_NoTitle(t *testing.T) {
	buf := bytes.Buffer{}
	config := createItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createItemOpts{
			viewer: true,
			draft:  true,
		},
	}

	err := runCreateItem(config)
	assert.EqualError(t, err, "--title is required with draft issues")
}
