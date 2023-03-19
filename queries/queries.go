package queries

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
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

const (
	LimitMax = 100 // https://docs.github.com/en/graphql/overview/resource-limitations#node-limit
)

// doQuery wraps calls to client.Query with a spinner
func doQuery(client api.GQLClient, name string, query interface{}, variables map[string]interface{}) error {
	// https://github.com/briandowns/spinner#available-character-sets
	dotStyle := spinner.CharSets[11]
	sp := spinner.New(dotStyle, 120*time.Millisecond, spinner.WithColor("fgCyan"))
	sp.Start()
	err := client.Query(name, query, variables)
	sp.Stop()
	return err
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
		TotalCount int
		Nodes      []ProjectField
		PageInfo   PageInfo
	} `graphql:"fields(first:100)"`
	Owner struct {
		TypeName string `graphql:"__typename"`
		User     struct {
			Login string
		} `graphql:"... on User"`
		Organization struct {
			Login string
		} `graphql:"... on Organization"`
	}
}

func (p Project) OwnerType() string {
	return p.Owner.TypeName
}

func (p Project) OwnerLogin() string {
	if p.OwnerType() == "User" {
		return p.Owner.User.Login
	}
	return p.Owner.Organization.Login
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
	Content     ProjectItemContent
	Id          string
	FieldValues struct {
		Nodes []FieldValueNodes
	} `graphql:"fieldValues(first: 100)"` // hardcoded to 100 for now on the assumption that this is a reasonable limit
}

type ProjectItemContent struct {
	TypeName    string      `graphql:"__typename"`
	DraftIssue  DraftIssue  `graphql:"... on DraftIssue"`
	PullRequest PullRequest `graphql:"... on PullRequest"`
	Issue       Issue       `graphql:"... on Issue"`
}
type ProjectWithFields struct {
	Fields struct {
		PageInfo   PageInfo
		Nodes      []ProjectField
		TotalCount int
	} `graphql:"fields(first: $first, after: $after)"`
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
		} `graphql:"reviewers(first: 10)"` // experienced issues with larger limits, 10 seems like enough for now
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

type DraftIssue struct {
	ID    string
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
	return p.Content.TypeName
}

// Title is the title of the project item.
func (p ProjectItem) Title() string {
	switch p.Content.TypeName {
	case "Issue":
		return p.Content.Issue.Title
	case "PullRequest":
		return p.Content.PullRequest.Title
	case "DraftIssue":
		return p.Content.DraftIssue.Title
	}
	return ""
}

// Body is the body of the project item.
func (p ProjectItem) Body() string {
	switch p.Content.TypeName {
	case "Issue":
		return p.Content.Issue.Body
	case "PullRequest":
		return p.Content.PullRequest.Body
	case "DraftIssue":
		return p.Content.DraftIssue.Body
	}
	return ""
}

// Number is the number of the project item. It is only valid for issues and pull requests.
func (p ProjectItem) Number() int {
	switch p.Content.TypeName {
	case "Issue":
		return p.Content.Issue.Number
	case "PullRequest":
		return p.Content.PullRequest.Number
	}

	return 0
}

func (p ProjectItem) ID() string {
	return p.Id
}

// Repo is the repository of the project item. It is only valid for issues and pull requests.
func (p ProjectItem) Repo() string {
	switch p.Content.TypeName {
	case "Issue":
		return p.Content.Issue.Repository.NameWithOwner
	case "PullRequest":
		return p.Content.PullRequest.Repository.NameWithOwner
	}
	return ""
}

