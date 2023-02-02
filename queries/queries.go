package queries

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/shurcooL/githubv4"
	"github.com/shurcooL/graphql"
)

type ClientOptions struct {
	Timeout time.Duration
}

func NewClient() (api.GQLClient, error) {
	timeout := 5 * time.Second

	apiOpts := api.ClientOptions{
		Timeout: timeout,
		Headers: map[string]string{
			"GraphQL-Features": "memex_copy_project,memex_project_fields_api,memex_project_delete_api",
		},
	}

	return gh.GQLClient(&apiOpts)
}

// Project is a ProjectV2 GraphQL object https://docs.github.com/en/graphql/reference/objects#projectv2.
type Project struct {
	Number           int
	URL              string
	ShortDescription string
	Closed           bool
	Title            string
	ID               string
	Owner            struct {
		User struct {
			Login string
		} `graphql:"... on User"`
		Organization struct {
			Login string
		} `graphql:"... on Organization"`
	}
}

// ProjectId returns the ID of a project. If OwnerType is VIEWER, no login is required.
func ProjectId(client api.GQLClient, o *Owner, number int) (string, error) {
	variables := map[string]interface{}{
		"login":  graphql.String(o.Login),
		"number": graphql.Int(number),
	}
	if o.Type == UserOwner {
		var query userOwner
		err := client.Query("UserProject", &query, variables)
		return query.Owner.Project.ID, err
	} else if o.Type == OrgOwner {
		var query orgOwner
		err := client.Query("OrgProject", &query, variables)
		return query.Owner.Project.ID, err
	} else if o.Type == ViewerOwner {
		var query viewerOwner
		err := client.Query("ViewerProject", &query, map[string]interface{}{"number": graphql.Int(number)})
		return query.Owner.Project.ID, err
	}
	return "", errors.New("unknown owner type")
}

// ProjectItem is a ProjectV2Item GraphQL object https://docs.github.com/en/graphql/reference/objects#projectv2item.
type ProjectItem struct {
	Id       string
	TypeName string `graphql:"type"`
	Content  struct {
		DraftIssue  DraftIssue  `graphql:"... on DraftIssue"`
		PullRequest PullRequest `graphql:"... on PullRequest"`
		Issue       Issue       `graphql:"... on Issue"`
	}
}

type DraftIssue struct {
	Body  string
	Title string
}

type PullRequest struct {
	Body       string
	Title      string
	Number     int
	Repository struct {
		NameWithOwner string
	}
}

type Issue struct {
	Body       string
	Title      string
	Number     int
	Repository struct {
		NameWithOwner string
	}
}

// Type is the underlying type of the project item.
func (p ProjectItem) Type() string {
	return p.TypeName
}

// Title is the title of the project item.
func (p ProjectItem) Title() string {
	if p.TypeName == "ISSUE" {
		return p.Content.Issue.Title
	} else if p.TypeName == "PULL_REQUEST" {
		return p.Content.PullRequest.Title
	} else if p.TypeName == "DRAFT_ISSUE" {
		return p.Content.DraftIssue.Title
	}
	return ""
}

// Body is the body of the project item.
func (p ProjectItem) Body() string {
	if p.TypeName == "ISSUE" {
		return p.Content.Issue.Body
	} else if p.TypeName == "PULL_REQUEST" {
		return p.Content.PullRequest.Body
	} else if p.TypeName == "DRAFT_ISSUE" {
		return p.Content.DraftIssue.Body
	}
	return ""
}

// Number is the number of the project item. It is only valid for issues and pull requests.
func (p ProjectItem) Number() int {
	if p.TypeName == "ISSUE" {
		return p.Content.Issue.Number
	} else if p.TypeName == "PULL_REQUEST" {
		return p.Content.PullRequest.Number
	}
	return 0
}

func (p ProjectItem) ID() string {
	return p.Id
}

