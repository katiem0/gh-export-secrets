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

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/katiem0/gh-export-secrets/internal/data"
	"github.com/katiem0/gh-export-secrets/internal/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type cmdFlags struct {
	app        string
	hostname   string
	token      string
	reportFile string
	debug      bool
}

func NewCmd() *cobra.Command {
	//var repository string
	cmdFlags := cmdFlags{}
	var authToken string

	cmd := cobra.Command{
		Use:   "gh export-secrets [flags] <organization> [repo ...] ",
		Short: "Generate a report of Actions, Dependabot, and Codespaces secrets for an organization and/or repositories.",
		Long:  "Generate a report of Actions, Dependabot, and Codespaces secrets for an organization and/or repositories.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var gqlClient *api.GraphQLClient
			var restClient *api.RESTClient

			// Reinitialize logging if debugging was enabled
			if cmdFlags.debug {
				logger, _ := log.NewLogger(cmdFlags.debug)
				defer logger.Sync() // nolint:errcheck
				zap.ReplaceGlobals(logger)
			}

			if cmdFlags.token != "" {
				authToken = cmdFlags.token
			} else {
				t, _ := auth.TokenForHost(cmdFlags.hostname)
				authToken = t
			}

			gqlClient, err = api.NewGraphQLClient(api.ClientOptions{
				Headers: map[string]string{
					"Accept": "application/vnd.github.hawkgirl-preview+json",
				},
				Host:      cmdFlags.hostname,
				AuthToken: authToken,
			})

			if err != nil {
				zap.S().Errorf("Error arose retrieving graphql client")
				return err
			}

			restClient, err = api.NewRESTClient(api.ClientOptions{
				Headers: map[string]string{
					"Accept": "application/vnd.github+json",
				},
				Host:      cmdFlags.hostname,
				AuthToken: authToken,
			})

			if err != nil {
				zap.S().Errorf("Error arose retrieving rest client")
				return err
			}

			owner := args[0]
			repos := args[1:]

			if _, err := os.Stat(cmdFlags.reportFile); errors.Is(err, os.ErrExist) {
				return err
			}

			reportWriter, err := os.OpenFile(cmdFlags.reportFile, os.O_WRONLY|os.O_CREATE, 0644)

			if err != nil {
				return err
			}

			return runCmd(owner, repos, &cmdFlags, data.NewAPIGetter(gqlClient, restClient), reportWriter)
		},
	}

	// Determine default report file based on current timestamp; for more info see https://pkg.go.dev/time#pkg-constants
	reportFileDefault := fmt.Sprintf("report-%s.csv", time.Now().Format("20060102150405"))

	// Configure flags for command

	cmd.PersistentFlags().StringVarP(&cmdFlags.app, "app", "a", "actions", "List secrets for a specific application or all: {all|actions|codespaces|dependabot}")
	cmd.PersistentFlags().StringVarP(&cmdFlags.token, "token", "t", "", `GitHub Personal Access Token (default "gh auth token")`)
	cmd.PersistentFlags().StringVarP(&cmdFlags.hostname, "hostname", "", "github.com", "GitHub Enterprise Server hostname")
	cmd.Flags().StringVarP(&cmdFlags.reportFile, "output-file", "o", reportFileDefault, "Name of file to write CSV report")
	cmd.PersistentFlags().BoolVarP(&cmdFlags.debug, "debug", "d", false, "To debug logging")
	//cmd.MarkPersistentFlagRequired("app")

	return &cmd
}

