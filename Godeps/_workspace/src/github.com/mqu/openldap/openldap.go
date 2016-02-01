/*
 * Openldap (2.4.30) binding in GO 
 * 
 * 
 *  link to ldap or ldap_r (for thread-safe binding)
 *
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
 * 
 */

package openldap

/*

#define LDAP_DEPRECATED 1
#include <stdlib.h>
#include <ldap.h>

static inline char* to_charptr(const void* s) { return (char*)s; }
static inline LDAPControl** to_ldapctrlptr(const void* s) {
	return (LDAPControl**) s;
}
*/
// #cgo CFLAGS: -DLDAP_DEPRECATED=1
// #cgo linux CFLAGS: -DLINUX=1
// #cgo LDFLAGS: -lldap -llber
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
	"strings"
	"strconv"
)

/* Intialize() open an LDAP connexion ; supported url formats :
 * 
 *   ldap://host:389/
 *   ldaps://secure-host:636/
 * 
 * return values :
 *  - on success : LDAP object, nil
 *  - on error : nil and error with error description.
 */
func Initialize(url string) (*Ldap, error) {
	_url := C.CString(url)
	defer C.free(unsafe.Pointer(_url))

	var ldap *C.LDAP

	// API: int ldap_initialize (LDAP **ldp, LDAP_CONST char *url )
	rv := C.ldap_initialize(&ldap, _url)

	if rv != 0 {
		err := errors.New(fmt.Sprintf("LDAP::Initialize() error (%d) : %s", rv, ErrorToString(int(rv))))
		return nil, err
	}

	return &Ldap{ldap}, nil
}

/*
 * StartTLS() is used for regular LDAP (not
 * LDAPS) connections to establish encryption
 * after the session is running.
 *
 * return value :
 *  - nil on success,
 *  - error with error description on error.
 */
func (self *Ldap) StartTLS() error {
	var rv int

	// API: int ldap_start_tls_s(LDAP *ld, LDAPControl **serverctrls, LDAPControl **clientctrls);
	rv = int(C.ldap_start_tls_s(self.conn,
		C.to_ldapctrlptr(unsafe.Pointer(nil)),
		C.to_ldapctrlptr(unsafe.Pointer(nil))))

	if rv == LDAP_OPT_SUCCESS {
		return nil
	}

	return errors.New(fmt.Sprintf("LDAP::StartTLS() error (%d) : %s", rv,
		ErrorToString(rv)))
}

/* 
 * Bind() is used for LDAP authentifications
 * 
 * if who is empty this is an anonymous bind
 * else this is an authentificated bind
 * 
 * return value : 
 *  - nil on succes,
 *  - error with error description on error.
 *
 */
func (self *Ldap) Bind(who, cred string) error {
	var rv int

	authmethod := C.int(LDAP_AUTH_SIMPLE)

	// DEPRECATED
	// API: int ldap_bind_s (LDAP *ld,	LDAP_CONST char *who, LDAP_CONST char *cred, int authmethod );
	if who == "" {
		_who := C.to_charptr(unsafe.Pointer(nil))
		_cred := C.to_charptr(unsafe.Pointer(nil))

		rv = int(C.ldap_bind_s(self.conn, _who, _cred, authmethod))
	} else {
		_who := C.CString(who)
		_cred := C.CString(cred)
		defer C.free(unsafe.Pointer(_who))
		rv = int(C.ldap_bind_s(self.conn, _who, _cred, authmethod))
	}

	if rv == LDAP_OPT_SUCCESS {
		return nil
	}

	self.conn = nil
	return errors.New(fmt.Sprintf("LDAP::Bind() error (%d) : %s", rv, ErrorToString(rv)))
}

/* 
 * close LDAP connexion
 * 
 * return value : 
 *  - nil on succes,
 *  - error with error description on error.
 *
 */
