package close

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/github/gh-projects/queries"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestRunClose_User(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get user project ID
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
					"projectV2": map[string]string{
						"id": "an ID",
					},
				},
			},
		})

	// close project
	gock.New("https://api.github.com").
		Post("/graphql").
		// this is the same as the below JSON, but for some reason gock doesn't match on the graphql boolean
		// TODO: would love to figure out why
		BodyString(`{"query":"mutation CloseProjectV2.*"variables":{"input":{"projectId":"an ID","closed":true}}}`).
		// JSON(map[string]interface{}{
		// 	"query": "mutation CloseProjectV2.*",
		// 	"variables": map[string]interface{}{
		// 		"input": map[string]interface{}{
		// 			"projectId": "an ID",
		// 			"closed":    true,
		// 		},
		// 	},
		// }).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": "a title",
						"url":   "http://a-url.com",
						"owner": map[string]string{
							"login": "monalisa",
						},
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := closeConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: closeOpts{
			number:    1,
			userOwner: "monalisa",
		},
		client: client,
	}

	err = runClose(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Closed project http://a-url.com\n",
		buf.String())
}

func TestRunClose_Org(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get org project ID
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
					"projectV2": map[string]string{
						"id": "an ID",
					},
				},
			},
		})

	// close project
	gock.New("https://api.github.com").
		Post("/graphql").
		// this is the same as the below JSON, but for some reason gock doesn't match on the graphql boolean
		// TODO: would love to figure out why
		BodyString(`{"query":"mutation CloseProjectV2.*"variables":{"input":{"projectId":"an ID","closed":true}}}`).
		// JSON(map[string]interface{}{
		// 	"query": "mutation CloseProjectV2.*",
		// 	"variables": map[string]interface{}{
		// 		"input": map[string]interface{}{
		// 			"projectId": "an ID",
		// 			"closed":  true,
		// 		},
		// 	},
		// }).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": "a title",
						"url":   "http://a-url.com",
						"owner": map[string]string{
							"login": "monalisa",
						},
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := closeConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: closeOpts{
			number:   1,
			orgOwner: "github",
		},
		client: client,
	}

	err = runClose(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Closed project http://a-url.com\n",
		buf.String())
}

func TestRunClose_Me(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// get viewer project ID
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
					"projectV2": map[string]string{
						"id": "an ID",
					},
				},
			},
		})

	// close project
	gock.New("https://api.github.com").
		Post("/graphql").
		// this is the same as the below JSON, but for some reason gock doesn't match on the graphql boolean
		// TODO: would love to figure out why
		BodyString(`{"query":"mutation CloseProjectV2.*"variables":{"input":{"projectId":"an ID","closed":true}}}`).
		// JSON(map[string]interface{}{
		// 	"query": "mutation CloseProjectV2.*",
		// 	"variables": map[string]interface{}{
		// 		"input": map[string]interface{}{
		// 			"projectId": "an ID",
		// 			"closed":    true,
		// 		},
		// 	},
		// }).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": "a title",
						"url":   "http://a-url.com",
						"owner": map[string]string{
							"login": "me",
						},
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := closeConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: closeOpts{
			number: 1,
			viewer: true,
		},
		client: client,
	}

	err = runClose(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Closed project http://a-url.com\n",
		buf.String())
}

func TestRunClose_Reopen(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// get user project ID
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
					"projectV2": map[string]string{
						"id": "an ID",
					},
				},
			},
		})

	// close project
	gock.New("https://api.github.com").
		Post("/graphql").
		// this is the same as the below JSON, but for some reason gock doesn't match on the graphql boolean
		// TODO: would love to figure out why
		BodyString(`{"query":"mutation CloseProjectV2.*"variables":{"input":{"projectId":"an ID","closed":false}}}`).
		// JSON(map[string]interface{}{
		// 	"query": "mutation CloseProjectV2.*",
		// 	"variables": map[string]interface{}{
		// 		"input": map[string]interface{}{
		// 			"projectId": "an ID",
		// 			"closed":    false,
		// 		},
		// 	},
		// }).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"title": "a title",
						"url":   "http://a-url.com",
						"owner": map[string]string{
							"login": "monalisa",
						},
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := closeConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: closeOpts{
			number:    1,
			userOwner: "monalisa",
			reopen:    true,
		},
		client: client,
	}

	err = runClose(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Reopened project http://a-url.com\n",
		buf.String())
}

func TestRunClose_NoOrgOrUserSpecified(t *testing.T) {
	buf := bytes.Buffer{}
	config := closeConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: closeOpts{
			number: 1,
		},
	}

	err := runClose(config)
	assert.EqualError(t, err, "one of --user, --org or --me is required")

}

func TestNewCmdClose(t *testing.T) {
	type args struct {
		f    *cmdutil.Factory
		runF func(config closeConfig) error
	}
	tests := []struct {
		name string
		args args
		want *cobra.Command
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCmdClose(tt.args.f, tt.args.runF); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCmdClose() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_runClose(t *testing.T) {
	type args struct {
		config closeConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := runClose(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("runClose() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_buildCloseQuery(t *testing.T) {
	type args struct {
		config closeConfig
	}
	tests := []struct {
		name  string
		args  args
		want  *queries.UpdateProjectMutation
		want1 map[string]interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := buildCloseQuery(tt.args.config)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildCloseQuery() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("buildCloseQuery() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_printResults(t *testing.T) {
	type args struct {
		config  closeConfig
		project queries.ProjectV2
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := printResults(tt.args.config, tt.args.project); (err != nil) != tt.wantErr {
				t.Errorf("printResults() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
