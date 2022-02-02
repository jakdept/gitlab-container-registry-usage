package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
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

type endpoint struct {
	baseurl   string
	authtoken string
	rl        *rate.Limiter
}

func NewGitlabEndpoint(baseurl, authtoken string, reqPerSec float64) *endpoint {
	if reqPerSec > 100 {
		reqPerSec = 100
	}

	baseurl = strings.TrimSuffix(baseurl, "/")
	baseurl = strings.TrimSuffix(baseurl, "api/v4")
	baseurl = strings.TrimSuffix(baseurl, "/")

	return &endpoint{
		baseurl:   strings.TrimRight(baseurl, "/") + "/api/v4",
		authtoken: authtoken,
		rl:        rate.NewLimiter(rate.Limit(reqPerSec), 1),
	}
}

type Group struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	Visibility string `json:"visibility"`
}

type ContainerRepository struct {
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

// runRequest runs a request to the Gitlab API and returns the response body
func (gitlab *endpoint) runRequest(ctx context.Context, url, method string,
	reqObj, respObj interface{}) (string, error) {

	var reqBody bytes.Buffer
	if reqObj != nil {
		if err := json.NewEncoder(&reqBody).Encode(reqObj); err != nil {
			return "", fmt.Errorf("could not encode body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, &reqBody)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+gitlab.authtoken)

	gitlab.rl.Wait(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed: %s\nreq %s %s", resp.Status, method, url)
	}

	buf := bytes.Buffer{}
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return "", fmt.Errorf("could not read body: %w", err)
	}

	if err := json.NewDecoder(&buf).Decode(&respObj); err != nil {
		return "", fmt.Errorf("could not decode body into json: %w\n%s", err, buf.String())
	}
	return nextGitlabPage(resp.Header.Get("link")), nil
}

// ListGroups lists all groups in a Gitlab install.
//
// Gitlab docs: https://docs.gitlab.com/ee/api/groups.html#list-groups
func (gitlab *endpoint) ListGroups(ctx context.Context) (groups []Group, err error) {
	next := gitlab.baseurl + "/groups"
	var newGroups []Group

	for next != "" {
		next, err = gitlab.runRequest(ctx, next, "GET", nil, &newGroups)
		if err != nil {
			err = fmt.Errorf("error listing groups: %w", err)
			return
		}
		groups = append(groups, newGroups...)
	}

	return
}

// ListRegistriesInGroup lists all container registries in a gitlab group.
//
// Gitlab docs: https://docs.gitlab.com/ee/api/container_registry.html#within-a-group
func (gitlab *endpoint) ListRegistriesInGroup(ctx context.Context, group Group,
) (repos []ContainerRepository, err error) {

	next := fmt.Sprintf("%s/%s/%d/%s",
		gitlab.baseurl, "groups", group.ID, "registry/repositories?tags=1")
	var newRepos []ContainerRepository

	for next != "" {
		next, err = gitlab.runRequest(ctx, next, "GET", nil, &newRepos)
		if err != nil {
			err = fmt.Errorf("error listing repos for group %s: %w", group.Path, err)
			return
		}
		repos = append(repos, newRepos...)
	}
	return
}

func (gitlab *endpoint) GetRegistryInfo(registry Tag) (err error) {

	return nil
}