// Repo is the repository of the project item. It is only valid for issues and pull requests.
func (p ProjectItem) Repo() string {
	if p.TypeName == "ISSUE" {
		return p.Content.Issue.Repository.NameWithOwner
	} else if p.TypeName == "PULL_REQUEST" {
		return p.Content.PullRequest.Repository.NameWithOwner
	}
	return ""
}

// ProjectItems returns the items of a project. If the OwnerType is VIEWER, no login is required.
func ProjectItems(client api.GQLClient, o *Owner, number int, first int) ([]ProjectItem, error) {
	variables := map[string]interface{}{
		"first":  graphql.Int(first),
		"number": graphql.Int(number),
	}
	if o.Type == UserOwner {
		variables["login"] = graphql.String(o.Login)
		var query userOwnerWithItems
		err := client.Query("UserProjectWithItems", &query, variables)
		return query.Owner.Project.Items.Nodes, err
	} else if o.Type == OrgOwner {
		variables["login"] = graphql.String(o.Login)
		var query orgOwnerWithItems
		err := client.Query("OrgProjectWithItems", &query, variables)
		return query.Owner.Project.Items.Nodes, err
	} else if o.Type == ViewerOwner {
		var query viewerOwnerWithItems
		err := client.Query("ViewerProjectWithItems", &query, variables)
		return query.Owner.Project.Items.Nodes, err
	}
	return []ProjectItem{}, errors.New("unknown owner type")
}

// ProjectField is a ProjectV2FieldConfiguration GraphQL object https://docs.github.com/en/graphql/reference/unions#projectv2fieldconfiguration.
type ProjectField struct {
	TypeName string `graphql:"__typename"`
	Field    struct {
		ID       string
		Name     string
		DataType string
	} `graphql:"... on ProjectV2Field"`
	IterationField struct {
		ID       string
		Name     string
		DataType string
	} `graphql:"... on ProjectV2IterationField"`
	SingleSelectField struct {
		ID       string
		Name     string
		DataType string
	} `graphql:"... on ProjectV2SingleSelectField"`
}

// ID is the ID of the project field.
func (p ProjectField) ID() string {
	if p.TypeName == "ProjectV2Field" {
		return p.Field.ID
	} else if p.TypeName == "ProjectV2IterationField" {
		return p.IterationField.ID
	} else if p.TypeName == "ProjectV2SingleSelectField" {
		return p.SingleSelectField.ID
	}
	return ""
}

// Name is the name of the project field.
func (p ProjectField) Name() string {
	if p.TypeName == "ProjectV2Field" {
		return p.Field.Name
	} else if p.TypeName == "ProjectV2IterationField" {
		return p.IterationField.Name
	} else if p.TypeName == "ProjectV2SingleSelectField" {
		return p.SingleSelectField.Name
	}
	return ""
}

// Type is the typename of the project field.
func (p ProjectField) Type() string {
	return p.TypeName
}

// ProjectFields returns the fields of a project. If the OwnerType is VIEWER, no login is required.
func ProjectFields(client api.GQLClient, o *Owner, number int, first int) ([]ProjectField, error) {
	variables := map[string]interface{}{
		"first":  graphql.Int(first),
		"number": graphql.Int(number),
	}
	if o.Type == UserOwner {
		variables["login"] = graphql.String(o.Login)
		var query userOwnerWithFields
		err := client.Query("UserProjectWithFields", &query, variables)
		return query.Owner.Project.Fields.Nodes, err
	} else if o.Type == OrgOwner {
		variables["login"] = graphql.String(o.Login)
		var query orgOwnerWithFields
		err := client.Query("OrgProjectWithFields", &query, variables)
		return query.Owner.Project.Fields.Nodes, err
	} else if o.Type == ViewerOwner {
		var query viewerOwnerWithFields
		err := client.Query("ViewerProjectWithFields", &query, variables)
		return query.Owner.Project.Fields.Nodes, err
	}
	return []ProjectField{}, errors.New("unknown owner type")
}

