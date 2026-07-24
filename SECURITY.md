# Security Release Process
Harbor is a large growing community devoted in creating a private enterprise-grade registry for all your cloud native assets. The community has adopted this security disclosure and response policy to ensure we responsibly handle critical issues.

## Supported Versions
The Harbor project maintains release branches for the three most recent minor releases. Applicable fixes, including security fixes, may be backported to those three release branches, depending on severity and feasibility. Please refer to [RELEASES.md](https://github.com/goharbor/harbor/blob/main/RELEASES.md) for details.

## Reporting a Vulnerability - Private Disclosure Process
Security is of the highest importance and all security vulnerabilities or suspected security vulnerabilities should be reported to Harbor privately, to minimize attacks against current users of Harbor before they are fixed. Vulnerabilities will be investigated and patched on the next patch (or minor) release as soon as possible. This information could be kept entirely internal to the project.

If you know of a publicly disclosed security vulnerability for Harbor, please **IMMEDIATELY** report it via [GitHub private vulnerability reporting](https://github.com/goharbor/harbor/security/advisories/new) to inform the Harbor Security Team.

**IMPORTANT: Do not file public issues on GitHub for security vulnerabilities**

To report a vulnerability or a security-related issue, use [GitHub private vulnerability reporting](https://github.com/goharbor/harbor/security/advisories/new): open the [Security tab](https://github.com/goharbor/harbor/security) of the affected repository and click **Report a vulnerability**. The report is visible only to you and the Harbor Security Team, which is made up of Harbor maintainers who have committer and release permissions. Reports will be addressed within 3 business days, including a detailed plan to investigate the issue and any potential workarounds to perform in the meantime. Do not report non-security-impacting bugs through this channel. Use [GitHub issues](https://github.com/goharbor/harbor/issues/new/choose) instead. Note that filing a report requires a GitHub account.

### Proposed Report Content
Provide a descriptive title and in the description of the report include the following information:
* Basic identity information, such as your name and your affiliation or company.
* Detailed steps to reproduce the vulnerability  (POC scripts, screenshots, and compressed packet captures are all helpful to us).
* Description of the effects of the vulnerability on Harbor and the related hardware and software configurations, so that the Harbor Security Team can reproduce it.
* How the vulnerability affects Harbor usage and an estimation of the attack surface, if there is one.
* List other projects or dependencies that were used in conjunction with Harbor to produce the vulnerability.

## When to report a vulnerability
* When you think Harbor has a potential security vulnerability.
* When you suspect a potential vulnerability, but you are unsure that it impacts Harbor.
* When you know of or suspect a potential vulnerability on another project that is used by Harbor. For example Harbor has a dependency on Docker, PGSql, Redis, Notary, Trivy, etc.

## Patch, Release, and Disclosure
The Harbor Security Team will respond to vulnerability reports as follows:

1.  The Security Team will investigate the vulnerability and determine its effects and criticality.
2.  If the issue is not deemed to be a vulnerability, the Security Team will close the report with a detailed reason for rejection.
3.  The Security Team will initiate a conversation with the reporter within 3 business days.
4.  If a vulnerability is acknowledged, the Security Team will accept the report as a draft security advisory, and the reporter is added as a collaborator on it. The Security Team will then work on a plan to communicate with the appropriate community, including identifying mitigating steps that affected users can take to protect themselves until the fix is rolled out.
5.  The Security Team will also assess the severity of the vulnerability with a [CVSS](https://www.first.org/cvss/specification-document) score, using the calculator built into the draft advisory. The Security Team makes the final call on the calculated CVSS; it is better to move quickly than making the CVSS perfect. A CVE will be requested from GitHub through the draft advisory; the CVE remains private until the advisory is published.
6.  The Security Team will work on fixing the vulnerability in a temporary private fork associated with the advisory, keeping the patch embargoed, and perform internal testing before preparing to roll out the fix.
7.  The Security Team will provide early disclosure of the vulnerability by emailing the cncf-harbor-distributors-announce@lists.cncf.io mailing list. Distributors can initially plan for the vulnerability patch ahead of the fix, and later can test the fix and provide feedback to the Harbor team. See the section **Early Disclosure to Harbor Distributors List** for details about how to join this mailing list.
8. A public disclosure date is negotiated by the Harbor Security Team, the bug submitter, and the distributors list. We prefer to fully disclose the bug as soon as possible once a user mitigation or patch is available. It is reasonable to delay disclosure when the bug or the fix is not yet fully understood, the solution is not well-tested, or for distributor coordination. The timeframe for disclosure is from immediate (especially if it’s already publicly known) to a few weeks. For a critical vulnerability with a straightforward mitigation, we expect report date to public disclosure date to be on the order of 14 business days. The Harbor Security Team holds the final say when setting a public disclosure date.
9.  Once the fix is confirmed, the Security Team will patch the vulnerability in the next patch or minor release, and backport a patch release into all earlier supported releases. Upon release of the patched version of Harbor, we will follow the **Public Disclosure Process**.

### Public Disclosure Process
The Security Team publishes the security [advisory](https://github.com/goharbor/harbor/security/advisories) to the Harbor community via GitHub, which also publishes the associated CVE. In most cases, additional communication via Slack, Twitter, CNCF lists, blog and other channels will assist in educating Harbor users and rolling out the patched release to affected users.

The Security Team will also publish any mitigating steps users can take until the fix can be applied to their Harbor instances. Harbor distributors will handle creating and publishing their own security advisories.

## Mailing lists
- Join cncf-harbor-distributors-announce@lists.cncf.io for early private information and vulnerability disclosure. Early disclosure may include mitigating steps and additional information on security patch releases. See below for information on how Harbor distributors or vendors can apply to join this list.

## Early Disclosure to Harbor Distributors List
This private list is intended to be used primarily to provide actionable information to multiple distributor projects at once. This list is not intended to inform individuals about security issues.

### Membership Criteria
To be eligible to join the cncf-harbor-distributors-announce@lists.cncf.io mailing list, you should:
1. Be an active distributor of Harbor.
2. Have a user base that is not limited to your own organization.
3. Have a publicly verifiable track record up to the present day of fixing security issues.
4. Not be a downstream or rebuild of another distributor.
5. Be a participant and active contributor in the Harbor community.
6. Accept the Embargo Policy that is outlined below.
7. Has someone who is already on the list vouch for the person requesting membership on behalf of your distribution.

**The terms and conditions of the Embargo Policy apply to all members of this mailing list. A request for membership represents your acceptance to the terms and conditions of the Embargo Policy**

### Embargo Policy
The information that members receive on cncf-harbor-distributors-announce@lists.cncf.io must not be made public, shared, or even hinted at anywhere beyond those who need to know within your specific team, unless you receive explicit approval to do so from the Harbor Security Team. This remains true until the public disclosure date/time agreed upon by the list. Members of the list and others cannot use the information for any reason other than to get the issue fixed for your respective distribution's users.
Before you share any information from the list with members of your team who are required to fix the issue, these team members must agree to the same terms, and only be provided with information on a need-to-know basis.

In the unfortunate event that you share information beyond what is permitted by this policy, you must urgently inform the Harbor Security Team via cncf-harbor-security@lists.cncf.io of exactly what information was leaked and to whom. If you continue to leak information and break the policy outlined here, you will be permanently removed from the list.

### Requesting to Join
Send new membership requests to cncf-harbor-security@lists.cncf.io.
In the body of your request please specify how you qualify for membership and fulfill each criterion listed in the Membership Criteria section above.

## Confidentiality, integrity and availability
We consider vulnerabilities leading to the compromise of data confidentiality, elevation of privilege, or integrity to be our highest priority concerns. Availability, in particular in areas relating to DoS and resource exhaustion, is also a serious security concern. The Harbor Security Team takes all vulnerabilities, potential vulnerabilities, and suspected vulnerabilities seriously and will investigate them in an urgent and expeditious manner.

Note that we do not currently consider the default settings for Harbor to be secure-by-default. It is necessary for operators to explicitly configure settings, role based access control, and other resource related features in Harbor to provide a hardened Harbor environment. We will not act on any security disclosure that relates to a lack of safe defaults. Over time, we will work towards improved safe-by-default configuration, taking into account backwards compatibility.
