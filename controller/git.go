package controller

import (
	"github.com/libgit2/git2go"
	log "github.com/sirupsen/logrus"
	"os"
	"io/ioutil"
	"os/exec"
	"io"
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
	PrivateKey string `json:"private_key"`
}

type MetadataArry struct {
	Version Ref `json:"version"`
	Metadata []map[string]string `json:"metadata"`
}

var (
	//path = "/tmp/git-resource-request/"
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
		updateRepo(path, branch)
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

func LastCommit(path string)([]map[string]string) {
	if path == "" {
		path = "/tmp/git-resource-request/"
	}
	allCommit := getLog(path)
	for i, c := range allCommit {
		if i == 0 {
			var lastCommit []map[string]string
			lastCommit = append(lastCommit, c)
			return lastCommit
		}
	}
	return nil
}

func getLog(path string) ([]map[string]string) {
	repo, err := git.OpenRepository(path)
	if err != nil {
		log.Fatal(err)
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
		commit := obj.Id().String()

		c := make(map[string]string)
		c["ref"] = commit
		result = append(result, c)

		return nil
	})

	return result
}

func updateRepo(path, branch string){
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

	checkoutBranch(repo, branch)
}

func CheckoutCommit(commit, path string){
	repo, err := git.OpenRepository(path)
	if err != nil {
		log.Fatal(err)
	}

	oid, err := git.NewOid(commit)
	if err != nil {
		log.Fatal(err)
	}

	repo.SetHeadDetached(oid)
	repo.CheckoutHead(&git.CheckoutOpts{Strategy: git.CheckoutForce})

}

func checkoutBranch(repo *git.Repository, branch string) error {
	remoteBranch, err := repo.References.Lookup("refs/remotes/origin/" + branch)
	if err != nil {
		return err
	}

	localCommit, err := repo.LookupCommit(remoteBranch.Target())
	if err != nil {
		return err
	}
	tree, err := repo.LookupTree(localCommit.TreeId())
	if err != nil {
		return err
	}
	checkoutOpts := &git.CheckoutOpts{
		Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing | git.CheckoutAllowConflicts | git.CheckoutUseTheirs,
	}

	err = repo.CheckoutTree(tree, checkoutOpts)
	if err != nil {
		return err
	}

	repo.SetHead("refs/heads/" + branch)

	return nil
}

func GetMetaData(ref, br, path string)[]map[string]string  {
	repo, err := git.OpenRepository(path)
	if err != nil {
		log.Fatal(err)
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
			branch["value"] = br
			branch["name"] = "branch"
			message["value"] = commit.Message()
			message["name"] = "message"
			result = append(result, message)
		}
		return nil
	})

	return result
}