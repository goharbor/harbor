package postprocessors

import (
	"testing"
	"time"

	"github.com/astaxie/beego"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanv2"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const sampleReport = `{
	"generated_at": "2020-08-01T18:28:49.072885592Z",
	"artifact": {
	  "repository": "library/ubuntu",
	  "digest": "sha256:d5b40885539615b9aeb7119516427959a158386af13e00d79a7da43ad1b3fb87",
	  "mime_type": "application/vnd.docker.distribution.manifest.v2+json"
	},
	"scanner": {
	  "name": "Trivy",
	  "vendor": "Aqua Security",
	  "version": "v0.9.1"
	},
	"severity": "Medium",
	"vulnerabilities": [
	  {
		"id": "CVE-2019-18276",
		"package": "bash",
		"version": "5.0-6ubuntu1.1",
		"severity": "Low",
		"description": "An issue was discovered in disable_priv_mode in shell.c in GNU Bash through 5.0 patch 11. By default, if Bash is run with its effective UID not equal to its real UID, it will drop privileges by setting its effective UID to its real UID. However, it does so incorrectly. On Linux and other systems that support \"saved UID\" functionality, the saved UID is not dropped. An attacker with command execution in the shell can use \"enable -f\" for runtime loading of a new builtin, which can be a shared object that calls setuid() and therefore regains privileges. However, binaries running with an effective UID of 0 are unaffected.",
		"links": [
		  "http://packetstormsecurity.com/files/155498/Bash-5.0-Patch-11-Privilege-Escalation.html",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2019-18276",
		  "https://github.com/bminor/bash/commit/951bdaad7a18cc0dc1036bba86b18b90874d39ff",
		  "https://security.netapp.com/advisory/ntap-20200430-0003/",
		  "https://www.youtube.com/watch?v=-wGtxJ8opa8"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2016-2781",
		"package": "coreutils",
		"version": "8.30-3ubuntu2",
		"severity": "Low",
		"description": "chroot in GNU coreutils, when used with --userspec, allows local users to escape to the parent session via a crafted TIOCSTI ioctl call, which pushes characters to the terminal's input buffer.",
		"links": [
		  "http://seclists.org/oss-sec/2016/q1/452",
		  "http://www.openwall.com/lists/oss-security/2016/02/28/2",
		  "http://www.openwall.com/lists/oss-security/2016/02/28/3",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2016-2781"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2017-8283",
		"package": "dpkg",
		"version": "1.19.7ubuntu3",
		"severity": "Low",
		"description": "dpkg-source in dpkg 1.3.0 through 1.18.23 is able to use a non-GNU patch program and does not offer a protection mechanism for blank-indented diff hunks, which allows remote attackers to conduct directory traversal attacks via a crafted Debian source package, as demonstrated by use of dpkg-source on NetBSD.",
		"links": [
		  "http://www.openwall.com/lists/oss-security/2017/04/20/2",
		  "http://www.securityfocus.com/bid/98064",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-8283"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2019-13050",
		"package": "gpgv",
		"version": "2.2.19-3ubuntu2",
		"severity": "Low",
		"description": "Interaction between the sks-keyserver code through 1.2.0 of the SKS keyserver network, and GnuPG through 2.2.16, makes it risky to have a GnuPG keyserver configuration line referring to a host on the SKS keyserver network. Retrieving data from this network may cause a persistent denial of service, because of a Certificate Spamming Attack.",
		"links": [
		  "http://lists.opensuse.org/opensuse-security-announce/2019-08/msg00039.html",
		  "https://access.redhat.com/articles/4264021",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2019-13050",
		  "https://gist.github.com/rjhansen/67ab921ffb4084c865b3618d6955275f",
		  "https://lists.fedoraproject.org/archives/list/package-announce@lists.fedoraproject.org/message/AUK2YRO6QIH64WP2LRA5D4LACTXQPPU4/",
		  "https://lists.fedoraproject.org/archives/list/package-announce@lists.fedoraproject.org/message/CP4ON34YEXEZDZOXXWV43KVGGO6WZLJ5/",
		  "https://lists.gnupg.org/pipermail/gnupg-announce/2019q3/000439.html",
		  "https://support.f5.com/csp/article/K08654551",
		  "https://support.f5.com/csp/article/K08654551?utm_source=f5support&amp;utm_medium=RSS",
		  "https://twitter.com/lambdafu/status/1147162583969009664"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2016-10228",
		"package": "libc-bin",
		"version": "2.31-0ubuntu9",
		"severity": "Low",
		"description": "The iconv program in the GNU C Library (aka glibc or libc6) 2.25 and earlier, when invoked with the -c option, enters an infinite loop when processing invalid multi-byte input sequences, leading to a denial of service.",
		"links": [
		  "http://openwall.com/lists/oss-security/2017/03/01/10",
		  "http://www.securityfocus.com/bid/96525",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2016-10228",
		  "https://sourceware.org/bugzilla/show_bug.cgi?id=19519"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2020-6096",
		"package": "libc-bin",
		"version": "2.31-0ubuntu9",
		"severity": "Low",
		"description": "An exploitable signed comparison vulnerability exists in the ARMv7 memcpy() implementation of GNU glibc 2.30.9000. Calling memcpy() (on ARMv7 targets that utilize the GNU glibc implementation) with a negative value for the 'num' parameter results in a signed comparison vulnerability. If an attacker underflows the 'num' parameter to memcpy(), this vulnerability could lead to undefined behavior such as writing to out-of-bounds memory and potentially remote code execution. Furthermore, this memcpy() implementation allows for program execution to continue in scenarios where a segmentation fault or crash should have occurred. The dangers occur in that subsequent execution and iterations of this code will be executed with this corrupted data.",
		"links": [
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2020-6096",
		  "https://lists.fedoraproject.org/archives/list/package-announce@lists.fedoraproject.org/message/URXOIA2LDUKHQXK4BE55BQBRI6ZZG3Y6/",
		  "https://sourceware.org/bugzilla/attachment.cgi?id=12334",
		  "https://sourceware.org/bugzilla/show_bug.cgi?id=25620",
		  "https://talosintelligence.com/vulnerability_reports/TALOS-2020-1019",
		  "https://www.talosintelligence.com/vulnerability_reports/TALOS-2020-1019"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2016-10228",
		"package": "libc6",
		"version": "2.31-0ubuntu9",
		"severity": "Low",
		"description": "The iconv program in the GNU C Library (aka glibc or libc6) 2.25 and earlier, when invoked with the -c option, enters an infinite loop when processing invalid multi-byte input sequences, leading to a denial of service.",
		"links": [
		  "http://openwall.com/lists/oss-security/2017/03/01/10",
		  "http://www.securityfocus.com/bid/96525",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2016-10228",
		  "https://sourceware.org/bugzilla/show_bug.cgi?id=19519"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2020-6096",
		"package": "libc6",
		"version": "2.31-0ubuntu9",
		"severity": "Low",
		"description": "An exploitable signed comparison vulnerability exists in the ARMv7 memcpy() implementation of GNU glibc 2.30.9000. Calling memcpy() (on ARMv7 targets that utilize the GNU glibc implementation) with a negative value for the 'num' parameter results in a signed comparison vulnerability. If an attacker underflows the 'num' parameter to memcpy(), this vulnerability could lead to undefined behavior such as writing to out-of-bounds memory and potentially remote code execution. Furthermore, this memcpy() implementation allows for program execution to continue in scenarios where a segmentation fault or crash should have occurred. The dangers occur in that subsequent execution and iterations of this code will be executed with this corrupted data.",
		"links": [
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2020-6096",
		  "https://lists.fedoraproject.org/archives/list/package-announce@lists.fedoraproject.org/message/URXOIA2LDUKHQXK4BE55BQBRI6ZZG3Y6/",
		  "https://sourceware.org/bugzilla/attachment.cgi?id=12334",
		  "https://sourceware.org/bugzilla/show_bug.cgi?id=25620",
		  "https://talosintelligence.com/vulnerability_reports/TALOS-2020-1019",
		  "https://www.talosintelligence.com/vulnerability_reports/TALOS-2020-1019"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2019-12904",
		"package": "libgcrypt20",
		"version": "1.8.5-5ubuntu1",
		"severity": "Low",
		"description": "In Libgcrypt 1.8.4, the C implementation of AES is vulnerable to a flush-and-reload side-channel attack because physical addresses are available to other processes. (The C implementation is used on platforms where an assembly-language implementation is unavailable.)",
		"links": [
		  "http://lists.opensuse.org/opensuse-security-announce/2019-07/msg00049.html",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2019-12904",
		  "https://dev.gnupg.org/T4541",
		  "https://github.com/gpg/libgcrypt/commit/a4c561aab1014c3630bc88faf6f5246fee16b020",
		  "https://github.com/gpg/libgcrypt/commit/daedbbb5541cd8ecda1459d3b843ea4d92788762",
		  "https://people.canonical.com/~ubuntu-security/cve/2019/CVE-2019-12904.html"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2017-11164",
		"package": "libpcre3",
		"version": "2:8.39-12build1",
		"severity": "Low",
		"description": "In PCRE 8.41, the OP_KETRMAX feature in the match function in pcre_exec.c allows stack exhaustion (uncontrolled recursion) when processing a crafted regular expression.",
		"links": [
		  "http://openwall.com/lists/oss-security/2017/07/11/3",
		  "http://www.securityfocus.com/bid/99575",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-11164"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2019-20838",
		"package": "libpcre3",
		"version": "2:8.39-12build1",
		"severity": "Low",
		"description": "libpcre in PCRE before 8.43 allows a subject buffer over-read in JIT when UTF is disabled, and \\X or \\R has more than one fixed quantifier, a related issue to CVE-2019-20454.",
		"links": [
		  "https://bugs.gentoo.org/717920",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2019-20838",
		  "https://www.pcre.org/original/changelog.txt"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2020-14155",
		"package": "libpcre3",
		"version": "2:8.39-12build1",
		"severity": "Low",
		"description": "libpcre in PCRE before 8.44 allows an integer overflow via a large number after a (?C substring.",
		"links": [
		  "https://about.gitlab.com/releases/2020/07/01/security-release-13-1-2-release/",
		  "https://bugs.gentoo.org/717920",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2020-14155",
		  "https://www.pcre.org/original/changelog.txt"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2018-20839",
		"package": "libsystemd0",
		"version": "245.4-4ubuntu3.1",
		"severity": "Medium",
		"description": "systemd 242 changes the VT1 mode upon a logout, which allows attackers to read cleartext passwords in certain circumstances, such as watching a shutdown, or using Ctrl-Alt-F1 and Ctrl-Alt-F2. This occurs because the KDGKBMODE (aka current keyboard mode) check is mishandled.",
		"links": [
		  "http://www.securityfocus.com/bid/108389",
		  "https://bugs.launchpad.net/ubuntu/+source/systemd/+bug/1803993",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2018-20839",
		  "https://github.com/systemd/systemd/commit/9725f1a10f80f5e0ae7d9b60547458622aeb322f",
		  "https://github.com/systemd/systemd/pull/12378",
		  "https://github.com/systemd/systemd/pull/13109",
		  "https://security.netapp.com/advisory/ntap-20190530-0002/"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2020-13776",
		"package": "libsystemd0",
		"version": "245.4-4ubuntu3.1",
		"severity": "Low",
		"description": "systemd through v245 mishandles numerical usernames such as ones composed of decimal digits or 0x followed by hex digits, as demonstrated by use of root privileges when privileges of the 0x0 user account were intended. NOTE: this issue exists because of an incomplete fix for CVE-2017-1000082.",
		"links": [
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2020-13776",
		  "https://github.com/systemd/systemd/issues/15985",
		  "https://lists.fedoraproject.org/archives/list/package-announce@lists.fedoraproject.org/message/IYGLFEKG45EYBJ7TPQMLWROWPTZBEU63/",
		  "https://security.netapp.com/advisory/ntap-20200611-0003/"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2018-1000654",
		"package": "libtasn1-6",
		"version": "4.16.0-2",
		"severity": "Low",
		"description": "GNU Libtasn1-4.13 libtasn1-4.13 version libtasn1-4.13, libtasn1-4.12 contains a DoS, specifically CPU usage will reach 100% when running asn1Paser against the POC due to an issue in _asn1_expand_object_id(p_tree), after a long time, the program will be killed. This attack appears to be exploitable via parsing a crafted file.",
		"links": [
		  "http://lists.opensuse.org/opensuse-security-announce/2019-06/msg00009.html",
		  "http://lists.opensuse.org/opensuse-security-announce/2019-06/msg00018.html",
		  "http://www.securityfocus.com/bid/105151",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2018-1000654",
		  "https://gitlab.com/gnutls/libtasn1/issues/4"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2018-20839",
		"package": "libudev1",
		"version": "245.4-4ubuntu3.1",
		"severity": "Medium",
		"description": "systemd 242 changes the VT1 mode upon a logout, which allows attackers to read cleartext passwords in certain circumstances, such as watching a shutdown, or using Ctrl-Alt-F1 and Ctrl-Alt-F2. This occurs because the KDGKBMODE (aka current keyboard mode) check is mishandled.",
		"links": [
		  "http://www.securityfocus.com/bid/108389",
		  "https://bugs.launchpad.net/ubuntu/+source/systemd/+bug/1803993",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2018-20839",
		  "https://github.com/systemd/systemd/commit/9725f1a10f80f5e0ae7d9b60547458622aeb322f",
		  "https://github.com/systemd/systemd/pull/12378",
		  "https://github.com/systemd/systemd/pull/13109",
		  "https://security.netapp.com/advisory/ntap-20190530-0002/"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2020-13776",
		"package": "libudev1",
		"version": "245.4-4ubuntu3.1",
		"severity": "Low",
		"description": "systemd through v245 mishandles numerical usernames such as ones composed of decimal digits or 0x followed by hex digits, as demonstrated by use of root privileges when privileges of the 0x0 user account were intended. NOTE: this issue exists because of an incomplete fix for CVE-2017-1000082.",
		"links": [
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2020-13776",
		  "https://github.com/systemd/systemd/issues/15985",
		  "https://lists.fedoraproject.org/archives/list/package-announce@lists.fedoraproject.org/message/IYGLFEKG45EYBJ7TPQMLWROWPTZBEU63/",
		  "https://security.netapp.com/advisory/ntap-20200611-0003/"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2013-4235",
		"package": "login",
		"version": "1:4.8.1-1ubuntu5.20.04",
		"severity": "Low",
		"description": "shadow: TOCTOU (time-of-check time-of-use) race condition when copying and removing directory trees",
		"links": [
		  "https://access.redhat.com/security/cve/cve-2013-4235",
		  "https://bugzilla.redhat.com/show_bug.cgi?id=CVE-2013-4235",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2013-4235",
		  "https://security-tracker.debian.org/tracker/CVE-2013-4235"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2018-7169",
		"package": "login",
		"version": "1:4.8.1-1ubuntu5.20.04",
		"severity": "Low",
		"description": "An issue was discovered in shadow 4.5. newgidmap (in shadow-utils) is setuid and allows an unprivileged user to be placed in a user namespace where setgroups(2) is permitted. This allows an attacker to remove themselves from a supplementary group, which may allow access to certain filesystem paths if the administrator has used \"group blacklisting\" (e.g., chmod g-rwx) to restrict access to paths. This flaw effectively reverts a security feature in the kernel (in particular, the /proc/self/setgroups knob) to prevent this sort of privilege escalation.",
		"links": [
		  "https://bugs.launchpad.net/ubuntu/+source/shadow/+bug/1729357",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2018-7169",
		  "https://github.com/shadow-maint/shadow/pull/97",
		  "https://security.gentoo.org/glsa/201805-09"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2013-4235",
		"package": "passwd",
		"version": "1:4.8.1-1ubuntu5.20.04",
		"severity": "Low",
		"description": "shadow: TOCTOU (time-of-check time-of-use) race condition when copying and removing directory trees",
		"links": [
		  "https://access.redhat.com/security/cve/cve-2013-4235",
		  "https://bugzilla.redhat.com/show_bug.cgi?id=CVE-2013-4235",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2013-4235",
		  "https://security-tracker.debian.org/tracker/CVE-2013-4235"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2018-7169",
		"package": "passwd",
		"version": "1:4.8.1-1ubuntu5.20.04",
		"severity": "Low",
		"description": "An issue was discovered in shadow 4.5. newgidmap (in shadow-utils) is setuid and allows an unprivileged user to be placed in a user namespace where setgroups(2) is permitted. This allows an attacker to remove themselves from a supplementary group, which may allow access to certain filesystem paths if the administrator has used \"group blacklisting\" (e.g., chmod g-rwx) to restrict access to paths. This flaw effectively reverts a security feature in the kernel (in particular, the /proc/self/setgroups knob) to prevent this sort of privilege escalation.",
		"links": [
		  "https://bugs.launchpad.net/ubuntu/+source/shadow/+bug/1729357",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2018-7169",
		  "https://github.com/shadow-maint/shadow/pull/97",
		  "https://security.gentoo.org/glsa/201805-09"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2020-10543",
		"package": "perl-base",
		"version": "5.30.0-9build1",
		"severity": "Low",
		"description": "Perl before 5.30.3 on 32-bit platforms allows a heap-based buffer overflow because nested regular expression quantifiers have an integer overflow.",
		"links": [
		  "http://lists.opensuse.org/opensuse-security-announce/2020-06/msg00044.html",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2020-10543",
		  "https://github.com/Perl/perl5/blob/blead/pod/perl5303delta.pod",
		  "https://github.com/Perl/perl5/compare/v5.30.2...v5.30.3",
		  "https://github.com/perl/perl5/commit/897d1f7fd515b828e4b198d8b8bef76c6faf03ed",
		  "https://lists.fedoraproject.org/archives/list/package-announce@lists.fedoraproject.org/message/IN3TTBO5KSGWE5IRIKDJ5JSQRH7ANNXE/",
		  "https://metacpan.org/pod/release/XSAWYERX/perl-5.28.3/pod/perldelta.pod",
		  "https://metacpan.org/pod/release/XSAWYERX/perl-5.30.3/pod/perldelta.pod",
		  "https://security.gentoo.org/glsa/202006-03",
		  "https://security.netapp.com/advisory/ntap-20200611-0001/"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2020-10878",
		"package": "perl-base",
		"version": "5.30.0-9build1",
		"severity": "Low",
		"description": "Perl before 5.30.3 has an integer overflow related to mishandling of a \"PL_regkind[OP(n)] == NOTHING\" situation. A crafted regular expression could lead to malformed bytecode with a possibility of instruction injection.",
		"links": [
		  "http://lists.opensuse.org/opensuse-security-announce/2020-06/msg00044.html",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2020-10878",
		  "https://github.com/Perl/perl5/blob/blead/pod/perl5303delta.pod",
		  "https://github.com/Perl/perl5/compare/v5.30.2...v5.30.3",
		  "https://github.com/perl/perl5/commit/0a320d753fe7fca03df259a4dfd8e641e51edaa8",
		  "https://github.com/perl/perl5/commit/3295b48defa0f8570114877b063fe546dd348b3c",
		  "https://lists.fedoraproject.org/archives/list/package-announce@lists.fedoraproject.org/message/IN3TTBO5KSGWE5IRIKDJ5JSQRH7ANNXE/",
		  "https://metacpan.org/pod/release/XSAWYERX/perl-5.28.3/pod/perldelta.pod",
		  "https://metacpan.org/pod/release/XSAWYERX/perl-5.30.3/pod/perldelta.pod",
		  "https://security.gentoo.org/glsa/202006-03",
		  "https://security.netapp.com/advisory/ntap-20200611-0001/"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2020-12723",
		"package": "perl-base",
		"version": "5.30.0-9build1",
		"severity": "Low",
		"description": "regcomp.c in Perl before 5.30.3 allows a buffer overflow via a crafted regular expression because of recursive S_study_chunk calls.",
		"links": [
		  "http://lists.opensuse.org/opensuse-security-announce/2020-06/msg00044.html",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2020-12723",
		  "https://github.com/Perl/perl5/blob/blead/pod/perl5303delta.pod",
		  "https://github.com/Perl/perl5/compare/v5.30.2...v5.30.3",
		  "https://github.com/Perl/perl5/issues/16947",
		  "https://github.com/Perl/perl5/issues/17743",
		  "https://github.com/perl/perl5/commit/66bbb51b93253a3f87d11c2695cfb7bdb782184a",
		  "https://lists.fedoraproject.org/archives/list/package-announce@lists.fedoraproject.org/message/IN3TTBO5KSGWE5IRIKDJ5JSQRH7ANNXE/",
		  "https://metacpan.org/pod/release/XSAWYERX/perl-5.28.3/pod/perldelta.pod",
		  "https://metacpan.org/pod/release/XSAWYERX/perl-5.30.3/pod/perldelta.pod",
		  "https://security.gentoo.org/glsa/202006-03",
		  "https://security.netapp.com/advisory/ntap-20200611-0001/"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  },
	  {
		"id": "CVE-2019-9923",
		"package": "tar",
		"version": "1.30+dfsg-7",
		"severity": "Low",
		"description": "pax_decode_header in sparse.c in GNU Tar before 1.32 had a NULL pointer dereference when parsing certain archives that have malformed extended headers.",
		"links": [
		  "http://git.savannah.gnu.org/cgit/tar.git/commit/?id=cb07844454d8cc9fb21f53ace75975f91185a120",
		  "http://lists.opensuse.org/opensuse-security-announce/2019-04/msg00077.html",
		  "http://savannah.gnu.org/bugs/?55369",
		  "https://bugs.launchpad.net/ubuntu/+source/tar/+bug/1810241",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2019-9923"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  }
	]
  }`

type TestReportConverterSuite struct {
	suite.Suite
	rc     ScanReportV1Converter
	rpUUID string
}

// SetupTest prepares env for test cases.
func (suite *TestReportConverterSuite) SetupTest() {

	suite.rpUUID = "reportUUID"
}

func TestReportConverterTests(t *testing.T) {
	suite.Run(t, &TestReportConverterSuite{})
}

func (suite *TestReportConverterSuite) SetupSuite() {
	/*
		os.Setenv("POSTGRESQL_HOST", "127.0.0.1")
		os.Setenv("POSTGRESQL_USR", "postgres")
		os.Setenv("POSTGRESQL_PORT", "5432")
		os.Setenv("POSTGRESQL_PWD", "example")
		os.Setenv("POSTGRESQL_DATABASE", "cvrs")
		os.Setenv("POSTGRES_MIGRATION_SCRIPTS_PATH", "/Users/prahaladd/Projects/harborsrc/src/github.com/goharbor/harbor/make/migrations/postgresql")
	*/

	rp := &scan.Report{
		Digest:           "d1000",
		RegistrationUUID: "ruuid",
		MimeType:         v1.MimeTypeNativeReport,
		TrackID:          "tid001",
		Requester:        "requester",
		Report:           sampleReport,
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(1000),
		UUID:             "reportUUID",
	}

	suite.rc = NewScanReportV1ToV2Converter()
	dao.PrepareTestForPostgresSQL()
	suite.create(rp)
	beego.SetLevel(beego.LevelDebug)
}

// TearDownTest clears test env for test cases.
func (suite *TestReportConverterSuite) TearDownTest() {
	// No delete method defined in manager as no requirement,
	// so, to clear env, call dao method here
	err := scan.DeleteReport(suite.rpUUID)
	require.NoError(suite.T(), err)
	delCount, err := scanv2.DeleteAllVulnerabilityRecordsForReport(suite.rpUUID)
	require.True(suite.T(), delCount > 0, "Failed to delete vulnerability records")
}
func (suite *TestReportConverterSuite) TestConvertReport() {
	rp := &scan.Report{
		Digest:           "d1000",
		RegistrationUUID: "ruuid",
		MimeType:         v1.MimeTypeNativeReport,
		TrackID:          "tid001",
		Requester:        "requester",
		Report:           sampleReport,
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(1000),
		UUID:             "reportUUID",
	}
	ruuid, err := suite.rc.Convert(rp)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), rp.UUID, ruuid)

}

func (suite *TestReportConverterSuite) create(r *scan.Report) {
	id, err := scan.CreateReport(r)
	require.NoError(suite.T(), err)
	require.Condition(suite.T(), func() (success bool) {
		success = id > 0
		return
	})
}
