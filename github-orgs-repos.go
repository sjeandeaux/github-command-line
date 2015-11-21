package main

import (
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"flag"
	"os/exec"
	"sync"
    "path/filepath"
)
//Configuration Application
type Config struct {
	token string
	organization string
	clone bool
	directory string
}

var config = new(Config)

func init(){
	flag.StringVar(&config.token, "token", "", "The token (https://github.com/settings/tokens/new)")
	flag.StringVar(&config.organization, "organization", "", "The organization")
	flag.BoolVar(&config.clone, "clone", false, "True we clone in current directory")
	flag.StringVar(&config.directory, "directory", "", "Where do we clone")
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

		fmt.Println("Name\tCloneURL\tGitURL\tSSHURL")
		wg := new(sync.WaitGroup)

		for _,repo := range repos {
			fmt.Printf("%s\t%s\t%s\t%s\n", *repo.Name, *repo.CloneURL,*repo.GitURL, *repo.SSHURL)
			if(config.clone){

			 clone(*repo.Name, *repo.SSHURL, wg)
			}
		}
		wg.Wait()
	}

}

func clone(name string, cloneURL string, wg *sync.WaitGroup)  {

	d := filepath.Join(config.directory, name)
	//TODO check dirc

	cmd := exec.Command("git", "clone", cloneURL, d)
	errCmd := cmd.Start()
	if errCmd != nil {
		fmt.Printf("error: %v\n\n", errCmd)
	}
	wg.Add(1)
	go func() {
		err := cmd.Wait()
		if err != nil {
			fmt.Printf("error: %v\n\n", err)
		}
		wg.Done()
		fmt.Printf("%s is done\n", cloneURL)
	}()

}
