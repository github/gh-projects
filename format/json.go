package format

import (
	"encoding/json"
	"strings"

	"github.com/github/gh-projects/queries"
)

// JSONProject serializes a Project to JSON.
func JSONProject(project queries.Project) ([]byte, error) {
	return json.Marshal(projectJSON{
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
// JSON fields are `totalCount` and `projects`.
func JSONProjects(projects []queries.Project, totalCount int) ([]byte, error) {
	var result []projectJSON
	for _, p := range projects {
		result = append(result, projectJSON{
			Number:           p.Number,
			URL:              p.URL,
			ShortDescription: p.ShortDescription,
			Public:           p.Public,
			Closed:           p.Closed,
			Title:            p.Title,
			ID:               p.ID,
			Readme:           p.Readme,
			Items: struct {
				TotalCount int `json:"totalCount"`
			}{
				TotalCount: p.Items.TotalCount,
			},
			Fields: struct {
				TotalCount int `json:"totalCount"`
			}{
				TotalCount: p.Fields.TotalCount,
			},
			Owner: struct {
				Type  string `json:"type"`
				Login string `json:"login"`
			}{
				Type:  p.OwnerType(),
				Login: p.OwnerLogin(),
			},
		})
	}

	return json.Marshal(struct {
		Projects   []projectJSON `json:"projects"`
		TotalCount int           `json:"totalCount"`
	}{
		Projects:   result,
		TotalCount: totalCount,
	})
}

type projectJSON struct {
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
	} `json:"owner"`
}

// JSONProjectField serializes a ProjectField to JSON.
func JSONProjectField(field queries.ProjectField) ([]byte, error) {
	return json.Marshal(projectFieldJSON{
		ID:   field.ID(),
		Name: field.Name(),
		Type: field.Type(),
	})
}

// JSONProjectFields serializes a slice of ProjectFields to JSON.
// JSON fields are `totalCount` and `fields`.
func JSONProjectFields(project *queries.Project) ([]byte, error) {
	var result []projectFieldJSON
	for _, f := range project.Fields.Nodes {
		result = append(result, projectFieldJSON{
			ID:   f.ID(),
			Name: f.Name(),
			Type: f.Type(),
		})
	}

	return json.Marshal(struct {
		Fields     []projectFieldJSON `json:"fields"`
		TotalCount int                `json:"totalCount"`
	}{
		Fields:     result,
		TotalCount: project.Fields.TotalCount,
	})
}

type projectFieldJSON struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// JSONProjectFieldsOptions serializes a slice of ProjectFieldsWithOptions to JSON.
// JSON fields are `totalCount` and `fields`.
func JSONProjectFieldsIterableOptions(completedIterations []queries.Iteration, iterations []queries.Iteration) ([]byte, error) {
	var result []projectFieldIterableOptionsJSON
	for _, f := range completedIterations {
		result = append(result, projectFieldIterableOptionsJSON{
			ID:        f.Id,
			Title:     f.Title,
			Duration:  f.Duration,
			StartDate: f.StartDate,
			Completed: true,
		})
	}
	for _, f := range iterations {
		result = append(result, projectFieldIterableOptionsJSON{
			ID:        f.Id,
			Title:     f.Title,
			Duration:  f.Duration,
			StartDate: f.StartDate,
			Completed: false,
		})
	}

	return json.Marshal(struct {
		Options    []projectFieldIterableOptionsJSON `json:"options"`
		TotalCount int                               `json:"totalCount"`
	}{
		Options:    result,
		TotalCount: len(completedIterations) + len(iterations),
	})
}

type projectFieldIterableOptionsJSON struct {
	ID        string                `json:"id"`
	Title     string                `json:"title"`
	Duration  int                   `json:"duration"`
	StartDate queries.IterationDate `json:"startDate"`
	Completed bool                  `json:"completed"`
}

func JSONProjectFieldsSingleSelectOptions(options []queries.SelectOption) ([]byte, error) {
	var result []projectFieldSingleSelectOptionsJSON

	for _, o := range options {
		result = append(result, projectFieldSingleSelectOptionsJSON{
			ID:   o.ID,
			Name: o.Name,
		})
	}

	return json.Marshal(struct {
		Options    []projectFieldSingleSelectOptionsJSON `json:"options"`
		TotalCount int                                   `json:"totalCount"`
	}{
		Options:    result,
		TotalCount: len(options),
	})
}

type projectFieldSingleSelectOptionsJSON struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// JSONProjectItem serializes a ProjectItem to JSON.
func JSONProjectItem(item queries.ProjectItem) ([]byte, error) {
	return json.Marshal(projectItemJSON{
		ID:    item.ID(),
		Title: item.Title(),
		Body:  item.Body(),
		Type:  item.Type(),
	})
}

type projectItemJSON struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Type  string `json:"type"`
}

// JSONProjectDraftIssue serializes a DraftIssue to JSON.
// This is needed because the field for
// https://docs.github.com/en/graphql/reference/mutations#updateprojectv2draftissue
// is a DraftIssue https://docs.github.com/en/graphql/reference/objects#draftissue
// and not a ProjectV2Item https://docs.github.com/en/graphql/reference/objects#projectv2item
func JSONProjectDraftIssue(item queries.DraftIssue) ([]byte, error) {

	return json.Marshal(draftIssueJSON{
		ID:    item.ID,
		Title: item.Title,
		Body:  item.Body,
		Type:  "DraftIssue",
	})
}

type draftIssueJSON struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Type  string `json:"type"`
}

