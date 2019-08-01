package resource

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"concourse-git-resource/common"
	"concourse-git-resource/git"
)

type OutPayload struct {
	common.Payload
	Params struct {
		Repository     string
		TagPath        string `json:"tag_path"`
		TagMessagePath string `json:"tag_message_path"`
	}
}

func NewOutPayload(stdin []byte) *OutPayload {
	var p OutPayload
	common.Parse(&p, stdin)

	return &p
}

func Out(payload *OutPayload, workdir string, printer *common.Printer) {
	var (
		err    error
		tag    string
		tagMsg string
	)

	wd := strings.TrimSuffix(workdir, string(filepath.Separator)) + string(filepath.Separator)

	if payload.Params.TagPath != "" {
		tag, err = getFileContent(wd + payload.Params.TagPath)
		if err != nil {
			panic(fmt.Sprintf("tag_path at \"%s\" not found", payload.Params.TagPath))
		}
	}

	if tag == "" {
		panic(fmt.Sprintf("tag_path at \"%s\" is empty", payload.Params.TagPath))
	}

	if payload.Params.TagMessagePath != "" {
		tagMsg, err = getFileContent(wd + payload.Params.TagMessagePath)
		if err != nil {
			panic(fmt.Sprintf("tag_message_path at \"%s\" not found", payload.Params.TagMessagePath))
		}
	}

	repo, err := git.Open(
		wd+payload.Params.Repository,
		payload.Source.Branch,
		git.RepositoryParams{
			RemoteUrl:     payload.Source.Url,
			HttpLogin:     payload.Source.Login,
			HttpPassword:  payload.Source.Password,
			SshPrivateKey: payload.Source.PrivateKey,
		},
	)
	if err != nil {
		panic(fmt.Sprint("Open repository error:", err))
	}
	defer repo.Close()

	commit := repo.CreateTag(tag, tagMsg)
	repo.PushTag(tag)

	var meta []map[string]string
	meta = append(meta, map[string]string{"name": "Commit", "value": commit.Id})
	meta = append(meta, map[string]string{"name": "Tag", "value": commit.Tag})
	meta = append(meta, map[string]string{"name": "Message", "value": strings.TrimSpace(commit.Message)})
	meta = append(meta, map[string]string{"name": "Date", "value": commit.Author.When.Format(time.RFC822)})
	meta = append(meta, map[string]string{"name": "Author", "value": fmt.Sprintf("%s <%s>", commit.Author.Name, commit.Author.Email)})

	printer.PrintData(map[string]interface{}{
		"version":  common.Version{Reference: commit.Id},
		"metadata": meta,
	})
}

// Read from file path and return trimmed text
func getFileContent(path string) (string, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(buf)), nil
}
