package main

import (
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"flag"
	"os/exec"
	"sync"
)
//Configuration Application
type Config struct {
	token string
	organization string
	clone bool
}

var config = new(Config)

func init(){
	flag.StringVar(&config.token, "token", "", "The token (https://github.com/settings/tokens/new)")
	flag.StringVar(&config.organization, "organization", "", "The organization")
	flag.BoolVar(&config.clone, "clone", false, "True we clone in current directory")
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
		wg := new(sync.WaitGroup)

		for _,repo := range repos {
			fmt.Printf("%s\t%s\n", *repo.CloneURL,*repo.GitURL)
			if(config.clone){
			 clone(*repo.SSHURL, wg)
			}
		}
		wg.Wait()
	}

}

func clone(cloneURL string, wg *sync.WaitGroup)  {
	wg.Add(1)
	cmd := exec.Command("git", "clone", cloneURL)
	err := cmd.Start()
	if err != nil {
		fmt.Printf("error: %v\n\n", err)
	}
	go func() {
		err := cmd.Wait()
		if err != nil {
			fmt.Printf("error: %v\n\n", err)
		}
		wg.Done()
		fmt.Printf("%s is done\n", cloneURL)
	}()

}
