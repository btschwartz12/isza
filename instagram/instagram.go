package instagram

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/btschwartz12/isza/repo"
	"go.uber.org/zap"
)

func ExecutePost(
	ctx context.Context,
	logger *zap.SugaredLogger,
	r *repo.Repo,
	workingDir string,
	username string,
	password string,
	post *repo.Post,
) error {
	logger.Infow("posting", "post", post.ID)

	captionPath := filepath.Join(workingDir, "caption.txt")
	err := os.WriteFile(captionPath, []byte(post.Caption), 0644)
	if err != nil {
		return fmt.Errorf("error writing caption file: %w", err)
	}
	defer os.Remove(captionPath)

	fullPaths := make([]string, len(post.ImageFilenames))
	for i, filename := range post.ImageFilenames {
		fullpath, err := filepath.Abs(r.GetPathForPost(filename))
		if err != nil {
			return fmt.Errorf("error getting absolute path for post: %w", err)
		}
		fullPaths[i] = fullpath
	}
	pathsArg := strings.Join(fullPaths, ",")

	pythonPath := "python3"
	scriptPath := filepath.Join(workingDir, "post.py")

	cmd := exec.Command(pythonPath, scriptPath, username, password, pathsArg, captionPath, "false")
	cmd.Dir = workingDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		logger.Errorw("error running post script", "err", err, "stdout", stdout.String(), "stderr", stderr.String())
		return fmt.Errorf("error running post script: %w", err)
	}
	logger.Infow("post complete", "post", post.ID)
	return nil
}
