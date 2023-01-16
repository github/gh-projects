package queries

import (
	"time"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
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

type ProjectViewerLogin struct {
	Viewer struct {
		Login string
	}
}