func projectItemContent(p queries.ProjectItem) any {
	switch p.Content.TypeName {
	case "DraftIssue":
		return struct {
			Type  string `json:"type"`
			Body  string `json:"body"`
			Title string `json:"title"`
		}{
			Type:  p.Content.TypeName,
			Body:  p.Content.DraftIssue.Body,
			Title: p.Content.DraftIssue.Title,
		}
	case "Issue":
		return struct {
			Type       string `json:"type"`
			Body       string `json:"body"`
			Title      string `json:"title"`
			Number     int    `json:"number"`
			Repository string `json:"repository"`
		}{
			Type:       p.Content.TypeName,
			Body:       p.Content.Issue.Body,
			Title:      p.Content.Issue.Title,
			Number:     p.Content.Issue.Number,
			Repository: p.Content.Issue.Repository.NameWithOwner,
		}
	case "PullRequest":
		return struct {
			Type       string `json:"type"`
			Body       string `json:"body"`
			Title      string `json:"title"`
			Number     int    `json:"number"`
			Repository string `json:"repository"`
		}{
			Type:       p.Content.TypeName,
			Body:       p.Content.PullRequest.Body,
			Title:      p.Content.PullRequest.Title,
			Number:     p.Content.PullRequest.Number,
			Repository: p.Content.PullRequest.Repository.NameWithOwner,
		}
	}

	return nil
}

func projectFieldValueData(v queries.FieldValueNodes) any {
	switch v.Type {
	case "ProjectV2ItemFieldDateValue":
		return v.ProjectV2ItemFieldDateValue.Date
	case "ProjectV2ItemFieldIterationValue":
		return struct {
			StartDate string `json:"startDate"`
			Duration  int    `json:"duration"`
		}{
			StartDate: v.ProjectV2ItemFieldIterationValue.StartDate,
			Duration:  v.ProjectV2ItemFieldIterationValue.Duration,
		}
	case "ProjectV2ItemFieldNumberValue":
		return v.ProjectV2ItemFieldNumberValue.Number
	case "ProjectV2ItemFieldSingleSelectValue":
		return v.ProjectV2ItemFieldSingleSelectValue.Name
	case "ProjectV2ItemFieldTextValue":
		return v.ProjectV2ItemFieldTextValue.Text
	case "ProjectV2ItemFieldMilestoneValue":
		return struct {
			Description string `json:"description"`
			DueOn       string `json:"dueOn"`
		}{
			Description: v.ProjectV2ItemFieldMilestoneValue.Milestone.Description,
			DueOn:       v.ProjectV2ItemFieldMilestoneValue.Milestone.DueOn,
		}
	case "ProjectV2ItemFieldLabelValue":
		names := make([]string, 0)
		for _, p := range v.ProjectV2ItemFieldLabelValue.Labels.Nodes {
			names = append(names, p.Name)
		}
		return names
	case "ProjectV2ItemFieldPullRequestValue":
		urls := make([]string, 0)
		for _, p := range v.ProjectV2ItemFieldPullRequestValue.PullRequests.Nodes {
			urls = append(urls, p.Url)
		}
		return urls
	case "ProjectV2ItemFieldRepositoryValue":
		return v.ProjectV2ItemFieldRepositoryValue.Repository.Url
	case "ProjectV2ItemFieldUserValue":
		logins := make([]string, 0)
		for _, p := range v.ProjectV2ItemFieldUserValue.Users.Nodes {
			logins = append(logins, p.Login)
		}
		return logins
	case "ProjectV2ItemFieldReviewerValue":
		names := make([]string, 0)
		for _, p := range v.ProjectV2ItemFieldReviewerValue.Reviewers.Nodes {
			if p.Type == "Team" {
				names = append(names, p.Team.Name)
			} else if p.Type == "User" {
				names = append(names, p.User.Login)
			}
		}
		return names

	}

	return nil
}

// serialize creates a map from field to field values
func serializeProjectWithItems(project *queries.Project) []map[string]any {
	fields := make(map[string]string)

	// make a map of fields by ID
	for _, f := range project.Fields.Nodes {
		fields[f.ID()] = CamelCase(f.Name())
	}
	itemsSlice := make([]map[string]any, 0)

	// for each value, look up the name by ID
	// and set the value to the field value
	for _, i := range project.Items.Nodes {
		o := make(map[string]any)
		o["id"] = i.Id
		o["content"] = projectItemContent(i)
		for _, v := range i.FieldValues.Nodes {
			id := v.ID()
			value := projectFieldValueData(v)

			o[fields[id]] = value
		}
		itemsSlice = append(itemsSlice, o)
	}
	return itemsSlice
}

// JSONProjectWithItems returns a detailed JSON representation of project items.
// JSON fields are `totalCount` and `items`.
func JSONProjectDetailedItems(project *queries.Project) ([]byte, error) {
	items := serializeProjectWithItems(project)
	return json.Marshal(struct {
		Items      []map[string]any `json:"items"`
		TotalCount int              `json:"totalCount"`
	}{
		Items:      items,
		TotalCount: project.Items.TotalCount,
	})
}

// CamelCase converts a string to camelCase, which is useful for turning Go field names to JSON keys.
func CamelCase(s string) string {
	if len(s) == 0 {
		return ""
	}

	if len(s) == 1 {
		return strings.ToLower(s)
	}
	return strings.ToLower(s[0:1]) + s[1:]
}
