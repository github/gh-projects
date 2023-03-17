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

	assert.Equal(t, `{"number":2,"url":"a url","shortDescription":"short description","public":true,"closed":false,"title":"","id":"123","readme":"readme","items":{"totalCount":1},"fields":{"totalCount":2},"owner":{"type":"User","login":"monalisa"}}`, string(b))
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

	assert.Equal(t, `{"number":2,"url":"a url","shortDescription":"short description","public":true,"closed":false,"title":"","id":"123","readme":"readme","items":{"totalCount":1},"fields":{"totalCount":2},"owner":{"type":"Organization","login":"github"}}`, string(b))
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
	b, err := JSONProjects([]queries.Project{userProject, orgProject}, 2)
	assert.NoError(t, err)

	assert.Equal(
		t,
		`{"projects":[{"number":2,"url":"a url","shortDescription":"short description","public":true,"closed":false,"title":"","id":"123","readme":"readme","items":{"totalCount":1},"fields":{"totalCount":2},"owner":{"type":"User","login":"monalisa"}},{"number":2,"url":"a url","shortDescription":"short description","public":true,"closed":false,"title":"","id":"123","readme":"readme","items":{"totalCount":1},"fields":{"totalCount":2},"owner":{"type":"Organization","login":"github"}}],"totalCount":2}`,
		string(b))
}

func TestJSONProjectField_FieldType(t *testing.T) {
	field := queries.ProjectField{}
	field.TypeName = "ProjectV2Field"
	field.Field.ID = "123"
	field.Field.Name = "name"

	b, err := JSONProjectField(field)
	assert.NoError(t, err)

	assert.Equal(t, `{"id":"123","name":"name","type":"ProjectV2Field"}`, string(b))
}

func TestJSONProjectField_SingleSelectType(t *testing.T) {
	field := queries.ProjectField{}
	field.TypeName = "ProjectV2SingleSelectField"
	field.SingleSelectField.ID = "123"
	field.SingleSelectField.Name = "name"

	b, err := JSONProjectField(field)
	assert.NoError(t, err)

	assert.Equal(t, `{"id":"123","name":"name","type":"ProjectV2SingleSelectField"}`, string(b))
}

func TestJSONProjectField_ProjectV2IterationField(t *testing.T) {
	field := queries.ProjectField{}
	field.TypeName = "ProjectV2IterationField"
	field.IterationField.ID = "123"
	field.IterationField.Name = "name"

	b, err := JSONProjectField(field)
	assert.NoError(t, err)

	assert.Equal(t, `{"id":"123","name":"name","type":"ProjectV2IterationField"}`, string(b))
}

func TestJSONProjectFields(t *testing.T) {
	field := queries.ProjectField{}
	field.TypeName = "ProjectV2Field"
	field.Field.ID = "123"
	field.Field.Name = "name"

	p := queries.ProjectWithFields{
		Fields: struct {
			PageInfo   queries.PageInfo
			Nodes      []queries.ProjectField
			TotalCount int
		}{
			Nodes:      []queries.ProjectField{field},
			TotalCount: 5,
		},
	}
	b, err := JSONProjectFields(p)
	assert.NoError(t, err)

	assert.Equal(t, `{"fields":[{"id":"123","name":"name","type":"ProjectV2Field"}],"totalCount":5}`, string(b))
}

func TestJSONProjectItem_DraftIssue(t *testing.T) {
	item := queries.ProjectItem{}
	item.Content.TypeName = "DraftIssue"
	item.Id = "123"
	item.Content.DraftIssue.Title = "title"
	item.Content.DraftIssue.Body = "a body"

	b, err := JSONProjectItem(item)
	assert.NoError(t, err)

	assert.Equal(t, `{"id":"123","title":"title","body":"a body","type":"DraftIssue"}`, string(b))
}

func TestJSONProjectItem_Issue(t *testing.T) {
	item := queries.ProjectItem{}
	item.Content.TypeName = "Issue"
	item.Id = "123"
	item.Content.Issue.Title = "title"
	item.Content.Issue.Body = "a body"

	b, err := JSONProjectItem(item)
	assert.NoError(t, err)

	assert.Equal(t, `{"id":"123","title":"title","body":"a body","type":"Issue"}`, string(b))
}

func TestJSONProjectDetailedItems(t *testing.T) {
	p := queries.ProjectWithItems{}
	p.Items.TotalCount = 5
	p.Items.Nodes = []queries.ProjectItem{
		{
			Id: "issueId",
			Content: queries.ProjectItemContent{
				TypeName: "Issue",
				Issue: queries.Issue{
					Title:  "Issue title",
					Body:   "a body",
					Number: 1,
					Repository: struct {
						NameWithOwner string
					}{
						NameWithOwner: "cli/go-gh",
					},
				},
			},
		},
		{
			Id: "pullRequestId",
			Content: queries.ProjectItemContent{
				TypeName: "PullRequest",
				PullRequest: queries.PullRequest{
					Title:  "Pull Request title",
					Body:   "a body",
					Number: 2,
					Repository: struct {
						NameWithOwner string
					}{
						NameWithOwner: "cli/go-gh",
					},
				},
			},
		},
		{
			Id: "draftIssueId",
			Content: queries.ProjectItemContent{
				TypeName: "DraftIssue",
				DraftIssue: queries.DraftIssue{
					Title: "Pull Request title",
					Body:  "a body",
				},
			},
		},
	}

	out, err := JSONProjectDetailedItems(p)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"{\"items\":[{\"content\":{\"type\":\"Issue\",\"body\":\"a body\",\"title\":\"Issue title\",\"number\":1,\"repository\":\"cli/go-gh\"},\"id\":\"issueId\"},{\"content\":{\"type\":\"PullRequest\",\"body\":\"a body\",\"title\":\"Pull Request title\",\"number\":2,\"repository\":\"cli/go-gh\"},\"id\":\"pullRequestId\"},{\"content\":{\"type\":\"DraftIssue\",\"body\":\"a body\",\"title\":\"Pull Request title\"},\"id\":\"draftIssueId\"}],\"totalCount\":5}",
		string(out))
}

func TestJSONProjectItem_PullRequest(t *testing.T) {
	item := queries.ProjectItem{}
	item.Content.TypeName = "PullRequest"
	item.Id = "123"
	item.Content.PullRequest.Title = "title"
	item.Content.PullRequest.Body = "a body"

	b, err := JSONProjectItem(item)
	assert.NoError(t, err)

	assert.Equal(t, `{"id":"123","title":"title","body":"a body","type":"PullRequest"}`, string(b))
}

func TestJSONProjectDraftIssue(t *testing.T) {
	item := queries.DraftIssue{}
	item.ID = "123"
	item.Title = "title"
	item.Body = "a body"

	b, err := JSONProjectDraftIssue(item)
	assert.NoError(t, err)

	assert.Equal(t, `{"id":"123","title":"title","body":"a body","type":"DraftIssue"}`, string(b))
}

func TestCamelCase(t *testing.T) {
	assert.Equal(t, "camelCase", CamelCase("camelCase"))
	assert.Equal(t, "camelCase", CamelCase("CamelCase"))
	assert.Equal(t, "c", CamelCase("C"))
	assert.Equal(t, "", CamelCase(""))
}
