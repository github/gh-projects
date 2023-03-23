package edit

import (
	"bytes"
	"testing"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestRunUpdate_User(t *testing.T) {
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

	// get user project ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query UserProject.*",
			"variables": map[string]interface{}{
				"login":       "monalisa",
				"number":      1,
				"firstItems":  100,
				"afterItems":  nil,
				"firstFields": 100,
				"afterFields": nil,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]string{
						"id": "an ID",
					},
				},
			},
		})

	// edit project
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation UpdateProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":100,"firstItems":100,"input":{"projectId":"an ID","title":"a new title","shortDescription":"a new description","readme":"a new readme","public":true}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": "a title",
						"url":   "http://a-url.com",
						"owner": map[string]string{
							"login": "monalisa",
						},
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := editConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: editOpts{
			number:           1,
			userOwner:        "monalisa",
			title:            "a new title",
			shortDescription: "a new description",
			visibility:       "PUBLIC",
			readme:           "a new readme",
		},
		client: client,
	}

	err = runEdit(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Updated project http://a-url.com\n",
		buf.String())
}

func TestRunUpdate_Org(t *testing.T) {
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

	// get org project ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query OrgProject.*",
			"variables": map[string]interface{}{
				"login":       "github",
				"number":      1,
				"firstItems":  100,
				"afterItems":  nil,
				"firstFields": 100,
				"afterFields": nil,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"organization": map[string]interface{}{
					"projectV2": map[string]string{
						"id": "an ID",
					},
				},
			},
		})

	// edit project
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation UpdateProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":100,"firstItems":100,"input":{"projectId":"an ID","title":"a new title","shortDescription":"a new description","readme":"a new readme","public":true}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": "a title",
						"url":   "http://a-url.com",
						"owner": map[string]string{
							"login": "monalisa",
						},
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := editConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: editOpts{
			number:           1,
			orgOwner:         "github",
			title:            "a new title",
			shortDescription: "a new description",
			visibility:       "PUBLIC",
			readme:           "a new readme",
		},
		client: client,
	}

	err = runEdit(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Updated project http://a-url.com\n",
		buf.String())
}

func TestRunUpdate_Me(t *testing.T) {
	defer gock.Off()
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

	gock.Observe(gock.DumpRequest)
	// get viewer project ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query ViewerProject.*",
			"variables": map[string]interface{}{
				"number":      1,
				"firstItems":  100,
				"afterItems":  nil,
				"firstFields": 100,
				"afterFields": nil,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"projectV2": map[string]string{
						"id": "an ID",
					},
				},
			},
		})

	// edit project
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation UpdateProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":100,"firstItems":100,"input":{"projectId":"an ID","title":"a new title","shortDescription":"a new description","readme":"a new readme","public":false}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": "a title",
						"url":   "http://a-url.com",
						"owner": map[string]string{
							"login": "me",
						},
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := editConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: editOpts{
			number:           1,
			userOwner:        "@me",
			title:            "a new title",
			shortDescription: "a new description",
			visibility:       "PRIVATE",
			readme:           "a new readme",
		},
		client: client,
	}

	err = runEdit(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Updated project http://a-url.com\n",
		buf.String())
}

func TestRunUpdate_OmitParams(t *testing.T) {
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

	// get user project ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query UserProject.*",
			"variables": map[string]interface{}{
				"login":       "monalisa",
				"number":      1,
				"firstItems":  100,
				"afterItems":  nil,
				"firstFields": 100,
				"afterFields": nil,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]string{
						"id": "an ID",
					},
				},
			},
		})

	// Update project
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation UpdateProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":100,"firstItems":100,"input":{"projectId":"an ID","title":"another title"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": "a title",
						"url":   "http://a-url.com",
						"owner": map[string]string{
							"login": "monalisa",
						},
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := editConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: editOpts{
			number:    1,
			userOwner: "monalisa",
			title:     "another title",
		},
		client: client,
	}

	err = runEdit(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Updated project http://a-url.com\n",
		buf.String())
}

func TestRunUpdate_EmptyUpdateParams(t *testing.T) {
	buf := bytes.Buffer{}
	config := editConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: editOpts{
			number:    1,
			userOwner: "monalisa",
		},
	}

	err := runEdit(config)
	assert.Error(t, err, "no fields to edit")
}
