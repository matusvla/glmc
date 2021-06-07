package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/matusvla/glmc/glconnector"
)

const gqlAddrFmt = "https://%s/"

func init() {
	flag.StringVar(&cliInput.destinationDir, "d", "", "root directory where the repos should be cloned")
	flag.StringVar(&cliInput.authToken, "t", "", "authentication token to connect to GitLab")
	flag.StringVar(&cliInput.gitLabAddr, "a", "https://gitlab.com", "GitLab instance address")
	flag.StringVar(&cliInput.groupName, "g", "", "GitLab group name")
	flag.BoolVar(&cliInput.verbose, "v", false, "verbose output of git")
	flag.BoolVar(&cliInput.quiet, "q", false, "quieter output")
}

type cliFlags struct {
	destinationDir string
	authToken      string
	gitLabAddr     string
	groupName      string
	verbose        bool
	quiet          bool
}

func (f *cliFlags) validate() ([]string, bool) {
	valid := true
	var invalidMsgs []string
	if f.destinationDir == "" {
		valid = false
		invalidMsgs = append(invalidMsgs, "Missing a mandatory value for CLI flag -d")
	}
	if f.authToken == "" { // todo no repetition for mandatory flags
		valid = false
		invalidMsgs = append(invalidMsgs, "Missing a mandatory value for CLI flag -t")
	}
	if f.groupName == "" { // todo no repetition for mandatory flags
		valid = false
		invalidMsgs = append(invalidMsgs, "Missing a mandatory value for CLI flag -g")
	}
	return invalidMsgs, valid
}

var cliInput cliFlags

func main() {
	flag.Parse()
	invalidMsgs, ok := cliInput.validate()
	if !ok {
		for _, msg := range invalidMsgs {
			fmt.Println(msg)
		}
		os.Exit(1)
	}

	repoURLs, err := glconnector.GetRepoList(cliInput.gitLabAddr, cliInput.groupName, cliInput.authToken)
	if err != nil {
		fmt.Printf("Unable to get the list of repositories, error: %v", err)
		return
	}
	if len(repoURLs) == 0 {
		fmt.Println("Did not find any repositories in the specified group")
		return
	}
	fmt.Printf("Found %v repositories\nCloning repositories...\n", len(repoURLs))

	var failures, processed int32
	var workerWg sync.WaitGroup

	gitLabAddr := strings.TrimPrefix(cliInput.gitLabAddr, "http://")
	gitLabAddr = strings.TrimPrefix(gitLabAddr, "https://")
	repoURLsCh := make(chan string)
	const workers = 8
	for i := 0; i < workers; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for repoURL := range repoURLsCh {

				destinationPath := path.Join(cliInput.destinationDir, strings.TrimPrefix(repoURL, "https://"+gitLabAddr))
				if !cliInput.quiet {
					fmt.Printf("(%d/%d) Cloning %s to %s\n", atomic.LoadInt32(&processed), len(repoURLs), repoURL, destinationPath)
				} else {
					fmt.Printf("\r (%d/%d)", atomic.LoadInt32(&processed), len(repoURLs))
				}

				err := cloneRepo(repoURL, destinationPath, cliInput.verbose && !cliInput.quiet)
				if err != nil {
					if !cliInput.quiet { // todo this could be done in a nicer way - as logging levels
						fmt.Printf("failed with error %v\n", err)
					}
					atomic.AddInt32(&failures, 1)
				}
				atomic.AddInt32(&processed, 1)
			}
		}()
	}

	for _, repoURL := range repoURLs {
		repoURLsCh <- repoURL
	}
	close(repoURLsCh)
	workerWg.Wait()
	if cliInput.quiet {
		fmt.Printf("\r (%d/%d)\n", len(repoURLs), len(repoURLs))
	}
	fmt.Printf("Successfully cloned %v repos, %v failed\n", processed-failures, failures)
}

func cloneRepo(repoURL, destinationPath string, verbose bool) error {
	args := []string{"clone", repoURL, destinationPath}
	if !verbose {
		args = append(args, "-q")
	}
	cmd := exec.Command("git", args...)
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stdout
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}
