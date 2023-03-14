package format

import (
	"encoding/json"

	"github.com/github/gh-projects/queries"
)

// JSONProject serializes a Project to JSON.
func JSONProject(project queries.Project) ([]byte, error) {
	type t struct {
		Number           int    `json:"number"`
		URL              string `json:"url"`
		ShortDescription string `json:"shortDescription"`
		Public           bool   `json:"public"`
		Closed           bool   `json:"closed"`
		Title            string `json:"title"`
		ID               string `json:"id"`
		Readme           string `json:"readme"`
		Items            struct {
			TotalCount int `json:"totalCount"`
		} `graphql:"items(first: 100)" json:"items"`
		Fields struct {
			TotalCount int `json:"totalCount"`
		} `graphql:"fields(first:100)" json:"fields"`
		Owner struct {
			Type  string `json:"type"`
			Login string `json:"login"`
		}
	}
	return json.Marshal(t{
		Number:           project.Number,
		URL:              project.URL,
		ShortDescription: project.ShortDescription,
		Public:           project.Public,
		Closed:           project.Closed,
		Title:            project.Title,
		ID:               project.ID,
		Readme:           project.Readme,
		Items: struct {
			TotalCount int `json:"totalCount"`
		}{
			TotalCount: project.Items.TotalCount,
		},
		Fields: struct {
			TotalCount int `json:"totalCount"`
		}{
			TotalCount: project.Fields.TotalCount,
		},
		Owner: struct {
			Type  string `json:"type"`
			Login string `json:"login"`
		}{
			Type:  project.OwnerType(),
			Login: project.OwnerLogin(),
		},
	})
}

// JSONProjects serializes a slice of Projects to JSON.
func JSONProjects(projects []queries.Project) ([]byte, error) {
	return json.Marshal(projects)
}

// JSONProjectField serializes a ProjectField to JSON.
func JSONProjectField(field queries.ProjectField) ([]byte, error) {
	type t struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	}

	return json.Marshal(t{
		ID:   field.ID(),
		Name: field.Name(),
		Type: field.Type(),
	})
}

// JSONProjectFields serializes a slice of ProjectFields to JSON.
func JSONProjectFields(fields []queries.ProjectField) ([]byte, error) {
	type t struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	}

	var result []t
	for _, f := range fields {
		result = append(result, t{
			ID:   f.ID(),
			Name: f.Name(),
			Type: f.Type(),
		})
	}

	return json.Marshal(result)
}

// JSONProjectItem serializes a ProjectItem to JSON.
func JSONProjectItem(item queries.ProjectItem) ([]byte, error) {
	type t struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Body  string `json:"body"`
	}

	return json.Marshal(t{
		ID:    item.ID(),
		Title: item.Title(),
		Body:  item.Body(),
	})
}

// JSONProjectDraftIssue serializes a DraftIssue to JSON.
// This is needed because the field for
// https://docs.github.com/en/graphql/reference/mutations#updateprojectv2draftissue
// is a DraftIssue https://docs.github.com/en/graphql/reference/objects#draftissue
// and not a ProjectV2Item https://docs.github.com/en/graphql/reference/objects#projectv2item
func JSONProjectDraftIssue(item queries.DraftIssue) ([]byte, error) {
	type t struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Body  string `json:"body"`
	}

	return json.Marshal(t{
		ID:    item.ID,
		Title: item.Title,
		Body:  item.Body,
	})
}
