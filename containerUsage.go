package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.uber.org/ratelimit"
)

// curl --header "Authorization: Bearer $GITLAB_API_TOKEN" https://git.liquidweb.com/api/v4/groups|jq|bat

func nextGitlabPage(header string) string {
	for _, link := range strings.Split(header, ",") {
		parts := strings.Split(link, ";")
		if len(parts) == 2 && strings.TrimSpace(parts[1]) == "rel=\"next\"" {
			return strings.Trim(strings.TrimSpace(parts[0]), " <>")
		}
	}
	return ""
}

type gitlabEndpoint struct {
	baseurl   string
	authtoken string
	rl        ratelimit.Limiter
}

func NewGitlabEndpoint(baseurl, authtoken string, reqPerSec int) *gitlabEndpoint {
	if reqPerSec > 100 {
		reqPerSec = 100
	}
	return &gitlabEndpoint{
		baseurl:   strings.TrimRight(baseurl, "/") + "/api/v4",
		authtoken: authtoken,
		rl:        ratelimit.New(reqPerSec),
	}
}

type GitlabGroup struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	Visibility string `json:"visibility"`
}

func (gitlab *gitlabEndpoint) ListGroups() (<-chan GitlabGroup, <-chan error) {

	next := gitlab.baseurl + "/groups"
	errChan := make(chan error)
	groupChan := make(chan GitlabGroup)
	var groups, respBody []GitlabGroup
	var err error

	gitlabGroupRequest := func(url string) ([]GitlabGroup, string, error) {
		respBody := respBody[:0]
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, "", err
		}
		req.Header.Add("Authorization", "Bearer "+gitlab.authtoken)

		gitlab.rl.Take()
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			errChan <- err
			return nil, "", err
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			errChan <- err
			return nil, "", err
		}
		return respBody, nextGitlabPage(resp.Header.Get("link")), nil
	}

	go func() {
		defer close(groupChan)
		defer close(errChan)
		for next != "" {
			groups, next, err = gitlabGroupRequest(next)
			if err != nil {
				errChan <- err
				return
			}
			for _, group := range groups {
				groupChan <- group
			}
		}
	}()

	return groupChan, errChan
}

func (gitlab *gitlabEndpoint) ListRepos() (<-chan GitlabGroup, <-chan error) {

	next := gitlab.baseurl + "/groups"
	errChan := make(chan error)
	groupChan := make(chan GitlabGroup)
	var groups, respBody []GitlabGroup
	var err error

	gitlabGroupRequest := func(url string) ([]GitlabGroup, string, error) {
		respBody := respBody[:0]
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, "", err
		}
		req.Header.Add("Authorization", "Bearer "+gitlab.authtoken)

		gitlab.rl.Take()
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			errChan <- err
			return nil, "", err
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			errChan <- err
			return nil, "", err
		}
		return respBody, nextGitlabPage(resp.Header.Get("link")), nil
	}

	go func() {
		defer close(groupChan)
		defer close(errChan)
		for next != "" {
			groups, next, err = gitlabGroupRequest(next)
			if err != nil {
				errChan <- err
				return
			}
			for _, group := range groups {
				groupChan <- group
			}
		}
	}()

	return groupChan, errChan
}
