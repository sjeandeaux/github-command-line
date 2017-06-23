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
	wg := new(sync.WaitGroup)

	fmt.Println("Name\tSSHURL")

	var err error
	var page int
	for page, err = 1, nil; page != 0; {
		page, err = listRepos(client, wg, page)
		if err != nil {
			log.Fatal(err)
		}
	}
	wg.Wait()

}

func listRepos(client *github.Client, wg *sync.WaitGroup, page int) (int, error) {
	var opt *github.RepositoryListByOrgOptions
	if page > 0 {
		opt = &github.RepositoryListByOrgOptions{"all", github.ListOptions{Page: page}}
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
				clone(*repo.Name, *repo.SSHURL, wg)
			}
		}
		return response.NextPage, nil
	}
}

func clone(name string, cloneURL string, wg *sync.WaitGroup) {

	d := filepath.Join(config.directory, name)
	//TODO check directory exists, permission

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
