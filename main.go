package main

import (
	"flag"

	"go.uber.org/zap"

	"CurtisM132/main/git"
)

func main() {
	logger := initLogger()

	commitManager := git.NewGitCommitManager(logger)
	commitVisualiser := git.NewGitCommitVisualiser(logger)

	gitFolder := flag.String("add", "", "Add a folder to be scanned")
	email := flag.String("email", "", "Email used in GIT commits")
	flag.Parse()

	if *gitFolder != "" {
		err := commitManager.AddAllReposInFolder(*gitFolder)
		if err != nil {
			logger.Error(err)
		}

		return
	}

	err := commitManager.PopulateCommitMap(*email)
	if err != nil {
		logger.Error(err)
	}

	err = commitVisualiser.VisualiseGitCommits(commitManager.GetCommitMap())
	if err != nil {
		logger.Error(err)
	}
}

func initLogger() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	return logger.Sugar()
}
