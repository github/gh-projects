package itemlist

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

	// list project items
	gock.New("https://api.github.com").
		Post("/graphql").
		JSON(map[string]interface{}{
			"query": "query UserProjectWithItems.*",
			"variables": map[string]interface{}{
				"first":  100,
				"login":  "monalisa",
				"number": 1,
				"after":  nil,
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
									"id": "issue ID",
									"content": map[string]interface{}{
										"__typename": "Issue",
										"title":      "an issue",
										"number":     1,
										"repository": map[string]string{
											"nameWithOwner": "cli/go-gh",
										},
									},
								},
								{
									"id": "pull request ID",
									"content": map[string]interface{}{
										"__typename": "PullRequest",
										"title":      "a pull request",
										"number":     2,
										"repository": map[string]string{
											"nameWithOwner": "cli/go-gh",
										},
									},
								},
								{
									"id": "draft issue ID",
									"content": map[string]interface{}{
										"title":      "draft issue",
										"__typename": "DraftIssue",
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
		"Type\tTitle\tNumber\tRepository\tID\nIssue\tan issue\t1\tcli/go-gh\tissue ID\nPullRequest\ta pull request\t2\tcli/go-gh\tpull request ID\nDraftIssue\tdraft issue\t - \t - \tdraft issue ID\n",
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

	// list project items
	gock.New("https://api.github.com").
		Post("/graphql").
		JSON(map[string]interface{}{
			"query": "query OrgProjectWithItems.*",
			"variables": map[string]interface{}{
				"first":  100,
				"login":  "github",
				"number": 1,
				"after":  nil,
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
									"id": "issue ID",
									"content": map[string]interface{}{
										"__typename": "Issue",
										"title":      "an issue",
										"number":     1,
										"repository": map[string]string{
											"nameWithOwner": "cli/go-gh",
										},
									},
								},
								{
									"id": "pull request ID",
									"content": map[string]interface{}{
										"__typename": "PullRequest",
										"title":      "a pull request",
										"number":     2,
										"repository": map[string]string{
											"nameWithOwner": "cli/go-gh",
										},
									},
								},
								{
									"id": "draft issue ID",
									"content": map[string]interface{}{
										"title":      "draft issue",
										"__typename": "DraftIssue",
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
		"Type\tTitle\tNumber\tRepository\tID\nIssue\tan issue\t1\tcli/go-gh\tissue ID\nPullRequest\ta pull request\t2\tcli/go-gh\tpull request ID\nDraftIssue\tdraft issue\t - \t - \tdraft issue ID\n",
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

	// list project items
	gock.New("https://api.github.com").
		Post("/graphql").
		JSON(map[string]interface{}{
			"query": "query ViewerProjectWithItems.*",
			"variables": map[string]interface{}{
				"first":  100,
				"number": 1,
				"after":  nil,
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
									"id": "issue ID",
									"content": map[string]interface{}{
										"__typename": "Issue",
										"title":      "an issue",
										"number":     1,
										"repository": map[string]string{
											"nameWithOwner": "cli/go-gh",
										},
									},
								},
								{
									"id": "pull request ID",
									"content": map[string]interface{}{
										"__typename": "PullRequest",
										"title":      "a pull request",
										"number":     2,
										"repository": map[string]string{
											"nameWithOwner": "cli/go-gh",
										},
									},
								},
								{
									"id": "draft issue ID",
									"content": map[string]interface{}{
										"title":      "draft issue",
										"__typename": "DraftIssue",
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
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Type\tTitle\tNumber\tRepository\tID\nIssue\tan issue\t1\tcli/go-gh\tissue ID\nPullRequest\ta pull request\t2\tcli/go-gh\tpull request ID\nDraftIssue\tdraft issue\t - \t - \tdraft issue ID\n",
		buf.String())
}

func TestRunList_JSON(t *testing.T) {
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

	// list project items
	gock.New("https://api.github.com").
		Post("/graphql").
		JSON(map[string]interface{}{
			"query": "query ViewerProjectWithItems.*",
			"variables": map[string]interface{}{
				"first":  100,
				"number": 1,
				"after":  nil,
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
									"id": "issue ID",
									"content": map[string]interface{}{
										"__typename": "Issue",
										"title":      "an issue",
										"body":       "an issue body",
										"number":     1,
										"repository": map[string]string{
											"nameWithOwner": "cli/go-gh",
										},
									},
								},
								{
									"id": "pull request ID",
									"content": map[string]interface{}{
										"__typename": "PullRequest",
										"title":      "a pull request",
										"body":       "a pull request body",
										"number":     2,
										"repository": map[string]string{
											"nameWithOwner": "cli/go-gh",
										},
									},
								},
								{
									"id": "draft issue ID",
									"content": map[string]interface{}{
										"__typename": "DraftIssue",
										"title":      "draft issue",
										"body":       "draft issue body",
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
			format:    "json",
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"[{\"content\":{\"type\":\"Issue\",\"body\":\"an issue body\",\"title\":\"an issue\",\"number\":1,\"repository\":\"cli/go-gh\"},\"id\":\"issue ID\"},{\"content\":{\"type\":\"PullRequest\",\"body\":\"a pull request body\",\"title\":\"a pull request\",\"number\":2,\"repository\":\"cli/go-gh\"},\"id\":\"pull request ID\"},{\"content\":{\"type\":\"DraftIssue\",\"body\":\"draft issue body\",\"title\":\"draft issue\"},\"id\":\"draft issue ID\"}]\n",
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

	// list project items
	gock.New("https://api.github.com").
		Post("/graphql").
		JSON(map[string]interface{}{
			"query": "query ViewerProjectWithItems.*",
			"variables": map[string]interface{}{
				"first":  100,
				"number": 1,
				"after":  nil,
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
			number:    1,
			userOwner: "@me",
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Project 1 for login @me has no items\n",
		buf.String())
}
