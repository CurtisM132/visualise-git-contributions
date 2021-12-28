package main

import (
	"flag"

	"go.uber.org/zap"

	"CurtisM132/main/git"
)

func main() {
	logger := initLogger()

	repos := git.NewGitRepos(logger)

	gitFolder := flag.String("add", "", "Add a folder to be scanned")
	email := flag.String("email", "", "Email used in GIT commits")
	flag.Parse()

	if *gitFolder != "" {
		err := repos.AddAllReposInFolder(*gitFolder)
		if err != nil {
			logger.Error(err)
		}

		return
	}

	err := repos.VisualiseGitCommits(*email)
	if err != nil {
		logger.Error(err)
	}
}

func initLogger() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	return logger.Sugar()
}
