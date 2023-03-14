package format

import (
	"encoding/json"

	"github.com/github/gh-projects/queries"
)

func JSONProject(project queries.Project) ([]byte, error) {
	return json.Marshal(project)
}

// JSONProjectField serializes a ProjectField to JSON.
func JSONProjectField(field queries.ProjectField) ([]byte, error) {
	type t struct {
		ID   string
		Name string
		Type string
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
		ID   string
		Name string
		Type string
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