func runCmd(owner string, repos []string, cmdFlags *cmdFlags, g *data.APIGetter, reportWriter io.Writer) error {
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

	if len(repos) > 0 {
		zap.S().Infof("Processing repos: %s", repos)

		for _, repo := range repos {

			zap.S().Debugf("Processing %s/%s", owner, repo)

			repoQuery, err := g.GetRepo(owner, repo)
			if err != nil {
				return err
			}
			allRepos = append(allRepos, repoQuery.Repository)
		}

	} else {
		// Prepare writer for outputting report
		for {
			zap.S().Debugf("Processing list of repositories for %s", owner)
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
	if len(repos) == 0 && (cmdFlags.app == "all" || cmdFlags.app == "actions") {
		orgSecrets, err := g.GetOrgActionSecrets(owner)
		if err != nil {
			return err
		}

		var oActionResponseObject data.SecretsResponse
		err = json.Unmarshal(orgSecrets, &oActionResponseObject)
		if err != nil {
			return err
		}
		if len(oActionResponseObject.Secrets) == 0 {
			zap.S().Debugf("No org level Actions Secrets for %s", owner)
		} else {
			zap.S().Debugf("Gathering Actions Secrets for %s", owner)
		}
		for _, orgSecret := range oActionResponseObject.Secrets {
			if orgSecret.Visibility == "selected" {
				zap.S().Debugf("Gathering Actions Secrets for %s that are scoped to specific repositories", owner)
				scoped_repo, err := g.GetScopedOrgActionSecrets(owner, orgSecret.Name)
				if err != nil {
					zap.S().Error("Error raised in writing output", zap.Error(err))
				}
				var responseOObject data.ScopedSecretsResponse
				err = json.Unmarshal(scoped_repo, &responseOObject)
				if err != nil {
					return err
				}
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
						zap.S().Error("Error raised in writing output", zap.Error(err))
					}
				}
			} else if orgSecret.Visibility == "private" {
				zap.S().Debugf("Gathering Actions Secret %s for %s that is accessible to all internal and private repositories.", orgSecret.Name, owner)
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
							zap.S().Error("Error raised in writing output", zap.Error(err))
						}
					}
				}
			} else {
				zap.S().Debugf("Gathering public Actions Secret %s for %s", orgSecret.Name, owner)
				err = csvWriter.Write([]string{
					"Organization",
					"Actions",
					orgSecret.Name,
					orgSecret.Visibility,
				})
				if err != nil {
					zap.S().Error("Error raised in writing output", zap.Error(err))
				}
			}
		}
	}

	// Writing to CSV Org level Dependabot secrets
	if len(repos) == 0 && (cmdFlags.app == "all" || cmdFlags.app == "dependabot") {

		orgDepSecrets, err := g.GetOrgDependabotSecrets(owner)
		if err != nil {
			return err
		}
		var oDepResponseObject data.SecretsResponse
		err = json.Unmarshal(orgDepSecrets, &oDepResponseObject)
		if err != nil {
			return err
		}

		if len(oDepResponseObject.Secrets) == 0 {
			zap.S().Debugf("No org level Dependabot Secrets for %s", owner)
		} else {
			zap.S().Debugf("Gathering Dependabot Secrets for %s", owner)
		}

		for _, orgDepSecret := range oDepResponseObject.Secrets {
			if orgDepSecret.Visibility == "selected" {
				zap.S().Debugf("Gathering Dependabot Secret %s for %s that is scoped to specific repositories", orgDepSecret.Name, owner)
				scoped_repo, err := g.GetScopedOrgDependabotSecrets(owner, orgDepSecret.Name)
				if err != nil {
					return err
				}
				var rDepResponseObject data.ScopedSecretsResponse
				err = json.Unmarshal(scoped_repo, &rDepResponseObject)
				if err != nil {
					return err
				}
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
						zap.S().Error("Error raised in writing output", zap.Error(err))
					}
				}
			} else if orgDepSecret.Visibility == "private" {
				zap.S().Debugf("Gathering Dependabot Secret %s for %s that is accessible to all internal and private repositories.", orgDepSecret.Name, owner)
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
							zap.S().Error("Error raised in writing output", zap.Error(err))
						}
					}
				}
			} else {
				zap.S().Debugf("Gathering public Dependabot Secret %s for %s", orgDepSecret.Name, owner)
				err = csvWriter.Write([]string{
					"Organization",
					"Dependabot",
					orgDepSecret.Name,
					orgDepSecret.Visibility,
				})
				if err != nil {
					zap.S().Error("Error raised in writing output", zap.Error(err))
				}
			}
		}
	}

	// Writing to CSV Org level Codespaces secrets
	if len(repos) == 0 && (cmdFlags.app == "all" || cmdFlags.app == "codespaces") {

		orgCodeSecrets, err := g.GetOrgCodespacesSecrets(owner)
		if err != nil {
			return err
		}
		var oCodeResponseObject data.SecretsResponse
		err = json.Unmarshal(orgCodeSecrets, &oCodeResponseObject)
		if err != nil {
			return err
		}
		if len(oCodeResponseObject.Secrets) == 0 {
			zap.S().Debugf("No org level Codespaces Secrets for %s", owner)
		} else {
			zap.S().Debugf("Gathering Codespaces Secrets for %s", owner)
		}

		for _, orgCodeSecret := range oCodeResponseObject.Secrets {
			zap.S().Debugf("Gathering Codespaces Secrets for %s that are scoped to specific repositories", owner)
			if orgCodeSecret.Visibility == "selected" {
				scoped_repo, err := g.GetScopedOrgCodespacesSecrets(owner, orgCodeSecret.Name)
				if err != nil {
					return err
				}
				var rCodeResponseObject data.ScopedSecretsResponse
				err = json.Unmarshal(scoped_repo, &rCodeResponseObject)
				if err != nil {
					return err
				}
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
						zap.S().Error("Error raised in writing output", zap.Error(err))
					}
				}
			} else if orgCodeSecret.Visibility == "private" {
				zap.S().Debugf("Gathering Codespaces Secret %s for %s that is accessible to all internal and private repositories.", orgCodeSecret.Name, owner)
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
							zap.S().Error("Error raised in writing output", zap.Error(err))
						}
					}
				}
			} else {
				zap.S().Debugf("Gathering public Codespaces Secret %s for %s", orgCodeSecret.Name, owner)
				err = csvWriter.Write([]string{
					"Organization",
					"Codespaces",
					orgCodeSecret.Name,
					orgCodeSecret.Visibility,
				})
				if err != nil {
					zap.S().Error("Error raised in writing output", zap.Error(err))
				}
			}
		}
	}

	// Writing to CSV repository level Secrets
	for _, singleRepo := range allRepos {
		// Writing to CSV repository level Actions secrets
		if cmdFlags.app == "all" || cmdFlags.app == "actions" {
			repoActionSecretsList, err := g.GetRepoActionSecrets(owner, singleRepo.Name)
			if err != nil {
				return err
			}
			var repoActionResponseObject data.SecretsResponse
			err = json.Unmarshal(repoActionSecretsList, &repoActionResponseObject)
			if err != nil {
				return err
			}
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
					zap.S().Error("Error raised in writing output", zap.Error(err))
				}
			}
		}
		// Writing to CSV repository level Dependabot secrets
		if cmdFlags.app == "all" || cmdFlags.app == "dependabot" {
			repoDepSecretsList, err := g.GetRepoDependabotSecrets(owner, singleRepo.Name)
			if err != nil {
				return err
			}
			var repoDepResponseObject data.SecretsResponse
			err = json.Unmarshal(repoDepSecretsList, &repoDepResponseObject)
			if err != nil {
				return err
			}
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
					zap.S().Error("Error raised in writing output", zap.Error(err))
				}
			}
		}
		// Writing to CSV repository level Codespaces secrets
		if cmdFlags.app == "all" || cmdFlags.app == "codespaces" {
			repoCodeSecretsList, err := g.GetRepoCodespacesSecrets(owner, singleRepo.Name)
			if err != nil {
				zap.S().Error("Error raised in writing output", zap.Error(err))
			}
			var repoCodeResponseObject data.SecretsResponse
			err = json.Unmarshal(repoCodeSecretsList, &repoCodeResponseObject)
			if err != nil {
				return err
			}
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
					zap.S().Error("Error raised in writing output", zap.Error(err))
				}
			}
		}
	}

	csvWriter.Flush()

	return nil

}
