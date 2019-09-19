 
## Supported Versions
 
| Version | Supported          |
| ------- | ------------------ |
| Harbor v1.7.x   | :white_check_mark: |
| Harbor v1.8.x   | :white_check_mark: |
| Harbor v1.9.x   | :white_check_mark: |
 
## Reporting a Vulnerability
 
Security is of the highest importance and all security vulnerabilities should be reported to Harbor privately, to minimize attacks against current users of Harbor before they are fixed. Vulnerabilities will be investigated and patched on the next patch (minor) release as soon as possible. This information could be kept entirely internal to the project.  
 
To report a CVE, please email the private address cncf-harbor-security@lists.cncf.io with the details of the vulnerability. The email will be fielded by the Harbor Security Team, which is made up of Harbor maintainers who have committer and release permissions. Emails will be addressed within 2 business days (according to Beijing time), including a detailed plan to rectify the issue and workarounds to perform in the meantime. Do not report bugs through this channel. Use GitHub issues instead.
 
#Mailing lists
cncf-harbor-security@lists.cncf.io: for any security concerns. Received by Product Security Team members, and used by this team to discuss security issues and fixes.
cncf-harbor-distributors-announce@lists.cncf.io: for early private information on security patch releases. See below for information about how Harbor distributors can apply to join this list.


#When to report a vulnerability
 
* When you think Harbor has a potential security vulnerability
* When you suspect a potential vulnerability but you are unsure that it impacts Harbor
* When you know of or suspect a potential vulnerability on another project that is used by Harbor.  For example, Docker, PGSql, Redis, Notary etc.
 
 
#Vulnerability Report Process
 
IMPORTANT: Do not file public issues on GitHub for security vulnerabilities. 

To report vulnerabilities, please send an email to cncf-harbor-security@lists.cncf.io. Provide a descriptive subject line and in the body of the email include the following information:
 
* Basic identity information, such as your name and your affiliation or company.
* Detailed steps to reproduce the vulnerability  (POC scripts, screenshots, and compressed packet captures are all helpful to us).
* Description of the effects of the vulnerability on Harbor and the related hardware and software configurations, so that the Harbor Security Team can reproduce it.
* How the vulnerability affects Harbor usage and an estimation of the attack surface, if there is one.
* List other projects or dependencies that were used in conjunction with Harbor to produce the vulnerability..
 
 
#Patch, Release, and Disclosure
 
The Harbor Security Team will respond to vulnerability reports as follows:
 
1.  The Security Team will investigate the vulnerability and determine its effects and criticality.
2.  If the issue is not deemed to be a vulnerability, the Security Team will follow up with a detailed reason for rejection.
3.  If a vulnerability is acknowledged and the timeline for a fix is determined, the Security Team will work on a plan to communicate with the appropriate community (to be completed within 1-7 days of the report of the vulnerability), including mitigating steps that affected users can take to protect themselves until the fix is rolled out.
4.  The Security Team will also create a CVSS (https://www.first.org/cvss/specification-document)  using the CVSS Calculator (https://www.first.org/cvss/calculator/3.0) . The Security Team makes the final call on the calculated CVSS; it is better to move quickly than making the CVSS perfect.
5.  The Security Team works on fixing the vulnerability and performs internal testing before preparing to roll out the fix.
6.  The Security Team will first email the fix to cncf-harbor-distributors-announce@lists.cncf.io, so that they can further test the fix and gather feedback. See the section ‘Disclosure to Private Distributors List’ for details about how to join this mailing list.
7.  Once the fix is confirmed, the Security Team will patch the vulnerability in the upcoming minor release and the next major release, and backport it into all earlier supported releases. In special cases, the fixes are cherry-picked into older unsupported releases for customers who cannot upgrade to a patched release.
8.  The Security Team publishes an advisory to the Harbor community via Github, Slack, blog, and assists in rolling out the patched release to affected users.  
 
#Disclosure to Private Distributors List

This list is intended to be used primarily to provide actionable information to multiple distributor projects at once. This list is not intended to inform individuals about security issues.

#Membership Criteria
To be eligible to join the cncf-harbor-distributors-announce@lists.cncf.io mailing list, your distribution should:
1. Be an active distributor of the Harbor component.
2. Have a user base that is not limited to your own organization.
3. Have a publicly verifiable track record up to the present day of fixing security issues.
4. Not be a downstream or rebuild of another distributor.
5. Be a participant and active contributor in the Harbor community.
6. Accept the Embargo Policy that is outlined below.
7. Have someone who is already on the list vouch for the person requesting membership on behalf of your distribution.
 
#Embargo Policy
The information that members receive on cncf-harbor-distributors-announce@lists.cncf.io must not be made public, shared, or even hinted at anywhere beyond those who need to know within your specific team, unless you receive explicit approval to do so from the list. This remains true until the public disclosure date/time agreed upon by the list. Members of the list and others cannot use the information for any reason other than to get the issue fixed for your respective distribution's users.
Before you share any information from the list with members of your team who are required to fix the issue, these team members must agree to the same terms, and only be provided with information on a need-to-know basis.
In the unfortunate event that you share information beyond what is permitted by this policy, you must urgently inform the cncf-harbor-security@lists.cncf.io mailing list of exactly what information was leaked and to whom.
If you continue to leak information and break the policy outlined here, you will be removed from the list.
 
#Requesting to Join
Send new membership requests to cncf-harbor-security@lists.cncf.io.
In the body of your request please specify how you qualify for membership and fulfill each criterion listed in the Membership Criteria section above.


