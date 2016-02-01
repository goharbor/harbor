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

/*#include <stdio.h>
#include <stdlib.h>
#include <ctype.h>
#include <ldap.h>

int _berval_get_len(struct berval **ber, int i){
	return ber[i]->bv_len;
}

char* _berval_get_value(struct berval **ber, int i){
	return ber[i]->bv_val;
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
)

// ------------------------------------------ RESULTS methods ---------------------------------------------
/*

	openldap C API : 

    int ldap_count_messages( LDAP *ld, LdapMessage *result )
    LdapMessage *ldap_first_message( LDAP *ld, LdapMessage *result )
    LdapMessage *ldap_next_message ( LDAP *ld, LdapMessage *message )

	int ldap_count_entries( LDAP *ld, LdapMessage *result )
	LdapMessage *ldap_first_entry( LDAP *ld, LdapMessage *result )
	LdapMessage *ldap_next_entry ( LDAP *ld, LdapMessage *entry )

    char *ldap_first_attribute(LDAP *ld, LdapMessage *entry, BerElement **berptr )
    char *ldap_next_attribute (LDAP *ld, LdapMessage *entry, BerElement *ber )

    char **ldap_get_values(LDAP *ld, LdapMessage *entry, char *attr)
    struct berval **ldap_get_values_len(LDAP *ld, LdapMessage *entry,char *attr)

    int ldap_count_values(char **vals)
    int ldap_count_values_len(struct berval **vals)
    void ldap_value_free(char **vals)
    void ldap_value_free_len(struct berval **vals)

*/

func (self *LdapMessage) Count() int {
	// API : int ldap_count_messages(LDAP *ld, LDAPMessage *chain )
	// err : (count = -1)
	count := int(C.ldap_count_messages(self.ldap.conn, self.msg))
	if count == -1 {
		panic("LDAP::Count() (ldap_count_messages) error (-1)")
	}
	return count

}

func (self *LdapMessage) FirstMessage() *LdapMessage {

	var msg *C.LDAPMessage
	msg = C.ldap_first_message(self.ldap.conn, self.msg)
	if msg == nil {
		return nil
	}
	_msg := new(LdapMessage)
	_msg.ldap = self.ldap
	_msg.errno = 0
	_msg.msg = msg
	return _msg
}

func (self *LdapMessage) NextMessage() *LdapMessage {
	var msg *C.LDAPMessage
	msg = C.ldap_next_message(self.ldap.conn, self.msg)

	if msg == nil {
		return nil
	}
	_msg := new(LdapMessage)
	_msg.ldap = self.ldap
	_msg.errno = 0
	_msg.msg = msg
	return _msg
}

/* an alias to ldap_count_message() ? */
func (self *LdapEntry) CountEntries() int {
	// API : int ldap_count_messages(LDAP *ld, LDAPMessage *chain )
	// err : (count = -1)
	return int(C.ldap_count_entries(self.ldap.conn, self.entry))
}

func (self *LdapMessage) FirstEntry() *LdapEntry {

	var msg *C.LDAPMessage
	// API: LdapMessage *ldap_first_entry( LDAP *ld, LdapMessage *result )
	msg = C.ldap_first_entry(self.ldap.conn, self.msg)
	if msg == nil {
		return nil
	}
	_msg := new(LdapEntry)
	_msg.ldap = self.ldap
	_msg.errno = 0
	_msg.entry = msg
	return _msg
}

func (self *LdapEntry) NextEntry() *LdapEntry {
	var msg *C.LDAPMessage
	// API: LdapMessage *ldap_next_entry ( LDAP *ld, LdapMessage *entry )
	msg = C.ldap_next_entry(self.ldap.conn, self.entry)

	if msg == nil {
		return nil
	}
	_msg := new(LdapEntry)
	_msg.ldap = self.ldap
	_msg.errno = 0
	_msg.entry = msg
	return _msg
}

func (self *LdapEntry) FirstAttribute() (string, error) {

	var ber *C.BerElement

	// API: char *ldap_first_attribute(LDAP *ld, LdapMessage *entry, BerElement **berptr )
	rv := C.ldap_first_attribute(self.ldap.conn, self.entry, &ber)

	if rv == nil {
		// error
		return "", nil
	}
	self.ber = ber
	return C.GoString(rv), nil
}

