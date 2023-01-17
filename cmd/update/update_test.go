package update

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

	// get user project ID
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
					"projectV2": map[string]string{
						"id": "an ID",
					},
				},
			},
		})

	// update project
	gock.New("https://api.github.com").
		Post("/graphql").
		// this is the same as the below JSON, but for some reason gock doesn't match on the graphql boolean
		// TODO: would love to figure out why
		BodyString(`{"query":"mutation UpdateProjectV2.*"variables":{"input":{"projectId":"an ID","title":"a new title","shortDescription":"a new description","readme":"a new readme","public":true}}}`).
		// JSON(map[string]interface{}{
		// 	"query": "mutation UpdateProjectV2.*",
		// 	"variables": map[string]interface{}{
		// 		"input": map[string]interface{}{
		// 			"projectId": "an ID",
		// 			"title":    "a new title",
		// 			"shortDescription":    "a new description",
		// 			"public":    true,
		// 			"readme":    "a new readme",
		// 		},
		// 	},
		// }).
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
	public := true
	buf := bytes.Buffer{}
	config := updateConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: updateOpts{
			number:           1,
			userOwner:        "monalisa",
			title:            "a new title",
			shortDescription: "a new description",
			public:           &public,
			readme:           "a new readme",
		},
		client: client,
	}

	err = runUpdate(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Updated project http://a-url.com\n",
		buf.String())
}

func TestRunUpdate_Org(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get org project ID
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
					"projectV2": map[string]string{
						"id": "an ID",
					},
				},
			},
		})

	// update project
	gock.New("https://api.github.com").
		Post("/graphql").
		// this is the same as the below JSON, but for some reason gock doesn't match on the graphql boolean
		// TODO: would love to figure out why
		BodyString(`{"query":"mutation UpdateProjectV2.*"variables":{"input":{"projectId":"an ID","title":"a new title","shortDescription":"a new description","readme":"a new readme","public":true}}}`).
		// JSON(map[string]interface{}{
		// 	"query": "mutation UpdateProjectV2.*",
		// 	"variables": map[string]interface{}{
		// 		"input": map[string]interface{}{
		// 			"projectId": "an ID",
		// 			"title":    "a new title",
		// 			"shortDescription":    "a new description",
		// 			"public":    true,
		// 			"readme":    "a new readme",
		// 		},
		// 	},
		// }).
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
	public := true
	buf := bytes.Buffer{}
	config := updateConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: updateOpts{
			number:           1,
			orgOwner:         "github",
			title:            "a new title",
			shortDescription: "a new description",
			public:           &public,
			readme:           "a new readme",
		},
		client: client,
	}

	err = runUpdate(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Updated project http://a-url.com\n",
		buf.String())
}

func TestRunUpdate_Me(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get viewer project ID
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
					"projectV2": map[string]string{
						"id": "an ID",
					},
				},
			},
		})

	// update project
	gock.New("https://api.github.com").
		Post("/graphql").
		// this is the same as the below JSON, but for some reason gock doesn't match on the graphql boolean
		// TODO: would love to figure out why
		BodyString(`{"query":"mutation UpdateProjectV2.*"variables":{"input":{"projectId":"an ID","title":"a new title","shortDescription":"a new description","readme":"a new readme","public":false}}}`).
		// JSON(map[string]interface{}{
		// 	"query": "mutation UpdateProjectV2.*",
		// 	"variables": map[string]interface{}{
		// 		"input": map[string]interface{}{
		// 			"projectId": "an ID",
		// 			"title":    "a new title",
		// 			"shortDescription":    "a new description",
		// 			"public":    false,
		// 			"readme":    "a new readme",
		// 		},
		// 	},
		// }).
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
	public := false
	buf := bytes.Buffer{}
	config := updateConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: updateOpts{
			number:           1,
			viewer:           true,
			title:            "a new title",
			shortDescription: "a new description",
			public:           &public,
			readme:           "a new readme",
		},
		client: client,
	}

	err = runUpdate(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Updated project http://a-url.com\n",
		buf.String())
}

func TestRunUpdate_NoParams(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get user project ID
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
					"projectV2": map[string]string{
						"id": "an ID",
					},
				},
			},
		})

	// Update project
	gock.New("https://api.github.com").
		Post("/graphql").
		// this is the same as the below JSON, but for some reason gock doesn't match on the graphql boolean
		// TODO: would love to figure out why
		BodyString(`{"query":"mutation UpdateProjectV2.*"variables":{"input":{"projectId":"an ID"}}}`).
		// JSON(map[string]interface{}{
		// 	"query": "mutation UpdateProjectV2.*",
		// 	"variables": map[string]interface{}{
		// 		"input": map[string]interface{}{
		// 			"projectId": "an ID",
		// 		},
		// 	},
		// }).
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
	config := updateConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: updateOpts{
			number:    1,
			userOwner: "monalisa",
		},
		client: client,
	}

	err = runUpdate(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Updated project http://a-url.com\n",
		buf.String())
}

func TestRunUpdate_NoOrgOrUserSpecified(t *testing.T) {
	buf := bytes.Buffer{}
	config := updateConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: updateOpts{
			number: 1,
		},
	}

	err := runUpdate(config)
	assert.EqualError(t, err, "one of --user, --org or --me is required")

}
