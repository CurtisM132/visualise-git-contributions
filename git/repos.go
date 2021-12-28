package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type Repos struct{}

const gitFolderPathsFile = "folders.txt"

func (r *Repos) AddAllReposInFolder(folder string) error {
	r.getGitReposInFolder(folder)

	err := r.writeToDefaultFile(folder)
	if err != nil {
		fmt.Printf("failed to write %s to default file: %s", folder, err.Error())
	}

	return nil
}

func (r *Repos) getGitReposInFolder(folder string) ([]string, error) {
	var repos []string

	repos = r.g(folder, repos)

	return repos, nil
}

func (r *Repos) g(path string, repos []string) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return repos
	}

	for _, f := range files {
		filename := f.Name()

		if filename == "git" {
			repos = append(repos, path)

		} else if f.IsDir() && filename != "node_modules" {
			path = fmt.Sprintf("%s/%s", path, filename)
			repos = r.g(path, repos)
		}
	}

	return repos
}

func (r *Repos) VisualiseGitContributions() error {
	r.getGitContributions()

	return nil
}

func (r *Repos) getGitContributions() (map[string]int, error) {
	repoList, err := r.readFromDefaultFile()
	if err != nil {
		return nil, err
	}

	numCommitsPerDay := make(map[string]int) // [time: number of commits]

	for _, repo := range repoList {
		// Open GIT repo using .git folder path
		r, err := git.PlainOpen(repo)
		if err != nil {
			return nil, fmt.Errorf("failed to open GIT repo (%s): %s", repo, err)
		}

		// HEAD reference
		ref, err := r.Head()
		if err != nil {
			return nil, fmt.Errorf("failed to get GIT repo (%s) HEAD: %s", repo, err)
		}

		// Commit history
		cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			return nil, fmt.Errorf("failed to get GIT repo (%s) commit history: %s", repo, err)
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

func (r *Repos) readFromDefaultFile() ([]string, error) {
	f, err := os.Open(gitFolderPathsFile)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	// Read the file into a buffer
	var buffer [4096]byte
	f.Read(buffer[:])

	// Split the file contents with each line being a new string
	splitter := regexp.MustCompile(`\n`)
	s := splitter.Split(string(buffer[:]), -1)

	return s, nil
}

func (r *Repos) writeToDefaultFile(s string) error {
	f, err := os.OpenFile(gitFolderPathsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(s + "\n")
	f.Sync()

	return nil
}