// ProjectItems returns the items of a project. If the OwnerType is VIEWER, no login is required.
func ProjectItems(client api.GQLClient, o *Owner, number int, limit int) (ProjectWithItems, error) {
	project := ProjectWithItems{}
	hasLimit := limit != 0
	variables := map[string]interface{}{
		"first":  graphql.Int(limit),
		"number": graphql.Int(number),
		"after":  (*githubv4.String)(nil),
	}

	// get the project by type
	if o.Type == UserOwner {
		variables["login"] = graphql.String(o.Login)
		var query userOwnerWithItems
		err := doQuery(client, "UserProjectWithItems", &query, variables)
		if err != nil {
			return project, err
		}
		project = query.Owner.Project
	} else if o.Type == OrgOwner {
		variables["login"] = graphql.String(o.Login)
		var query orgOwnerWithItems
		err := doQuery(client, "OrgProjectWithItems", &query, variables)
		if err != nil {
			return project, err
		}
		project = query.Owner.Project
	} else if o.Type == ViewerOwner {
		var query viewerOwnerWithItems
		err := doQuery(client, "ViewerProjectWithItems", &query, variables)
		if err != nil {
			return project, err
		}
		project = query.Owner.Project
	} else {
		return project, errors.New("unknown owner type")
	}
	// get the remaining items if there are any
	// and append them to the project items
	hasNextPage := project.Items.PageInfo.HasNextPage
	cursor := project.Items.PageInfo.EndCursor
	// reset to the default batch size on loops after the first
	variables["first"] = graphql.Int(LimitMax)

	for {
		if !hasNextPage || (hasLimit && len(project.Items.Nodes) >= limit) {
			return project, nil
		}
		// set the cursor to the end of the last page
		variables["after"] = (*githubv4.String)(&cursor)
		if o.Type == UserOwner {
			variables["login"] = graphql.String(o.Login)
			var query userOwnerWithItems
			err := doQuery(client, "UserProjectWithItems", &query, variables)
			if err != nil {
				return project, err
			}

			project.Items.Nodes = append(project.Items.Nodes, query.Owner.Project.Items.Nodes...)
			hasNextPage = query.Owner.Project.Items.PageInfo.HasNextPage
			cursor = query.Owner.Project.Items.PageInfo.EndCursor
		} else if o.Type == OrgOwner {
			variables["login"] = graphql.String(o.Login)
			var query orgOwnerWithItems
			err := doQuery(client, "OrgProjectWithItems", &query, variables)
			if err != nil {
				return project, err
			}

			project.Items.Nodes = append(project.Items.Nodes, query.Owner.Project.Items.Nodes...)
			hasNextPage = query.Owner.Project.Items.PageInfo.HasNextPage
			cursor = query.Owner.Project.Items.PageInfo.EndCursor
		} else if o.Type == ViewerOwner {
			var query viewerOwnerWithItems
			err := doQuery(client, "ViewerProjectWithItems", &query, variables)
			if err != nil {
				return project, err
			}

			project.Items.Nodes = append(project.Items.Nodes, query.Owner.Project.Items.Nodes...)
			hasNextPage = query.Owner.Project.Items.PageInfo.HasNextPage
			cursor = query.Owner.Project.Items.PageInfo.EndCursor
		}
	}
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

// ProjectFields returns a project with fields. If the OwnerType is VIEWER, no login is required.
func ProjectFields(client api.GQLClient, o *Owner, number int, limit int) (ProjectWithFields, error) {
	project := ProjectWithFields{}
	hasLimit := limit != 0
	variables := map[string]interface{}{
		"first":  graphql.Int(limit),
		"number": graphql.Int(number),
		"after":  (*githubv4.String)(nil),
	}

	if o.Type == UserOwner {
		variables["login"] = graphql.String(o.Login)
		var query userOwnerWithFields
		err := doQuery(client, "UserProjectWithFields", &query, variables)
		if err != nil {
			return project, err
		}

		project = query.Owner.Project
	} else if o.Type == OrgOwner {
		variables["login"] = graphql.String(o.Login)
		var query orgOwnerWithFields
		err := doQuery(client, "OrgProjectWithFields", &query, variables)
		if err != nil {
			return project, err
		}

		project = query.Owner.Project
	} else if o.Type == ViewerOwner {
		var query viewerOwnerWithFields
		err := doQuery(client, "ViewerProjectWithFields", &query, variables)
		if err != nil {
			return project, err
		}

		project = query.Owner.Project
	} else {
		return project, errors.New("unknown owner type")
	}

	// get the remaining items if there are any
	// and append them to the project items
	hasNextPage := project.Fields.PageInfo.HasNextPage
	cursor := project.Fields.PageInfo.EndCursor
	// reset to the default batch size on loops after the first
	variables["first"] = graphql.Int(LimitMax)

	for {
		if !hasNextPage || (hasLimit && len(project.Fields.Nodes) >= limit) {
			return project, nil
		}

		// set the cursor to the end of the last page
		variables["after"] = (*githubv4.String)(&cursor)
		if o.Type == UserOwner {
			variables["login"] = graphql.String(o.Login)
			var query userOwnerWithFields
			err := doQuery(client, "UserProjectWithFields", &query, variables)
			if err != nil {
				return project, err
			}

			project.Fields.Nodes = append(project.Fields.Nodes, query.Owner.Project.Fields.Nodes...)
			hasNextPage = query.Owner.Project.Fields.PageInfo.HasNextPage
			cursor = query.Owner.Project.Fields.PageInfo.EndCursor
		} else if o.Type == OrgOwner {
			variables["login"] = graphql.String(o.Login)
			var query orgOwnerWithFields
			err := doQuery(client, "OrgProjectWithFields", &query, variables)
			if err != nil {
				return project, err
			}

			project.Fields.Nodes = append(project.Fields.Nodes, query.Owner.Project.Fields.Nodes...)
			hasNextPage = query.Owner.Project.Fields.PageInfo.HasNextPage
			cursor = query.Owner.Project.Fields.PageInfo.EndCursor
		} else if o.Type == ViewerOwner {
			var query viewerOwnerWithFields
			err := doQuery(client, "ViewerProjectWithFields", &query, variables)
			if err != nil {
				return project, err
			}

			project.Fields.Nodes = append(project.Fields.Nodes, query.Owner.Project.Fields.Nodes...)
			hasNextPage = query.Owner.Project.Fields.PageInfo.HasNextPage
			cursor = query.Owner.Project.Fields.PageInfo.EndCursor
		}
	}
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
		Project ProjectWithFields `graphql:"projectV2(number: $number)"`
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
		Project ProjectWithFields `graphql:"projectV2(number: $number)"`
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
		Project ProjectWithFields `graphql:"projectV2(number: $number)"`
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
	err := doQuery(client, "Viewer", &query, map[string]interface{}{})
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
		err := doQuery(client, "UserLogin", &query, variables)
		return query.User.Id, err
	} else if t == OrgOwner {
		var query orgLogin
		err := doQuery(client, "OrgLogin", &query, variables)
		return query.Organization.Id, err
	} else if t == ViewerOwner {
		var query viewerLogin
		err := doQuery(client, "ViewerLogin", &query, nil)
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
	err = doQuery(client, "GetIssueOrPullRequest", &query, variables)
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
			TotalCount int
			PageInfo   PageInfo
			Nodes      []Project
		} `graphql:"projectsV2(first: $first, after: $after)"`
		Login string
	} `graphql:"user(login: $login)"`
}

