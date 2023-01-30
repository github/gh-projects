package edit

import (
	"bytes"
	"testing"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestRunItemEdit(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	// edit item
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation EditDraftIssueItem.*","variables":{"input":{"draftIssueId":"DI_item_id","title":"a title","body":"a new body"}}}`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2DraftIssue": map[string]interface{}{
					"draftIssue": map[string]interface{}{
						"title": "a title",
						"body":  "a new body",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := editItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: editItemOpts{
			title:  "a title",
			body:   "a new body",
			itemID: "DI_item_id",
		},
		client: client,
	}

	err = runEditItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Title\tBody\na title\ta new body\n",
		buf.String())
}

func TestRunItemEdit_NoChanges(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := editItemConfig{
		tp:     tableprinter.New(&buf, false, 0),
		opts:   editItemOpts{},
		client: client,
	}

	err = runEditItem(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"No changes to make",
		buf.String())
}

func TestRunItemEdit_InavlidID(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := editItemConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: editItemOpts{
			title:  "a title",
			body:   "a new body",
			itemID: "item_id",
		},
		client: client,
	}

	err = runEditItem(config)
	assert.Error(t, err, "ID must be the ID of the draft issue content which is prefixed with `DI_`")
}
