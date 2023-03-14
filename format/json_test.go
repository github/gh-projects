package format

import (
	"testing"

	"github.com/github/gh-projects/queries"

	"github.com/stretchr/testify/assert"
)

func TestJSONProject_User(t *testing.T) {
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

	assert.Equal(t, `{"number":2,"url":"a url","shortDescription":"short description","public":true,"closed":false,"title":"","id":"123","readme":"readme","items":{"totalCount":1},"fields":{"totalCount":2},"Owner":{"type":"User","login":"monalisa"}}`, string(b))
}

func TestJSONProject_Org(t *testing.T) {
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
	project.Owner.TypeName = "Organization"
	project.Owner.Organization.Login = "github"
	b, err := JSONProject(project)
	assert.NoError(t, err)

	assert.Equal(t, `{"number":2,"url":"a url","shortDescription":"short description","public":true,"closed":false,"title":"","id":"123","readme":"readme","items":{"totalCount":1},"fields":{"totalCount":2},"Owner":{"type":"Organization","login":"github"}}`, string(b))
}

func TestJSONProjects(t *testing.T) {
	userProject := queries.Project{
		ID:               "123",
		Number:           2,
		URL:              "a url",
		ShortDescription: "short description",
		Public:           true,
		Readme:           "readme",
	}

	userProject.Items.TotalCount = 1
	userProject.Fields.TotalCount = 2
	userProject.Owner.TypeName = "User"
	userProject.Owner.User.Login = "monalisa"

	orgProject := queries.Project{
		ID:               "123",
		Number:           2,
		URL:              "a url",
		ShortDescription: "short description",
		Public:           true,
		Readme:           "readme",
	}

	orgProject.Items.TotalCount = 1
	orgProject.Fields.TotalCount = 2
	orgProject.Owner.TypeName = "Organization"
	orgProject.Owner.Organization.Login = "github"
	b, err := JSONProjects([]queries.Project{userProject, orgProject})
	assert.NoError(t, err)

	assert.Equal(t, `[{"number":2,"url":"a url","shortDescription":"short description","public":true,"closed":false,"title":"","id":"123","readme":"readme","items":{"totalCount":1},"fields":{"totalCount":2},"Owner":{"type":"User","login":"monalisa"}},{"number":2,"url":"a url","shortDescription":"short description","public":true,"closed":false,"title":"","id":"123","readme":"readme","items":{"totalCount":1},"fields":{"totalCount":2},"Owner":{"type":"Organization","login":"github"}}]`, string(b))
}
