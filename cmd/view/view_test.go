package view

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

	url, err := buildURL(viewConfig{
		opts: viewOpts{
			number:    1,
			userOwner: "@me",
		},
		client: client,
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/users/theviewer/projects/1", url)
}

func TestBuildURLUser(t *testing.T) {
	url, err := buildURL(viewConfig{
		opts: viewOpts{
			userOwner: "monalisa",
			number:    1,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/users/monalisa/projects/1", url)
}

func TestBuildURLOrg(t *testing.T) {
	url, err := buildURL(viewConfig{
		opts: viewOpts{
			orgOwner: "github",
			number:   1,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/orgs/github/projects/1", url)
}

func TestRunView_User(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/graphql").
		Reply(200).
		JSON(`
			{"data":
				{"user":
					{
						"login":"monalisa",
						"projectV2": {
							"number": 1,
							"items": {
								"totalCount": 10
							},
							"readme": null,
							"fields": {
								"nodes": [
									{
										"name": "Title"
									}
								]
							}
						}
					}
				}
			}
		`)

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := viewConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: viewOpts{
			userOwner: "monalisa",
			number:    1,
		},
		client: client,
	}

	err = runView(config)
	assert.NoError(t, err)

}

func TestRunView_Viewer(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/graphql").
		Reply(200).
		JSON(`
			{"data":
				{"viewer":
					{
						"login":"monalisa",
						"projectV2": {
							"number": 1,
							"items": {
								"totalCount": 10
							},
							"readme": null,
							"fields": {
								"nodes": [
									{
										"name": "Title"
									}
								]
							}
						}
					}
				}
			}
		`)

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := viewConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: viewOpts{
			userOwner: "@me",
			number:    1,
		},
		client: client,
	}

	err = runView(config)
	assert.NoError(t, err)
}

func TestRunView_Org(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/graphql").
		Reply(200).
		JSON(`
			{"data":
				{"organization":
					{
						"login":"monalisa",
						"projectV2": {
							"number": 1,
							"items": {
								"totalCount": 10
							},
							"readme": null,
							"fields": {
								"nodes": [
									{
										"name": "Title"
									}
								]
							}
						}
					}
				}
			}
		`)

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := viewConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: viewOpts{
			orgOwner: "github",
			number:   1,
		},
		client: client,
	}

	err = runView(config)
	assert.NoError(t, err)
}

func TestRunViewWeb(t *testing.T) {
	buf := bytes.Buffer{}
	config := viewConfig{
		opts: viewOpts{
			userOwner: "monalisa",
			web:       true,
			number:    8,
		},
		URLOpener: func(url string) error {
			buf.WriteString(url)
			return nil
		},
	}

	err := runView(config)
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/users/monalisa/projects/8", buf.String())
}
