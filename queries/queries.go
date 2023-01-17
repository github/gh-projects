package queries

import (
	"errors"
	"time"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/shurcooL/graphql"
)

type ClientOptions struct {
	Timeout time.Duration
}

func NewClient() (api.GQLClient, error) {
	timeout := 5 * time.Second

	apiOpts := api.ClientOptions{
		Timeout: timeout,
	}

	return gh.GQLClient(&apiOpts)
}

// General Queries

type ProjectViewerLogin struct {
	Viewer struct {
		Login string
		Id    string
	}
}

type ProjectUserLogin struct {
	User struct {
		Login string
		Id    string
	} `graphql:"user(login: $login)"`
}

type ProjectOrgLogin struct {
	Organization struct {
		Login string
		Id    string
	} `graphql:"organization(login: $login)"`
}

type OwnerType string

const UserOwner OwnerType = "USER"
const OrgOwner OwnerType = "ORGANIZATION"
const ViewerOwner OwnerType = "VIEWER"

func GetOwnerId(client api.GQLClient, login string, t OwnerType) (string, error) {
	variables := map[string]interface{}{
		"login": graphql.String(login),
	}
	if t == UserOwner {
		var query ProjectUserLogin
		err := client.Query("UserLogin", &query, variables)
		return query.User.Id, err
	} else if t == OrgOwner {
		var query ProjectOrgLogin
		err := client.Query("OrgLogin", &query, variables)
		return query.Organization.Id, err
	} else if t == ViewerOwner {
		var query ProjectViewerLogin
		err := client.Query("ViewerLogin", &query, nil)
		return query.Viewer.Id, err
	}
	return "", errors.New("unknown owner type")
}

// List Queries

type Projects struct {
	Nodes []ProjectNode
}

type ProjectNode struct {
	Title            string
	Number           int
	URL              string
	ShortDescription string
	Closed           bool
}

// userQuery, organizationQuery, and viewerQuery will all satisfy the query interface
type ProjectQuery interface {
	Projects() Projects
	Login() string
}

type ProjectUserQuery struct {
	Owner struct {
		Projects Projects `graphql:"projectsV2(first: $first)"`
		Login    string
	} `graphql:"user(login: $login)"`
}

func (u ProjectUserQuery) Projects() Projects {
	return u.Owner.Projects
}

func (u ProjectUserQuery) Login() string {
	return u.Owner.Login
}

type ProjectOrganizationQuery struct {
	Owner struct {
		Projects Projects `graphql:"projectsV2(first: $first)"`
		Login    string
	} `graphql:"organization(login: $login)"`
}

func (o ProjectOrganizationQuery) Projects() Projects {
	return o.Owner.Projects
}

func (o ProjectOrganizationQuery) Login() string {
	return o.Owner.Login
}

type ProjectViewerQuery struct {
	Owner struct {
		Projects Projects `graphql:"projectsV2(first: $first)"`
		Login    string
	} `graphql:"viewer"`
}

func (v ProjectViewerQuery) Projects() Projects {
	return v.Owner.Projects
}

func (v ProjectViewerQuery) Login() string {
	return v.Owner.Login
}

// Create Queries

type CreateProject struct {
	OwnerId      string
	Title        string
	TeamId       string
	RepositoryId string
}

type CreateProjectMutation struct {
	CreateProjectV2 struct {
		ProjectV2 ProjectV2 `graphql:"projectV2"`
	} `graphql:"createProjectV2(input:$input)"`
}

type ProjectV2 struct {
	Title string
	Id    string
	Url   string
	Owner struct {
		User struct {
			Login string
		} `graphql:"... on User"`
		Organization struct {
			Login string
		} `graphql:"... on Organization"`
	}
}
