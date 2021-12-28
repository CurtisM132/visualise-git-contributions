package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"CurtisM132/main/utils"
)

type Repos struct {
	logger *zap.SugaredLogger
}

const gitFolderPathsFile = "folders.txt"

func NewGitRepos(logger *zap.SugaredLogger) *Repos {
	return &Repos{
		logger: logger,
	}
}

func (r *Repos) AddAllReposInFolder(folder string) error {
	var gitRepos []string
	r.findGitReposInPath(folder, &gitRepos)

	err := r.storeRepoList(gitRepos)
	if err != nil {
		r.logger.Errorf("failed to store GIT repos to file system: %s", err)
	}

	return nil
}

func (r *Repos) findGitReposInPath(path string, repos *[]string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}

	for _, f := range files {
		filename := f.Name()

		if filename == ".git" {
			r.logger.Infof("Found GIT Repo: %s", path)
			*repos = append(*repos, path)

		} else if f.IsDir() && filename != "node_modules" {
			r.findGitReposInPath(fmt.Sprintf("%s\\%s", path, filename), repos)
		}
	}
}

func (r *Repos) VisualiseGitContributions() error {
	r.getGitContributions()

	return nil
}

func (r *Repos) getGitContributions() (map[string]int, error) {
	repoList, err := r.readRepoList()
	if err != nil {
		return nil, fmt.Errorf("failed to read repo list from file: %s", err)
	}

	numCommitsPerDay := make(map[string]int) // [time: number of commits]

	for _, repoPath := range repoList {
		r.logger.Infof("Getting GIT commits for %s", repoPath)

		// Open GIT repo using .git folder path
		gitRepo, err := git.PlainOpen(repoPath)
		if err != nil {
			r.logger.Errorf("failed to open GIT repo (%s): %s", repoPath, err)
			continue
		}

		// HEAD reference
		ref, err := gitRepo.Head()
		if err != nil {
			r.logger.Errorf("failed to get GIT repo (%s) HEAD: %s", repoPath, err)
			continue
		}

		// Commit history
		cIter, err := gitRepo.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			r.logger.Errorf("failed to get GIT repo (%s) commit history: %s", repoPath, err)
			continue
		}

		_ = cIter.ForEach(func(c *object.Commit) error {
			// TODO
			// if c.Author.Name == Me

			commitDate := c.Author.When.Format("01/02/2006")

			numCommitsPerDay[commitDate] = numCommitsPerDay[commitDate] + 1

			return nil
		})

	}

	return numCommitsPerDay, nil
}

func (r *Repos) readRepoList() ([]string, error) {
	f, err := os.Open(gitFolderPathsFile)
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

	// File terminates with a new line so ignore the last element
	return s[:len(s)-1], nil
}

func (r *Repos) storeRepoList(repoList []string) error {
	var existingRepoList []string
	existingRepoList, _ = r.readRepoList()

	f, err := os.OpenFile(gitFolderPathsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
