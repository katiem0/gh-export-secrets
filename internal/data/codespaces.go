package data

import (
	"fmt"
	"io/ioutil"
	"log"
)

func (g *APIGetter) GetOrgCodespacesSecrets(owner string) ([]byte, error) {
	url := fmt.Sprintf("orgs/%s/codespaces/secrets", owner)

	resp, err := g.restClient.Request("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	return responseData, err
}

func (g *APIGetter) GetRepoCodespacesSecrets(owner string, repo string) ([]byte, error) {
	url := fmt.Sprintf("repos/%s/%s/codespaces/secrets", owner, repo)

	resp, err := g.restClient.Request("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	return responseData, err
}

func (g *APIGetter) GetScopedOrgCodespacesSecrets(owner string, secret string) ([]byte, error) {
	url := fmt.Sprintf("orgs/%s/codespaces/secrets/%s/repositories", owner, secret)

	resp, err := g.restClient.Request("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	return responseData, err
}
