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

func TestRunCreate_User(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

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

	// create project
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation CreateProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"ownerId":"an ID","title":"a title"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"createProjectV2": map[string]interface{}{
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
	config := createConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createOpts{
			title:     "a title",
			userOwner: "monalisa",
		},
		client: client,
	}

	err = runCreate(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created project 'a title'\nhttp://a-url.com\n",
		buf.String())
}

func TestRunCreate_Org(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
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

	// create project
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation CreateProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"ownerId":"an ID","title":"a title"}}}`).Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"createProjectV2": map[string]interface{}{
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
	config := createConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createOpts{
			title:    "a title",
			orgOwner: "github",
		},
		client: client,
	}

	err = runCreate(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created project 'a title'\nhttp://a-url.com\n",
		buf.String())
}

func TestRunCreate_Me(t *testing.T) {
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
					"id":    "an ID",
					"login": "me",
				},
			},
		})

	// create project
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation CreateProjectV2.*"variables":{"afterFields":null,"afterItems":null,"firstFields":0,"firstItems":0,"input":{"ownerId":"an ID","title":"a title"}}}`).Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"createProjectV2": map[string]interface{}{
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
	config := createConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createOpts{
			title:     "a title",
			userOwner: "@me",
		},
		client: client,
	}

	err = runCreate(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created project 'a title'\nhttp://a-url.com\n",
		buf.String())
}
