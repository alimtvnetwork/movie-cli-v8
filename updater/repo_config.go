package updater

import (
	"database/sql"
	"errors"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

const repoPathConfigKey = "RepoPath"

func loadSavedRepoPath() (string, error) {
	database, err := db.Open()
	if err != nil {
		return "", appfault.Wrap("open config database", err)
	}
	defer database.Close()

	repoPath, err := database.GetConfig(repoPathConfigKey)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", appfault.Wrap("read saved repo path", err)
	}
	return repoPath, nil
}

func saveRepoPath(repoPath string) error {
	database, err := db.Open()
	if err != nil {
		return appfault.Wrap("open config database", err)
	}
	defer database.Close()

	if err := database.SetConfig(repoPathConfigKey, repoPath); err != nil {
		return appfault.Wrap("save repo path", err)
	}
	return nil
}
