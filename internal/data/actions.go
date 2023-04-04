package data

import (
	"fmt"
	"io"
	"log"
)

func (g *APIGetter) GetOrgActionSecrets(owner string) ([]byte, error) {
	url := fmt.Sprintf("orgs/%s/actions/secrets", owner)

	resp, err := g.restClient.Request("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return responseData, err
}

func (g *APIGetter) GetRepoActionSecrets(owner string, repo string) ([]byte, error) {
	url := fmt.Sprintf("repos/%s/%s/actions/secrets", owner, repo)

	resp, err := g.restClient.Request("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return responseData, err
}

func (g *APIGetter) GetScopedOrgActionSecrets(owner string, secret string) ([]byte, error) {
	url := fmt.Sprintf("orgs/%s/actions/secrets/%s/repositories", owner, secret)

	resp, err := g.restClient.Request("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return responseData, err
}
