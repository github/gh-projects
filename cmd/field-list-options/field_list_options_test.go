package fieldlistvalues

import (
	"bytes"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/stretchr/testify/assert"
	"testing"

	"gopkg.in/h2non/gock.v1"
)

func TestRunList_User_(t *testing.T) {
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
									"configuration": map[string]interface{}{
										"iterations": []map[string]interface{}{
											{
												"startDate": "2022-05-29",
												"id":        "iteration ID 2",
												"title":     "current",
												"duration":  1,
											},
										},
										"completedIterations": []map[string]interface{}{
											{
												"startDate": "2022-05-29",
												"id":        "iteration ID 1",
												"title":     "past",
												"duration":  1,
											},
										},
									},
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
			itemID:    "iteration ID",
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"ID\tTitle\tStart Date\tDuration\tCompleted\niteration ID 1\tpast\t2022-05-29\t1\ttrue\niteration ID 2\tcurrent\t2022-05-29\t1\n",
		buf.String())
}

func TestRunList_Org(t *testing.T) {
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
									"configuration": map[string]interface{}{
										"iterations": []map[string]interface{}{
											{
												"startDate": "2022-05-29",
												"id":        "iteration ID 2",
												"title":     "current",
												"duration":  1,
											},
										},
										"completedIterations": []map[string]interface{}{
											{
												"startDate": "2022-05-29",
												"id":        "iteration ID 1",
												"title":     "past",
												"duration":  1,
											},
										},
									},
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
			itemID:   "iteration ID",
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"ID\tTitle\tStart Date\tDuration\tCompleted\niteration ID 1\tpast\t2022-05-29\t1\ttrue\niteration ID 2\tcurrent\t2022-05-29\t1\n",
		buf.String())
}

func TestRunList_Me(t *testing.T) {
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
									"configuration": map[string]interface{}{
										"iterations": []map[string]interface{}{
											{
												"startDate": "2022-05-29",
												"id":        "iteration ID 2",
												"title":     "current",
												"duration":  1,
											},
										},
										"completedIterations": []map[string]interface{}{
											{
												"startDate": "2022-05-29",
												"id":        "iteration ID 1",
												"title":     "past",
												"duration":  1,
											},
										},
									},
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
			userOwner: "@me",
			itemID:    "iteration ID",
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"ID\tTitle\tStart Date\tDuration\tCompleted\niteration ID 1\tpast\t2022-05-29\t1\ttrue\niteration ID 2\tcurrent\t2022-05-29\t1\n",
		buf.String())
}

func TestRunList_Empty(t *testing.T) {
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
			number:    1,
			userOwner: "@me",
			itemID:    "itemID",
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Project 1 for login @me has no fields with given ID.\n",
		buf.String())
}
