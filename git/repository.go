package git

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"gopkg.in/libgit2/git2go.v27"
)

type RepositoryParams struct {
	RemoteUrl     string
	HttpLogin     string
	HttpPassword  string
	SshPrivateKey string
}

type Repository struct {
	gitRepository *git.Repository
	params        RepositoryParams
	path          string
}

type Commit struct {
	Id        string
	Tag       string
	Files     []string
	Message   string
	Author    *git.Signature
	Committer *git.Signature
	Tagger    *git.Signature
}

// Check err and panic with message if err!=nil
func checkPanic(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s\n", msg, err))
	}
}

// Create and return git.RemoteCallbacks with authentication
func (repo Repository) createRemoteCallbacks() git.RemoteCallbacks {
	var credentialsCallback git.CredentialsCallback

	if repo.params.SshPrivateKey != "" {
		credentialsCallback = func(
			url string,
			username string,
			allowedTypes git.CredType,
		) (git.ErrorCode, *git.Cred) {
			ret, cred := git.NewCredSshKeyFromMemory(username, "", repo.params.SshPrivateKey, "")
			return git.ErrorCode(ret), &cred
		}
	} else {
		credentialsCallback = func(
			url string,
			username string,
			allowedTypes git.CredType,
		) (git.ErrorCode, *git.Cred) {
			if repo.params.HttpLogin != "" {
				username = repo.params.HttpLogin
			}
			ret, cred := git.NewCredUserpassPlaintext(username, repo.params.HttpPassword)
			return git.ErrorCode(ret), &cred
		}
	}

	return git.RemoteCallbacks{
		CertificateCheckCallback: func(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
			return 0
		},
		CredentialsCallback: credentialsCallback,
	}
}

// Open or clone git repository on path
func Open(path string, branch string, params RepositoryParams) *Repository {
	var (
		err  error
		repo = Repository{
			params: params,
			path:   path,
		}
	)

	fetchOptions := &git.FetchOptions{
		Prune:           git.FetchPruneUnspecified,
		DownloadTags:    git.DownloadTagsAll,
		RemoteCallbacks: repo.createRemoteCallbacks(),
	}

	repo.gitRepository, err = git.OpenRepository(repo.path)
	if err == nil {
		h, err := repo.gitRepository.Head()
		checkPanic(err, "Getting HEAD error")
		defer h.Free()

		if branch == "" {
			branch, err = repo.getLocalBranch().Name()
			checkPanic(err, "Get branch name error")
		}
		r, err := repo.gitRepository.Remotes.Lookup("origin")
		checkPanic(err, "Remote origin lookup error")
		defer r.Free()

		checkPanic(
			r.Fetch([]string{}, fetchOptions, ""),
			"Remote fetch error",
		)

		rb, err := repo.gitRepository.References.Lookup("refs/remotes/origin/" + branch)
		checkPanic(err, "Remote branch lookup error")
		defer rb.Free()

		remoteTarget := rb.Target()
		if h.Target().String() == remoteTarget.String() {
			return &repo
		}

		rc, err := repo.gitRepository.LookupCommit(remoteTarget)
		checkPanic(err, "Remote commit lookup error")
		defer rc.Free()

		checkPanic(
			repo.gitRepository.ResetToCommit(rc, git.ResetHard, &git.CheckoutOpts{}),
			"Reset to r commit error",
		)

		return &repo
	}

	repo.gitRepository, err = git.Clone(params.RemoteUrl, path, &git.CloneOptions{
		CheckoutBranch: branch,
		FetchOptions:   fetchOptions,
	})
	checkPanic(err, "Clone repo error")

	return &repo
}

// Checkout repository to commit
func (repo Repository) CheckoutCommit(c string) *Commit {
	id, err := git.NewOid(c)
	checkPanic(err, "Commit ID error")

	return repo.doCheckout(id)
}

// Checkout repository to tag
func (repo Repository) CheckoutTag(t string) *Commit {
	refs := repo.gitRepository.References
	ref, err := refs.Dwim(t)
	checkPanic(err, "Reference dwim error")
	defer ref.Free()

	return repo.doCheckout(ref.Target())
}

// Force checkout repository to target ID
func (repo Repository) doCheckout(oid *git.Oid) *Commit {
	err := repo.gitRepository.SetHeadDetached(oid)
	checkPanic(err, "Set HEAD error")

	checkPanic(
		repo.gitRepository.CheckoutHead(&git.CheckoutOpts{Strategy: git.CheckoutForce}),
		"CheckoutTag HEAD error",
	)

	o, err := repo.gitRepository.Lookup(oid)
	checkPanic(err, "Lookup object by Oid error")
	defer o.Free()

	if o.Type() == git.ObjectTag {
		t, err := o.AsTag()
		checkPanic(err, "Get object as tag error")
		defer t.Free()

		o = t.Target()
	}

	c, err := o.AsCommit()
	checkPanic(err, "Get commit by object error")
	defer c.Free()

	return &Commit{
		Id:        oid.String(),
		Message:   c.Message(),
		Author:    c.Author(),
		Committer: c.Committer(),
	}
}

