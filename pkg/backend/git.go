package backend

import (
	"context"
	"io"
	"os/exec"
)

// execGitUploadPack executes git-upload-pack with the given reader and writer.
func execGitUploadPack(ctx context.Context, repoPath string, r io.Reader, w io.Writer) error {
	cmd := exec.CommandContext(ctx, "git", "upload-pack", "--stateless-rpc", repoPath)
	cmd.Stdin = r
	cmd.Stdout = w
	return cmd.Run()
}

// execGitUploadPackAdvertise executes git-upload-pack to advertise refs.
func execGitUploadPackAdvertise(ctx context.Context, repoPath string, w io.Writer) error {
	cmd := exec.CommandContext(ctx, "git", "upload-pack", "--stateless-rpc", "--advertise-refs", repoPath)
	cmd.Stdout = w
	return cmd.Run()
}
