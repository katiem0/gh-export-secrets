package data

import (
	"fmt"
	"io/ioutil"
	"log"
)

func (g *APIGetter) GetOrgDependabotSecrets(owner string) ([]byte, error) {
	url := fmt.Sprintf("orgs/%s/dependabot/secrets", owner)

	resp, err := g.restClient.Request("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	return responseData, err
}

func (g *APIGetter) GetRepoDependabotSecrets(owner string, repo string) ([]byte, error) {
	url := fmt.Sprintf("repos/%s/%s/dependabot/secrets", owner, repo)

	resp, err := g.restClient.Request("GET", url, nil)

	responseData, err := ioutil.ReadAll(resp.Body)
	return responseData, err
}

func (g *APIGetter) GetScopedOrgDependabotSecrets(owner string, secret string) ([]byte, error) {
	url := fmt.Sprintf("orgs/%s/dependabot/secrets/%s/repositories", owner, secret)

	resp, err := g.restClient.Request("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	return responseData, err
}
