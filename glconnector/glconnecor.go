package glconnector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type RepoResponse struct {
	Data struct {
		Group struct {
			Projects struct {
				Pageinfo struct {
					Endcursor   string `json:"endCursor"`
					Hasnextpage bool   `json:"hasNextPage"`
				} `json:"pageInfo"`
				Nodes []struct {
					Httpurltorepo string `json:"httpUrlToRepo"`
				} `json:"nodes"`
			} `json:"projects"`
		} `json:"group"`
	} `json:"data"`
}

const (
	gqlQueryFmt = `{"query": "query groupProjects{group(fullPath:\"%s\"){projects(includeSubgroups:true,after:\"%s\"){pageInfo{endCursor hasNextPage}nodes{httpUrlToRepo}}}}","variables": {}}`
	gqlAddrFmt  = "https://%s/api/graphql"
)

func GetRepoList(gitLabAddr, groupName, authToken string) ([]string, error) {
	var responses []RepoResponse
	var cursor string
	client := &http.Client{}

	gitLabAddr = strings.TrimPrefix(gitLabAddr, "http://")
	gitLabAddr = strings.TrimPrefix(gitLabAddr, "https://")
	gitLabURL := fmt.Sprintf(gqlAddrFmt, gitLabAddr)

	for {
		gqlRequest := []byte(fmt.Sprintf(gqlQueryFmt, groupName, cursor))
		body := bytes.NewBuffer(gqlRequest)

		// Create request
		req, err := http.NewRequest("POST", gitLabURL, body)
		if err != nil {
			return nil, err
		}

		// Headers
		authString := fmt.Sprintf("Bearer %s", authToken)
		req.Header.Add("Authorization", authString) // todo from CLI
		req.Header.Add("Content-Type", "application/json")

		// Fetch Request
		resp, err := client.Do(req)

		if err != nil {
			return nil, err
		}
		// Read Response Body
		var res RepoResponse
		d := json.NewDecoder(resp.Body)
		err = d.Decode(&res)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Sprintf("Unexpected status: %v", resp.Status))
		}

		responses = append(responses, res)
		pi := res.Data.Group.Projects.Pageinfo
		if !pi.Hasnextpage {
			break
		}
		cursor = pi.Endcursor
	}

	var repoURLs []string
	for _, resp := range responses {
		for _, node := range resp.Data.Group.Projects.Nodes {
			repoURLs = append(repoURLs, node.Httpurltorepo)
		}
	}
	return repoURLs, nil
}
