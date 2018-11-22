package controller

func Check(branch, tagFilter, path string) RefResult {
	if branch == "" {
		branch = "master"
	}
	if path == "" {
		path = "/tmp/git-resource-request/"
	}
	if tagFilter != "" {
		return LastTag(path, tagFilter)
	} else {
		return LastCommit(path, branch)
	}
	return nil
}
