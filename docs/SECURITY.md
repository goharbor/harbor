## Supported Versions
 
| Version | Supported          |
| ------- | ------------------ |
| Harbor v1.5.x   | :white_check_mark: |
| Harbor v1.6.x   | :white_check_mark: |
| Harbor v1.7.x   | :white_check_mark: |
| Harbor v1.8.x   | :white_check_mark: |
| Harbor v1.9.x   | :white_check_mark: |
 
## Reporting a Vulnerability
 
Security is of the highest importance and all security vulnerabilities should be reported to Harbor privately to minimize attacks against current users of Harbor before they are fixed.  Vulnerabilities will be investigated and patched on the next patch(minor) release ASAP and this information could be kept internal to the project entirely.  
 
To report a CVE, please email the private cncf-harbor-security@lists.cncf.io with the vulnerability details which will be fielded by the Harbor security team made up of Harbor maintainers who have committer and release permissions.  Emails will be addressed within 2 business days (according to Beijing time) including detailed plan to rectify the issue and workarounds in the meantime.  Bugs should not be reported through this channel and instead go through GitHub issues.
 
#Mailing lists
cncf-harbor-security@lists.cncf.io: for any security concerns. Received by Product Security Team members, and used by this Team to discuss security issues and fixes.
cncf-harbor-distributors-announce@lists.cncf.io: for early private information on Security patch releases. See below how Harbor distributors can apply for this list.


#When to report a vulnerability
 
-      When you think Harbor has a potential security vulnerability
-      When you suspect a potential vulnerability but are unsure if it impacts Harbor
-      When you know or suspect a potential vulnerability on another project being leveraged by Harbor.  For ex., Docker, PGSql, Redis, Notary etc.
 
 
#Vulnerability Report Process
 
Please do not file a public issue on GitHub for security vulnerabilities.  Instead please email cncf-harbor-security@lists.cncf.io and use a descriptive subject line and in the body of the email include information such as  
 
-      Basic information such as your name and affiliation / company
-      Detailed steps taken to reproduce the vulnerability  (POC scripts, screenshots, and compressed packet captures are all helpful to us)
-      Description of its effects on Harbor and how related hardware and software configurations so that Harbor security team can reproduce it
-      How the vulnerability impacts Harbor usage and an estimation of the attack surface if there is one
-      What other projects or dependencies that were used in conjunction to produce the vulnerability 
 
 
#Patch, Release, and Disclosure
 
Harbor Security team will respond to vulnerability reports as follows
 
-      Security team will Investigate the vulnerability and determine its effects and criticality.
-      If deemed not a vulnerability or vulnerability, security team will follow-up with a detailed reason for rejection.
-      If vulnerability is acknowledged and timeline for fix is determined, security team will work on communication plan to the appropriate community (completed within 1-7 days of vulnerability reported) including mitigating steps that affected users can take to protect themselves until fix is rolled out.
-      Security team to work on fixing vulnerability and perform internal testing before preparing to roll out release.
-      The security team will email the fix to cncf-harbor-distributors-announce@lists.cncf.io first to further test out security fix and gather feedback.  Please see details in the ‘Disclosure to Private Distributors List’ section detail on how to join mailing list.
-      Once fix is confirmed, security team will patch vulnerability on the upcoming minor as well for the next major release and backport it all earlier supported releases.  In special cases, they are cherry picked to older non-related releases for customers that do not have the ability to upgrade to a patched release.
-      Publish advisory on Harbor community (on Github, Slack, blog) and assist in rolling out patched release for affected users.  
 
#Disclosure to Private Distributors List

This list is intended to be used primarily to provide actionable information to multiple distributor projects at once. This list is not intended for individuals to find out about security issues.
#Membership Criteria
To be eligible for the cncf-harbor-distributors-announce@lists.cncf.io mailing list, your distribution should:
Be an active distributor of Harbor component.
Have a user base not limited to your own organization.
Have a publicly verifiable track record up to present day of fixing security issues.
Not be a downstream or rebuild of another distributor.
Be a participant and active contributor in the community.
Accept the Embargo Policy that is outlined below.
Have someone already on the list vouch for the person requesting membership on behalf of your distribution.
 
#Embargo Policy
The information members receive on cncf-harbor-distributors-announce@lists.cncf.io must not be made public, shared, nor even hinted at anywhere beyond the need-to-know within your specific team except with the list's explicit approval. This holds true until the public disclosure date/time that was agreed upon by the list. Members of the list and others may not use the information for anything other than getting the issue fixed for your respective distribution's users.
Before any information from the list is shared with respective members of your team required to fix said issue, they must agree to the same terms and only find out information on a need-to-know basis.
In the unfortunate event you share the information beyond what is allowed by this policy, you must urgently inform the cncf-harbor-security@lists.cncf.io mailing list of exactly what information leaked and to whom.
If you continue to leak information and break the policy outlined here, you will be removed from the list.
 
#Requesting to Join
New membership requests are sent to cncf-harbor-security@lists.cncf.io.
In the body of your request please specify how you qualify and fulfill each criterion listed in Membership Criteria section above.

