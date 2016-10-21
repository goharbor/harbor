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

/*
# include <ldap.h>

*/
// #cgo CFLAGS: -DLDAP_DEPRECATED=1
// #cgo linux CFLAGS: -DLINUX=1
// #cgo LDFLAGS: -lldap -llber
import "C"

type Ldap struct {
	conn *C.LDAP
}

type LdapMessage struct {
	ldap *Ldap
	// conn *C.LDAP
	msg   *C.LDAPMessage
	errno int
}

type LdapAttribute struct{
	name string
	values []string
}


type LdapEntry struct {
	ldap *Ldap
	// conn  *C.LDAP
	entry *C.LDAPMessage
	errno int
	ber   *C.BerElement

	dn string
	values []LdapAttribute
}

type LdapSearchResult struct{
	ldap *Ldap

	scope int
	filter string
	base string
	attributes []string
	
	entries []LdapEntry
}
