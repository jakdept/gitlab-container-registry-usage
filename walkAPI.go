package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
)

var (
	url    = kingpin.Flag("url", "target url").Required().String()
	method = kingpin.Flag("method", "HTTP method").Default("GET").Enum("GET", "POST", "PUT", "DELETE").String()

	followGitlabNext = kingpin.Flag("gitlab-next", "Follow the Gitlab 'next' header").Bool()
	authToken        = kingpin.Flag("gitlab-token", "Gitlab API token").Envar("GITLAB_API_TOKEN").Default("").String()
)

func nextGitlabPage(header string) string {
	for _, link := range strings.Split(header, ",") {
		parts := strings.Split(link, ";")
		if len(parts) == 2 && strings.TrimSpace(parts[1]) == "rel=\"next\"" {
			return strings.Trim(strings.TrimSpace(parts[0]), " <>")
		}
	}
	return ""
}
func main() {
	time.Sleep(time.Microsecond * 100)
	req, err := http.NewRequest(*method, *url, os.Stdin)
	if err != nil {
		log.Fatalf("failed to create request - %w", err)
	}

	if authToken != nil && *authToken != "" {
		req.Header.Set("Authorization", "Bearer "+*authToken)
	}

	http.DefaultClient.Do(req)
}
