package main

import (
	"context"
	"log"

	"github.com/alecthomas/kingpin"
	"github.com/fatih/color"

	"github.com/jakdept/gitlab-container-registry-usage/lib/gitlab"
)

// curl --header "Authorization: Bearer $GITLAB_API_TOKEN" https://git.liquidweb.com/api/v4/groups|jq|bat

var (
	url  = kingpin.Flag("url", "target url").Required().String()
	freq = kingpin.Flag("freq", "requests per second (hz)").Default("10").Float64()

	authToken = kingpin.Flag("gitlab-token", "Gitlab API token").Envar("GITLAB_API_TOKEN").Default("").String()
)

func main() {
	kingpin.Parse()

	ctx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	endpoint := gitlab.NewGitlabEndpoint(*url, *authToken, *freq)
	groups, err := endpoint.ListGroups(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	// color.Blue(spew.Sdump(groups))
	var groupTotal, registryTotal int

	for _, group := range groups {
		groupTotal = 0
		regs, err := endpoint.ListRegistriesInGroup(ctx, group)
		// color.Blue(spew.Sdump(regs))
		if err != nil {
			log.Fatalln(err)
		}
		for _, reg := range regs {
			registryTotal = 0
			// color.Yellow("Registry: %s", reg.Path)
			for _, tag := range reg.Tags {
				if err := endpoint.GetRegistryTagInfo(ctx, reg, &tag); err != nil {
					log.Fatalln(err)
				}
				color.Green("\t\tTag: %s Size: %d", tag.Path, tag.TotalSize)
				registryTotal += tag.TotalSize
			}
			color.Cyan("\tProject [%s] Usage: %d", reg.Path, registryTotal)
			groupTotal += registryTotal
		}
		color.Yellow("Group [%s] total: %d", group.Path, groupTotal)
	}
}
