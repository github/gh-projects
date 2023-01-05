package cmd

type projects struct {
	Nodes []projectNode
}

type projectNode struct {
	Title            string
	Number           int
	URL              string
	ShortDescription string
	Closed           bool
}

// userQuery, organizationQuery, and viewerQuery will all satisfy the query interface
type query interface {
	projects() projects
	login() string
}

type userQuery struct {
	Owner struct {
		Projects projects `graphql:"projectsV2(first: $first)"`
		Login    string
	} `graphql:"user(login: $login)"`
}

func (u userQuery) projects() projects {
	return u.Owner.Projects
}

func (u userQuery) login() string {
	return u.Owner.Login
}

type organizationQuery struct {
	Owner struct {
		Projects projects `graphql:"projectsV2(first: $first)"`
		Login    string
	} `graphql:"organization(login: $login)"`
}

func (o organizationQuery) projects() projects {
	return o.Owner.Projects
}

func (o organizationQuery) login() string {
	return o.Owner.Login
}

type viewerQuery struct {
	Owner struct {
		Projects projects `graphql:"projectsV2(first: $first)"`
		Login    string
	} `graphql:"viewer"`
}

func (v viewerQuery) projects() projects {
	return v.Owner.Projects
}

func (v viewerQuery) login() string {
	return v.Owner.Login
}

type viewerLogin struct {
	Viewer struct {
		Login string
	}
}
