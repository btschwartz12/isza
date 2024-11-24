package repo

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	_ "modernc.org/sqlite"

	"github.com/btschwartz12/isza/repo/db"
)

const (
	postUploadDir  = "posts"
	dbName         = "isza.db"
	maxStorageSize = 5000 << 20 // 5GB
)

type Repo struct {
	logger *zap.SugaredLogger
	db     *sql.DB
	varDir string
}

func NewRepo(logger *zap.SugaredLogger, varDir string) (*Repo, error) {
	r := &Repo{
		logger: logger,
	}

	if err := os.MkdirAll(varDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating var dir: %w", err)
	}
	r.varDir = varDir

	if err := os.MkdirAll(filepath.Join(varDir, postUploadDir), 0755); err != nil {
		return nil, fmt.Errorf("error creating post upload dir: %w", err)
	}

	conn, err := sql.Open("sqlite", filepath.Join(varDir, dbName))
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	if _, err := conn.Exec(string(db.Schema)); err != nil {
		return nil, fmt.Errorf("error executing schema: %w", err)
	}

	r.db = conn
	return r, nil
}

func (r *Repo) storageFull() bool {
	var stat os.FileInfo
	var err error
	if stat, err = os.Stat(r.varDir); err != nil {
		r.logger.Errorw("error getting var dir info", "error", err)
		return true
	}
	if stat.Size() > maxStorageSize {
		r.logger.Errorw("storage full", "size", stat.Size())
		return true
	}
	return false
}
