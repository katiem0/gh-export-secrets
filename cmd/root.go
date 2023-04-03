package cmd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	gh "github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/katiem0/gh-export-secrets/internal/data"
	"github.com/spf13/cobra"
)

type cmdFlags struct {
	all        bool
	actions    bool
	dependabot bool
	codespaces bool
	reportFile string
}

func NewCmd() *cobra.Command {
	var repository string
	cmdFlags := cmdFlags{}

	cmd := cobra.Command{
		Use:   "gh export-secrets [flags] <organization> [repository] ",
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
			if len(args) > 1 {
				repository = args[1]
			}

			if _, err := os.Stat(cmdFlags.reportFile); errors.Is(err, os.ErrExist) {
				return err
			}

			reportWriter, err := os.OpenFile(cmdFlags.reportFile, os.O_WRONLY|os.O_CREATE, 0644)

			if err != nil {
				return err
			}

			return runCmd(owner, repository, &cmdFlags, data.NewAPIGetter(gqlClient, restClient), reportWriter)
		},
	}

	// Determine default report file based on current timestamp; for more info see https://pkg.go.dev/time#pkg-constants
	reportFileDefault := fmt.Sprintf("report-%s.csv", time.Now().Format("20060102150405"))

	// Configure flags for command
	cmd.PersistentFlags().BoolVarP(&cmdFlags.all, "all", "a", false, "Whether to retrieve all secrets types")
	cmd.PersistentFlags().BoolVarP(&cmdFlags.actions, "actionsSecrets", "b", false, "Whether to retrieve Actions secrets")
	cmd.PersistentFlags().BoolVarP(&cmdFlags.dependabot, "dependabotSecrets", "d", false, "Whether to retrieve Dependabot secrets")
	cmd.PersistentFlags().BoolVarP(&cmdFlags.codespaces, "codespacesSecrets", "c", false, "Whether to retrieve Codespaces secrets")
	cmd.Flags().StringVarP(&cmdFlags.reportFile, "output-file", "o", reportFileDefault, "Name of file to write CSV report")

	return &cmd
}

