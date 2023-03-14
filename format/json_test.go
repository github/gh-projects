package format

import (
	"fmt"
	"testing"

	"github.com/github/gh-projects/queries"

	"github.com/stretchr/testify/assert"
)

func TestJSONProject(t *testing.T) {
	project := queries.Project{
		ID:               "123",
		Number:           2,
		URL:              "a url",
		ShortDescription: "short description",
		Public:           true,
		Readme:           "readme",
	}

	project.Items.TotalCount = 1
	project.Fields.TotalCount = 2
	project.Owner.TypeName = "User"
	project.Owner.User.Login = "monalisa"
	b, err := JSONProject(project)
	assert.NoError(t, err)
	fmt.Println(string(b))
	assert.Equal(t, `{"number":2,"url":"a url","shortDescription":"short description","public":true,"closed":false,"title":"","id":"123","readme":"readme","items":{"totalCount":1},"fields":{"totalCount":2},"Owner":{"type":"User","login":"monalisa"}}`, string(b))
}