// viewerLogin is used to query the Login of the viewer.
type viewerLogin struct {
	Viewer struct {
		Login string
		Id    string
	}
}

// userLogin is used to query the Login of a user.
type userLogin struct {
	User struct {
		Login string
		Id    string
	} `graphql:"user(login: $login)"`
}

// orgLogin is used to query the Login of an organization.
type orgLogin struct {
	Organization struct {
		Login string
		Id    string
	} `graphql:"organization(login: $login)"`
}

type viewerLoginOrgs struct {
	Viewer struct {
		Login         string
		ID            string
		Organizations struct {
			Nodes []struct {
				Login                   string
				ViewerCanCreateProjects bool
				ID                      string
			}
		} `graphql:"organizations(first: 100)"`
	}
}

// userOwner is used to query the project of a user.
type userOwner struct {
	Owner struct {
		Project Project `graphql:"projectV2(number: $number)"`
		Login   string
	} `graphql:"user(login: $login)"`
}

// userOwnerWithItems is used to query the project of a user with its items.
type userOwnerWithItems struct {
	Owner struct {
		Project struct {
			Items struct {
				Nodes []ProjectItem
			} `graphql:"items(first: $first)"`
		} `graphql:"projectV2(number: $number)"`
	} `graphql:"user(login: $login)"`
}

// userOwnerWithFields is used to query the project of a user with its fields.
type userOwnerWithFields struct {
	Owner struct {
		Project struct {
			Fields struct {
				Nodes []ProjectField
			} `graphql:"fields(first: $first)"`
		} `graphql:"projectV2(number: $number)"`
	} `graphql:"user(login: $login)"`
}

// orgOwner is used to query the project of an organization.
type orgOwner struct {
	Owner struct {
		Project Project `graphql:"projectV2(number: $number)"`
		Login   string
	} `graphql:"organization(login: $login)"`
}

// orgOwnerWithItems is used to query the project of an organization with its items.
type orgOwnerWithItems struct {
	Owner struct {
		Project struct {
			Items struct {
				Nodes []ProjectItem
			} `graphql:"items(first: $first)"`
		} `graphql:"projectV2(number: $number)"`
	} `graphql:"organization(login: $login)"`
}

// orgOwnerWithFields is used to query the project of an organization with its fields.
type orgOwnerWithFields struct {
	Owner struct {
		Project struct {
			Fields struct {
				Nodes []ProjectField
			} `graphql:"fields(first: $first)"`
		} `graphql:"projectV2(number: $number)"`
	} `graphql:"organization(login: $login)"`
}

// viewerOwner is used to query the project of the viewer.
type viewerOwner struct {
	Owner struct {
		Project Project `graphql:"projectV2(number: $number)"`
		Login   string
	} `graphql:"viewer"`
}

// viewerOwnerWithItems is used to query the project of the viewer with its items.
type viewerOwnerWithItems struct {
	Owner struct {
		Project struct {
			Items struct {
				Nodes []ProjectItem
			} `graphql:"items(first: $first)"`
		} `graphql:"projectV2(number: $number)"`
	} `graphql:"viewer"`
}

// viewerOwnerWithFields is used to query the project of the viewer with its fields.
type viewerOwnerWithFields struct {
	Owner struct {
		Project struct {
			Fields struct {
				Nodes []ProjectField
			} `graphql:"fields(first: $first)"`
		} `graphql:"projectV2(number: $number)"`
	} `graphql:"viewer"`
}

// OwnerType is the type of the owner of a project, which can be either a user or an organization. Viewer is the current user.
type OwnerType string

const UserOwner OwnerType = "USER"
const OrgOwner OwnerType = "ORGANIZATION"
const ViewerOwner OwnerType = "VIEWER"

// ViewerLoginName returns the login name of the viewer.
func ViewerLoginName(client api.GQLClient) (string, error) {
	var query viewerLogin
	err := client.Query("Viewer", &query, map[string]interface{}{})
	if err != nil {
		return "", err
	}
	return query.Viewer.Login, nil
}

