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

	// list project items
	gock.New("https://api.github.com").
		Post("/graphql").
		JSON(map[string]interface{}{
			"query": "query UserProjectWithItems.*",
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
						"items": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"type": "ISSUE",
									"content": map[string]interface{}{
										"title":  "an issue",
										"body":   "a body",
										"number": 1,
										"repository": map[string]string{
											"nameWithOwner": "cli/go-gh",
										},
									},
								},
								{
									"type": "PULL_REQUEST",
									"content": map[string]interface{}{
										"title":  "a pull request",
										"body":   "the body",
										"number": 2,
										"repository": map[string]string{
											"nameWithOwner": "cli/go-gh",
										},
									},
								},
								{
									"type": "DRAFT_ISSUE",
									"content": map[string]interface{}{
										"title": "draft issue",
										"body":  "",
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
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Type\tTitle\tBody\tNumber\tRepository\nISSUE\tan issue\ta body\t1\tcli/go-gh\nPULL_REQUEST\ta pull request\tthe body\t2\tcli/go-gh\nDRAFT_ISSUE\tdraft issue\t - \t - \t - \n",
		buf.String())
}

func TestRunList_Org(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// list project items
	gock.New("https://api.github.com").
		Post("/graphql").
		JSON(map[string]interface{}{
			"query": "query OrgProjectWithItems.*",
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
						"items": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"type": "ISSUE",
									"content": map[string]interface{}{
										"title":  "an issue",
										"body":   "a body",
										"number": 1,
										"repository": map[string]string{
											"nameWithOwner": "cli/go-gh",
										},
									},
								},
								{
									"type": "PULL_REQUEST",
									"content": map[string]interface{}{
										"title":  "a pull request",
										"body":   "the body",
										"number": 2,
										"repository": map[string]string{
											"nameWithOwner": "cli/go-gh",
										},
									},
								},
								{
									"type": "DRAFT_ISSUE",
									"content": map[string]interface{}{
										"title": "draft issue",
										"body":  "",
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
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Type\tTitle\tBody\tNumber\tRepository\nISSUE\tan issue\ta body\t1\tcli/go-gh\nPULL_REQUEST\ta pull request\tthe body\t2\tcli/go-gh\nDRAFT_ISSUE\tdraft issue\t - \t - \t - \n",
		buf.String())
}

func TestRunList_Me(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// list project items
	gock.New("https://api.github.com").
		Post("/graphql").
		JSON(map[string]interface{}{
			"query": "query ViewerProjectWithItems.*",
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
						"items": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"type": "ISSUE",
									"content": map[string]interface{}{
										"title":  "an issue",
										"body":   "a body",
										"number": 1,
										"repository": map[string]string{
											"nameWithOwner": "cli/go-gh",
										},
									},
								},
								{
									"type": "PULL_REQUEST",
									"content": map[string]interface{}{
										"title":  "a pull request",
										"body":   "the body",
										"number": 2,
										"repository": map[string]string{
											"nameWithOwner": "cli/go-gh",
										},
									},
								},
								{
									"type": "DRAFT_ISSUE",
									"content": map[string]interface{}{
										"title": "draft issue",
										"body":  "",
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
			number: 1,
			viewer: true,
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Type\tTitle\tBody\tNumber\tRepository\nISSUE\tan issue\ta body\t1\tcli/go-gh\nPULL_REQUEST\ta pull request\tthe body\t2\tcli/go-gh\nDRAFT_ISSUE\tdraft issue\t - \t - \t - \n",
		buf.String())
}

func TestRunList_Empty(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// list project items
	gock.New("https://api.github.com").
		Post("/graphql").
		JSON(map[string]interface{}{
			"query": "query ViewerProjectWithItems.*",
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
						"items": map[string]interface{}{
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
		"Project 1 for login me has no items\n",
		buf.String())
}