func (self *Ldap) Close() error {

	// DEPRECATED
	// API: int ldap_unbind(LDAP *ld)
	rv := C.ldap_unbind(self.conn)

	if rv == LDAP_OPT_SUCCESS {
		return nil
	}

	self.conn = nil
	return errors.New(fmt.Sprintf("LDAP::Close() error (%d) : %s", int(rv), ErrorToString(int(rv))))

}
/* 
 * Unbind() close LDAP connexion
 * 
 * an alias to Ldap::Close()
 *
 */
func (self *Ldap) Unbind() error {
	return self.Close()
}

/* Search() is used to search LDAP server
  - base is where search is starting
  - scope allows local or deep search. Supported values :
     - LDAP_SCOPE_BASE
	   - LDAP_SCOPE_ONELEVEL
     - LDAP_SCOPE_SUBTREE
  - filter is an LDAP search expression,
  - attributes is an array of string telling with LDAP attribute to get from this request
*/
func (self *Ldap) Search(base string, scope int, filter string, attributes []string) (*LdapMessage, error) {

	var attrsonly int = 0 // false: returns all, true, returns only attributes without values

	_base := C.CString(base)
	defer C.free(unsafe.Pointer(_base))

	_filter := C.CString(filter)
	defer C.free(unsafe.Pointer(_filter))

	// transform []string to C.char** null terminated array (attributes argument)
	_attributes := make([]*C.char, len(attributes)+1) // default set to nil (NULL in C)

	for i, arg := range attributes {
		_attributes[i] = C.CString(arg)
		defer C.free(unsafe.Pointer(_attributes[i]))
	}

	var msg *C.LDAPMessage

	// DEPRECATED
	// API: int ldap_search_s (LDAP *ld, char *base, int scope, char *filter, char **attrs, int attrsonly, LdapMessage * ldap_res)
	rv := int(C.ldap_search_s(self.conn, _base, C.int(scope), _filter, &_attributes[0], C.int(attrsonly), &msg))

	if rv == LDAP_OPT_SUCCESS {
		_msg := new(LdapMessage)
		_msg.ldap = self
		_msg.errno = rv
		_msg.msg = msg
		return _msg, nil
	}

	return nil, errors.New(fmt.Sprintf("LDAP::Search() error : %d (%s)", rv, ErrorToString(rv)))
}

// ------------------------------------- Ldap* method (object oriented) -------------------------------------------------------------------

// Create a new LdapAttribute entry with name and values.
func LdapAttributeNew(name string, values []string)(*LdapAttribute){
	a := new(LdapAttribute)
	a.values = values
	a.name = name
	return a
}

// Append() adds an LdapAttribute to self LdapEntry
func (self *LdapEntry) Append(a LdapAttribute){
	self.values = append(self.values, a)
}

// String() is used for fmt.Println(self)
//
func (self *LdapAttribute) String() string{
	return self.ToText()
}

// ToText() returns a text string representation of LdapAttribute
// avoiding displaying binary data.
//
func (self *LdapAttribute) ToText() string{
	
	var list []string
	
	for _, a := range self.Values() {
		if (!_isPrint(a)) {
			list = append(list, fmt.Sprintf("binary-data[%d]", len(a)))
		} else {
			list = append(list, a)
		}
	}
	if len(list) > 1 {
		return fmt.Sprintf("%s: (%d)[%s]", self.name, len(list), strings.Join(list, ", "))
	}
	return fmt.Sprintf("%s: [%s]", self.name, strings.Join(list, ", "))
}

// Name() return attribute name
func (self *LdapAttribute) Name() string{
	return self.name
}

// Values() returns array values for self LdapAttribute 
//
func (self *LdapAttribute) Values() []string{
	return self.values
}

// _isPrint() returns true if str is printable
//
// @private method
func _isPrint(str string) bool{
	for _, c := range str{
		
		if !strconv.IsPrint(rune(c)) {
			return false
		}
	}
	
	return true
}

// IsPrint() returns true is self LdapAttribute is printable.
func (self *LdapAttribute) IsPrint() bool{
	for _, a := range self.Values() {
		if (!_isPrint(a)) {
			return false
		}
	}
	return true
}

// Dn() returns DN (Distinguish Name) for self LdapEntry
func (self *LdapEntry) Dn() string{
	return self.dn
}

