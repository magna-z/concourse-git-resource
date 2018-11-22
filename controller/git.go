package controller

import (
	"github.com/libgit2/git2go"
	log "github.com/sirupsen/logrus"
	"os"
	"io/ioutil"
	"os/exec"
	"io"
	"regexp"
	"sort"
	"strings"
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
	PathSearch string `json:"path_search"`
	PrivateKey string `json:"private_key"`
}

type MetadataJson struct {
	Version Ref `json:"version"`
	Metadata []map[string]string `json:"metadata"`
}

type Tag struct {
	Name string
	Commit string
	When int64
}

type RefResult []map[string]string

type MetadataResult []Metadata

type Metadata struct {
	commit string
	author string
	date string
	commiter string
	message string
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

func LastCommit(path, branch string) RefResult {
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
	var result RefResult
	c := make(map[string]string)
	c["ref"] = commit.Id().String()
	result = append(result, c)
	defer repo.Free()
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

func GetMetaData(ref, path string) []map[string]string {
	repo, err := git.OpenRepository(path)
	if err != nil {
		log.Fatal(err)
	}
	var oid *git.Oid
	obj, _ := repo.References.Lookup("refs/tags/" + ref)
	if obj != nil {
		oid = obj.Target()
	} else {
		oid, _ = git.NewOid(ref)
	}

	o, err := repo.LookupCommit(oid)
	var result []map[string]string

	commit := make(map[string]string)
	commit["name"] = "commit"
	commit["value"] = o.Id().String()

	author := make(map[string]string)
	author["name"] =  "author"
	author["value"] = o.Committer().Name

	whenCommit := make(map[string]string)
	whenCommit["name"] = "date"
	whenCommit["value"] = o.Committer().When.String()

	branch := make(map[string]string)
	branch["name"] = "branch"
	branch["value"] = ""

	tag := make(map[string]string)
	tag["name"] = "tag"
	tag["value"] = ""

	message := make(map[string]string)
	message["name"] = "message"
	message["value"] = o.Message()



	result = append(result, commit, author, whenCommit, branch, tag, message)
	defer repo.Free()
	return result
}

func tagWhen(repo *git.Repository, oid *git.Oid)int64 {
	obj, err := repo.Lookup(oid)
	if err != nil {
		log.Fatal(err)
	}
	if obj.Type() == git.ObjectTag {
		o, err := repo.LookupTag(oid)
		if err != nil {
			log.Fatal(err)
		}
		return o.Tagger().When.Unix()
	}
	if obj.Type() == git.ObjectCommit {
		o, err := repo.LookupCommit(oid)
		if err != nil {
			log.Fatal(err)
		}
		return o.Committer().When.Unix()
	}
	return 0
}

func listTag(path, tagFilter string) []Tag {
	repo, err := git.OpenRepository(path)
	if err != nil {
		log.Fatal(err)
	}
	var result []Tag
	err = repo.Tags.Foreach(func(name string, id *git.Oid) error {
		t := Tag{Name:name, Commit:id.String(), When: tagWhen(repo, id)}
		re := regexp.MustCompile(tagFilter)

		if re.MatchString(name) {
			result = append(result, t)
		}
		return nil
	})
	defer repo.Free()
	return result

}

func lastTags(listTag []Tag) RefResult {
	sort.Slice(listTag, func(i, j int) bool {
		if listTag[i].When < listTag[j].When {
			return true
		}
		if listTag[i].When > listTag[j].When {
			return false
		}
		return listTag[i].When < listTag[j].When
	})

	lt := listTag[len(listTag)-1]

	lastTag := make(map[string]string)
	lastTag["ref"]= strings.Trim(lt.Name, "refs/tags/")
	var result RefResult
	result = append(result, lastTag)
	return result
}

func LastTag(path, tagFilter string)RefResult  {
	list := listTag(path, tagFilter)
	lastTag := lastTags(list)
	return lastTag
}

func Checkout(path, ref string) {
	repo, err := git.OpenRepository(path)
	if err != nil {
		log.Fatal(err)
	}
	var oid *git.Oid
	obj, _ := repo.References.Lookup("refs/tags/"+ref)
	if obj != nil {
		oid = obj.Target()
	} else {
		oid, _ = git.NewOid(ref)
	}
	repo.SetHeadDetached(oid)
	repo.CheckoutHead(&git.CheckoutOpts{Strategy: git.CheckoutForce})
	defer repo.Free()

}