/*
 *
 * Copyright (C) 2012 - Marc Quinton.
 *
 * Use of this source code is governed by the MIT Licence :
 *  http://opensource.org/licenses/mit-license.php
 *
 * Permission is hereby granted, free of charge, to any person obtaining
 * a copy of this software and associated documentation files (the
 * "Software"), to deal in the Software without restriction, including
 * without limitation the rights to use, copy, modify, merge, publish,
 * distribute, sublicense, and/or sell copies of the Software, and to
 * permit persons to whom the Software is furnished to do so, subject to
 * the following conditions:
 * 
 * The above copyright notice and this permission notice shall be
 * included in all copies or substantial portions of the Software.
 * 
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 * EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 * MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
 * NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY
 * CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
 * TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package openldap

const (
	// first version for this GO API binding
	OPENLDAP_API_BINDING_VERSION = "0.2"
)

const (
	LDAP_VERSION1 = 1
	LDAP_VERSION2 = 2
	LDAP_VERSION3 = 3
)

const (
	LDAP_VERSION_MIN = LDAP_VERSION2
	LDAP_VERSION     = LDAP_VERSION2
	LDAP_VERSION_MAX = LDAP_VERSION3
)

const (
	LDAP_API_VERSION = 3001
	LDAP_VENDOR_NAME = "OpenLDAP"
)

const (
	LDAP_PORT  = 389
	LDAPS_PORT = 636
)

const (
	LDAP_OPT_SUCCESS = 0
	LDAP_OPT_ERROR   = -1
)

// search scopes
const (
	LDAP_SCOPE_BASE        = 0x0000
	LDAP_SCOPE_ONELEVEL    = 0x0001
	LDAP_SCOPE_SUBTREE     = 0x0002
	LDAP_SCOPE_SUBORDINATE = 0x0003 // OpenLDAP extension
	LDAP_SCOPE_DEFAULT     = -1     // OpenLDAP extension
)

const (
	LDAP_SCOPE_BASEOBJECT = LDAP_SCOPE_BASE
	LDAP_SCOPE_ONE        = LDAP_SCOPE_ONELEVEL
	LDAP_SCOPE_SUB        = LDAP_SCOPE_SUBTREE
	LDAP_SCOPE_CHILDREN   = LDAP_SCOPE_SUBORDINATE
)

const (
	LDAP_RES_ANY         = -1
	LDAP_RES_UNSOLICITED = 0
)

//const (
//LDAP_API_FEATURE_THREAD_SAFE           = 1
//LDAP_API_FEATURE_SESSION_THREAD_SAFE   = 1
//LDAP_API_FEATURE_OPERATION_THREAD_SAFE = 1
//)

const (
	LDAP_SUCCESS                   = 0x00
	LDAP_OPERATIONS_ERROR          = 0x01
	LDAP_PROTOCOL_ERROR            = 0x02
	LDAP_TIMELIMIT_EXCEEDED        = 0x03
	LDAP_SIZELIMIT_EXCEEDED        = 0x04
	LDAP_COMPARE_FALSE             = 0x05
	LDAP_COMPARE_TRUE              = 0x06
	LDAP_AUTH_METHOD_NOT_SUPPORTED = 0x07
	LDAP_STRONG_AUTH_REQUIRED      = 0x08
	// Not used in LDAPv3
	LDAP_PARTIAL_RESULTS = 0x09

	// Next 5 new in LDAPv3
	LDAP_REFERRAL                       = 0x0a
	LDAP_ADMINLIMIT_EXCEEDED            = 0x0b
	LDAP_UNAVAILABLE_CRITICAL_EXTENSION = 0x0c
	LDAP_CONFIDENTIALITY_REQUIRED       = 0x0d
	LDAP_SASL_BIND_INPROGRESS           = 0x0e

	LDAP_NO_SUCH_ATTRIBUTE      = 0x10
	LDAP_UNDEFINED_TYPE         = 0x11
	LDAP_INAPPROPRIATE_MATCHING = 0x12
	LDAP_CONSTRAINT_VIOLATION   = 0x13
	LDAP_TYPE_OR_VALUE_EXISTS   = 0x14
	LDAP_INVALID_SYNTAX         = 0x15

	LDAP_NO_SUCH_OBJECT    = 0x20 /* 32 */
	LDAP_ALIAS_PROBLEM     = 0x21
	LDAP_INVALID_DN_SYNTAX = 0x22
	// Next two not used in LDAPv3
	LDAP_IS_LEAF             = 0x23
	LDAP_ALIAS_DEREF_PROBLEM = 0x24

	LDAP_INAPPROPRIATE_AUTH   = 0x30 /* 48 */
	LDAP_INVALID_CREDENTIALS  = 0x31 /* 49 */
	LDAP_INSUFFICIENT_ACCESS  = 0x32
	LDAP_BUSY                 = 0x33
	LDAP_UNAVAILABLE          = 0x34
	LDAP_UNWILLING_TO_PERFORM = 0x35
	LDAP_LOOP_DETECT          = 0x36

	LDAP_SORT_CONTROL_MISSING = 0x3C /* 60 */
	LDAP_INDEX_RANGE_ERROR    = 0x3D /* 61 */

	LDAP_NAMING_VIOLATION       = 0x40
	LDAP_OBJECT_CLASS_VIOLATION = 0x41
	LDAP_NOT_ALLOWED_ON_NONLEAF = 0x42
	LDAP_NOT_ALLOWED_ON_RDN     = 0x43
	LDAP_ALREADY_EXISTS         = 0x44 /* 68 */
	LDAP_NO_OBJECT_CLASS_MODS   = 0x45
	LDAP_RESULTS_TOO_LARGE      = 0x46
	// Next two for LDAPv3
	LDAP_AFFECTS_MULTIPLE_DSAS = 0x47
	LDAP_OTHER                 = 0x50

	// Used by some APIs
	LDAP_SERVER_DOWN    = 0x51
	LDAP_LOCAL_ERROR    = 0x52
	LDAP_ENCODING_ERROR = 0x53
	LDAP_DECODING_ERROR = 0x54
	LDAP_TIMEOUT        = 0x55
	LDAP_AUTH_UNKNOWN   = 0x56
	LDAP_FILTER_ERROR   = 0x57 /* 87 */
	LDAP_USER_CANCELLED = 0x58
	LDAP_PARAM_ERROR    = 0x59
	LDAP_NO_MEMORY      = 0x5a

	// Preliminary LDAPv3 codes
	LDAP_CONNECT_ERROR           = 0x5b
	LDAP_NOT_SUPPORTED           = 0x5c
	LDAP_CONTROL_NOT_FOUND       = 0x5d
	LDAP_NO_RESULTS_RETURNED     = 0x5e
	LDAP_MORE_RESULTS_TO_RETURN  = 0x5f
	LDAP_CLIENT_LOOP             = 0x60
	LDAP_REFERRAL_LIMIT_EXCEEDED = 0x61
)