// OwnerID returns the ID of an OwnerType. If the OwnerType is VIEWER, no login is required.
func OwnerID(client api.GQLClient, login string, t OwnerType) (string, error) {
	variables := map[string]interface{}{
		"login": graphql.String(login),
	}
	if t == UserOwner {
		var query userLogin
		err := client.Query("UserLogin", &query, variables)
		return query.User.Id, err
	} else if t == OrgOwner {
		var query orgLogin
		err := client.Query("OrgLogin", &query, variables)
		return query.Organization.Id, err
	} else if t == ViewerOwner {
		var query viewerLogin
		err := client.Query("ViewerLogin", &query, nil)
		return query.Viewer.Id, err
	}
	return "", errors.New("unknown owner type")
}

// issueOrPullRequest is used to query the global id of an issue or pull request by its URL.
type issueOrPullRequest struct {
	Resource struct {
		Typename string `graphql:"__typename"`
		Issue    struct {
			ID string
		} `graphql:"... on Issue"`
		PullRequest struct {
			ID string
		} `graphql:"... on PullRequest"`
	} `graphql:"resource(url: $url)"`
}

// IssueOrPullRequestID returns the ID of the issue or pull request from a URL.
func IssueOrPullRequestID(client api.GQLClient, rawURL string) (string, error) {
	uri, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	variables := map[string]interface{}{
		"url": githubv4.URI{URL: uri},
	}
	var query issueOrPullRequest
	err = client.Query("GetIssueOrPullRequest", &query, variables)
	if err != nil {
		return "", err
	}
	if query.Resource.Typename == "Issue" {
		return query.Resource.Issue.ID, nil
	} else if query.Resource.Typename == "PullRequest" {
		return query.Resource.PullRequest.ID, nil
	}
	return "", errors.New("unknown resource type")
}

// userProjects queries the $first projects of a user.
type userProjects struct {
	Owner struct {
		Projects struct {
			Nodes []Project
		} `graphql:"projectsV2(first: $first)"`
		Login string
	} `graphql:"user(login: $login)"`
}

// orgProjects queries the $first projects of an organization.
type orgProjects struct {
	Owner struct {
		Projects struct {
			Nodes []Project
		} `graphql:"projectsV2(first: $first)"`
		Login string
	} `graphql:"organization(login: $login)"`
}

// viewerProjects queries the $first projects of the viewer.
type viewerProjects struct {
	Owner struct {
		Projects struct {
			Nodes []Project
		} `graphql:"projectsV2(first: $first)"`
		Login string
	} `graphql:"viewer"`
}

type loginTypes struct {
	Login string
	Type  OwnerType
	ID    string
}

func logins(client api.GQLClient) ([]loginTypes, error) {
	l := []loginTypes{}
	var v viewerLoginOrgs
	err := client.Query("ViewerLoginAndOrgs", &v, nil)
	if err != nil {
		return l, err
	}
	l = append(l, loginTypes{
		Login: v.Viewer.Login,
		Type:  ViewerOwner,
		ID:    v.Viewer.ID,
	})
	for _, org := range v.Viewer.Organizations.Nodes {
		if org.ViewerCanCreateProjects {
			l = append(l, loginTypes{
				Login: org.Login,
				Type:  OrgOwner,
				ID:    org.ID,
			})
		}
	}
	return l, nil
}

type Owner struct {
	Login string
	Type  OwnerType
	ID    string
}

