package main

import (
	"flag"

	"CurtisM132/main/git"

	"go.uber.org/zap"
)

func main() {
	logger := initLogger()

	repos := git.NewGitRepos(logger)

	var gitFolder string
	flag.StringVar(&gitFolder, "add", "", "Add a folder to be scanned")
	flag.Parse()

	if gitFolder != "" {
		err := repos.AddAllReposInFolder(gitFolder)
		if err != nil {
			logger.Error(err)
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

func initLogger() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	return logger.Sugar()
}
