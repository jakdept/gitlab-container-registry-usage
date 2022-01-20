package main

import (
	"context"
	"log"

	"github.com/alecthomas/kingpin"
	"github.com/davecgh/go-spew/spew"

	"github.com/jakdept/gitlab-container-registry-usage/lib/gitlab"
)

// curl --header "Authorization: Bearer $GITLAB_API_TOKEN" https://git.liquidweb.com/api/v4/groups|jq|bat

var (
	url  = kingpin.Flag("url", "target url").Required().String()
	freq = kingpin.Flag("freq", "requests per second (hz)").Default("10").Float64()

	authToken = kingpin.Flag("gitlab-token", "Gitlab API token").Envar("GITLAB_API_TOKEN").Default("").String()
)

func main() {
	ctx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	gitlab := gitlab.NewGitlabEndpoint(*url, *authToken, *freq)
	groups, err := gitlab.ListGroups(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	spew.Dump(groups)

}
