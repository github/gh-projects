package list

import (
	"bytes"
	"testing"

	gh "github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/github/gh-projects/queries"
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
	assert.Equal(t, &queries.ProjectsViewerQuery{}, query)
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
	assert.Equal(t, &queries.ProjectsUserQuery{}, query)
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
	assert.Equal(t, &queries.ProjectsOrganizationQuery{}, query)
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

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

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
								{"title": "Project 1", "shortDescription": "Short description 1", "url": "url1", "closed": false},
								{"title": "Project 2", "shortDescription": "", "url": "url2", "closed": true}
							]
						}
					}
				}
			}
		`)

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := listConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: listOpts{
			login:     "monalisa",
			userOwner: true,
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Title\tDescription\tURL\nProject 1\tShort description 1\turl1\n",
		buf.String())
}

func TestRunListViewer(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/graphql").
		// BodyString(`{"query":"query ProjectsQuery($first:Int!){viewer{projectsV2(first: $first){nodes{title,number,url,shortDescription,closed}},login}}","variables":{"first":100}}`).
		Reply(200).
		JSON(`
			{"data":
				{"viewer":
					{
						"login":"monalisa",
						"projectsV2": {
							"nodes": [
								{"title": "Project 1", "shortDescription": "Short description 1", "url": "url1", "closed": false},
								{"title": "Project 2", "shortDescription": "", "url": "url2", "closed": true}
							]
						}
					}
				}
			}
		`)

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := listConfig{
		tp:     tableprinter.New(&buf, false, 0),
		opts:   listOpts{},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Title\tDescription\tURL\nProject 1\tShort description 1\turl1\n",
		buf.String())
}

func TestRunListOrg(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/graphql").
		Reply(200).
		JSON(`
			{"data":
				{"organization":
					{
						"login":"monalisa",
						"projectsV2": {
							"nodes": [
								{"title": "Project 1", "shortDescription": "Short description 1", "url": "url1", "closed": false},
								{"title": "Project 2", "shortDescription": "", "url": "url2", "closed": true}
							]
						}
					}
				}
			}
		`)

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := listConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: listOpts{
			login:    "github",
			orgOwner: true,
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Title\tDescription\tURL\nProject 1\tShort description 1\turl1\n",
		buf.String())
}

func TestRunListEmpty(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/graphql").
		Reply(200).
		JSON(`
			{"data":
				{"viewer":
					{
						"login":"monalisa",
						"projectsV2": {
							"nodes": []
						}
					}
				}
			}
		`)

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := listConfig{
		tp:     tableprinter.New(&buf, false, 0),
		opts:   listOpts{},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"No projects found for monalisa\n",
		buf.String())
}

func TestRunListWithClosed(t *testing.T) {
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
								{"title": "Project 1", "shortDescription": "Short description 1", "url": "url1", "closed": false},
								{"title": "Project 2", "shortDescription": "", "url": "url2", "closed": true}
							]
						}
					}
				}
			}
		`)

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := listConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: listOpts{
			login:     "monalisa",
			userOwner: true,
			closed:    true,
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Title\tDescription\tURL\tState\nProject 1\tShort description 1\turl1\topen\nProject 2\t - \turl2\tclosed\n",
		buf.String())
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

	err := runList(config)
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/users/monalisa/projects", buf.String())
}

func TestRunListErrorOnlyLogin(t *testing.T) {
	config := listConfig{
		opts: listOpts{
			login: "monalisa",
		},
	}

	err := runList(config)
	assert.Error(t, err, "one of --user or --org is required with --login")
}

func TestRunListErrorUserAndOrg(t *testing.T) {
	config := listConfig{
		opts: listOpts{
			login:     "monalisa",
			userOwner: true,
			orgOwner:  true,
		},
	}

	err := runList(config)
	assert.Error(t, err, "only one of --user or --org can be set")
}