func runCmd(owner string, repository string, cmdFlags *cmdFlags, g *data.APIGetter, reportWriter io.Writer) error {
	var reposCursor *string
	var allRepos []data.RepoInfo

	csvWriter := csv.NewWriter(reportWriter)

	err := csvWriter.Write([]string{
		"SecretLevel",
		"SecretType",
		"SecretName",
		"SecretAccess",
		"RepositoryName",
		"RepositoryID",
	})

	if err != nil {
		return err
	}

	if len(repository) > 0 {
		repoQuery, err := g.GetRepo(owner, repository)
		if err != nil {
			return err
		}
		allRepos = append(allRepos, repoQuery.Repository)
	} else {
		// Prepare writer for outputting report
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
	}

	// Writing to CSV Org level Actions secrets
	if len(repository) == 0 && (cmdFlags.actions || cmdFlags.all) {
		orgSecrets, err := g.GetOrgActionSecrets(owner)
		if err != nil {
			return err
		}
		var oActionResponseObject data.SecretsResponse
		json.Unmarshal(orgSecrets, &oActionResponseObject)

		for _, orgSecret := range oActionResponseObject.Secrets {
			if orgSecret.Visibility == "selected" {
				scoped_repo, err := g.GetScopedOrgActionSecrets(owner, orgSecret.Name)
				if err != nil {
					return err
				}
				var responseOObject data.ScopedSecretsResponse
				json.Unmarshal(scoped_repo, &responseOObject)
				for _, scopeSecret := range responseOObject.Repositories {
					err = csvWriter.Write([]string{
						"Organization",
						"Actions",
						orgSecret.Name,
						orgSecret.Visibility,
						scopeSecret.Name,
						strconv.Itoa(scopeSecret.ID),
					})
					if err != nil {
						return err
					}
				}
			} else if orgSecret.Visibility == "private" {
				for _, repoActPrivateSecret := range allRepos {
					if repoActPrivateSecret.Visibility != "public" {
						err = csvWriter.Write([]string{
							"Organization",
							"Actions",
							orgSecret.Name,
							orgSecret.Visibility,
							repoActPrivateSecret.Name,
							strconv.Itoa(repoActPrivateSecret.DatabaseId),
						})
						if err != nil {
							return err
						}
					}
				}
			} else {
				err = csvWriter.Write([]string{
					"Organization",
					"Actions",
					orgSecret.Name,
					orgSecret.Visibility,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	// Writing to CSV Org level Dependabot secrets
	if len(repository) == 0 && (cmdFlags.dependabot || cmdFlags.all) {

		orgDepSecrets, err := g.GetOrgDependabotSecrets(owner)
		if err != nil {
			return err
		}
		var oDepResponseObject data.SecretsResponse
		json.Unmarshal(orgDepSecrets, &oDepResponseObject)
		//fmt.Println(responseObject.Secrets)

		for _, orgDepSecret := range oDepResponseObject.Secrets {
			if orgDepSecret.Visibility == "selected" {
				scoped_repo, err := g.GetScopedOrgDependabotSecrets(owner, orgDepSecret.Name)
				if err != nil {
					return err
				}
				var rDepResponseObject data.ScopedSecretsResponse
				json.Unmarshal(scoped_repo, &rDepResponseObject)
				for _, depScopeSecret := range rDepResponseObject.Repositories {
					err = csvWriter.Write([]string{
						"Organization",
						"Dependabot",
						orgDepSecret.Name,
						orgDepSecret.Visibility,
						depScopeSecret.Name,
						strconv.Itoa(depScopeSecret.ID),
					})
					if err != nil {
						return err
					}
				}
			} else if orgDepSecret.Visibility == "private" {
				for _, repoPrivateSecret := range allRepos {
					if repoPrivateSecret.Visibility != "public" {
						err = csvWriter.Write([]string{
							"Organization",
							"Dependabot",
							orgDepSecret.Name,
							orgDepSecret.Visibility,
							repoPrivateSecret.Name,
							strconv.Itoa(repoPrivateSecret.DatabaseId),
						})
						if err != nil {
							return err
						}
					}
				}
			} else {
				err = csvWriter.Write([]string{
					"Organization",
					"Dependabot",
					orgDepSecret.Name,
					orgDepSecret.Visibility,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	// Writing to CSV Org level Codespaces secrets
	if len(repository) == 0 && (cmdFlags.codespaces || cmdFlags.all) {

		orgCodeSecrets, err := g.GetOrgCodespacesSecrets(owner)
		if err != nil {
			return err
		}
		var oCodeResponseObject data.SecretsResponse
		json.Unmarshal(orgCodeSecrets, &oCodeResponseObject)
		//fmt.Println(responseObject.Secrets)

		for _, orgCodeSecret := range oCodeResponseObject.Secrets {
			if orgCodeSecret.Visibility == "selected" {
				scoped_repo, err := g.GetScopedOrgCodespacesSecrets(owner, orgCodeSecret.Name)
				if err != nil {
					return err
				}
				var rCodeResponseObject data.ScopedSecretsResponse
				json.Unmarshal(scoped_repo, &rCodeResponseObject)
				for _, codeScopeSecret := range rCodeResponseObject.Repositories {
					err = csvWriter.Write([]string{
						"Organization",
						"Codespaces",
						orgCodeSecret.Name,
						orgCodeSecret.Visibility,
						codeScopeSecret.Name,
						strconv.Itoa(codeScopeSecret.ID),
					})
					if err != nil {
						return err
					}
				}
			} else if orgCodeSecret.Visibility == "private" {
				for _, repoCodePrivateSecret := range allRepos {
					if repoCodePrivateSecret.Visibility != "public" {
						err = csvWriter.Write([]string{
							"Organization",
							"Codespaces",
							orgCodeSecret.Name,
							orgCodeSecret.Visibility,
							repoCodePrivateSecret.Name,
							strconv.Itoa(repoCodePrivateSecret.DatabaseId),
						})
						if err != nil {
							return err
						}
					}
				}
			} else {
				err = csvWriter.Write([]string{
					"Organization",
					"Codespaces",
					orgCodeSecret.Name,
					orgCodeSecret.Visibility,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	// Writing to CSV repository level Secrets
	for _, singleRepo := range allRepos {
		// Writing to CSV repository level Actions secrets
		if cmdFlags.actions || cmdFlags.all {
			repoActionSecretsList, err := g.GetRepoActionSecrets(owner, singleRepo.Name)
			if err != nil {
				return err
			}
			var repoActionResponseObject data.SecretsResponse
			json.Unmarshal(repoActionSecretsList, &repoActionResponseObject)
			for _, repoActionsSecret := range repoActionResponseObject.Secrets {
				err = csvWriter.Write([]string{
					"Repository",
					"Actions",
					repoActionsSecret.Name,
					"RepoOnly",
					singleRepo.Name,
					strconv.Itoa(singleRepo.DatabaseId),
				})
				if err != nil {
					return err
				}
			}
		}
		// Writing to CSV repository level Dependabot secrets
		if cmdFlags.dependabot || cmdFlags.all {
			repoDepSecretsList, err := g.GetRepoDependabotSecrets(owner, singleRepo.Name)
			if err != nil {
				return err
			}
			var repoDepResponseObject data.SecretsResponse
			json.Unmarshal(repoDepSecretsList, &repoDepResponseObject)
			for _, repoDepSecret := range repoDepResponseObject.Secrets {
				err = csvWriter.Write([]string{
					"Repository",
					"Dependabot",
					repoDepSecret.Name,
					"RepoOnly",
					singleRepo.Name,
					strconv.Itoa(singleRepo.DatabaseId),
				})
				if err != nil {
					return err
				}
			}
		}
		// Writing to CSV repository level Codespaces secrets
		if cmdFlags.codespaces || cmdFlags.all {
			repoCodeSecretsList, err := g.GetRepoCodespacesSecrets(owner, singleRepo.Name)
			if err != nil {
				return err
			}
			var repoCodeResponseObject data.SecretsResponse
			json.Unmarshal(repoCodeSecretsList, &repoCodeResponseObject)
			for _, repoCodeSecret := range repoCodeResponseObject.Secrets {
				err = csvWriter.Write([]string{
					"Repository",
					"Codespaces",
					repoCodeSecret.Name,
					"RepoOnly",
					singleRepo.Name,
					strconv.Itoa(singleRepo.DatabaseId),
				})
				if err != nil {
					return err
				}
			}
		}
	}

	csvWriter.Flush()

	return nil

}
