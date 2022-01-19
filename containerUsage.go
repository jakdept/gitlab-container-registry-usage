package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"go.uber.org/ratelimit"
)

// curl --header "Authorization: Bearer $GITLAB_API_TOKEN" https://git.liquidweb.com/api/v4/groups|jq|bat

var (
	url    = kingpin.Flag("url", "target url").Required().String()
	method = kingpin.Flag("method", "HTTP method").Default("GET").Enum("GET", "POST", "PUT", "DELETE")
	sleep  = kingpin.Flag("sleep", "Sleep between requests").Default("100").Duration()

	authToken = kingpin.Flag("gitlab-token", "Gitlab API token").Envar("GITLAB_API_TOKEN").Default("")
)

type jsonDateTime struct {
	time.Time
}

func (t *jsonDateTime) UnmarshalJSON(buf []byte) error {
	tt, err := time.Parse(time.RFC3339Nano, strings.Trim(string(buf), `"`))
	if err != nil {
		return err
	}
	t.Time = tt
	return nil
}

func (t *jsonDateTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", t.Format(time.RFC3339Nano))), nil
}

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

type Group struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	Visibility string `json:"visibility"`
}

type Repo struct {
	ID                     int64  `json:"id"`
	Name                   string `json:"name"`
	Path                   string `json:"path"`
	ProjectID              int64  `json:"project_id"`
	Location               string `json:"location"`
	CreatedAt              string `json:"created_at"`
	CleanupPolicyStartedAt string `json:"cleanup_policy_started_at"`
	TagsCount              int64  `json:"tags_count"`
	Tags                   []Tag  `json:"tags"`
}

type Tag struct {
	Name      string       `json:"name"`
	Path      string       `json:"path"`
	Location  string       `json:"location"`
	CreatedAt jsonDateTime `json:"created_at"`
	TotalSize int          `json:"total_size"`
}

func (gitlab *gitlabEndpoint) ListGroups() (groups []Group, err error) {

	next := gitlab.baseurl + "/api/v4/groups"
	var newGroups []Group

	gitlabGroupRequest := func(url string) ([]Group, string, error) {
		var respBody []Group
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, "", err
		}
		req.Header.Add("Authorization", "Bearer "+gitlab.authtoken)

		gitlab.rl.Take()
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, "", err
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return nil, "", err
		}
		return respBody, nextGitlabPage(resp.Header.Get("link")), nil
	}

	for next != "" {
		newGroups, next, err = gitlabGroupRequest(next)
		if err != nil {
			err = fmt.Errorf("error listing groups: %w", err)
			return
		}
		groups = append(groups, newGroups...)
	}

	return
}

func (gitlab *gitlabEndpoint) ListRepos(group Group) (repos []Repo, err error) {

	next := fmt.Sprintf("%s/%s/%d/%s",
		strings.TrimSuffix(gitlab.baseurl, "/"),
		"api/v4/groups",
		group.ID, "registry/repositories?tags=1")
	var newRepos []Repo

	gitlabGroupRequest := func(url string) ([]Repo, string, error) {
		var respBody []Repo
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, "", err
		}
		req.Header.Add("Authorization", "Bearer "+gitlab.authtoken)

		gitlab.rl.Take()
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, "", err
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return nil, "", err
		}
		return respBody, nextGitlabPage(resp.Header.Get("link")), nil
	}

	for next != "" {
		newRepos, next, err = gitlabGroupRequest(next)
		if err != nil {
			err = fmt.Errorf("error listing repos for group %s: %w", group.Path, err)
			return
		}
		repos = append(repos, newRepos...)
	}
	return
}
