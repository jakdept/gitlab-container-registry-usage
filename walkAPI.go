package main

// var (
// 	url    = kingpin.Flag("url", "target url").Required().String()
// 	method = kingpin.Flag("method", "HTTP method").Default("GET").Enum("GET", "POST", "PUT", "DELETE").String()
// 	sleep  = kingpin.Flag("sleep", "Sleep between requests").Default("100").Milliseconds()

// 	followGitlabNext = kingpin.Flag("gitlab-next", "Follow the Gitlab 'next' header").Bool()
// 	authToken        = kingpin.Flag("gitlab-token", "Gitlab API token").Envar("GITLAB_API_TOKEN").Default("").String()
// )

// func runRequest(method, url, body string) {
// 	time.Sleep(time.Microsecond * 100)
// 	req, err := http.NewRequest(method, url, strings.NewReader(body))
// 	if err != nil {
// 		log.Fatalf("failed to create request - %w", err)
// 	}

// 	if authToken != nil && *authToken != "" {
// 		req.Header.Set("Authorization", "Bearer "+*authToken)
// 	}
// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		log.Fatalf("request failed - %w", err)
// 	}
// 	if resp.StatusCode != 200 {
// 		log.Fatalf("bad response - %w", err)
// 	}
// 	if err = resp.Write(os.Stdout); err != nil {
// 		log.Fatalf("cannot write body - %w", err)
// 	}

// 	next := nextGitlabPage(resp.Header.Get("link"))
// 	if next != nil {

// 	}

// }

// func main() {
// 	time.Sleep(time.Microsecond * 100)
// 	req, err := http.NewRequest(*method, *url, os.Stdin)
// 	if err != nil {
// 		log.Fatalf("failed to create request - %w", err)
// 	}

// 	if authToken != nil && *authToken != "" {
// 		req.Header.Set("Authorization", "Bearer "+*authToken)
// 	}
// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		log.Fatalf("request failed - %w", err)
// 	}
// 	if resp.StatusCode != 200 {
// 		log.Fatalf("bad response - %w", err)
// 	}
// 	if err = resp.Write(os.Stdout); err != nil {
// 		log.Fatalf("cannot write body - %w", err)
// 	}

// 	next := nextGitlabPage(resp.Header.Get("link"))
// 	if next != nil {

// 	}
// }
