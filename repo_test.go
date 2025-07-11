package simulator

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
	require.Equal(t, "4e9da30", headCommit.Hash.String()[:7])
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
	require.Equal(t, "4e9da30", ref.Hash().String()[:7])
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
	require.Equal(t, "aeec35c", ref.Hash().String()[:7])
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
	require.Equal(t, "40338e3", headCommit.Hash.String()[:7])
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
	require.Equal(t, "4e9da30", headCommit.Hash.String()[:7])
}

func TestAdvanceTime(t *testing.T) {
	path := setupRepo(t, []RepoOperation{
		InitRepo{DefaultBranch: "main"},
		AdvanceTime{Duration: 10 * time.Second},
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
	require.Equal(t, "1c0002a", headCommit.Hash.String()[:7])
	require.Equal(t,
		headCommit.Author.When.UTC(),
		time.Date(2025, 6, 7, 1, 49, 20, 0, time.UTC).UTC(),
	)
}

func TestSetDefaultBranch(t *testing.T) {
	path := setupRepo(t, []RepoOperation{
		InitRepo{DefaultBranch: "main"},
		NewFile{Path: "file1.txt", Content: "content1"},
		Commit{Message: "commit1"},
		BranchAndCheckout{Name: "branch1"},
		SetDefaultBranch{DefaultBranch: "branch1"},
	})

	r, err := git.PlainOpen(path)
	require.NoError(t, err)
	headRef, err := r.Head()
	require.NoError(t, err)
	require.Equal(t, "branch1", headRef.Name().Short())
}

func TestCheckoutTag(t *testing.T) {
	path := setupRepo(t, []RepoOperation{
		InitRepo{DefaultBranch: "main"},
		NewFile{Path: "file1.txt", Content: "content1"},
		Commit{Message: "commit1"},
		Tag{Name: "v1.0.0"},
		NewFile{Path: "file2.txt", Content: "content2"},
		Commit{Message: "commit2"},
		CheckoutTag{Name: "v1.0.0"},
	})

	r, err := git.PlainOpen(path)
	require.NoError(t, err)

	// Get HEAD reference
	headRef, err := r.Head()
	require.NoError(t, err)

	// Get the commit at HEAD
	headCommit, err := r.CommitObject(headRef.Hash())
	require.NoError(t, err)

	// Get the tag reference
	tagRef, err := r.Tag("v1.0.0")
	require.NoError(t, err)

	// Verify that HEAD is at the commit where the tag was created
	require.Equal(t, tagRef.Hash().String(), headRef.Hash().String())

	// Verify that we're in a detached HEAD state (HEAD is a hash reference, not a symbolic reference)
	require.Equal(t, plumbing.HashReference, headRef.Type())

	// Additional verification that we're at the first commit
	require.Equal(t, "commit1", headCommit.Message)
	require.Equal(t, "4e9da30", headCommit.Hash.String()[:7])
}
