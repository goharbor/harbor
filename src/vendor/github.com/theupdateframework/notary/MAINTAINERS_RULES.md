# Maintainers Rules

This document lays out some basic rules and guidelines all maintainers are expected to follow.
Changes to the [Acceptance Criteria](#hard-acceptance-criteria-for-merging-a-pr) for merging PRs require a ceiling(two-thirds) supermajority from the maintainers.
Changes to the [Repo Guidelines](#repo-guidelines) require a simple majority.

## Hard Acceptance Criteria for merging a PR:

- 2 LGTMs are required when merging a PR
- If there is obviously still discussion going on in the PR, even with 2 LGTMs, let the discussion resolve before merging. If you’re not sure, reach out to the maintainers involved in the discussion.
- All checks must be green
    - There are limited mitigating circumstances for this, like if the docs builds are just broken and that’s the only test failing.
    - Adding or removing a check requires simple majority approval from the maintainers.

## Repo Guidelines:

- Consistency is vital to keep complexity low and understandable.
- Automate as much as possible (we don’t have guidelines about coding style for example because we’ve automated fmt, vet, lint, etc…).
- Try to keep PRs small and focussed (this is not always possible, i.e. builder refactor, storage refactor, etc… but a good target).

## Process for becoming a maintainer:

- Invitation is proposed by an existing maintainer.
- Ceiling(two-thirds) supermajority approval from existing maintainers (including vote of proposing maintainer) required to accept proposal.
- Newly approved maintainer submits PR adding themselves to the MAINTAINERS file.
- Existing maintainers publicly mark their approval on the PR.
- Existing maintainer updates repository permissions to grant write access to new maintainer.
- New maintainer merges their PR.

## Removing maintainers

It is preferrable that a maintainer gracefully removes themselves from the MAINTAINERS file if they are
aware they will no longer have the time or motivation to contribute to the project. Maintainers that
have been inactive in the repo for a period of at least one year should be contacted to ask if they
wish to be removed.

In the case that an inactive maintainer is unresponsive for any reason, a ceiling(two-thirds) supermajority
vote of the existing maintainers can be used to approve their removal from the MAINTAINERS file, and revoke
their merge permissions on the repository.