// NewOwner creates a project Owner
// When userLogin == "@me", userLogin becomes the current viewer
// If userLogin is not empty, it is used to lookup the user owner
// If orgLogin is not empty, it is used to lookup the org owner
// If both userLogin and orgLogin are empty, interative mode is used to select an owner
// from the current viewer and their organizations
func NewOwner(client api.GQLClient, userLogin, orgLogin string) (*Owner, error) {
	if userLogin == "@me" {
		id, err := OwnerID(client, userLogin, ViewerOwner)
		if err != nil {
			return nil, err
		}

		return &Owner{
			Login: userLogin,
			Type:  ViewerOwner,
			ID:    id,
		}, nil
	} else if userLogin != "" {
		id, err := OwnerID(client, userLogin, UserOwner)
		if err != nil {
			return nil, err
		}

		return &Owner{
			Login: userLogin,
			Type:  UserOwner,
			ID:    id,
		}, nil
	} else if orgLogin != "" {
		id, err := OwnerID(client, orgLogin, OrgOwner)
		if err != nil {
			return nil, err
		}

		return &Owner{
			Login: orgLogin,
			Type:  OrgOwner,
			ID:    id,
		}, nil
	}

	logins, err := logins(client)
	if err != nil {
		return nil, err
	}

	options := make([]string, 0, len(logins))
	for _, l := range logins {
		options = append(options, l.Login)
	}

	var q = []*survey.Question{
		{
			Name: "owner",
			Prompt: &survey.Select{
				Message: "Which owner would you like to use?",
				Options: options,
			},
			Validate: survey.Required,
		},
	}

	answerIndex := 0
	err = survey.Ask(q, &answerIndex)
	if err != nil {
		return nil, err
	}

	l := logins[answerIndex]
	return &Owner{
		Login: l.Login,
		Type:  l.Type,
		ID:    l.ID,
	}, nil
}

// NewProject creates a project based on the owner and project number
// if number is 0 it will prompt the user to select a project interactively
// otherwise it will make a request to get the project by number
func NewProject(client api.GQLClient, o *Owner, number int) (*Project, error) {
	if number != 0 {
		variables := map[string]interface{}{
			"login":  graphql.String(o.Login),
			"number": graphql.Int(number),
		}
		if o.Type == UserOwner {
			var query userOwner
			err := client.Query("UserProject", &query, variables)
			return &query.Owner.Project, err
		} else if o.Type == OrgOwner {
			var query orgOwner
			err := client.Query("OrgProject", &query, variables)
			return &query.Owner.Project, err
		} else if o.Type == ViewerOwner {
			var query viewerOwner
			err := client.Query("ViewerProject", &query, map[string]interface{}{"number": graphql.Int(number)})
			return &query.Owner.Project, err
		}
		return nil, errors.New("unknown owner type")
	}
	// TODO: pagination
	projects, err := Projects(client, o.Login, o.Type, 100)
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("no projects found for %s", o.Login)
	}

	options := make([]string, 0, len(projects))
	for _, p := range projects {
		title := fmt.Sprintf("%s (#%d)", p.Title, p.Number)
		options = append(options, title)
	}

	var q = []*survey.Question{
		{
			Name: "project",
			Prompt: &survey.Select{
				Message: "Which project would you like to use?",
				Options: options,
			},
			Validate: survey.Required,
		},
	}

	answerIndex := 0
	err = survey.Ask(q, &answerIndex)
	if err != nil {
		return nil, err
	}

	return &projects[answerIndex], nil
}

// Projects returns the projects for an Owner. If the OwnerType is VIEWER, no login is required.
func Projects(client api.GQLClient, login string, t OwnerType, first int) ([]Project, error) {
	variables := map[string]interface{}{
		"login": graphql.String(login),
		"first": graphql.Int(first),
	}
	if t == UserOwner {
		var query userProjects
		err := client.Query("UserProjects", &query, variables)
		return query.Owner.Projects.Nodes, err
	} else if t == OrgOwner {
		var query orgProjects
		err := client.Query("OrgProjects", &query, variables)
		return query.Owner.Projects.Nodes, err
	} else if t == ViewerOwner {
		var query viewerProjects
		err := client.Query("ViewerProjects", &query, map[string]interface{}{"first": graphql.Int(first)})
		return query.Owner.Projects.Nodes, err
	}
	return []Project{}, errors.New("unknown owner type")
}
