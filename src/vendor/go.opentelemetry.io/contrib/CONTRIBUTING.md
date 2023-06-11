# Contributing to opentelemetry-go-contrib

Welcome to the OpenTelemetry Go contrib repository!
Thank you for your interest in contributing to this project.
Before you start please be sure to read through these contributing requirements and recommendations.

## Become a Contributor

All contributions to this project MUST be licensed under this project's [license](LICENSE).
You will need to sign the [CNCF CLA](https://identity.linuxfoundation.org/projects/cncf) before your contributions will be accepted.

## Filing Issues

Sensitive security-related issues should be reported to <cncf-opentelemetry-tc@lists.cncf.io>. See the [security policy](https://github.com/open-telemetry/opentelemetry-go-contrib/security/policy) for details.

When reporting bugs, please be sure to include the following.

- What version of Go and opentelemetry-go-contrib are you using?
- What operating system and processor architecture are you using?
- What did you do?
- What did you expect to see?
- What did you see instead?

For instrumentation requests, please see the [instrumentation documentation](./instrumentation/README.md#new-instrumentation).

## Contributing Code

The project welcomes code patches, but to make sure things are well coordinated you should discuss any significant change before starting the work.
It's recommended that you signal your intention to contribute in the issue tracker, either by [filing a new issue](https://github.com/open-telemetry/opentelemetry-go-contrib/issues/new) or by claiming an [existing one](https://github.com/open-telemetry/opentelemetry-go-contrib/issues).

### Style Guide

* Code should conform to the [opentelemetry-go Style Guide](https://github.com/open-telemetry/opentelemetry-go/blob/main/CONTRIBUTING.md#style-guide).
* Make sure to run `make precommit` - this will find and fix issues with the code formatting.

### Pull Requests

All pull requests need to be made from [a fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo) of this repository.
Changes should be made using [the GitHub flow](https://guides.github.com/introduction/flow/) and submitted as a pull request to this repository.

A pull request is considered ready to merge when the following criteria are meet.

* It has received two approvals from Approvers/Maintainers (at different companies).
* All feedback has been addressed. Be sure to "Resolve" all comments that have been addressed to signal this.
* Any substantive changes submitted after an Approval removes that Approval.
  You will need to manually clear these prior Approval reviews to indicate to the reviewer that they need to resubmit their review.
  This includes changes resulting from other feedback.
  Unless the approver explicitly stated that their approval will persist across changes it should be assumed that the pull request needs their review again.
  Other project members (e.g. approvers, maintainers) can help with this if there are any questions or if you forget to clear reviews.
* If the changes are not trivial, cosmetic, or for documentation or dependencies only, the pull request will need to be open for review for at least one working day.
  This gives people reasonable time to review.
* `CHANGELOG.md` has been updated to reflect what has been added, changed, removed, or fixed from the end users perspective.
  See [how to keep a changelog](https://keepachangelog.com/en/1.0.0/).
* Urgent fixes can take exception as long as it has been actively communicated.

Any Maintainer can merge the pull request once it is ready to merge.

### Draft Pull Requests

It can be helpful at times to publish your incomplete changes.
To do this create [a draft pull request](https://github.blog/2019-02-14-introducing-draft-pull-requests/).
Additionally, you can prefix the pull request title with `[WIP]`.

## Where to Get Help

You can connect with us in our [slack channel](https://cloud-native.slack.com/archives/C01NPAXACKT).

The Go special interest group (SIG) meets regularly.
See the OpenTelemetry [community](https://github.com/open-telemetry/community#golang-sdk) repo for information on this and other language SIGs.
See the [public meeting notes](https://docs.google.com/document/d/1A63zSWX0x2CyCK_LoNhmQC4rqhLpYXJzXbEPDUQ2n6w/edit#heading=h.9tngw7jdwd6b) for a summary description of past meetings.

## Approvers and Maintainers

Approvers:

- [Evan Torrie](https://github.com/evantorrie), Verizon Media
- [Josh MacDonald](https://github.com/jmacd), LightStep
- [Sam Xie](https://github.com/XSAM), Cisco/AppDynamics
- [David Ashpole](https://github.com/dashpole), Google
- [Gustavo Silva Paiva](https://github.com/paivagustavo), LightStep
- [Robert Pająk](https://github.com/pellared), Splunk
- [Chester Cheung](https://github.com/hanyuancheung), Tencent

Maintainers:

- [Aaron Clawson](https://github.com/MadVikingGod), LightStep
- [Anthony Mirabella](https://github.com/Aneurysm9), Amazon
- [Tyler Yahn](https://github.com/MrAlias), Splunk

### Become an Approver or a Maintainer

See the [community membership document in OpenTelemetry community
repo](https://github.com/open-telemetry/community/blob/main/community-membership.md).
