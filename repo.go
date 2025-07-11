package simulator

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type RepoState struct {
	Path     string
	Repo     *git.Repository
	Worktree *git.Worktree
	LastHash plumbing.Hash
	Now      time.Time

	DefaultAuthor      object.Signature
	DefaultBranch      string
	DefaultAdvanceTime time.Duration
}

func NewRepoState() *RepoState {
	return &RepoState{}
}

func (state *RepoState) WithPath(path string) *RepoState {
	state.Path = path
	return state
}

func (state *RepoState) WithNow(now time.Time) *RepoState {
	state.Now = now
	return state
}

func (state *RepoState) WithDefaultAuthor(name, email string) *RepoState {
	state.DefaultAuthor = object.Signature{
		Name:  name,
		Email: email,
		When:  state.Now,
	}
	return state
}

func (state *RepoState) WithDefaultBranch(name string) *RepoState {
	state.DefaultBranch = name
	return state
}

func (state *RepoState) WithDefaultAdvanceTime(duration time.Duration) *RepoState {
	state.DefaultAdvanceTime = duration
	return state
}

type RepoOperation interface {
	Apply(t *testing.T, state *RepoState)
}

type InitRepo struct {
	DefaultBranch string
	AdvanceTimeBy *time.Duration
}

func (op InitRepo) Apply(t *testing.T, state *RepoState) {
	r, err := git.PlainInitWithOptions(
		state.Path,
		&git.PlainInitOptions{
			InitOptions: git.InitOptions{
				DefaultBranch: plumbing.ReferenceName("refs/heads/" + op.DefaultBranch),
			},
			Bare: false,
		})
	require.NoError(t, err)
	wt, err := r.Worktree()
	require.NoError(t, err)

	state.Repo = r
	state.Worktree = wt

	state.DefaultBranch = op.DefaultBranch

	if op.AdvanceTimeBy != nil {
		state.Now = state.Now.Add(*op.AdvanceTimeBy)
	} else {
		state.Now = state.Now.Add(state.DefaultAdvanceTime)
	}
}

type NewFile struct {
	Path          string
	Content       string
	AdvanceTimeBy *time.Duration
}

func (op NewFile) Apply(t *testing.T, state *RepoState) {
	fullPath := filepath.Join(state.Path, op.Path)
	err := os.WriteFile(fullPath, []byte(op.Content), 0644)
	require.NoError(t, err)

	_, err = state.Worktree.Add(op.Path)
	require.NoError(t, err)

	if op.AdvanceTimeBy != nil {
		state.Now = state.Now.Add(*op.AdvanceTimeBy)
	} else {
		state.Now = state.Now.Add(state.DefaultAdvanceTime)
	}
}

type Commit struct {
	Message       string
	Author        *object.Signature
	AdvanceTimeBy *time.Duration
}

func (op Commit) Apply(t *testing.T, state *RepoState) {
	author := op.Author
	if author == nil {
		// Create a new signature with the current time
		author = &object.Signature{
			Name:  state.DefaultAuthor.Name,
			Email: state.DefaultAuthor.Email,
			When:  state.Now,
		}
	} else {
		// Update the When field of the provided author
		author.When = state.Now
	}
	hash, err := state.Worktree.Commit(op.Message, &git.CommitOptions{
		Author: author,
	})
	require.NoError(t, err)

	state.LastHash = hash

	if op.AdvanceTimeBy != nil {
		state.Now = state.Now.Add(*op.AdvanceTimeBy)
	} else {
		state.Now = state.Now.Add(state.DefaultAdvanceTime)
	}
}

type Tag struct {
	Name          string
	AdvanceTimeBy *time.Duration
}

func (op Tag) Apply(t *testing.T, state *RepoState) {
	_, err := state.Repo.CreateTag(op.Name, state.LastHash, nil)
	require.NoError(t, err)

	if op.AdvanceTimeBy != nil {
		state.Now = state.Now.Add(*op.AdvanceTimeBy)
	} else {
		state.Now = state.Now.Add(state.DefaultAdvanceTime)
	}
}

type TagAnnotated struct {
	Name          string
	Message       string
	Author        struct{ Name, Email string }
	AdvanceTimeBy *time.Duration
}

func (op TagAnnotated) Apply(t *testing.T, state *RepoState) {
	// Create a new tagger with the current time
	tagger := &object.Signature{
		Name:  op.Author.Name,
		Email: op.Author.Email,
		When:  state.Now,
	}
	_, err := state.Repo.CreateTag(op.Name, state.LastHash, &git.CreateTagOptions{
		Message: op.Message,
		Tagger:  tagger,
	})
	require.NoError(t, err)

	if op.AdvanceTimeBy != nil {
		state.Now = state.Now.Add(*op.AdvanceTimeBy)
	} else {
		state.Now = state.Now.Add(state.DefaultAdvanceTime)
	}
}

type BranchAndCheckout struct {
	Name          string
	AdvanceTimeBy *time.Duration
}

func (op BranchAndCheckout) Apply(t *testing.T, state *RepoState) {
	err := state.Worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(op.Name),
		Create: true,
	})
	require.NoError(t, err)

	if op.AdvanceTimeBy != nil {
		state.Now = state.Now.Add(*op.AdvanceTimeBy)
	} else {
		state.Now = state.Now.Add(state.DefaultAdvanceTime)
	}
}

type Checkout struct {
	Name          string
	AdvanceTimeBy *time.Duration
}

func (op Checkout) Apply(t *testing.T, state *RepoState) {
	err := state.Worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(op.Name),
		Create: false,
	})
	require.NoError(t, err)
	if op.AdvanceTimeBy != nil {
		state.Now = state.Now.Add(*op.AdvanceTimeBy)
	} else {
		state.Now = state.Now.Add(state.DefaultAdvanceTime)
	}
}

type AdvanceTime struct {
	Duration time.Duration
}

func (op AdvanceTime) Apply(_ *testing.T, state *RepoState) {
	state.Now = state.Now.Add(op.Duration)
}

type SetDefaultBranch struct {
	DefaultBranch string
	AdvanceTimeBy *time.Duration
}

func (op SetDefaultBranch) Apply(t *testing.T, state *RepoState) {
	state.DefaultBranch = op.DefaultBranch

	ref := plumbing.NewSymbolicReference(
		plumbing.HEAD,
		plumbing.NewBranchReferenceName(op.DefaultBranch),
	)

	err := state.Repo.Storer.SetReference(ref)
	require.NoError(t, err)

	if op.AdvanceTimeBy != nil {
		state.Now = state.Now.Add(*op.AdvanceTimeBy)
	} else {
		state.Now = state.Now.Add(state.DefaultAdvanceTime)
	}
}

type CheckoutTag struct {
	Name string
}

func (op CheckoutTag) Apply(t *testing.T, state *RepoState) {
	r, err := git.PlainOpen(state.Path)
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	tagRef, err := r.Tag(op.Name)
	if err != nil {
		t.Fatalf("Failed to find tag %s: %v", op.Name, err)
	}

	w, err := r.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash: tagRef.Hash(),
	})
	if err != nil {
		t.Fatalf("Failed to checkout tag %s: %v", op.Name, err)
	}
}