// Attributes() returns an array of LdapAttribute
func (self *LdapEntry) Attributes() []LdapAttribute{
	return self.values
}

// Print() allow printing self LdapEntry with fmt.Println()
func (self *LdapEntry) String() string {
	return self.ToText()
}

// GetValuesByName() get a list of values for self LdapEntry, using "name" attribute
func (self *LdapEntry) GetValuesByName(attrib string) []string{
	
	for _, a := range self.values{
		if a.Name() == attrib {
			return a.values
		}
	}
	
	return []string{}
}
// GetOneValueByName() ; a quick way to get a single attribute value
func (self *LdapEntry) GetOneValueByName(attrib string) (string, error){
	
	for _, a := range self.values{
		if a.Name() == attrib {
			return a.values[0], nil
		}
	}
	
	return "", errors.New(fmt.Sprintf("LdapEntry::GetOneValueByName() error : attribute %s not found", attrib))
}

// ToText() return a string representating self LdapEntry
func (self *LdapEntry) ToText() string{

	txt := fmt.Sprintf("dn: %s\n", self.dn)
	
	for _, a := range self.values{
		txt = txt + fmt.Sprintf("%s\n", a.ToText())
	}

	return txt
}

// Append() add e to LdapSearchResult array
func (self *LdapSearchResult) Append(e LdapEntry){
	self.entries = append(self.entries, e)
}

// ToText() : a quick way to print an LdapSearchResult
func (self *LdapSearchResult) ToText() string{

	txt := fmt.Sprintf("# query : %s\n", self.filter)
	txt = txt + fmt.Sprintf("# num results : %d\n", self.Count())
	txt = txt + fmt.Sprintf("# search : %s\n", self.Filter())
	txt = txt + fmt.Sprintf("# base : %s\n", self.Base())
	txt = txt + fmt.Sprintf("# attributes : [%s]\n", strings.Join(self.Attributes(), ", "))

	for _, e := range self.entries{
		txt = txt + fmt.Sprintf("%s\n", e.ToText())
	}

	return txt
}

// String() : used for fmt.Println(self)
func (self *LdapSearchResult) String() string{
	return self.ToText()
}

// Entries() : returns an array of LdapEntry for self
func (self *LdapSearchResult) Entries() []LdapEntry{
	return self.entries
}

// Count() : returns number of results for self search.
func (self *LdapSearchResult) Count() int{
	return len(self.entries)
}

// Filter() : returns filter for self search
func (self *LdapSearchResult) Filter() string{
	return self.filter
}

// Filter() : returns base DN for self search
func (self *LdapSearchResult) Base() string{
	return self.base
}

// Filter() : returns scope for self search
func (self *LdapSearchResult) Scope() int{
	return self.scope
}

// Filter() : returns an array of attributes used for this actual search
func (self *LdapSearchResult) Attributes() []string{
	return self.attributes
}

// SearchAll() : a quick way to make search. This method returns an LdapSearchResult with all necessary methods to
// access data. Result is a collection (tree) of []LdapEntry / []LdapAttribute.
//
func (self *Ldap) SearchAll(base string, scope int, filter string, attributes []string) (*LdapSearchResult, error) {

	sr := new(LdapSearchResult)

	sr.ldap   = self
	sr.base   = base
	sr.scope  = scope
	sr.filter = filter
	sr.attributes = attributes

	// Search(base string, scope int, filter string, attributes []string) (*LDAPMessage, error)	
	result, err := self.Search(base, scope, filter, attributes)

	if err != nil {
		fmt.Println(err)
		return sr, err
	}

	// Free LDAP::Result() allocated data
	defer result.MsgFree()

	e := result.FirstEntry()

	for e != nil {
		_e := new(LdapEntry)
		
		_e.dn = e.GetDn()

		attr, _ := e.FirstAttribute()
		for attr != "" {

			_attr := LdapAttributeNew(attr, e.GetValues(attr))
			_e.Append(*_attr)

			attr, _ = e.NextAttribute()

		}

		sr.Append(*_e)

		e = e.NextEntry()
	}
	
	return sr, nil
}
