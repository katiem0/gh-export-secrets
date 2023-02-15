package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	gh "github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/shurcooL/graphql"
	"github.com/spf13/cobra"
)

type cmdFlags struct {
	all            bool
	actionsOnly    bool
	dependabotOnly bool
	codespacesOnly bool
	reportFile     string
}

func NewCmd() *cobra.Command {
	cmdFlags := cmdFlags{}

	cmd := cobra.Command{
		Use:   "gh-export-secrets <organization> [flags]",
		Short: "Generate a report of Actions, Dependabot, and Codespaces secrets for an organization.",
		Long:  "Generate a report of Actions, Dependabot, and Codespaces secrets for an organization.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var gqlClient api.GQLClient
			var restClient api.RESTClient

			gqlClient, err = gh.GQLClient(&api.ClientOptions{
				Headers: map[string]string{
					"Accept": "application/vnd.github.hawkgirl-preview+json",
				},
			})

			restClient, err = gh.RESTClient(&api.ClientOptions{
				Headers: map[string]string{
					"Accept": "application/vnd.github+json",
				},
			})

			owner := args[0]

			if _, err := os.Stat(cmdFlags.reportFile); errors.Is(err, os.ErrExist) {
				return err
			}

			reportWriter, err := os.OpenFile(cmdFlags.reportFile, os.O_WRONLY|os.O_CREATE, 0644)

			if err != nil {
				return err
			}

			return runCmd(owner, newAPIGetter(gqlClient, restClient), reportWriter)
		},
	}

	// Determine default report file based on current timestamp; for more info see https://pkg.go.dev/time#pkg-constants
	reportFileDefault := fmt.Sprintf("report-%s.csv", time.Now().Format("20060102150405"))

	// Configure flags for command
	cmd.PersistentFlags().BoolVarP(&cmdFlags.all, "all", "a", true, "Whether to retrieve all secrets types")
	cmd.PersistentFlags().BoolVarP(&cmdFlags.actionsOnly, "actionsOnly", "w", false, "Whether to retrieve Actions secrets only")
	cmd.PersistentFlags().BoolVarP(&cmdFlags.dependabotOnly, "dependabotOnly", "d", false, "Whether to retrieve Dependabot secrets only")
	cmd.PersistentFlags().BoolVarP(&cmdFlags.codespacesOnly, "codespacesOnly", "c", false, "Whether to retrieve Codespaces secrets only")
	cmd.Flags().StringVarP(&cmdFlags.reportFile, "output-file", "o", reportFileDefault, "Name of file to write CSV report")

	return &cmd
}

func runCmd(owner string, g Getter, reportWriter io.Writer) error {
	var reposCursor *string
	var allRepos []repoinfo

	for {
		reposQuery, err := g.GetReposList(owner, reposCursor)

		if err != nil {
			return err
		}

		allRepos = append(allRepos, reposQuery.Organization.Repositories.Nodes...)

		reposCursor = &reposQuery.Organization.Repositories.PageInfo.EndCursor

		if !reposQuery.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
	}
	resp, err := g.GetOrgActionSecrets(owner)
	if err != nil {
		return err
	}
	var responseObject response
	json.Unmarshal(resp, &responseObject)
	fmt.Println(responseObject.Secrets)
	return nil
}

type repoinfo struct {
	DatabaseId int
	Name       string
	UpdatedAt  time.Time
	Visibility string
}

type reposQuery struct {
	Organization struct {
		Repositories struct {
			TotalCount int
			Nodes      []repoinfo
			PageInfo   struct {
				EndCursor   string
				HasNextPage bool
			}
		} `graphql:"repositories(first: 100, after: $endCursor)"`
	} `graphql:"organization(login: $owner)"`
}

type Getter interface {
	GetReposList(owner string, endCursor *string) (*reposQuery, error)
	GetOrgActionSecrets(owner string) ([]byte, error)
}

type APIGetter struct {
	gqlClient  api.GQLClient
	restClient api.RESTClient
}

func newAPIGetter(gqlClient api.GQLClient, restClient api.RESTClient) *APIGetter {
	return &APIGetter{
		gqlClient:  gqlClient,
		restClient: restClient,
	}
}

func (g *APIGetter) GetReposList(owner string, endCursor *string) (*reposQuery, error) {
	query := new(reposQuery)
	variables := map[string]interface{}{
		"endCursor": (*graphql.String)(endCursor),
		"owner":     graphql.String(owner),
	}

	err := g.gqlClient.Query("getRepos", &query, variables)

	return query, err
}

type response struct {
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

func (g *APIGetter) GetOrgActionSecrets(owner string) ([]byte, error) {
	url := fmt.Sprintf("orgs/%s/actions/secrets", owner)

	resp, err := g.restClient.Request("GET", url, nil)

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return responseData, err
}
