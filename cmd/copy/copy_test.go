package copy

import (
	"bytes"
	"testing"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestRunCopy_User(t *testing.T) {
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

	// get user ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query UserLogin.*",
			"variables": map[string]string{
				"login": "monalisa",
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"id":    "an ID",
					"login": "monalisa",
				},
			},
		})

	// Copy project
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation CopyProjectV2.*","variables":{"input":{"ownerId":"an ID","projectId":"an ID","title":"a title","includeDraftIssues":false}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"copyProjectV2": map[string]interface{}{
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
	config := copyConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: copyOpts{
			title:           "a title",
			sourceUserOwner: "monalisa",
			targetUserOwner: "monalisa",
			number:          1,
		},
		client: client,
	}

	err = runCopy(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created project copy 'a title'\nhttp://a-url.com\n",
		buf.String())
}

func TestRunCopy_Org(t *testing.T) {
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
	// get org ID
	gock.New("https://api.github.com").
		Post("/graphql").
		MatchType("json").
		JSON(map[string]interface{}{
			"query": "query OrgLogin.*",
			"variables": map[string]string{
				"login": "github",
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"organization": map[string]interface{}{
					"id":    "an ID",
					"login": "github",
				},
			},
		})

	// Copy project
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation CopyProjectV2.*","variables":{"input":{"ownerId":"an ID","projectId":"an ID","title":"a title","includeDraftIssues":false}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"copyProjectV2": map[string]interface{}{
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
	config := copyConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: copyOpts{
			title:          "a title",
			sourceOrgOwner: "github",
			targetOrgOwner: "github",
			number:         1,
		},
		client: client,
	}

	err = runCopy(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created project copy 'a title'\nhttp://a-url.com\n",
		buf.String())
}

func TestRunCopy_Me(t *testing.T) {
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
					"id":    "an ID",
					"login": "me",
				},
			},
		})

	// Copy project
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation CopyProjectV2.*","variables":{"input":{"ownerId":"an ID","projectId":"an ID","title":"a title","includeDraftIssues":false}}}`).Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"copyProjectV2": map[string]interface{}{
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
	config := copyConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: copyOpts{
			title:        "a title",
			sourceViewer: true,
			targetViewer: true,
			number:       1,
		},
		client: client,
	}

	err = runCopy(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created project copy 'a title'\nhttp://a-url.com\n",
		buf.String())
}

func TestRunCopy_NoSourceOrgOrUserSpecified(t *testing.T) {
	buf := bytes.Buffer{}
	config := copyConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: copyOpts{
			title: "a title",
		},
	}

	err := runCopy(config)
	assert.EqualError(t, err, "one of --source-user, --source-org or --source-me is required")
}

func TestRunCopy_NoTargetOrgOrUserSpecified(t *testing.T) {
	buf := bytes.Buffer{}
	config := copyConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: copyOpts{
			title:        "a title",
			sourceViewer: true,
		},
	}

	err := runCopy(config)
	assert.EqualError(t, err, "one of --target-user, --target-org or --target-me is required")
}
