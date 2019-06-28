Concourse Git Resource on Golang + git2go(libgit2)
---

Track the commits or tags in a [git](http://git-scm.com) repository.  
Builded and based on Alpine with extend libraries [musl](https://musl-libc.org) and [libgit2-0.27.7](https://libgit2.org).

[![](https://images.microbadger.com/badges/image/devinotelecom/concourse-git-resource.svg)](https://microbadger.com/images/devinotelecom/concourse-git-resource)

## Source configuration

* `url`: ***Required***. The location of the repository. Scheme may be as `http(s)://...` for HTTP or `ssh://...` for SSH.
Connect protocol is determined by next parameters. If `private_key` is specified will be used SSH, otherwise - HTTP.

* `private_key`: *Optional*. Content of ssh private key file to user when pulling/pushing by SSH protocol.
Now may be used **only RSA keys** - [git2go.v27](https://github.com/libgit2/git2go) > [libgit2 v0.27](https://github.com/libgit2/libgit2) > [libssh2 v1.8](https://libssh2.org)
when keys as `ECDSA` or `ED25519` is not supported. *Waiting release git2go with libssh2 v1.9*.

* `login`: *Optional*. User login for `http(s)://...` scheme in `url`. If not specified the login try to used from `url` - `https://(login)@github.com/...`. 

* `password`: *Optional*. Password for `http(s)://...` scheme in `url`.

* `branch`: *Optional*. The branch to track. If unset for [get](https://concourse-ci.org/get-step.html) step, the repository's default branch is used.
Usually master but [could be different](https://help.github.com/articles/setting-the-default-branch/).

* `tag_regex`: *Optional*. If specified, the resource will only detect commits that have a tag matching the expression that have been made against the branch.
Patterns are [Golang regex](https://golang.org/pkg/regexp/) compatible.

* `paths`: *Optional*. If specified (as a list of [glob](https://github.com/gobwas/glob) patterns), only changes to the specified files will yield new versions from check.

#### Resource configuration examples:
```yaml
resource_types:
- name: git
  type: docker-image
  source: {repository: devinotelecom/concourse-git-resource, tag: latest}


resources:
# All commits in branch master from SSH repository
- name: ssh-repo-master
  type: git
  source:
    url: ssh://git@github.com:devinotelecom/concourse-git-resource.git
    private_key: |
      -----BEGIN RSA PRIVATE KEY-----
      MIIEpQIBAAKCAQEA6iKXxSSUfcgv0KrfaN+nE1xBUEGonDcw94d+pigs36SgmbZX
      ...
      vT1XKXFNTSlGXkPA9lcLUjwvSPorz/oZaasVw+06kBhhMP4QubI81MU=
      -----END RSA PRIVATE KEY-----
    branch: master

# Version tags from HTTP repository with auth
- name: http-repo-tags
  type: git
  source:
    url: https://github.com/devinotelecom/concourse-git-resource.git
    login: concourse
    password: secret_password
    tag_regex: ^v[0-9.]+

# All commits for README.md in HTTP public repository  
- name: http-repo-path
  type: git
  source:
    url: https://github.com/devinotelecom/concourse-git-resource.git
    branch: master
    path: README.md
```

## Behavior

### `check`: Check for changes in repository as new commits or tags.

The repository is cloned (or updated if already present), and any commits/tags from the given version on are returned.
If no old version is given, the only ref for HEAD is returned.

### `in`: Clone the repository, at the given from `check` ref.

Clones the repository to the destination, and locks it down to a given ref. It will return the same given ref as version.

### `out`: Add tag and push to a remote repository.

#### `out` parameters

* `repository`:  ***Required***. The local path of the repository to push to the remote repository.

* `tag_path`: *Optional*. The value should be a path to a file containing the name of the tag. If this is set then HEAD will be tagged.

* `tag_message_path`: *Optional*. If specified the tag will be an annotated by message. The value should be a path to a file containing the annotation message.

#### `out` parameters configuration example:
```yaml
jobs:
- name: Add version tag
  plan:
  - get: ssh-repo-master
  - task: add version tag with message
    config:
     ...
  - put: ssh-repo-master
    params:
      repository: 
      tag_path: git/tag
      tag_message_path: git/tag_message
```

## TO DO:

* Update to [libssh2 v1.9](https://libssh2.org) to usage `ECDSA` or `ED25519`.

* `check`, `in`, `out`: Proxy support.

* `in`: Depth support.

* `in`: Submodule support.

* `in`: GPG verification support.

* `out`: Commit and push support with merge/rebase.

### Contributing

Please make all pull requests to the `master` branch.
