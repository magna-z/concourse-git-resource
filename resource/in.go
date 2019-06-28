package resource

import (
	"fmt"
	"strings"
	"time"

	"concourse-git-resource/common"
	"concourse-git-resource/git"
)

type InPayload struct {
	common.Payload
}

func NewInPayload(stdin []byte) *InPayload {
	var p InPayload
	common.Parse(&p, stdin)

	return &p
}

func In(payload *InPayload, path string, printer *common.Printer) {
	repo := git.Open(path, payload.Source.Branch, git.RepositoryParams{
		RemoteUrl:     payload.Source.Url,
		HttpLogin:     payload.Source.Login,
		HttpPassword:  payload.Source.Password,
		SshPrivateKey: payload.Source.PrivateKey,
	})
	defer repo.Close()

	var commit *git.Commit
	if payload.Source.TagRegex != "" {
		commit = repo.CheckoutTag(payload.Version.Reference)
	} else {
		commit = repo.CheckoutCommit(payload.Version.Reference)
	}

	var meta []map[string]string
	meta = append(meta, map[string]string{"name": "Commit", "value": commit.Id})
	meta = append(meta, map[string]string{"name": "Message", "value": strings.TrimSpace(commit.Message)})
	meta = append(meta, map[string]string{"name": "Date", "value": commit.Author.When.Format(time.RFC822)})
	meta = append(meta, map[string]string{"name": "Author", "value": fmt.Sprintf("%s <%s>", commit.Author.Name, commit.Author.Email)})

	printer.PrintData(map[string]interface{}{
		"version":  payload.Version,
		"metadata": meta,
	})
}
