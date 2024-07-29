package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/gitutil"
)

type repoLookup interface {
	GetRootDirectory(wd string) (string, error)
	GetBranchName() string
	RemoteURL() (string, error)
	GetRepoRoot() string
}

func newRepoLookup(wd string) (repoLookup, error) {
	repo, err := git.PlainOpenWithOptions(wd, &git.PlainOpenOptions{DetectDotGit: true})
	switch {
	case errors.Is(err, git.ErrRepositoryNotExists):
		return &noRepoLookupImpl{}, nil
	case err != nil:
		return nil, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	h, err := repo.Head()
	if err != nil {
		return nil, err
	}

	return &repoLookupImpl{
		RepoRoot: worktree.Filesystem.Root(),
		Repo:     repo,
		Head:     h,
	}, nil
}

type repoLookupImpl struct {
	RepoRoot string
	Repo     *git.Repository
	Head     *plumbing.Reference
}

func (r *repoLookupImpl) GetRootDirectory(wd string) (string, error) {
	dir, err := filepath.Rel(r.RepoRoot, wd)

	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(dir, string(os.PathSeparator), "/"), err
}

func (r *repoLookupImpl) GetBranchName() string {
	if r.Head == nil {
		return ""
	}
	return r.Head.Name().String()
}

func (r *repoLookupImpl) RemoteURL() (string, error) {
	if r.Repo == nil {
		return "", nil
	}
	return gitutil.GetGitRemoteURL(r.Repo, "origin")
}

func (r *repoLookupImpl) GetRepoRoot() string {
	return r.RepoRoot
}

type noRepoLookupImpl struct{}

func (r *noRepoLookupImpl) GetRootDirectory(wd string) (string, error) {
	return ".", nil
}

func (r *noRepoLookupImpl) GetBranchName() string {
	return ""
}

func (r *noRepoLookupImpl) RemoteURL() (string, error) {
	return "", nil
}

func (r *noRepoLookupImpl) GetRepoRoot() string {
	return ""
}

func main() {
	ctx := context.Background()

	repo := auto.GitRepo{
		URL:         "https://github.com/pulumi/test-repo.git",
		ProjectPath: "goproj",
		Shallow:     true,
		Branch:      "master",
	}
	ws, err := auto.NewLocalWorkspace(ctx, auto.Repo(repo))
	if err != nil {
		panic(err)
	}

	fmt.Println("WorkDir", ws.WorkDir())

	rl, err := newRepoLookup(ws.WorkDir())
	if err != nil {
		panic(err)
	}

	dir, err := rl.GetRootDirectory(filepath.Join(ws.WorkDir(), "something"))
	if err != nil {
		panic(err)
	}
	fmt.Println("GetRootDirectory", "goproj/something", dir)

	branch := rl.GetBranchName()
	if err != nil {
		panic(err)
	}
	fmt.Println("GetBranchName", "refs/heads/master", branch)

	remote, err := rl.RemoteURL()
	if err != nil {
		panic(err)
	}
	fmt.Println("RemoteURL", "https://github.com/pulumi/test-repo.git", remote)

	fmt.Println("GetRepoRoot", path.Dir(ws.WorkDir()), rl.GetRepoRoot())
}
