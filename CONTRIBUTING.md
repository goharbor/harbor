# Contributing to Harbor

## Welcome

Welcome to Harbor! This guide provides information on filing issues and guidelines for open source contributors. **Please leave comments / suggestions if you find something is missing or incorrect.**

Contributors are encouraged to collaborate using the following resources in addition to the GitHub [issue tacker](https://github.com/vmware/harbor/issues):
* [Slack](https://vmwarecode.slack.com/messages/harbor): If you don't have an @vmware.com or @emc.com email, please sign up at [VMware {code}](https://code.vmware.com/join/) to get a Slack invite.
* Mail group: Use harbor-dev@googlegroups.com for discussion on Harbor development and contribution. To subscribe, send an email to harbor-dev+subscribe@googlegroups.com .

## Getting Started

### Sign the CLA

Before doing contributions, you must sign the CLA. If it is the first time you're making a PR, please make sure to sign the contributor license agreement (CLA) online. A bot will automatically update the PR for the CLA process.

### Fork Repository

Fork the Harbor repository on GitHub to your personal account.
```
#Set golang environment
export GOPATH=$HOME/go
mkdir -p $GOPATH/src/github.com/vmware

#Get code
go get github.com/vmware/harbor
cd $GOPATH/src/github.com/vmware/harbor

#Track repository under your personal account
git config push.default nothing # Anything to avoid pushing to vmware/harbor by default
git remote rename origin vmware 
git remote add $USER git@github.com:$USER/harbor.git
git fetch $USER

```
**NOTES:** Note that GOPATH can be any directory, the example above uses $HOME/go. Change $USER above to your own GitHub username.

To build the project, please refer the [build](docs/compile_guide.md) guideline.

### Repository Structure

Here are the basic structure of the harbor code base. Some of the key folders / files are commented for your references.
```
.
...
├── Makefile      #Make file for compile and build code
├── contrib       #<TODO>
├── docs          #Keep documents here
├── make          #Resource for build and setup Harbor environment
...
├── src           #Source code folder
├── tests         #Test cases for API / e2e testings
└── tools         #<TODO>
...
```

The folder graph below shows the structure of the source code folder `harbor/src`, which will be your primary working directory. The key folders are also commented.
```
#[TODO]
.
├── adminserver          # Source code for the admin server component
│   ├── api
│   ├── auth
│   ├── client
│   ├── handlers
│   ├── systemcfg
│   └── systeminfo
├── common              # Source code for some general components like dao etc.
│   ├── api
│   ├── config
│   ├── dao
│   ├── models
│   ├── notifier
│   ├── scheduler
│   ├── secret
│   ├── security
│   └── utils
├── jobservice          # Source code for the job service component
│   ├── api
│   ├── config
│   ├── job
│   ├── replication
│   ├── scan
│   └── utils
├── ui                  # Source code for the harbor service component
│   ├── api
│   ├── auth
│   ├── config
│   ├── controllers
│   ├── filter
│   ├── promgr
│   ├── proxy
│   ├── service
│   ├── static
│   ├── utils
│   └── views
├── ui_ng               # The code of harbor web UI
│   ├── e2e
│   ├── lib             # Source code of harbor-ui npm library which includes the main UI components of web UI
│   └── src             # General web page UI code of Harbor
└── vendor              # Go code dependencies
    ├── github.com
    ├── golang.org
    ├── google.golang.org
    └── gopkg.in
```

### Setup Development Environment

#### Go
Harbor backend is written in [Go](http://golang.org/). If you don't have a Harbor backend service development environment, please [set one up](https://golang.org/doc/code.html).

|  Harbor  |  Requires Go  |
|----------|---------------|
|   1.1    |    1.7.3      |
|   1.2    |    1.7.3      |
|   1.3    |    1.9.2      |

Ensure your GOPATH and PATH have been configured in accordance with the Go environment instructions.

**Dependency Management:** Harbor uses [dep](https://github.com/golang/dep) for dependency management of go code.  The official maintainers will take the responsibility for managing the code in `vendor` directory.  Please don't try to submit PR to update the dependency code, open an issue instead.  If your PR requires a change in the vendor code please make sure you discuss it with maintainers in advance.

#### Web

Harbor web UI is built based on [Clarity](https://vmware.github.io/clarity/) and [Angular](https://angular.io/) web framework. To setup web UI development environment, please make sure the [npm](https://www.npmjs.com/get-npm) tool is installed firstly.

|  Harbor  |  Requires Angular  |  Requires Clarity  |
|----------|--------------------|--------------------|
|   1.1    |      2.4.1         |       0.8.7        |
|   1.2    |      4.1.3         |       0.9.8        |
|   1.3    |      4.3.0         |       0.10.17      |

**Npm Package Dependency:** Run the following commands to restore the package dependencies.
```
#For the web UI
cd $REPO_DIR/src/ui_ng
npm install

#For the UI library
cd $REPO_DIR/src/ui_ng/lib
npm install
```


To run the code, please refer the [build](docs/compile_guide.md) guideline.

## Contribute Workflow

Pull requests (PR) are always welcome, even they are small fixes like typos or a few lines of code changes. If there will be significant effort, please first document as an issue and get the discussion going before starting to work on it.

Please submit a PR to contain changes bit by bit. A PR consisting of a lot features and code changes may be hard to review. It is recommended to submit PRs in a incremental fasion.

The graphic shown below describes the overall workflow about how to contribute code to Harbor repository.
![contribute workflow](docs/img/workflow.png)

### Fork and clone

Fork the Harbor repository and clone the code to your local workspace. Per Go's [workspace instructions](https://golang.org/doc/code.html#Workspaces), place Harbor's code on your `GOPATH`. Refer section [Fork Repository](#fork-repository) for details.

Define a local working directory:
```
working_dir=$GOPATH/src/github.com/vmware
```

Set user to match your github profile name:
```
user={your github profile name}
```

Both `$working_dir` and `$user` are mentioned in the figure above.

### Branch
Changes should be made on your own fork in a new branch. The branch should be named  `XXX-description` where XXX is the number of the issue. Pull requests should be rebased on top of master without multiple branches mixed into the PR. If your pull requests do not merge cleanly, use commands listed below to get it up to date.

```
#vmware is the origin upstream

cd $working_dir/kubernetes
git fetch vmware
git checkout master
git rebase vmware/master
```

Branch from the updated `master` branch:

```
git checkout -b my_feature
```

### Develop, Build and Test

Write code on the new branch in your fork. The coding style used in Harbor is suggested by the Golang community. See the [style doc](https://github.com/golang/go/wiki/CodeReviewComments) for details.

Try to limit column width to 120 characters for both code and markdown documents such as this one.

Always run [golint](https://github.com/golang/lint) on source code before
committing your changes.
```
#In the #working_dir/harbor, run

go list ./... | grep -v -E 'vendor|tests' | xargs -L1 fgt golint

```

Unit test cases should be added to cover the new code. Unit test framework for backend services is using [go testing](https://golang.org/doc/code.html#Testing). The UI library test framework is built based on [Jasmine](http://jasmine.github.io/2.4/introduction.html) and [Karma](https://karma-runner.github.io/1.0/index.html), please refer [Angular Testing](https://angular.io/guide/testing) for more details.

Run go test cases:
```
#cd #working_dir/src/[package]
go test -v ./...
```

Run UI library test cases:
```
#cd #working_dir/src/ui_ng/lib
npm run test
```

To build code, please refer [build](docs/compile_guide.md) guideline.

###  Keep sync with upstream

Once your branch gets out of sync with the vmware/master branch, use the following commands to update:
```
git checkout my_feature
git fetch -a
git rebase vmware/master

```

Please don't use `git pull` instead of the above `fetch / rebase`. `git pull` does a merge, which leaves merge commits. These make the commit history messy and violate the principle that commits ought to be individually understandable and useful (see below). You can also consider changing your `.git/config` file via git config `branch.autoSetupRebase` always to change the behavior of `git pull`.

### Commit

Commit your changes if they're ready:
```
#git add -A
git commit
git push --force-with-lease $user my_feature
```

The commit message should follow the convention on [How to Write a Git Commit Message](http://chris.beams.io/posts/git-commit/). Be sure to include any related GitHub issue references in the commit message. See [GFM syntax](https://guides.github.com/features/mastering-markdown/#GitHub-flavored-markdown) for referencing issues and commits.

To help write conforming commit messages, it is recommended to set up the [git-good-commit](https://github.com/tommarshall/git-good-commit) commit hook. Run this command in the Harbor repo's root directory:

```
curl https://cdn.rawgit.com/tommarshall/git-good-commit/v0.6.1/hook.sh > .git/hooks/commit-msg && chmod +x .git/hooks/commit-msg
```
### Squash Commits

Before you submitting any pull request, always squash your commits into logical units of change. A logical unit of change is defined as a set of codes and documents that should be treated as a whole. When possible, compact your commits into one. The commands to use are `git rebase -i` and/or `git push -f`. 

### Automated Testing
[TODO:]

### Push and Create PR
When ready for review, push your branch to your fork repository on `github.com`:
```
git push --force-with-lease $user my_feature

```

Then visit your fork at https://github.com/$user/harbor and click the `Compare & Pull Request` button next to your `my_feature` branch to create a new pull request (PR). Description of a pull request should refer to all the issues that it addresses. Remember to put a reference to issues (such as `Closes #XXX` and `Fixes #XXX`) in commits so that the issues can be closed when the PR is merged.

Once your pull request has been opened it will be assigned to one or more reviewers. Those reviewers will do a thorough code review, looking for correctness, bugs, opportunities for improvement, documentation and comments, and style.

Commit changes made in response to review comments to the same branch on your fork.

## Reporting issues

It is a great way to contribute to Harbor by reporting an issue. Well-written and complete bug reports are always welcome! Please open an issue on Github and follow the template to fill in required information.

Before opening any issue, please look up the existing [issues](https://github.com/vmware/harbor/issues) to avoid submitting a duplication.
If you find a match, you can "subscribe" it to get notified on updates. If you have additional helpful information about the issue, please leave a comment.

When reporting issues, always include:

* Version of docker engine and docker-compose
* Configuration files of Harbor
* Log files in /var/log/harbor/ 

Because the issues are open to the public, when submitting the log and configuration files, be sure to remove any sensitive information, e.g. user name, password, IP address, and company name. You can
replace those parts with "REDACTED" or other strings like "****".

Be sure to include the steps to reproduce the problem if applicable. It can help us understand and fix your issue faster.

## Documenting 

Update the documentation if you are creating or changing features. Good documentation is as important as the code itself.

The main location for the document is `docs/` folder. The images referred in documents can be placed in `docs/img`.

Document is written with Markdown text. See [Writting on GitHub](https://help.github.com/categories/writing-on-github/) for more details.

## Design new features

You can propose new designs for existing Harbor features. You can also design
entirely new features. Please do open an issue on Github for discussion first. This is necessary to ensure the overall architecture is consistent and to avoid duplicated work in the roadmap.
