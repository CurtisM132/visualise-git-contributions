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

type GitCommitManager struct {
	logger       *zap.SugaredLogger
	gitCommitMap map[string]int
}

func NewGitCommitManager(logger *zap.SugaredLogger) *GitCommitManager {
	return &GitCommitManager{
		logger: logger,
	}
}

// Scan the supplied folder recursively and add all applicable GIT repos to a persistently stored file
func (m *GitCommitManager) AddAllReposInFolder(folder string) error {
	var gitRepos []string
	m.findGitReposInPath(folder, &gitRepos)

	if len(gitRepos) > 0 {
		err := m.storeRepoListInFS(gitRepos)
		if err != nil {
			return fmt.Errorf("failed to store GIT repos to file system: %s", err)
		}
	}

	return nil
}

func (m *GitCommitManager) PopulateCommitMap(authorEmail string) error {
	m.gitCommitMap = m.createGitCommitMap()

	repos, err := m.getRepoListFromFS()
	if err != nil {
		return fmt.Errorf("failed to read repo list from file: %s", err)
	}

	for _, path := range repos {
		cIter, err := m.getCommitHistory(path)
		if err != nil {
			m.logger.Error(err)
			continue
		}

		m.addGitCommitsToMap(cIter, authorEmail)
	}

	return nil
}

func (m *GitCommitManager) GetCommitMap() *map[string]int {
	return &m.gitCommitMap
}

func (m *GitCommitManager) findGitReposInPath(path string, repos *[]string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		// Swallow error
		return
	}

	for _, f := range files {
		filename := f.Name()

		if filename == ".git" {
			m.logger.Infof("Found GIT Repo: %s", path)

			*repos = append(*repos, path)
		} else if f.IsDir() && filename != "node_modules" {
			// Search next folder
			m.findGitReposInPath(fmt.Sprintf("%s\\%s", path, filename), repos)
		}
	}
}

func (m *GitCommitManager) addGitCommitsToMap(cIter object.CommitIter, authorEmail string) {
	_ = cIter.ForEach(func(c *object.Commit) error {
		// Add commit if no email is supplied or if an email is supplied and it matches the commit author
		if authorEmail == "" || (authorEmail != "" && c.Author.Email == authorEmail) {
			commitDate := c.Author.When.Format(DateFormat)

			m.gitCommitMap[commitDate] = m.gitCommitMap[commitDate] + 1
		}

		return nil
	})
}

// Create and populate a map to hold the amount of GIT commits against a specific date
func (m *GitCommitManager) createGitCommitMap() map[string]int {
	// [time: number of commits]
	gitCommitMap := make(map[string]int, commitCutOffDays)

	// Zero initialise the map for the past X day
	cutOffDate := time.Now().Add(-(commitCutOffDays * (24 * time.Hour)))
	for i := 1; i < commitCutOffDays; i++ {
		gitCommitMap[cutOffDate.Format(DateFormat)] = 0

		cutOffDate = cutOffDate.Add(24 * time.Hour)
	}

	return gitCommitMap
}

// Get GIT commit history for a specific GIT repo (using go-git)
func (m *GitCommitManager) getCommitHistory(path string) (object.CommitIter, error) {
	m.logger.Infof("Getting GIT commit history for %s", path)

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

// Get the persisted previously discovered GIT repos
func (m *GitCommitManager) getRepoListFromFS() ([]string, error) {
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

// Store a list of GIT repos in the file system
func (m *GitCommitManager) storeRepoListInFS(repoList []string) error {
	// Get the existing list of GIT repos (if the file actually exists)
	existingRepoList, _ := m.getRepoListFromFS()

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
