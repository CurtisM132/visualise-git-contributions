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

const (
	ansiBlack      string = "\033[0;37;30m"
	ansiLightGreen string = "\033[1;30;47m"
	ansiGreen      string = "\033[1;30;43m"
	ansiDarkGreen  string = "\033[1;30;42m"
)

type GitCommitVisualiser struct {
	logger *zap.SugaredLogger
}

func NewGitCommitVisualiser(logger *zap.SugaredLogger) *GitCommitVisualiser {
	return &GitCommitVisualiser{
		logger: logger,
	}
}

// Scan the supplied folder recursively and add all applicable GIT repos to a persistently stored file
func (r *GitCommitVisualiser) AddAllReposInFolder(folder string) error {
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

func (r *GitCommitVisualiser) VisualiseGitCommits(gitCommitMap *map[string]int) error {
	r.printGitCommits(gitCommitMap)

	return nil
}

func (r *GitCommitVisualiser) findGitReposInPath(path string, repos *[]string) {
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

func (r *GitCommitVisualiser) addGitCommitsToMap(cIter object.CommitIter, authorEmail string, gitCommitMap *map[string]int) {
	_ = cIter.ForEach(func(c *object.Commit) error {
		// Add commit if no email is supplied or if an email is supplied and it matches the commit author
		if authorEmail == "" || (authorEmail != "" && c.Author.Email == authorEmail) {
			commitDate := c.Author.When.Format(DateFormat)

			(*gitCommitMap)[commitDate] = (*gitCommitMap)[commitDate] + 1
		}

		return nil
	})
}

// Create and populate a map to hold the amount of GIT commits against a specific date
func (r *GitCommitVisualiser) createGitCommitMap() map[string]int {
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
func (r *GitCommitVisualiser) getCommitHistory(path string) (object.CommitIter, error) {
	r.logger.Infof("Getting GIT commit history for %s", path)

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

// Pretty print the GIT commits in a tabula colour-coded format
func (r *GitCommitVisualiser) printGitCommits(gitCommitMap *map[string]int) {
	rollingDate := time.Now().Add(-(commitCutOffDays * (24 * time.Hour)))

	for i := 1; i <= commitCutOffDays; i++ {
		commitNum := (*gitCommitMap)[rollingDate.Format(DateFormat)]

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

		// Start new line every 30 days
		if i%30 == 0 {
			fmt.Println(output)
		} else {
			fmt.Print(output)
		}

		rollingDate = rollingDate.Add(24 * time.Hour)
	}
}

// Get the persisted previously discovered GIT repos
func (r *GitCommitVisualiser) getRepoListFromFS() ([]string, error) {
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
func (r *GitCommitVisualiser) storeRepoListInFS(repoList []string) error {
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