const (
	LDAP_DEREF_NEVER     = 0
	LDAP_DEREF_SEARCHING = 1
	LDAP_DEREF_FINDING   = 2
	LDAP_DEREF_ALWAYS    = 3
)

const (
	LDAP_NO_LIMIT = 0
)

const (
	LDAP_MSG_ONE      = 0
	LDAP_MSG_ALL      = 1
	LDAP_MSG_RECEIVED = 2
)

// LDAP_OPTions
// 0x0000 - 0x0fff reserved for api options
// 0x1000 - 0x3fff reserved for api extended options
// 0x4000 - 0x7fff reserved for private and experimental options

const (
	LDAP_OPT_API_INFO  = 0x0000
	LDAP_OPT_DESC      = 0x0001 // historic
	LDAP_OPT_DEREF     = 0x0002
	LDAP_OPT_SIZELIMIT = 0x0003
	LDAP_OPT_TIMELIMIT = 0x0004
	// 0x05 - 0x07 not defined

	LDAP_OPT_REFERRALS = 0x0008
	LDAP_OPT_RESTART   = 0x0009
	// 0x0a - 0x10 not defined

	LDAP_OPT_PROTOCOL_VERSION = 0x0011
	LDAP_OPT_SERVER_CONTROLS  = 0x0012
	LDAP_OPT_CLIENT_CONTROLS  = 0x0013
	// 0x14 not defined

	LDAP_OPT_API_FEATURE_INFO = 0x0015
	// 0x16 - 0x2f not defined

	LDAP_OPT_HOST_NAME          = 0x0030
	LDAP_OPT_RESULT_CODE        = 0x0031
	LDAP_OPT_ERROR_NUMBER       = LDAP_OPT_RESULT_CODE
	LDAP_OPT_DIAGNOSTIC_MESSAGE = 0x0032
	LDAP_OPT_ERROR_STRING       = LDAP_OPT_DIAGNOSTIC_MESSAGE
	LDAP_OPT_MATCHED_DN         = 0x0033
	// 0x0034 - 0x3fff not defined

	// 0x0091 used by Microsoft for LDAP_OPT_AUTO_RECONNECT

	LDAP_OPT_SSPI_FLAGS = 0x0092
	// 0x0093 used by Microsoft for LDAP_OPT_SSL_INFO

	// 0x0094 used by Microsoft for LDAP_OPT_REF_DEREF_CONN_PER_MSG

	LDAP_OPT_SIGN        = 0x0095
	LDAP_OPT_ENCRYPT     = 0x0096
	LDAP_OPT_SASL_METHOD = 0x0097
	// 0x0098 used by Microsoft for LDAP_OPT_AREC_EXCLUSIVE

	LDAP_OPT_SECURITY_CONTEXT = 0x0099

// 0x009A used by Microsoft for LDAP_OPT_ROOTDSE_CACHE

// 0x009B - 0x3fff not defined

)

