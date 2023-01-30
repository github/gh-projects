package list

import (
	"bytes"
	"testing"

	gh "github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

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
			viewer: true,
		},
		client: client,
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/users/theviewer/projects", url)
}

func TestBuildURLUser(t *testing.T) {
	url, err := buildURL(listConfig{
		opts: listOpts{
			userOwner: "monalisa",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/users/monalisa/projects", url)
}

func TestBuildURLOrg(t *testing.T) {
	url, err := buildURL(listConfig{
		opts: listOpts{
			orgOwner: "github",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/orgs/github/projects", url)
}

func TestBuildURLWithClosed(t *testing.T) {
	url, err := buildURL(listConfig{
		opts: listOpts{
			orgOwner: "github",
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
								{"title": "Project 1", "shortDescription": "Short description 1", "url": "url1", "closed": false, "ID": "1"},
								{"title": "Project 2", "shortDescription": "", "url": "url2", "closed": true, "ID": "2"}
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
			userOwner: "monalisa",
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Title\tDescription\tURL\tID\nProject 1\tShort description 1\turl1\t1\n",
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
								{"title": "Project 1", "shortDescription": "Short description 1", "url": "url1", "closed": false, "ID": "1"},
								{"title": "Project 2", "shortDescription": "", "url": "url2", "closed": true, "ID": "2"}
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
			viewer: true,
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Title\tDescription\tURL\tID\nProject 1\tShort description 1\turl1\t1\n",
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
								{"title": "Project 1", "shortDescription": "Short description 1", "url": "url1", "closed": false, "ID": "1"},
								{"title": "Project 2", "shortDescription": "", "url": "url2", "closed": true, "ID": "2"}
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
			orgOwner: "github",
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Title\tDescription\tURL\tID\nProject 1\tShort description 1\turl1\t1\n",
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
		tp: tableprinter.New(&buf, false, 0),
		opts: listOpts{
			viewer: true,
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"No projects found for me\n",
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
								{"title": "Project 1", "shortDescription": "Short description 1", "url": "url1", "closed": false, "ID": "1"},
								{"title": "Project 2", "shortDescription": "", "url": "url2", "closed": true, "ID": "2"}
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
			userOwner: "monalisa",
			closed:    true,
		},
		client: client,
	}

	err = runList(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Title\tDescription\tURL\tState\tID\nProject 1\tShort description 1\turl1\topen\t1\nProject 2\t - \turl2\tclosed\t2\n",
		buf.String())
}

func TestRunListWeb(t *testing.T) {
	buf := bytes.Buffer{}
	config := listConfig{
		opts: listOpts{
			userOwner: "monalisa",
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

func TestRunListErrorNoLogin(t *testing.T) {
	config := listConfig{
		opts: listOpts{},
	}

	err := runList(config)
	assert.Error(t, err, "one of --user, --org or --me is required")
}
