package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	//DotFile The file in HOME user
	DotFile = ".github-command-line"
)

//Config Configuration Application
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
	runtime.GOMAXPROCS(runtime.NumCPU())

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)

	fmt.Println("Name\t\t\t\t\tclone url")

	var wg sync.WaitGroup
	var err error
	var page int
	for page, err = 1, nil; page != 0; {
		page, err = listRepos(&wg, client, page)
		if err != nil {
			fmt.Printf("%s", err)
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
	}
	//clone all repositories
	for _, repo := range repos {
		p := Project{name: *repo.Name, cloneURL: *repo.SSHURL}
		fmt.Printf("%q\t\t\t\t\t%q\n", p.name, p.cloneURL)
		if config.clone {
			p.clone(wg)
		}
	}
	return response.NextPage, nil
}

//Project the name and url to clone
type Project struct {
	name     string
	cloneURL string
}

func (p Project) clone(wg *sync.WaitGroup) {

	d := filepath.Join(config.directory, p.name)
	if _, err := os.Stat(d); os.IsNotExist(err) {
		wg.Add(1)
		cmd := exec.Command("git", "clone", p.cloneURL, d)

		go func(wg *sync.WaitGroup) {
			errCmd := cmd.Start()
			defer wg.Done()
			if errCmd != nil {
				fmt.Printf("error: %v\n\n", errCmd)
			}
			err := cmd.Wait()
			if err != nil {
				fmt.Printf("%s", err)
			}
			fmt.Printf("%q\t\t\t\t\t%q done\n", p.name, p.cloneURL)

		}(wg)
	} else {
		fmt.Printf("%q\t\t\t\t\t%q exists\n", p.name, p.cloneURL)
		//TODO stash and pull
	}
}
