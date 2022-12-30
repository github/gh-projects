package cmd

import (
	"bytes"
	"testing"

	gh "github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/shurcooL/graphql"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestBuildQueryViewer(t *testing.T) {
	query, variables := buildQuery(listConfig{
		opts: listOpts{
			// login is empty
			// first is empty
		},
	})
	assert.Equal(t, &viewerQuery{}, query)
	assert.Equal(t, graphql.Int(100), variables["first"])
	assert.Empty(t, variables["login"])
}

func TestBuildQueryOwner(t *testing.T) {
	query, variables := buildQuery(listConfig{
		opts: listOpts{
			userOwner: true,
			login:     "monalisa",
			// first is empty
		},
	})
	assert.Equal(t, &userQuery{}, query)
	assert.Equal(t, graphql.Int(100), variables["first"])
	assert.Equal(t, graphql.String("monalisa"), variables["login"])
}

func TestBuildQueryOrganization(t *testing.T) {
	query, variables := buildQuery(listConfig{
		opts: listOpts{
			orgOwner: true,
			login:    "github",
			// first is empty
		},
	})
	assert.Equal(t, &organizationQuery{}, query)
	assert.Equal(t, graphql.Int(100), variables["first"])
	assert.Equal(t, graphql.String("github"), variables["login"])
}

func TestBuildURLViewer(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/graphql").
		Reply(200).
		JSON(`
			{"data":
				{"viewer":
					{
						"login":"theviewer"
					}
				}
			}
		`)

	client, _ := gh.GQLClient(nil)

	url, err := buildURL(listConfig{
		opts: listOpts{
			// login is empty
		},
		client: client,
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/users/theviewer/projects", url)
}

func TestBuildURLUser(t *testing.T) {
	url, err := buildURL(listConfig{
		opts: listOpts{
			userOwner: true,
			login:     "monalisa",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/users/monalisa/projects", url)
}

func TestBuildURLOrg(t *testing.T) {
	url, err := buildURL(listConfig{
		opts: listOpts{
			orgOwner: true,
			login:    "github",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/orgs/github/projects", url)
}

func TestBuildURLWithClosed(t *testing.T) {
	url, err := buildURL(listConfig{
		opts: listOpts{
			orgOwner: true,
			login:    "github",
			closed:   true,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/orgs/github/projects?query=is%3Aclosed", url)
}

func TestPrintNoResults(t *testing.T) {
	projects := []projectNode{}
	buf := bytes.Buffer{}
	config := listConfig{
		out: &buf,
	}

	printResults(config, projects, "monalisa")
	assert.Equal(t, "No projects found for monalisa\n", buf.String())
}
func TestPrintResults(t *testing.T) {
	projects := []projectNode{
		{
			Title:            "Project 1",
			ShortDescription: "Short description 1",
			URL:              "url",
			Closed:           false,
		},
	}
	buf := bytes.Buffer{}
	config := listConfig{
		tp: tableprinter.New(&buf, false, 0),
	}

	printResults(config, projects, "monalisa")
	assert.Equal(t, "Title\tDescription\tURL\nProject 1\tShort description 1\turl\n", buf.String())
}

func TestPrintResultsClosed(t *testing.T) {
	projects := []projectNode{
		{
			Title:            "Project 1",
			ShortDescription: "Short description 1",
			URL:              "url1",
			Closed:           false,
		},
		{
			Title:            "Project 2",
			ShortDescription: "",
			URL:              "url2",
			Closed:           true,
		},
	}
	buf := bytes.Buffer{}
	config := listConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: listOpts{
			closed: true,
		},
	}

	printResults(config, projects, "monalisa")
	assert.Equal(
		t,
		"Title\tDescription\tURL\tState\nProject 1\tShort description 1\turl1\topen\nProject 2\t - \turl2\tclosed\n",
		buf.String())
}

func TestRunList(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/graphql").
		Reply(200).
		JSON(`
			{"data":
				{"user":
					{
						"login":"monalisa",
						"projectsV2": {
							"nodes": [
								{"title": "Project 1", "shortDescription": "Short description 1", "url": "url", "closed": false}
							]
						}
					}
				}
			}
		`)

	client, _ := gh.GQLClient(nil)

	buf := bytes.Buffer{}
	config := listConfig{
		tp:  tableprinter.New(&buf, false, 0),
		out: &buf,
		opts: listOpts{
			login:     "monalisa",
			userOwner: true,
		},
		client: client,
	}

	runList(config)
	assert.Equal(t, "Title\tDescription\tURL\nProject 1\tShort description 1\turl\n", buf.String())
}

func TestRunListWeb(t *testing.T) {
	buf := bytes.Buffer{}
	config := listConfig{
		opts: listOpts{
			login:     "monalisa",
			userOwner: true,
			web:       true,
		},
		URLOpener: func(url string) error {
			buf.WriteString(url)
			return nil
		},
	}

	runList(config)
	assert.Equal(t, "https://github.com/users/monalisa/projects", buf.String())
}
