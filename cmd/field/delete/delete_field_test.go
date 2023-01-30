package delete

import (
	"bytes"
	"testing"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestRunDeleteField(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// delete Field
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation DeleteField.*","variables":{"input":{"fieldId":"an ID"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"deleteProjectV2Field": map[string]interface{}{
					"projectV2Field": map[string]interface{}{
						"id": "Field ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := deleteFieldConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: deleteFieldOpts{
			fieldID: "an ID",
		},
		client: client,
	}

	err = runDeleteField(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Deleted field\n",
		buf.String())
}