// API Extensions

const LDAP_OPT_API_EXTENSION_BASE = 0x4000 // API extensions

// private and experimental options

// OpenLDAP specific options

const (
	LDAP_OPT_DEBUG_LEVEL     = 0x5001 // debug level
	LDAP_OPT_TIMEOUT         = 0x5002 // default timeout
	LDAP_OPT_REFHOPLIMIT     = 0x5003 // ref hop limit
	LDAP_OPT_NETWORK_TIMEOUT = 0x5005 // socket level timeout
	LDAP_OPT_URI             = 0x5006
	LDAP_OPT_REFERRAL_URLS   = 0x5007 // Referral URLs
	LDAP_OPT_SOCKBUF         = 0x5008 // sockbuf
	LDAP_OPT_DEFBASE         = 0x5009 // searchbase
	LDAP_OPT_CONNECT_ASYNC   = 0x5010 // create connections asynchronously
	LDAP_OPT_CONNECT_CB      = 0x5011 // connection callbacks
	LDAP_OPT_SESSION_REFCNT  = 0x5012 // session reference count
)

// OpenLDAP TLS options

const (
	LDAP_OPT_X_TLS              = 0x6000
	LDAP_OPT_X_TLS_CTX          = 0x6001 // OpenSSL CTX*
	LDAP_OPT_X_TLS_CACERTFILE   = 0x6002
	LDAP_OPT_X_TLS_CACERTDIR    = 0x6003
	LDAP_OPT_X_TLS_CERTFILE     = 0x6004
	LDAP_OPT_X_TLS_KEYFILE      = 0x6005
	LDAP_OPT_X_TLS_REQUIRE_CERT = 0x6006
	LDAP_OPT_X_TLS_PROTOCOL_MIN = 0x6007
	LDAP_OPT_X_TLS_CIPHER_SUITE = 0x6008
	LDAP_OPT_X_TLS_RANDOM_FILE  = 0x6009
	LDAP_OPT_X_TLS_SSL_CTX      = 0x600a // OpenSSL SSL*
	LDAP_OPT_X_TLS_CRLCHECK     = 0x600b
	LDAP_OPT_X_TLS_CONNECT_CB   = 0x600c
	LDAP_OPT_X_TLS_CONNECT_ARG  = 0x600d
	LDAP_OPT_X_TLS_DHFILE       = 0x600e
	LDAP_OPT_X_TLS_NEWCTX       = 0x600f
	LDAP_OPT_X_TLS_CRLFILE      = 0x6010 // GNUtls only
	LDAP_OPT_X_TLS_PACKAGE      = 0x6011
)

