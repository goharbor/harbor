# Versioning and Release
This document describes the versioning and release process of Harbor. This document is a living document, contents will be updated according to each releases.

## Releases
Harbor releases will be versioned using dotted triples, similar to [Semantic Version](http://semver.org/). For this specific document, we will refer to the respective components of this triple as `<major>.<minor>.<patch>`. The version number may have additional information, such as "-rc1,-rc2,-rc3" to mark release candidate builds for earlier access. Such releases will be considered as "pre-releases".

### Major and Minor Releases
Major and minor releases of Harbor will be branched from master when the release reaches to `RC(release candidate)` state. The branch format should follow `release-<major>.<minor>.0`. For example, once the release `v1.0.0` reaches to RC, a branch will be created with the format `release-1.0.0`. When the release reaches to `GA(General Available)` state, The tag with format `v<major>.<minor>.<patch>` and should be made with command `git tag -s v<major>.<minor>.<patch>`. The release cadency is around 3 months, might be adjusted based on open source event, but will communicate it clearly.

### Patch releases
Patch releases are based on the major/minor release branch, the release cadency for patch release of recent minor release is one month to solve critical community and security issues. The cadency for patch release of recent minus two minor releases are on-demand driven based on the severity of the issue to be fixed.

### Pre-releases
`Pre-releases:mainly the different RC builds` will be compiled from their corresponding branches. Please note they are done to assist in the stabilization process, no guarantees are provided.

### Minor Release Support Matrix
| Version | Supported          |
| ------- | ------------------ |
| Harbor v1.8.x   | :white_check_mark: |
| Harbor v1.9.x   | :white_check_mark: |
| Harbor v1.10.x   | :white_check_mark: |

### Upgrade path and support policy
The upgrade path for Harbor is (1) 1.0.x patch releases are always compatible with its major and minor version. For example, previous released 1.8.x can be upgraded to most recent 1.8.4 release. (2) Harbor only supports two previous minor releases to upgrade to current minor release. For example, 1.9.0 will only support 1.7.0 and 1.8.0 to upgrade from, 1.6.0 to 1.9.0 is not supported. One should upgrade to 1.8.0 first, then to 1.9.0.
The Harbor project maintains release branches for the three most recent minor releases, each minor release will be maintained for approximately 9 months. There is no mandated timeline for major versions and there are currently no criteria for shipping a new major version (i.e. Harbor 2.0.0).

### Next Release
The activity for next release will be tracked in the [up-to-date project board](https://github.com/orgs/goharbor/projects/1). If your issue is not present in the corresponding release, please reach out to the maintainers to add the issue to the project board.
