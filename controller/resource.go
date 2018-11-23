package controller

func Check(branch, tagFilter, pathSearech, ref, path string) RefResult {
	if branch == "" {
		branch = "master"
	}
	if path == "" {
		path = "/tmp/git-resource-request/"
	}
	if tagFilter != "" {
		return LastTag(path, tagFilter)
	}
	if pathSearech != "" {
		return CheckPaths(path, branch, ref, pathSearech)
	} else {
		return LastCommit(path, branch)
	}
	return nil
}
