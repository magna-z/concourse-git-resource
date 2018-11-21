package controller

import (
	"github.com/libgit2/git2go"
	log "github.com/sirupsen/logrus"
	"os"
	"io/ioutil"
	"os/exec"
	"io"
	"strings"
	"regexp"
)

type Payload struct {
	Source Source `json:"source"`
	Version Ref `json:"version"`
}

type Ref struct {
	Ref string `json:"ref"`
}

type Source struct {
	Url    string `json:"url"`
	Branch string `json:"branch"`
	TagFilter string `json:"tag_filter"`
	PrivateKey string `json:"private_key"`
}

type MetadataArry struct {
	Version Ref `json:"version"`
	Metadata []map[string]string `json:"metadata"`
}

var (
	sshKeyPath = "/root/.ssh/"
)

func createSshPubKey() {
	cmd := exec.Command("ssh-keygen", "-y", "-f", "/root/.ssh/id_rsa")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, "values written to stdin are passed to cmd's standard input")
	}()
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(sshKeyPath+"id_rsa.pub", []byte(out),0644)
}

func credentialsCallback(url string, username string, allowedTypes git.CredType) (git.ErrorCode, *git.Cred) {
	ret, cred := git.NewCredSshKey("git", sshKeyPath + "id_rsa.pub", sshKeyPath + "id_rsa", "")
	return git.ErrorCode(ret), &cred
}

func certificateCheckCallback(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
	return 0
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil || os.IsNotExist(err) {
		return false
	}
	return true
}

func Init(url, branch, privateSshKey, path string) {
	if path == "" {
		path = "/tmp/git-resource-request/"
	}
	if !exists(sshKeyPath) {
		os.MkdirAll(sshKeyPath, 0755)
		ioutil.WriteFile(sshKeyPath+"id_rsa", []byte(privateSshKey),0600)
		createSshPubKey()

	}
	if exists(path + "/.git") {
		updateRepo(path)
	} else {
		getRepo(url, branch, path)
	}
}

func getRepo(url, branch, path string) {
	cloneOptions := &git.CloneOptions{}
	cloneOptions.CheckoutBranch = branch
	cloneOptions.FetchOptions = &git.FetchOptions{
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback:      credentialsCallback,
			CertificateCheckCallback: certificateCheckCallback,
		},
	}
	_, err := git.Clone(url, path, cloneOptions)
	if err != nil {
		log.Fatal(err, url)
	}
}

func LastCommit(path, branch string) []map[string]string {
	if path == "" {
		path = "/tmp/git-resource-request/"
	}
	repo, err := git.OpenRepository(path)
	if err != nil {
		log.Fatal(err)
	}
	remoteBranch, err := repo.References.Lookup("refs/remotes/origin/" + branch)
	if err != nil {
		log.Fatal(err)
	}
	commit, err := repo.LookupCommit(remoteBranch.Target())
	if err != nil {
		log.Fatal(err)
	}
	var result []map[string]string

	c := make(map[string]string)
	c["ref"] = commit.Id().String()
	result = append(result, c)
	return result
}

func updateRepo(path string){
	FetchOptions := &git.FetchOptions{
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback:      credentialsCallback,
			CertificateCheckCallback: certificateCheckCallback,
		},
	}
	repo, err := git.OpenRepository(path)
	if err != nil {
		log.Fatal(err)
	}
	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		log.Fatal(err)
	}
	err = remote.Fetch(nil, FetchOptions, "")
	if err != nil {
		log.Fatal(err)
	}
}

func tagResolv(repo *git.Repository, tag string) string {
	if tag != "" {
		var actualNames string
		var actualOid string
		repo.Tags.Foreach(func(name string, id *git.Oid) error {
			actualNames = name
			actualOid = id.String()
			return nil
		})
		re := regexp.MustCompile(tag)
		if re.MatchString(actualNames) {
			commit := actualOid
			return commit
		}
	}
	return ""
}

func CheckoutCommit(commit, tagFilter, path string){
	repo, err := git.OpenRepository(path)
	if err != nil {
		log.Fatal(err)
	}
	if tagFilter != "" {
		var actualNames string
		var actualOid string
		repo.Tags.Foreach(func(name string, id *git.Oid) error {
			actualNames = name
			actualOid = id.String()
			return nil
		})
		re := regexp.MustCompile(tagFilter)
		if re.MatchString(actualNames) {
			commit = actualOid
		}
	}
	oid, err := git.NewOid(commit)
	if err != nil {
		log.Fatal(err)
	}
	repo.SetHeadDetached(oid)
	repo.CheckoutHead(&git.CheckoutOpts{Strategy: git.CheckoutForce})
}

func GetMetaData(ref, tagFilter, br, path string)[]map[string]string  {
	repo, err := git.OpenRepository(path)
	if err != nil {
		log.Fatal(err)
	}
	if tagFilter != "" {
		var actualNames string
		var actualOid string
		repo.Tags.Foreach(func(name string, id *git.Oid) error {
			actualNames = name
			actualOid = id.String()
			return nil
		})
		re := regexp.MustCompile(tagFilter)
		if re.MatchString(actualNames) {
			ref = actualOid
		}
	}
	odb, err := repo.Odb()
	if err != nil {
		log.Fatal(err)
	}
	var result []map[string]string
	odb.ForEach(func(oid *git.Oid) error {
		obj, err := repo.Lookup(oid)
		if err != nil {
			log.Fatal(err)
		}
		if ref == obj.Id().String() {
			commit, err := obj.AsCommit()
			if err != nil {
				log.Fatal(err)
			}
			message := make(map[string]string)
			branch := make(map[string]string)
			committer := make(map[string]string)
			committer["value"] = commit.Committer().Name
			committer["name"] = "committer"
			branch["value"] = br
			branch["name"] = "branch"
			message["value"] = commit.Message()
			message["name"] = "message"
			result = append(result, message, branch, committer)
		}
		return nil
	})
	return result
}

func LastTag(path string) []map[string]string {
	if path == "" {
		path = "/tmp/git-resource-request/"
	}
	repo, err := git.OpenRepository(path)
	if err != nil {
		log.Fatal(err)
	}
	var actualNames string
	err = repo.Tags.Foreach(func(name string, oid *git.Oid) error {
		actualNames = name
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	t := strings.Trim(actualNames, "[]")
	var result []map[string]string
	c := make(map[string]string)
	c["ref"] = strings.Trim(t, "refs/tags/")
	result = append(result, c)
	return result
}