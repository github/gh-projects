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
	timeout := 15 * time.Second

	apiOpts := api.ClientOptions{
		Timeout: timeout,
		Headers: map[string]string{},
	}

	return gh.GQLClient(&apiOpts)
}

// PageInfo is a PageInfo GraphQL object https://docs.github.com/en/graphql/reference/objects#pageinfo.
type PageInfo struct {
	EndCursor   githubv4.String
	HasNextPage bool
}

// Project is a ProjectV2 GraphQL object https://docs.github.com/en/graphql/reference/objects#projectv2.
type Project struct {
	Number           int
	URL              string
	ShortDescription string
	Public           bool
	Closed           bool
	Title            string
	ID               string
	Readme           string
	Items            struct {
		TotalCount int
	} `graphql:"items(first: 100)"`
	Fields struct {
		Nodes []ProjectField
	} `graphql:"fields(first:100)"`
	Owner struct {
		User struct {
			Login string
		} `graphql:"... on User"`
		Organization struct {
			Login string
		} `graphql:"... on Organization"`
	}
}

// ProjectWithItems is for fetching all of the items in a single project with pagination
// it fetches a lot of data, be careful with it!
type ProjectWithItems struct {
	Number           int
	URL              string
	ShortDescription string
	Public           bool
	Closed           bool
	Title            string
	ID               string
	Readme           string
	Items            struct {
		PageInfo   PageInfo
		TotalCount int
		Nodes      []ProjectItem
	} `graphql:"items(first: $first, after: $after)"`
	Fields struct {
		Nodes []ProjectField
	} `graphql:"fields(first:100)"`
	Owner struct {
		User struct {
			Login string
		} `graphql:"... on User"`
		Organization struct {
			Login string
		} `graphql:"... on Organization"`
	}
}

// ProjectItem is a ProjectV2Item GraphQL object https://docs.github.com/en/graphql/reference/objects#projectv2item.
type ProjectItem struct {
	Content struct {
		TypeName    string      `graphql:"__typename"`
		DraftIssue  DraftIssue  `graphql:"... on DraftIssue"`
		PullRequest PullRequest `graphql:"... on PullRequest"`
		Issue       Issue       `graphql:"... on Issue"`
	}
	Id          string
	TypeName    string `graphql:"type"`
	FieldValues struct {
		Nodes []FieldValueNodes
	} `graphql:"fieldValues(first: 100)"` // hardcoded to 100 for now on the assumption that this is a reasonable limit
}

func (p ProjectItem) Data() any {
	switch p.Content.TypeName {
	case "DraftIssue":
		return struct {
			TypeName string
			Body     string
			Title    string
		}{
			TypeName: p.Content.TypeName,
			Body:     p.Content.DraftIssue.Body,
			Title:    p.Content.DraftIssue.Title,
		}
	case "Issue":
		return struct {
			TypeName   string
			Body       string
			Title      string
			Number     int
			Repository string
		}{
			TypeName:   p.Content.TypeName,
			Body:       p.Content.Issue.Body,
			Title:      p.Content.Issue.Title,
			Number:     p.Content.Issue.Number,
			Repository: p.Content.Issue.Repository.NameWithOwner,
		}
	case "PullRequest":
		return struct {
			TypeName   string
			Body       string
			Title      string
			Number     int
			Repository string
		}{
			TypeName:   p.Content.TypeName,
			Body:       p.Content.PullRequest.Body,
			Title:      p.Content.PullRequest.Title,
			Number:     p.Content.PullRequest.Number,
			Repository: p.Content.PullRequest.Repository.NameWithOwner,
		}
	}

	return nil
}

