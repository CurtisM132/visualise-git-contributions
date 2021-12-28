package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"CurtisM132/main/utils"
)

const gitRepoPathsFile = "repos.txt"
const commitCutOffDays = 180

const (
	ansiBlack      string = "\033[0;37;30m"
	ansiLightGreen string = "\033[1;30;47m"
	ansiGreen      string = "\033[1;30;43m"
	ansiDarkGreen  string = "\033[1;30;42m"
)

type Repos struct {
	logger *zap.SugaredLogger
}

func NewGitRepos(logger *zap.SugaredLogger) *Repos {
	return &Repos{
		logger: logger,
	}
}

func (r *Repos) AddAllReposInFolder(folder string) error {
	var gitRepos []string
	r.findGitReposInPath(folder, &gitRepos)

	if len(gitRepos) > 0 {
		err := r.storeRepoListInFS(gitRepos)
		if err != nil {
			return fmt.Errorf("failed to store GIT repos to file system: %s", err)
		}
	}

	return nil
}

func (r *Repos) VisualiseGitCommits(authorEmail string) error {
	c, err := r.getGitCommits(authorEmail)
	if err != nil {
		return fmt.Errorf("failed to get GIT contributions: %s", err)
	}

	r.printGitCommits(c)

	return nil
}

func (r *Repos) findGitReposInPath(path string, repos *[]string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		// Swallow error
		return
	}

	for _, f := range files {
		filename := f.Name()

		if filename == ".git" {
			r.logger.Infof("Found GIT Repo: %s", path)

			*repos = append(*repos, path)
		} else if f.IsDir() && filename != "node_modules" {
			// Search next folder
			r.findGitReposInPath(fmt.Sprintf("%s\\%s", path, filename), repos)
		}
	}
}

func (r *Repos) getGitCommits(authorEmail string) (map[string]int, error) {
	repoList, err := r.getRepoListFromFS()
	if err != nil {
		return nil, fmt.Errorf("failed to read repo list from file: %s", err)
	}

	gitCommitMap := r.createGitCommitMap()

	for _, repoPath := range repoList {
		r.logger.Infof("Getting GIT commits for %s", repoPath)

		cIter, err := r.getCommitHistory(repoPath)
		if err != nil {
			r.logger.Error(err)
			continue
		}

		_ = cIter.ForEach(func(c *object.Commit) error {
			// Add commit if no email is supplied or if an email is supplied and it matches the commit author
			if authorEmail == "" || (authorEmail != "" && c.Author.Email == authorEmail) {
				commitDate := c.Author.When.Format("02/01/2006")

				gitCommitMap[commitDate] = gitCommitMap[commitDate] + 1
			}

			return nil
		})
	}

	return gitCommitMap, nil
}

func (r *Repos) createGitCommitMap() map[string]int {
	// [time: number of commits]
	gitCommitMap := make(map[string]int, commitCutOffDays)

	// Zero initialise the map for the past X day
	cutOffDate := time.Now().Add(-(commitCutOffDays * (24 * time.Hour)))
	for i := 1; i < commitCutOffDays; i++ {
		gitCommitMap[cutOffDate.Format("02/01/2006")] = 0

		cutOffDate = cutOffDate.Add(24 * time.Hour)
	}

	return gitCommitMap
}

func (r *Repos) getCommitHistory(path string) (object.CommitIter, error) {
	// Open GIT repo using .git folder path
	gitRepo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open GIT repo (%s): %s", path, err)
	}

	// HEAD reference
	ref, err := gitRepo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get GIT repo (%s) HEAD: %s", path, err)
	}

	// Commit history
	cIter, err := gitRepo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to get GIT repo (%s) commit history: %s", path, err)
	}

	return cIter, nil
}

func (r *Repos) printGitCommits(commits map[string]int) {
	rollingDate := time.Now().Add(-(commitCutOffDays * (24 * time.Hour)))

	for i := 1; i <= commitCutOffDays; i++ {
		commitNum := commits[rollingDate.Format("02/01/2006")]

		colourCode := ansiBlack
		switch {
		case commitNum >= 10:
			colourCode = ansiDarkGreen
		case commitNum >= 5:
			colourCode = ansiGreen
		case commitNum > 0:
			colourCode = ansiLightGreen
		}

		output := fmt.Sprintf("%s %d %s", colourCode, commitNum, "\033[0m")
		if commitNum == 0 {
			output = fmt.Sprintf("%s %s %s", colourCode, "-", "\033[0m")
		}

		// Start new line when a line reaches 30 days
		if i%30 == 0 {
			fmt.Println(output)
		} else {
			fmt.Print(output)
		}

		rollingDate = rollingDate.Add(24 * time.Hour)
	}
}

func (r *Repos) getRepoListFromFS() ([]string, error) {
	f, err := os.Open(gitRepoPathsFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read file
	var buffer [4096]byte
	f.Read(buffer[:])

	// Split file contents by carriage returns
	splitter := regexp.MustCompile(`\n`)
	s := splitter.Split(string(buffer[:]), -1)

	return s[:len(s)-1], nil
}

func (r *Repos) storeRepoListInFS(repoList []string) error {
	// Get the existing list of GIT repos (if the file actually exists)
	existingRepoList, _ := r.getRepoListFromFS()

	f, err := os.OpenFile(gitRepoPathsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, s := range repoList {
		if !utils.Contains(existingRepoList, s) {
			f.WriteString(s + "\n")
		}
	}

	f.Sync()

	return nil
}