// Get local branch
func (repo Repository) getLocalBranch() *git.Branch {
	h, err := repo.gitRepository.Head()
	checkPanic(err, "Getting HEAD error")
	defer h.Free()

	branch := h.Branch()
	if branch.IsBranch() {
		return branch
	}

	branchIterator, err := repo.gitRepository.NewBranchIterator(git.BranchLocal)
	checkPanic(err, "Local branch iterator error")
	defer branchIterator.Free()

	branch, _, err = branchIterator.Next()
	checkPanic(err, "Getting branch from iterator error")
	defer branch.Free()

	return branch
}

// Get files list changed on commit
func (repo Repository) getChangedFiles(c *git.Commit) []string {
	var (
		pc    *git.Commit
		files []string
	)

	ct, err := c.Tree()
	checkPanic(err, "Get commit tree error")
	defer ct.Free()

	if c.ParentCount() == 0 {
		checkPanic(
			ct.Walk(func(path string, ent *git.TreeEntry) int {
				if ent.Type == git.ObjectBlob {
					files = append(files, path+ent.Name)
				}
				return 0
			}),
			"Commits tree walk error",
		)
		return files
	}

	// Only first parent commit
	pc = c.Parent(0)
	defer pc.Free()

	pct, err := pc.Tree()
	checkPanic(err, "Get commit parent tree error")
	defer pct.Free()

	diff, err := repo.gitRepository.DiffTreeToTree(ct, pct, &git.DiffOptions{})
	checkPanic(err, "Get tree diff error")
	defer checkPanic(diff.Free(), "Diff tree release error")

	checkPanic(
		diff.ForEach(func(diffDetail git.DiffDelta, diff float64) (git.DiffForEachHunkCallback, error) {
			files = append(files, diffDetail.NewFile.Path)
			return nil, nil
		}, 0),
		"Diff tree foreach error",
	)

	return files
}

// Listing all commits using topological sorting
func (repo Repository) ListCommits() []*Commit {
	w, err := repo.gitRepository.Walk()
	checkPanic(err, "Commits walk error")
	defer w.Free()

	w.Sorting(git.SortTopological)
	checkPanic(w.PushHead(), "Commits walk push head error")

	var commits []*Commit
	checkPanic(
		w.Iterate(func(c *git.Commit) bool {
			commits = append(commits, &Commit{
				Id:    c.Id().String(),
				Files: repo.getChangedFiles(c),
			})

			return true
		}),
		"Commits iterate error:",
	)

	return commits
}

// Listing all tags with sorting by commit timestamp
func (repo Repository) ListTags() []string {
	ri, err := repo.gitRepository.NewReferenceIterator()
	checkPanic(err, "Reference iterator error")
	defer ri.Free()

	var keys []int
	refs := make(map[int]string)
	for {
		r, err := ri.Next()
		if err != nil {
			break
		}
		defer r.Free()

		if !r.IsTag() {
			continue
		}

		tag := strings.TrimPrefix(r.Name(), "refs/tags/")
		t, err := repo.gitRepository.Lookup(r.Target())
		checkPanic(err, "Tag target lookup error")
		defer t.Free()

		var c *git.Commit
		if t.Type() == git.ObjectTag {
			o, err := t.AsTag()
			checkPanic(err, "Object as tag error")
			defer o.Free()

			c, err = o.Target().AsCommit()
			checkPanic(err, "Commit from object error")
		} else {
			c, err = t.AsCommit()
			checkPanic(err, "Object as commit error")
		}
		defer c.Free()

		ts := int(c.Committer().When.Unix())
		keys = append(keys, ts)
		refs[ts] = tag
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] > keys[j]
	})

	var tags []string
	for _, v := range keys {
		tags = append(tags, refs[v])
	}

	return tags
}

// Create tag on HEAD
func (repo Repository) CreateTag(tag string, msg string) *Commit {
	h, err := repo.gitRepository.Head()
	checkPanic(err, "Getting HEAD error")
	defer h.Free()

	c, err := repo.gitRepository.LookupCommit(h.Target())
	checkPanic(err, "Lookup HEAD commit error")
	defer c.Free()

	var tagSignature *git.Signature

	if msg == "" {
		_, err = repo.gitRepository.Tags.CreateLightweight(tag, c, false)
		checkPanic(err, "Create lightweight tag error")
	} else {
		tagSignature = &git.Signature{
			Name:  "git",
			Email: "git@localhost",
			When:  time.Now(),
		}

		_, err = repo.gitRepository.Tags.Create(tag, c, tagSignature, msg)
		checkPanic(err, "Create tag with message error")
	}

	return &Commit{
		Id:        c.Id().String(),
		Tag:       tag,
		Message:   c.Message(),
		Author:    c.Author(),
		Committer: c.Committer(),
		Tagger:    tagSignature,
	}
}

func (repo Repository) PushTag(tag string) {
	r, err := repo.gitRepository.Remotes.Lookup("origin")
	checkPanic(err, "Lookup remotes error")
	checkPanic(
		r.Push([]string{"refs/tags/" + tag}, &git.PushOptions{
			RemoteCallbacks: repo.createRemoteCallbacks(),
		}),
		"Tag push error",
	)
}

// Close operations in git repository
func (repo Repository) Close() {
	repo.gitRepository.Free()
}
