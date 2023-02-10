package fieldedit

import (
	"bytes"
	"testing"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestRunEditField_Name(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)
	// edit Field
	gock.New("https://api.github.com").
		Post("/graphql").
		BodyString(`{"query":"mutation EditField.*","variables":{"input":{"fieldId":"an ID","name":"a name"`).
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"updateProjectV2Field": map[string]interface{}{
					"projectV2Field": map[string]interface{}{
						"id": "Field ID",
					},
				},
			},
		})

	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
	assert.NoError(t, err)

	buf := bytes.Buffer{}
	config := editFieldConfig{
		tp: tableprinter.New(&buf, false, 0),
		opts: editFieldOpts{
			name:    "a name",
			fieldID: "an ID",
		},
		client: client,
	}

	err = runEditField(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"Edited field\n",
		buf.String())
}

// gock does not recognize the singleSelectOptions as equal, even though they are
// uncomment this test and delete TestRunEditField_Name in the future

// func TestRunEditField(t *testing.T) {
// 	defer gock.Off()
// 	gock.Observe(gock.DumpRequest)
// 	// edit Field
// 	gock.New("https://api.github.com").
// 		Post("/graphql").
// 		BodyString(`{"query":"mutation EditField.*","variables":{"input":{"fieldId":"an ID","name":"a name","singleSelectOptions":["one","two","three"]}}}`).
// 		Reply(200).
// 		JSON(map[string]interface{}{
// 			"data": map[string]interface{}{
// 				"updateProjectV2Field": map[string]interface{}{
// 					"projectV2Field": map[string]interface{}{
// 						"id": "Field ID",
// 					},
// 				},
// 			},
// 		})

// 	client, err := gh.GQLClient(&api.ClientOptions{AuthToken: "token"})
// 	assert.NoError(t, err)

// 	buf := bytes.Buffer{}
// 	config := editFieldConfig{
// 		tp: tableprinter.New(&buf, false, 0),
// 		opts: editFieldOpts{
// 			name:                "a name",
// 			fieldID:             "an ID",
// 			singleSelectOptions: []string{"one", "two", "three"},
// 		},
// 		client: client,
// 	}

// 	err = runEditField(config)
// 	assert.NoError(t, err)
// 	assert.Equal(
// 		t,
// 		"Edited field\n",
// 		buf.String())
// }

func TestRunEditField_NoOptions(t *testing.T) {
	buf := bytes.Buffer{}
	config := editFieldConfig{
		tp:   tableprinter.New(&buf, false, 0),
		opts: editFieldOpts{},
	}

	err := runEditField(config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"No changes to make",
		buf.String())
}
