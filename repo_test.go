package simulator

import (
	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func setupRepo(t *testing.T, ops []RepoOperation) string {
	path := t.TempDir()

	state := NewRepoState().
		WithPath(path).
		WithDefaultAuthor("John Doe", "jd@example.com").
		WithNow(time.Date(2025, 6, 7, 1, 49, 0, 0, time.UTC)).
		WithDefaultAdvanceTime(5 * time.Second)

	for _, op := range ops {
		op.Apply(t, state)
	}

	return path
}

func TestFirstFile(t *testing.T) {
	path := setupRepo(t, []RepoOperation{
		InitRepo{DefaultBranch: "main"},
		NewFile{Path: "file1.txt", Content: "content1"},
		Commit{Message: "commit1"},
	})

	r, err := git.PlainOpen(path)
	require.NoError(t, err)
	headRef, err := r.Head()
	require.NoError(t, err)
	require.Equal(t, "main", headRef.Name().Short())
	headCommit, err := r.CommitObject(headRef.Hash())
	require.NoError(t, err)
	require.Equal(t, "commit1", headCommit.Message)
	require.Equal(t, "8617236", headCommit.Hash.String()[:7])
}

func TestTagExists(t *testing.T) {
	path := setupRepo(t, []RepoOperation{
		InitRepo{DefaultBranch: "main"},
		NewFile{Path: "file1.txt", Content: "content1"},
		Commit{Message: "commit1"},
		Tag{Name: "v1.0.0"},
		NewFile{Path: "file2.txt", Content: "content2"},
		Commit{Message: "commit2"},
	})

	r, err := git.PlainOpen(path)
	require.NoError(t, err)
	ref, err := r.Tag("v1.0.0")
	require.NoError(t, err)
	require.Equal(t, "8617236", ref.Hash().String()[:7])
}

func TestTagAnnotatedExists(t *testing.T) {
	path := setupRepo(t, []RepoOperation{
		InitRepo{DefaultBranch: "main"},
		NewFile{Path: "file1.txt", Content: "content1"},
		Commit{Message: "commit1"},
		TagAnnotated{Name: "v1.0.0", Message: "message1"},
		NewFile{Path: "file2.txt", Content: "content2"},
		Commit{Message: "commit2"},
	})

	r, err := git.PlainOpen(path)
	require.NoError(t, err)
	ref, err := r.Tag("v1.0.0")
	require.NoError(t, err)
	require.Equal(t, "86f1655", ref.Hash().String()[:7])
}

func TestBranchAndCheckout(t *testing.T) {
	path := setupRepo(t, []RepoOperation{
		InitRepo{DefaultBranch: "main"},
		NewFile{Path: "file1.txt", Content: "content1"},
		Commit{Message: "commit1"},
		BranchAndCheckout{Name: "branch1"},
		NewFile{Path: "file2.txt", Content: "content2"},
		Commit{Message: "commit2"},
	})

	r, err := git.PlainOpen(path)
	require.NoError(t, err)
	headRef, err := r.Head()
	require.NoError(t, err)
	require.Equal(t, "branch1", headRef.Name().Short())
	headCommit, err := r.CommitObject(headRef.Hash())
	require.NoError(t, err)
	require.Equal(t, "commit2", headCommit.Message)
	require.Equal(t, "1810fa6", headCommit.Hash.String()[:7])
}

func TestCheckout(t *testing.T) {
	path := setupRepo(t, []RepoOperation{
		InitRepo{DefaultBranch: "main"},
		NewFile{Path: "file1.txt", Content: "content1"},
		Commit{Message: "commit1"},
		BranchAndCheckout{Name: "branch1"},
		NewFile{Path: "file2.txt", Content: "content2"},
		Commit{Message: "commit2"},
		Checkout{Name: "main"},
	})

	r, err := git.PlainOpen(path)
	require.NoError(t, err)
	headRef, err := r.Head()
	require.NoError(t, err)
	require.Equal(t, "main", headRef.Name().Short())
	headCommit, err := r.CommitObject(headRef.Hash())
	require.NoError(t, err)
	require.Equal(t, "commit1", headCommit.Message)
	require.Equal(t, "8617236", headCommit.Hash.String()[:7])
}
