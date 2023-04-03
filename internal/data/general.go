package data

import (
	"time"

	"github.com/cli/go-gh/pkg/api"
	"github.com/shurcooL/graphql"
)

type Getter interface {
	GetReposList(owner string, endCursor *string) ([]ReposQuery, error)
	GetRepo(owner string, name string) ([]RepoQuery, error)
	GetOrgActionSecrets(owner string) ([]byte, error)
	GetRepoActionSecrets(owner string, repo string) ([]byte, error)
	GetScopedOrgActionSecrets(owner string, secret string) ([]byte, error)
	GetOrgDependabotSecrets(owner string) ([]byte, error)
	GetRepoDependabotSecrets(owner string, repo string) ([]byte, error)
	GetScopedOrgDependabotSecrets(owner string, secret string) ([]byte, error)
	GetOrgCodespacesSecrets(owner string) ([]byte, error)
	GetRepoCodespacesSecrets(owner string, repo string) ([]byte, error)
	GetScopedOrgCodespacesSecrets(owner string, secret string) ([]byte, error)
}

type APIGetter struct {
	gqlClient  api.GQLClient
	restClient api.RESTClient
}

func NewAPIGetter(gqlClient api.GQLClient, restClient api.RESTClient) *APIGetter {
	return &APIGetter{
		gqlClient:  gqlClient,
		restClient: restClient,
	}
}

type SecretExport struct {
	SecretLevel    string
	SecretType     string
	SecretName     string
	SecretAccess   string
	RepositoryName string
	RepositoryID   int
}

type RepoInfo struct {
	DatabaseId int       `json:"databaseId"`
	Name       string    `json:"name"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Visibility string    `json:"visibility"`
}

type ReposQuery struct {
	Organization struct {
		Repositories struct {
			TotalCount int
			Nodes      []RepoInfo
			PageInfo   struct {
				EndCursor   string
				HasNextPage bool
			}
		} `graphql:"repositories(first: 100, after: $endCursor)"`
	} `graphql:"organization(login: $owner)"`
}

type RepoQuery struct {
	Repository RepoInfo `graphql:"repository(owner: $owner, name: $name)"`
}

func (g *APIGetter) GetReposList(owner string, endCursor *string) (*ReposQuery, error) {
	query := new(ReposQuery)
	variables := map[string]interface{}{
		"endCursor": (*graphql.String)(endCursor),
		"owner":     graphql.String(owner),
	}

	err := g.gqlClient.Query("getRepos", &query, variables)

	return query, err
}

func (g *APIGetter) GetRepo(owner string, name string) (*RepoQuery, error) {
	query := new(RepoQuery)
	variables := map[string]interface{}{
		"owner": graphql.String(owner),
		"name":  graphql.String(name),
	}

	err := g.gqlClient.Query("getRepo", &query, variables)
	return query, err
}

type SecretsResponse struct {
	TotalCount int      `json:"total_count"`
	Secrets    []Secret `json:"secrets"`
}

type Secret struct {
	Name          string    `json:"name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Visibility    string    `json:"visibility"`
	SelectedRepos string    `json:"selected_repositories_url"`
}

type ScopedSecretsResponse struct {
	TotalCount   int                `json:"total_count"`
	Repositories []ScopedRepository `json:"repositories"`
}

type ScopedRepository struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