func (self *LdapEntry) NextAttribute() (string, error) {

	// API: char *ldap_next_attribute (LDAP *ld, LdapMessage *entry, BerElement *ber )
	rv := C.ldap_next_attribute(self.ldap.conn, self.entry, self.ber)

	if rv == nil {
		// error
		return "", nil
	}
	return C.GoString(rv), nil
}

// private func
func sptr(p uintptr) *C.char {
	return *(**C.char)(unsafe.Pointer(p))
}

// private func used to convert null terminated char*[] to go []string
func cstrings_array(x **C.char) []string {
	var s []string
	for p := uintptr(unsafe.Pointer(x)); sptr(p) != nil; p += unsafe.Sizeof(uintptr(0)) {
		s = append(s, C.GoString(sptr(p)))
	}
	return s
}

// GetValues() return an array of string containing values for LDAP attribute "attr".
// Binary data are supported.
func (self *LdapEntry) GetValues(attr string) []string {
	var s []string

	_attr := C.CString(attr)
	defer C.free(unsafe.Pointer(_attr))

	var bv **C.struct_berval
	
	//API: struct berval **ldap_get_values_len(LDAP *ld, LDAPMessage *entry, char *attr)
	bv = C.ldap_get_values_len(self.ldap.conn, self.entry, _attr)

	var i int
	count := int(C.ldap_count_values_len(bv))

	for i = 0 ; i < count; i++ {
		s = append(s, C.GoStringN(C._berval_get_value(bv, C.int(i)), C._berval_get_len(bv, C.int(i))))
	}

	// free allocated array (bv)
	C.ldap_value_free_len(bv)

	return s
}

// ------------------------------------------------ RESULTS -----------------------------------------------
/*
    int ldap_result ( LDAP *ld, int msgid, int all, struct timeval *timeout, LdapMessage **result );
	int ldap_msgfree( LdapMessage *msg );
	int ldap_msgtype( LdapMessage *msg );
	int ldap_msgid  ( LdapMessage *msg );

*/

// Result()
// take care to free LdapMessage result with MsgFree()
//
func (self *Ldap) Result() (*LdapMessage, error) {

	var msgid int = 1
	var all int = 1

	var tv C.struct_timeval
	tv.tv_sec = 30

	var msg *C.LDAPMessage

	// API: int ldap_result( LDAP *ld, int msgid, int all, struct timeval *timeout, LDAPMessage **result );
	rv := C.ldap_result(self.conn, C.int(msgid), C.int(all), &tv, &msg)

	if rv != LDAP_OPT_SUCCESS {
		return nil, errors.New(fmt.Sprintf("LDAP::Result() error :  %d (%s)", rv, ErrorToString(int(rv))))
	}

	_msg := new(LdapMessage)
	_msg.ldap = self
	_msg.errno = int(rv)
	_msg.msg = msg

	return _msg, nil
}

// MsgFree() is used to free LDAP::Result() allocated data
//
// returns -1 on error.
//
func (self *LdapMessage) MsgFree() int{
        if self.msg != nil {
                rv := C.ldap_msgfree(self.msg)
                self.msg = nil
                return int(rv)
        }
        return -1
}


//  ---------------------------------------- DN Methods ---------------------------------------------------
/*

	char *ldap_get_dn( LDAP *ld, LdapMessage *entry)
	int   ldap_str2dn( const char *str, LDAPDN *dn, unsigned flags)
	void  ldap_dnfree( LDAPDN dn)
	int   ldap_dn2str( LDAPDN dn, char **str, unsigned flags)

	char **ldap_explode_dn( const char *dn, int notypes)
	char **ldap_explode_rdn( const char *rdn, int notypes)

	char *ldap_dn2ufn  ( const char * dn )
	char *ldap_dn2dcedn( const char * dn )
	char *ldap_dcedn2dn( const char * dn )
	char *ldap_dn2ad_canonical( const char * dn )

*/

// GetDn() return the DN (Distinguish Name) for self LdapEntry
func (self *LdapEntry) GetDn() string {
	// API: char *ldap_get_dn( LDAP *ld, LDAPMessage *entry )
	rv := C.ldap_get_dn(self.ldap.conn, self.entry)
	defer C.free(unsafe.Pointer(rv))

	if rv == nil {
		err := self.ldap.Errno()
		panic(fmt.Sprintf("LDAP::GetDn() error %d (%s)", err, ErrorToString(err)))
	}

	val := C.GoString(rv)
	return val
}