const (
	LDAP_OPT_X_TLS_NEVER  = 0
	LDAP_OPT_X_TLS_HARD   = 1
	LDAP_OPT_X_TLS_DEMAND = 2
	LDAP_OPT_X_TLS_ALLOW  = 3
	LDAP_OPT_X_TLS_TRY    = 4
)

const (
	LDAP_OPT_X_TLS_CRL_NONE = 0
	LDAP_OPT_X_TLS_CRL_PEER = 1
	LDAP_OPT_X_TLS_CRL_ALL  = 2
)

// for LDAP_OPT_X_TLS_PROTOCOL_MIN

//!!! const (
//!!! LDAP_OPT_X_TLS_PROTOCOL(maj,min) = (((maj) << 8) + (min))
//!!! LDAP_OPT_X_TLS_PROTOCOL_SSL2 = (2 << 8)
//!!! LDAP_OPT_X_TLS_PROTOCOL_SSL3 = (3 << 8)
//!!! LDAP_OPT_X_TLS_PROTOCOL_TLS1_0 = ((3 << 8) + 1)
//!!! LDAP_OPT_X_TLS_PROTOCOL_TLS1_1 = ((3 << 8) + 2)
//!!! LDAP_OPT_X_TLS_PROTOCOL_TLS1_2 = ((3 << 8) + 3)
//!!! )

// OpenLDAP SASL options

const (
	LDAP_OPT_X_SASL_MECH         = 0x6100
	LDAP_OPT_X_SASL_REALM        = 0x6101
	LDAP_OPT_X_SASL_AUTHCID      = 0x6102
	LDAP_OPT_X_SASL_AUTHZID      = 0x6103
	LDAP_OPT_X_SASL_SSF          = 0x6104 // read-only
	LDAP_OPT_X_SASL_SSF_EXTERNAL = 0x6105 // write-only
	LDAP_OPT_X_SASL_SECPROPS     = 0x6106 // write-only
	LDAP_OPT_X_SASL_SSF_MIN      = 0x6107
	LDAP_OPT_X_SASL_SSF_MAX      = 0x6108
	LDAP_OPT_X_SASL_MAXBUFSIZE   = 0x6109
	LDAP_OPT_X_SASL_MECHLIST     = 0x610a // read-only
	LDAP_OPT_X_SASL_NOCANON      = 0x610b
	LDAP_OPT_X_SASL_USERNAME     = 0x610c // read-only
	LDAP_OPT_X_SASL_GSS_CREDS    = 0x610d
)

// OpenLDAP GSSAPI options

const (
	LDAP_OPT_X_GSSAPI_DO_NOT_FREE_CONTEXT    = 0x6200
	LDAP_OPT_X_GSSAPI_ALLOW_REMOTE_PRINCIPAL = 0x6201
)

// 
// OpenLDAP per connection tcp-keepalive settings
// (Linux only, ignored where unsupported)
const (
	LDAP_OPT_X_KEEPALIVE_IDLE     = 0x6300
	LDAP_OPT_X_KEEPALIVE_PROBES   = 0x6301
	LDAP_OPT_X_KEEPALIVE_INTERVAL = 0x6302
)

/* authentication methods available */
const (
	LDAP_AUTH_NONE   = 0x00 // no authentication
	LDAP_AUTH_SIMPLE = 0x80 // context specific + primitive
	LDAP_AUTH_SASL   = 0xa3 // context specific + constructed
	LDAP_AUTH_KRBV4  = 0xff // means do both of the following
	LDAP_AUTH_KRBV41 = 0x81 // context specific + primitive
	LDAP_AUTH_KRBV42 = 0x82 // context specific + primitive
)
