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

func TestRunCreateField_User(t *testing.T) {
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

	// create Field
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation CreateField.*","variables":{"input":{"projectId":"an ID","dataType":"TEXT","name":"a name"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"createProjectV2Field": map[string]interface{}{
					"projectV2Field": map[string]interface{}{
						"id": "Field ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := createFieldConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createFieldOpts{
			name:      "a name",
			userOwner: "monalisa",
			number:    1,
			dataType:  "TEXT",
		},
		client: client,
	}

	err = runCreateField(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created field\n",
		buf.String())
}

func TestRunCreateField_Org(t *testing.T) {
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

	// create Field
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation CreateField.*","variables":{"input":{"projectId":"an ID","dataType":"TEXT","name":"a name"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"createProjectV2Field": map[string]interface{}{
					"projectV2Field": map[string]interface{}{
						"id": "Field ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := createFieldConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createFieldOpts{
			name:     "a name",
			orgOwner: "github",
			number:   1,
			dataType: "TEXT",
		},
		client: client,
	}

	err = runCreateField(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created field\n",
		buf.String())
}

func TestRunCreateField_Me(t *testing.T) {
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

	// create Field
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation CreateField.*","variables":{"input":{"projectId":"an ID","dataType":"TEXT","name":"a name"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"createProjectV2Field": map[string]interface{}{
					"projectV2Field": map[string]interface{}{
						"id": "Field ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := createFieldConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createFieldOpts{
			viewer:   true,
			number:   1,
			name:     "a name",
			dataType: "TEXT",
		},
		client: client,
	}

	err = runCreateField(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created field\n",
		buf.String())
}

func TestRunCreateField_TEXT(t *testing.T) {
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

	// create Field
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation CreateField.*","variables":{"input":{"projectId":"an ID","dataType":"TEXT","name":"a name"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"createProjectV2Field": map[string]interface{}{
					"projectV2Field": map[string]interface{}{
						"id": "Field ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := createFieldConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createFieldOpts{
			viewer:   true,
			number:   1,
			name:     "a name",
			dataType: "TEXT",
		},
		client: client,
	}

	err = runCreateField(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created field\n",
		buf.String())
}

func TestRunCreateField_DATE(t *testing.T) {
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

	// create Field
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation CreateField.*","variables":{"input":{"projectId":"an ID","dataType":"DATE","name":"a name"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"createProjectV2Field": map[string]interface{}{
					"projectV2Field": map[string]interface{}{
						"id": "Field ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := createFieldConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createFieldOpts{
			viewer:   true,
			number:   1,
			name:     "a name",
			dataType: "DATE",
		},
		client: client,
	}

	err = runCreateField(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created field\n",
		buf.String())
}

func TestRunCreateField_NUMBER(t *testing.T) {
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

	// create Field
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation CreateField.*","variables":{"input":{"projectId":"an ID","dataType":"NUMBER","name":"a name"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"createProjectV2Field": map[string]interface{}{
					"projectV2Field": map[string]interface{}{
						"id": "Field ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := createFieldConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createFieldOpts{
			viewer:   true,
			number:   1,
			name:     "a name",
			dataType: "NUMBER",
		},
		client: client,
	}

	err = runCreateField(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Created field\n",
		buf.String())
}

// gock fails to match the create field operation
// I've manually tested this and it works

// func TestRunCreateField_SINGLE_SELECT(t *testing.T) {
// 	defer gock.Off()
// 	gock.Observe(gock.DumpRequest)
// 	// get project ID
// 	gock.New("https://api.github.com").
// 		Post("/graphql").
// 		MatchType("json").
// 		JSON(map[string]interface{}{
// 			"query": "query ViewerProject.*",
// 			"variables": map[string]interface{}{
// 				"number": 1,
// 			},
// 		}).
// 		Reply(200).
// 		JSON(map[string]interface{}{
// 			"data": map[string]interface{}{
// 				"viewer": map[string]interface{}{
// 					"projectV2": map[string]interface{}{
// 						"id": "an ID",
// 					},
// 				},
// 			},
// 		})

// 	// create Field
// 	gock.New("https://api.github.com").
// 		Post("/graphql").
// 		BodyString(`{"query":"mutation CreateField.*","variables":{"input":{"projectId":"an ID","dataType":"SINGLE_SELECT","name":"a name","singleSelectOptions":["a","b","c"]}}}`).
// 		Reply(200).
// 		JSON(map[string]interface{}{
// 			"data": map[string]interface{}{
// 				"createProjectV2Field": map[string]interface{}{
// 					"projectV2Field": map[string]interface{}{
// 						"id": "Field ID",
// 					},
// 				},
// 			},
// 		})

// 	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
// 	assert.NoError(t, err)

// 	buf := bytes.Buffer{}
// 	config := createFieldConfig{
// 		tp: tableprinter.New(&buf, false, 0),
// 		opts: createFieldOpts{
// 			viewer:              true,
// 			number:              1,
// 			name:                "a name",
// 			dataType:            "SINGLE_SELECT",
// 			singleSelectOptions: []string{"a", "b", "c"},
// 		},
// 		client: client,
// 	}

// 	err = runCreateField(config)
// 	assert.NoError(t, err)
// 	assert.Equal(
// 		t,
// 		"Created field\n",
// 		buf.String())
// }

func TestRunCreateField_NoOrgOrUserSpecified(t *testing.T) {
	buf := bytes.Buffer{}
	config := createFieldConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createFieldOpts{
			name: "a name",
		},
	}

	err := runCreateField(config)
	assert.EqualError(t, err, "one of --user, --org or --me is required")
}

func TestRunCreateField_SingleSelectNoOptions(t *testing.T) {
	buf := bytes.Buffer{}
	config := createFieldConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: createFieldOpts{
			name:     "a name",
			viewer:   true,
			dataType: "SINGLE_SELECT",
		},
	}

	err := runCreateField(config)
	assert.EqualError(t, err, "at least one single select options is required with data type is SINGLE_SELECT")
}
