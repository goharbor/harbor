# Internal Notes for Groq

## Releasing

In order to publish a release, simpy manually create a release via the Github UI from the `groq_release` branch. You should pick an appropriate semver tag for the release based on the upstream `goharbor/harbor` release, with a simple monotonically increasing `-groqN` suffix. For example, if the `groq_release` branch has been rebased on the upstream `v2.13.1` release tag, and this is the first Groq release, the release tag would be `v2.13.1-groq1`. 


## Syncing Upstream

> Do not simply sync changes from `main`, pick a specific release tag from the `goharbor/harbor` repository (e.g., `v2.13.1`).

To synchronize changes from upstream, rebase the `groq_release` branch on an appropriate upstream release tag. 

For example, assuming you have the upstream `goharbor/harbor` remote added and want to rebase onto `v2.13.1`:
```sh
git fetch -a --tags && git rebase -i tags/v2.13.1
```


## Changes

We saw it important to maintain a separate fork of Harbor to handle a couple of specific concerns:

1. We need to add support for Google GAR and GCP Workload Identity Federation ([upstream PR](https://github.com/goharbor/harbor/pull/22091)).

2. We're concerned that Harbor does not currenty coalesce concurrent remote requests for the same image reference (upstream PR WIP). 
