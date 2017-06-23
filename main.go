package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	DotFile = ".github-command-line"
)

//Configuration Application
type Config struct {
	Token        string
	organization string
	clone        bool
	directory    string
}

var config = new(Config)

func init() {
	flag.StringVar(&config.Token, "token", "", "The token (https://github.com/settings/tokens/new)")
	flag.StringVar(&config.organization, "organization", "", "The organization")
	flag.BoolVar(&config.clone, "clone", false, "True we clone in current directory")
	flag.StringVar(&config.directory, "directory", "", "Where do we clone")
	flag.Parse()
	//TODO validate
	readToken()
}

//read token from HOME/.github-command-line if -token is empty
func readToken() {
	if config.Token == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		value, _ := ioutil.ReadFile(filepath.Join(usr.HomeDir, DotFile))
		var tmpConfig Config
		json.Unmarshal(value, &tmpConfig)
		config.Token = tmpConfig.Token
	}
}

func main() {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)

	fmt.Println("Name\tSSHURL")

	var wg sync.WaitGroup
	var err error
	var page int
	for page, err = 1, nil; page != 0; {
		page, err = listRepos(&wg, client, page)
		if err != nil {
			fmt.Printf("error")
		}
	}
	wg.Wait()
}

func listRepos(wg *sync.WaitGroup, client *github.Client, page int) (int, error) {
	var opt *github.RepositoryListByOrgOptions
	if page > 0 {
		opt = &github.RepositoryListByOrgOptions{Type: "all", ListOptions: github.ListOptions{Page: page}}
	} else {
		opt = nil
	}

	repos, response, err := client.Repositories.ListByOrg(context.Background(), config.organization, opt)
	if err != nil {
		return 0, err
	} else {
		for _, repo := range repos {
			fmt.Printf("%s\t%s\n", *repo.Name, *repo.SSHURL)
			if config.clone {
				clone(wg, *repo.Name, *repo.SSHURL)
			}
		}
		return response.NextPage, nil
	}
}

func clone(wg *sync.WaitGroup, name string, cloneURL string) {

	wg.Add(1)
	d := filepath.Join(config.directory, name)
	//TODO check directory exists, permission

	cmd := exec.Command("git", "clone", cloneURL, d)

	go func(wg *sync.WaitGroup) {
		errCmd := cmd.Start()
		defer wg.Done()
		if errCmd != nil {
			fmt.Printf("error: %v\n\n", errCmd)
		}
		err := cmd.Wait()
		if err != nil {
			fmt.Printf("error")
		}
		fmt.Printf("%s\t%s done\n", name, cloneURL)

	}(wg)

}