type FieldValueNodes struct {
	Type                        string `graphql:"__typename"`
	ProjectV2ItemFieldDateValue struct {
		Date  string
		Field ProjectField
	} `graphql:"... on ProjectV2ItemFieldDateValue"`
	ProjectV2ItemFieldIterationValue struct {
		StartDate string
		Duration  int
		Field     ProjectField
	} `graphql:"... on ProjectV2ItemFieldIterationValue"`
	ProjectV2ItemFieldLabelValue struct {
		Labels struct {
			Nodes []struct {
				Name string
			}
		} `graphql:"labels(first: 10)"` // experienced issues with larger limits, 10 seems like enough for now
		Field ProjectField
	} `graphql:"... on ProjectV2ItemFieldLabelValue"`
	ProjectV2ItemFieldNumberValue struct {
		Number float32
		Field  ProjectField
	} `graphql:"... on ProjectV2ItemFieldNumberValue"`
	ProjectV2ItemFieldSingleSelectValue struct {
		Name  string
		Field ProjectField
	} `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
	ProjectV2ItemFieldTextValue struct {
		Text  string
		Field ProjectField
	} `graphql:"... on ProjectV2ItemFieldTextValue"`
	ProjectV2ItemFieldMilestoneValue struct {
		Milestone struct {
			Description string
			DueOn       string
		}
		Field ProjectField
	} `graphql:"... on ProjectV2ItemFieldMilestoneValue"`
	ProjectV2ItemFieldPullRequestValue struct {
		PullRequests struct {
			Nodes []struct {
				Url string
			}
		} `graphql:"pullRequests(first:10)"` // experienced issues with larger limits, 10 seems like enough for now
		Field ProjectField
	} `graphql:"... on ProjectV2ItemFieldPullRequestValue"`
	ProjectV2ItemFieldRepositoryValue struct {
		Repository struct {
			Url string
		}
		Field ProjectField
	} `graphql:"... on ProjectV2ItemFieldRepositoryValue"`
	ProjectV2ItemFieldUserValue struct {
		Users struct {
			Nodes []struct {
				Login string
			}
		} `graphql:"users(first: 10)"` // experienced issues with larger limits, 10 seems like enough for now
		Field ProjectField
	} `graphql:"... on ProjectV2ItemFieldUserValue"`
	ProjectV2ItemFieldReviewerValue struct {
		Reviewers struct {
			Nodes []struct {
				Type string `graphql:"__typename"`
				Team struct {
					Name string
				} `graphql:"... on Team"`
				User struct {
					Login string
				} `graphql:"... on User"`
			}
		}
		Field ProjectField
	} `graphql:"... on ProjectV2ItemFieldReviewerValue"`
}

func (v FieldValueNodes) ID() string {
	switch v.Type {
	case "ProjectV2ItemFieldDateValue":
		return v.ProjectV2ItemFieldDateValue.Field.ID()
	case "ProjectV2ItemFieldIterationValue":
		return v.ProjectV2ItemFieldIterationValue.Field.ID()
	case "ProjectV2ItemFieldNumberValue":
		return v.ProjectV2ItemFieldNumberValue.Field.ID()
	case "ProjectV2ItemFieldSingleSelectValue":
		return v.ProjectV2ItemFieldSingleSelectValue.Field.ID()
	case "ProjectV2ItemFieldTextValue":
		return v.ProjectV2ItemFieldTextValue.Field.ID()
	case "ProjectV2ItemFieldMilestoneValue":
		return v.ProjectV2ItemFieldMilestoneValue.Field.ID()
	case "ProjectV2ItemFieldLabelValue":
		return v.ProjectV2ItemFieldLabelValue.Field.ID()
	case "ProjectV2ItemFieldPullRequestValue":
		return v.ProjectV2ItemFieldPullRequestValue.Field.ID()
	case "ProjectV2ItemFieldRepositoryValue":
		return v.ProjectV2ItemFieldRepositoryValue.Field.ID()
	case "ProjectV2ItemFieldUserValue":
		return v.ProjectV2ItemFieldUserValue.Field.ID()
	case "ProjectV2ItemFieldReviewerValue":
		return v.ProjectV2ItemFieldReviewerValue.Field.ID()
	}

	return ""
}

func (v FieldValueNodes) Data() any {
	switch v.Type {
	case "ProjectV2ItemFieldDateValue":
		return v.ProjectV2ItemFieldDateValue.Date
	case "ProjectV2ItemFieldIterationValue":
		return struct {
			StartDate string
			Duration  int
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
			Description string
			DueOn       string
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
func ProjectItems(client api.GQLClient, o *Owner, number int, first int) (ProjectWithItems, error) {
	variables := map[string]interface{}{
		"first":  graphql.Int(first),
		"number": graphql.Int(number),
		"after":  (*githubv4.String)(nil),
	}

	project := ProjectWithItems{}

	// get the project by type
	if o.Type == UserOwner {
		variables["login"] = graphql.String(o.Login)
		var query userOwnerWithItems
		err := client.Query("UserProjectWithItems", &query, variables)
		if err != nil {
			return project, err
		}
		project = query.Owner.Project
	} else if o.Type == OrgOwner {
		variables["login"] = graphql.String(o.Login)
		var query orgOwnerWithItems
		err := client.Query("OrgProjectWithItems", &query, variables)
		if err != nil {
			return project, err
		}
		project = query.Owner.Project
	} else if o.Type == ViewerOwner {
		var query viewerOwnerWithItems
		err := client.Query("ViewerProjectWithItems", &query, variables)
		if err != nil {
			return project, err
		}
		project = query.Owner.Project
	} else {
		return project, errors.New("unknown owner type")
	}
	// get the remaining items if there are any
	// and append them to the project items
	hasNext := project.Items.PageInfo.HasNextPage
	cursor := project.Items.PageInfo.EndCursor
	for {
		if !hasNext {
			break
		}
		// set the cursor to the end of the last page
		variables["after"] = (*githubv4.String)(&cursor)
		if o.Type == UserOwner {
			variables["login"] = graphql.String(o.Login)
			var query userOwnerWithItems
			err := client.Query("UserProjectWithItems", &query, variables)
			if err != nil {
				return project, err
			}

			project.Items.Nodes = append(project.Items.Nodes, query.Owner.Project.Items.Nodes...)
			hasNext = query.Owner.Project.Items.PageInfo.HasNextPage
			cursor = query.Owner.Project.Items.PageInfo.EndCursor
		} else if o.Type == OrgOwner {
			variables["login"] = graphql.String(o.Login)
			var query orgOwnerWithItems
			err := client.Query("OrgProjectWithItems", &query, variables)
			if err != nil {
				return project, err
			}

			project.Items.Nodes = append(project.Items.Nodes, query.Owner.Project.Items.Nodes...)
			hasNext = query.Owner.Project.Items.PageInfo.HasNextPage
			cursor = query.Owner.Project.Items.PageInfo.EndCursor
		} else if o.Type == ViewerOwner {
			var query viewerOwnerWithItems
			err := client.Query("ViewerProjectWithItems", &query, variables)
			if err != nil {
				return project, err
			}

			project.Items.Nodes = append(project.Items.Nodes, query.Owner.Project.Items.Nodes...)
			hasNext = query.Owner.Project.Items.PageInfo.HasNextPage
			cursor = query.Owner.Project.Items.PageInfo.EndCursor
		}
	}
	return project, nil
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
			PageInfo PageInfo
			Nodes    []struct {
				Login                   string
				ViewerCanCreateProjects bool
				ID                      string
			}
		} `graphql:"organizations(first: 100, after: $after)"`
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
		Project ProjectWithItems `graphql:"projectV2(number: $number)"`
	} `graphql:"user(login: $login)"`
}

// userOwnerWithFields is used to query the project of a user with its fields.
type userOwnerWithFields struct {
	Owner struct {
		Project struct {
			Fields struct {
				Nodes []ProjectField
			} `graphql:"fields(first: 100)"`
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
		Project ProjectWithItems `graphql:"projectV2(number: $number)"`
	} `graphql:"organization(login: $login)"`
}

// orgOwnerWithFields is used to query the project of an organization with its fields.
type orgOwnerWithFields struct {
	Owner struct {
		Project struct {
			Fields struct {
				Nodes []ProjectField
			} `graphql:"fields(first: 100)"`
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
		Project ProjectWithItems `graphql:"projectV2(number: $number)"`
	} `graphql:"viewer"`
}

// viewerOwnerWithFields is used to query the project of the viewer with its fields.
type viewerOwnerWithFields struct {
	Owner struct {
		Project struct {
			Fields struct {
				Nodes []ProjectField
			} `graphql:"fields(first: 100)"`
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
	return "", errors.New("resource not found, please check the URL")
}

// userProjects queries the $first projects of a user.
type userProjects struct {
	Owner struct {
		Projects struct {
			PageInfo PageInfo
			Nodes    []Project
		} `graphql:"projectsV2(first: $first, after: $after)"`
		Login string
	} `graphql:"user(login: $login)"`
}

// orgProjects queries the $first projects of an organization.
type orgProjects struct {
	Owner struct {
		Projects struct {
			PageInfo PageInfo
			Nodes    []Project
		} `graphql:"projectsV2(first: $first, after: $after)"`
		Login string
	} `graphql:"organization(login: $login)"`
}

// viewerProjects queries the $first projects of the viewer.
type viewerProjects struct {
	Owner struct {
		Projects struct {
			PageInfo PageInfo
			Nodes    []Project
		} `graphql:"projectsV2(first: $first, after: $after)"`
		Login string
	} `graphql:"viewer"`
}

type loginTypes struct {
	Login string
	Type  OwnerType
	ID    string
}

// userOrgLogins gets all the logins of the viewer and the organizations the viewer is a member of.
func userOrgLogins(client api.GQLClient) ([]loginTypes, error) {
	l := make([]loginTypes, 0)
	var v viewerLoginOrgs
	variables := map[string]interface{}{
		"after": (*githubv4.String)(nil),
	}

	err := client.Query("ViewerLoginAndOrgs", &v, variables)
	if err != nil {
		return l, err
	}

	// add the user
	l = append(l, loginTypes{
		Login: v.Viewer.Login,
		Type:  ViewerOwner,
		ID:    v.Viewer.ID,
	})

	// add orgs where the user can create projects
	for _, org := range v.Viewer.Organizations.Nodes {
		if org.ViewerCanCreateProjects {
			l = append(l, loginTypes{
				Login: org.Login,
				Type:  OrgOwner,
				ID:    org.ID,
			})
		}
	}

	// this seem unlikely, but if there are more org logins, paginate the rest
	if v.Viewer.Organizations.PageInfo.HasNextPage {
		return paginateOrgLogins(client, l, string(v.Viewer.Organizations.PageInfo.EndCursor))
	}

	return l, nil
}

// paginateOrgLogins after cursor and append them to the list of logins.
func paginateOrgLogins(client api.GQLClient, l []loginTypes, cursor string) ([]loginTypes, error) {
	var v viewerLoginOrgs
	variables := map[string]interface{}{
		"after": (graphql.String)(cursor),
	}

	err := client.Query("ViewerLoginAndOrgs", &v, variables)
	if err != nil {
		return l, err
	}

	for _, org := range v.Viewer.Organizations.Nodes {
		if org.ViewerCanCreateProjects {
			l = append(l, loginTypes{
				Login: org.Login,
				Type:  OrgOwner,
				ID:    org.ID,
			})
		}
	}

	if v.Viewer.Organizations.PageInfo.HasNextPage {
		return paginateOrgLogins(client, l, string(v.Viewer.Organizations.PageInfo.EndCursor))
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

	logins, err := userOrgLogins(client)
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

	projects, err := Projects(client, o.Login, o.Type)
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

// ProjectsLimit returns up to limit projects for an Owner. If the OwnerType is VIEWER, no login is required.
func ProjectsLimit(client api.GQLClient, login string, t OwnerType, limit int) ([]Project, error) {
	variables := map[string]interface{}{
		"login": graphql.String(login),
		"first": graphql.Int(limit),
		"after": (*graphql.String)(nil),
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
		delete(variables, "login")
		// remove the login from viewer query
		var query viewerProjects
		err := client.Query("ViewerProjects", &query, variables)
		return query.Owner.Projects.Nodes, err
	}
	return []Project{}, errors.New("unknown owner type")
}

// Projects returns all the projects for an Owner. If the OwnerType is VIEWER, no login is required.
func Projects(client api.GQLClient, login string, t OwnerType) ([]Project, error) {
	projects := make([]Project, 0)
	cursor := (*githubv4.String)(nil)
	hasNextPage := false

	// loop until we get all the projects
	for {
		// the code below is very repetitive, the only real difference being the type of the query
		// and the query variables. I couldn't figure out a way to make this cleaner that was worth
		// the cost.
		if t == UserOwner {
			var query userProjects
			variables := map[string]interface{}{
				"login": graphql.String(login),
				"first": graphql.Int(100),
				"after": cursor,
			}
			if err := client.Query("UserProjects", &query, variables); err != nil {
				return projects, err
			}
			projects = append(projects, query.Owner.Projects.Nodes...)
			hasNextPage = query.Owner.Projects.PageInfo.HasNextPage
			cursor = &query.Owner.Projects.PageInfo.EndCursor
		} else if t == OrgOwner {
			var query orgProjects
			variables := map[string]interface{}{
				"login": graphql.String(login),
				"first": graphql.Int(100),
				"after": cursor,
			}
			if err := client.Query("OrgProjects", &query, variables); err != nil {
				return projects, err
			}
			projects = append(projects, query.Owner.Projects.Nodes...)
			hasNextPage = query.Owner.Projects.PageInfo.HasNextPage
			cursor = &query.Owner.Projects.PageInfo.EndCursor
		} else if t == ViewerOwner {
			var query viewerProjects
			variables := map[string]interface{}{
				"first": graphql.Int(100),
				"after": cursor,
			}
			if err := client.Query("ViewerProjects", &query, variables); err != nil {
				return projects, err
			}
			projects = append(projects, query.Owner.Projects.Nodes...)
			hasNextPage = query.Owner.Projects.PageInfo.HasNextPage
			cursor = &query.Owner.Projects.PageInfo.EndCursor
		}

		if !hasNextPage {
			return projects, nil
		}
	}
}
