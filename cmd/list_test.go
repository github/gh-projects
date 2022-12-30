package cmd

import (
	"bytes"
	"testing"

	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/shurcooL/graphql"
	"github.com/stretchr/testify/assert"
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

type client struct {
}

func (c client) Query(name string, i interface{}, variables map[string]interface{}) error {
	if name == "Viewer" {
		queryViewer.Viewer.Login = "theviewer"
	}
	return nil
}

func TestBuildURLViewer(t *testing.T) {
	url, err := buildURL(listConfig{
		opts: listOpts{
			// login is empty
		},
		client: client{},
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
		client: client{},
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
		client: client{},
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
		client: client{},
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
