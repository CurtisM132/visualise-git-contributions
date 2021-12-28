package main

import (
	"flag"
	"fmt"

	"CurtisM132/main/git"
)

func main() {
	repos := git.Repos{}

	var gitFolder string
	flag.StringVar(&gitFolder, "add", "", "Add a folder to be scanned")
	flag.Parse()

	if gitFolder != "" {
		err := repos.AddAllReposInFolder(gitFolder)
		if err != nil {
			fmt.Print(err)
		}

		return
	}

	repos.VisualiseGitContributions()

	// _, err := readFromDefaultFile()
	// if err != nil {
	// 	fmt.Print("failed to read contents of default file: " + err.Error())
	// 	return
	// }
}
