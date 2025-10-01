# Versioning and Release
This document describes the versioning and release process of Harbor. This is a living document and will be updated with each release.

## Releases
Harbor releases use a three-part version number, similar to [Semantic Versioning](http://semver.org/): `<major>.<minor>.<patch>`. The version number may include additional information, such as `-rc1`, `-rc2`, or `-rc3` to indicate release candidate builds for early access. These are considered pre-releases.

### Major and Minor Releases
Major and minor releases are branched from `main` when the release reaches the RC (release candidate) stage. The branch should be named `release-<major>.<minor>.0`. For example, when release `v1.0.0` reaches RC, a branch named `release-1.0.0` was created. When the release reaches GA (General Availability), a tag in the format `v<major>.<minor>.<patch>` should be created using the command `git tag -s v<major>.<minor>.<patch>`. The release cadence is approximately every 3 months, but may be adjusted based on events. Any changes to the cadence will be communicated clearly.

### Patch Releases
Patch releases are based on the major/minor release branch. The cadence for patch releases of the most recent minor release is one month, to address critical community and security issues. Patch releases for the two previous minor releases are made on demand, depending on the severity of the issues to be fixed.

### Pre-releases
Pre-releases, mainly the different RC builds, are compiled from their corresponding branches. These builds are intended to help with stabilization, but no guarantees are provided.

### Support Policy

The Harbor project maintains the three latest minor releases. Each minor release is supported for approximately 9 months.

### Supported Versions

| Version        | Supported          |
|----------------|--------------------|
| Harbor v2.14.x | :white_check_mark: |
| Harbor v2.13.x | :white_check_mark: |
| Harbor v2.12.x | :white_check_mark: |
| Harbor v2.11.x | :x:                |


### Upgrade Policy

Upgrades to a new minor release are only supported from the two previous minor releases. 
For example, to upgrade to `v2.14.x`, you must be on a `v2.13.x` or `v2.12.x` release. 
Upgrading directly from `v2.11.x` to `v2.14.x` is not supported; 
you must first upgrade to `v2.12.x`, then to `v2.14.x`.

### Next Release
The activity and release dates of the next release are tracked in the [release plan wiki](https://github.com/goharbor/harbor/wiki/Release-plans).


### Publishing a New Release

The following steps outline what to do when planning and publishing a release. Depending on the type of release (major, minor, or patch), not all steps may be required.

1. Prepare information about what's new in the release.
   * For every release, update the documentation to reflect changes. See the [goharbor/website](https://github.com/goharbor/website) repo for details on creating release documentation. All documentation should be published by the time the release is out.
   * For every release, write release notes. See [previous releases](https://github.com/goharbor/harbor/releases) for examples of what to include.
   * For major and minor releases, write a blog post highlighting new features. Plan to publish this on the same day as the release. Highlight the main themes or focus areas, such as security, bug fixes, or feature improvements. If new features or workflows are introduced, consider writing additional blog posts to help users learn about them. These can be published after the release date (not all blogs need to be published at once).
2. Release the new version. Make the new version, documentation updates, and blog posts available.
3. Announce the release and thank contributors. For all releases, do the following:
   * In all community messages, include a brief list of highlights and links to the new release blog, release notes, or download location. Also, give shoutouts to community members whose contributions are included in the release.
   * Send an email to the community via the [mailing list](https://lists.cncf.io/g/harbor-users).
   * Post a message in the Harbor [Slack channel](https://cloud-native.slack.com/archives/CC1E09J6S).
   * Post to social media. Maintainers are encouraged to post or repost from the Harbor account to help spread the word.
