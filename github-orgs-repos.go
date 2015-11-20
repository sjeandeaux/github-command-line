package main

import (
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"flag"
)
//Configuration Application
type Config struct {
	token string
	organization string
}

var config = new(Config)

func init(){
	flag.StringVar(&config.token, "token", "", "The token (https://github.com/settings/tokens/new)")
	flag.StringVar(&config.organization, "organization", "", "The organization")
	flag.Parse()
	//TODO validate
}

func main() {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)

	repos, _, err := client.Repositories.ListByOrg(config.organization, nil)
	if err != nil {
		fmt.Printf("error: %v\n\n", err)
	} else {

		fmt.Println("CloneURL\tGitURL")
		for _,repo := range repos {
			fmt.Printf("%s\t%s", *repo.CloneURL,*repo.GitURL)
			fmt.Println()

		}
	}
}
