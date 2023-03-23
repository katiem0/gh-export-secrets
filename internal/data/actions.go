package data

import (
	"fmt"
	"io/ioutil"
	"log"
)

func (g *APIGetter) GetOrgActionSecrets(owner string) ([]byte, error) {
	url := fmt.Sprintf("orgs/%s/actions/secrets", owner)

	resp, err := g.restClient.Request("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	return responseData, err
}

func (g *APIGetter) GetRepoActionSecrets(owner string, repo string) ([]byte, error) {
	url := fmt.Sprintf("repos/%s/%s/actions/secrets", owner, repo)

	resp, err := g.restClient.Request("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	return responseData, err
}

func (g *APIGetter) GetScopedOrgActionSecrets(owner string, secret string) ([]byte, error) {
	url := fmt.Sprintf("orgs/%s/actions/secrets/%s/repositories", owner, secret)

	resp, err := g.restClient.Request("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	return responseData, err
}
