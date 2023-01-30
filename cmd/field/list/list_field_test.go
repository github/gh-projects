package list

import (
	"bytes"
	"testing"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestRunList_User(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// list project fields
	gock.New("https://api.github.com").
		Post("/graphql").
		JSON(map[string]interface{}{
			"query": "query UserProjectWithFields.*",
			"variables": map[string]interface{}{
				"first":  100,
				"login":  "monalisa",
				"number": 1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"fields": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"__typename": "ProjectV2Field",
									"name":       "FieldTitle",
									"id":         "field ID",
								},
								{
									"__typename": "ProjectV2SingleSelectField",
									"name":       "Status",
									"id":         "status ID",
								},
								{
									"__typename": "ProjectV2IterationField",
									"name":       "Iterations",
									"id":         "iteration ID",
								},
							},
						},
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := listConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: listOpts{
			number:    1,
			userOwner: "monalisa",
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Name\tDataType\tID\nFieldTitle\tProjectV2Field\tfield ID\nStatus\tProjectV2SingleSelectField\tstatus ID\nIterations\tProjectV2IterationField\titeration ID\n",
		buf.String())
}

func TestRunList_Org(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// list project fields
	gock.New("https://api.github.com").
		Post("/graphql").
		JSON(map[string]interface{}{
			"query": "query OrgProjectWithFields.*",
			"variables": map[string]interface{}{
				"first":  100,
				"login":  "github",
				"number": 1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"organization": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"fields": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"__typename": "ProjectV2Field",
									"name":       "FieldTitle",
									"id":         "field ID",
								},
								{
									"__typename": "ProjectV2SingleSelectField",
									"name":       "Status",
									"id":         "status ID",
								},
								{
									"__typename": "ProjectV2IterationField",
									"name":       "Iterations",
									"id":         "iteration ID",
								},
							},
						},
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := listConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: listOpts{
			number:   1,
			orgOwner: "github",
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Name\tDataType\tID\nFieldTitle\tProjectV2Field\tfield ID\nStatus\tProjectV2SingleSelectField\tstatus ID\nIterations\tProjectV2IterationField\titeration ID\n",
		buf.String())
}

func TestRunList_Me(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// list project fields
	gock.New("https://api.github.com").
		Post("/graphql").
		JSON(map[string]interface{}{
			"query": "query ViewerProjectWithFields.*",
			"variables": map[string]interface{}{
				"first":  100,
				"number": 1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"fields": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"__typename": "ProjectV2Field",
									"name":       "FieldTitle",
									"id":         "field ID",
								},
								{
									"__typename": "ProjectV2SingleSelectField",
									"name":       "Status",
									"id":         "status ID",
								},
								{
									"__typename": "ProjectV2IterationField",
									"name":       "Iterations",
									"id":         "iteration ID",
								},
							},
						},
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := listConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: listOpts{
			number: 1,
			viewer: true,
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Name\tDataType\tID\nFieldTitle\tProjectV2Field\tfield ID\nStatus\tProjectV2SingleSelectField\tstatus ID\nIterations\tProjectV2IterationField\titeration ID\n",
		buf.String())
}

func TestRunList_Empty(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// list project fields
	gock.New("https://api.github.com").
		Post("/graphql").
		JSON(map[string]interface{}{
			"query": "query ViewerProjectWithFields.*",
			"variables": map[string]interface{}{
				"first":  100,
				"number": 1,
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"fields": map[string]interface{}{
							"nodes": nil,
						},
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := listConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: listOpts{
			number: 1,
			viewer: true,
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Project 1 for login me has no fields\n",
		buf.String())
}
