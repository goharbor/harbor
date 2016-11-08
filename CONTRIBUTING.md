# Contributing to Harbor

This guide provides information on filing issues and guidelines for open source contributors.

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


## Contribution guidelines

### Pull requests

Pull requests (PR) are always welcome, even they are small fixes like typos or a few lines of code changes. If there will be significant effort, please first document as an issue and get the discussion going before starting to work on it.

Please submit a PR to contain changes bit by bit. A PR consisting of a lot features and code changes may be hard to review. It is recommended to submit PRs in a incremental fasion.

### Design new features

You can propose new designs for existing Harbor features. You can also design
entirely new features. Please do open an issue on Github for discussion first. This is necessary to ensure the overall architecture is consistent and to avoid duplicated work in the roadmap.

### Conventions

Fork Harbor's repository and make changes on your own fork in a new branch. The branch should be named  XXX-description where XXX is the number of the issue. Please run the full test suite on your branch before creating a PR.

Always run [golint](https://github.com/golang/lint) on source code before
committing your changes.

Pull requests should be rebased on top of master without multiple branches mixed into the PR. If your pull requests do not merge cleanly, use `rebase master` in your branch rather than `merge master`.

Update the documentation if you are creating or changing features. Good documentation is as important as the code itself.

Before you submitting any pull request, always squash your commits into logical units of change. A logical unit of change is defined as a set of codes and documents that should be treated as a whole. When possible, compact your commits into one. The commands to use are `git rebase -i` and/or `git push -f`. 

Description of a pull request should refer to all the issues that it addresses. Remember to put a reference to issues (such as `Closes #XXX` and `Fixes #XXX`) in commits so that the issues can be closed when the PR is merged.

### Signing CLA
If it is the first time you are making a PR, be sure to sign the contributor license agreement (CLA) online. A bot will automatically update the PR for the CLA process.