// orgProjects queries the $first projects of an organization.
type orgProjects struct {
	Owner struct {
		Projects struct {
			TotalCount int
			PageInfo   PageInfo
			Nodes      []Project
		} `graphql:"projectsV2(first: $first, after: $after)"`
		Login string
	} `graphql:"organization(login: $login)"`
}

// viewerProjects queries the $first projects of the viewer.
type viewerProjects struct {
	Owner struct {
		Projects struct {
			TotalCount int
			PageInfo   PageInfo
			Nodes      []Project
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

	err := doQuery(client, "ViewerLoginAndOrgs", &v, variables)
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

	err := doQuery(client, "ViewerLoginAndOrgs", &v, variables)
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
			err := doQuery(client, "UserProject", &query, variables)
			return &query.Owner.Project, err
		} else if o.Type == OrgOwner {
			var query orgOwner
			err := doQuery(client, "OrgProject", &query, variables)
			return &query.Owner.Project, err
		} else if o.Type == ViewerOwner {
			var query viewerOwner
			err := doQuery(client, "ViewerProject", &query, map[string]interface{}{"number": graphql.Int(number)})
			return &query.Owner.Project, err
		}
		return nil, errors.New("unknown owner type")
	}

	projects, _, err := Projects(client, o.Login, o.Type, 0)
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

// Projects returns all the projects for an Owner. If the OwnerType is VIEWER, no login is required.
func Projects(client api.GQLClient, login string, t OwnerType, limit int) ([]Project, int, error) {
	projects := make([]Project, 0)
	cursor := (*githubv4.String)(nil)
	hasNextPage := false
	hasLimit := limit != 0
	totalCount := 0

	// the api limits batches to 100. We want to use the maximum batch size unless the user
	// requested a lower limit.
	first := LimitMax
	if hasLimit && limit < first {
		first = limit
	}
	// loop until we get all the projects
	for {
		// the code below is very repetitive, the only real difference being the type of the query
		// and the query variables. I couldn't figure out a way to make this cleaner that was worth
		// the cost.
		if t == UserOwner {
			var query userProjects
			variables := map[string]interface{}{
				"login": graphql.String(login),
				"first": graphql.Int(first),
				"after": cursor,
			}
			if err := doQuery(client, "UserProjects", &query, variables); err != nil {
				return projects, 0, err
			}
			projects = append(projects, query.Owner.Projects.Nodes...)
			hasNextPage = query.Owner.Projects.PageInfo.HasNextPage
			cursor = &query.Owner.Projects.PageInfo.EndCursor
			totalCount = query.Owner.Projects.TotalCount
		} else if t == OrgOwner {
			var query orgProjects
			variables := map[string]interface{}{
				"login": graphql.String(login),
				"first": graphql.Int(first),
				"after": cursor,
			}
			if err := doQuery(client, "OrgProjects", &query, variables); err != nil {
				return projects, 0, err
			}
			projects = append(projects, query.Owner.Projects.Nodes...)
			hasNextPage = query.Owner.Projects.PageInfo.HasNextPage
			cursor = &query.Owner.Projects.PageInfo.EndCursor
			totalCount = query.Owner.Projects.TotalCount
		} else if t == ViewerOwner {
			var query viewerProjects
			variables := map[string]interface{}{
				"first": graphql.Int(first),
				"after": cursor,
			}
			if err := doQuery(client, "ViewerProjects", &query, variables); err != nil {
				return projects, 0, err
			}
			projects = append(projects, query.Owner.Projects.Nodes...)
			hasNextPage = query.Owner.Projects.PageInfo.HasNextPage
			cursor = &query.Owner.Projects.PageInfo.EndCursor
			totalCount = query.Owner.Projects.TotalCount
		}

		if !hasNextPage || (hasLimit && len(projects) >= limit) {
			return projects, totalCount, nil
		}
	}
}
