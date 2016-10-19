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

/* #include <stdlib.h>
#include <stdio.h>
#define LDAP_DEPRECATED 1
#include <ldap.h>

// goc can not use union on structs ; so create a new type with same attributes and size
// fixme : support binary mods (mod_bvalues)
typedef struct ldapmod_str {
	int		       mod_op;
	char		   *mod_type;
	char           **mod_vals;
} LDAPModStr;


int _ldap_add(LDAP *ld, char* dn, LDAPModStr **attrs){

	//API: int ldap_add_ext_s(LDAP *ld, char *dn, LDAPMod **attrs, LDAPControl *sctrls, LDAPControl *cctrls );
	// nota : cast (LDAPMod **) is possible because structs have same size
	return ldap_add_ext_s(ld, dn, (LDAPMod **)attrs, NULL, NULL);
}

int _ldap_modify(LDAP *ld, char* dn, LDAPModStr **mods ){
 
	// nota : cast (LDAPMod **) is possible because structs have same size
	return ldap_modify_ext_s( ld, dn, (LDAPMod **)mods, NULL, NULL);
}

int _ldap_rename (LDAP *ld, char *dn, char *newrdn, char *newSuperior, int deleteoldrdn){
	//API: int ldap_rename_s( ld, dn, newrdn, newparent, deleteoldrdn, sctrls[], cctrls[])

	return ldap_rename_s(ld, dn, newrdn, newSuperior, deleteoldrdn, NULL, NULL);
}

void _ldap_mods_free (LDAPModStr **mods, int freemods){
	//API: void ldap_mods_free(LDAPMod **mods, int freemods);
	return ldap_mods_free((LDAPMod **)mods, freemods);
}


*/
// #cgo CFLAGS: -DLDAP_DEPRECATED=1
// #cgo linux CFLAGS: -DLINUX=1
// #cgo LDFLAGS: -lldap -llber
import "C"

import (
	"errors"
	"unsafe"
	"fmt"
)


func (self *Ldap) doModify(dn string, attrs map[string][]string, changeType int, full_add bool) (int){

	_dn := C.CString(dn)
	defer C.free(unsafe.Pointer(_dn))
	
	mods := make([]*C.LDAPModStr, len(attrs)+1)
	// mods[len] = nil by default

	count:= 0
	for key, values := range attrs {

		// transform []string to C.char** null terminated array (attributes argument)
		_values := make([]*C.char, len(values)+1) // default set to nil (NULL in C)

		for i, value := range values {
			_values[i] = C.CString(value)
			defer C.free(unsafe.Pointer(_values[i]))
		}

		var mod C.LDAPModStr

		mod.mod_op = C.int(changeType)
		mod.mod_type = C.CString(key)
		mod.mod_vals = &_values[0]
		
		defer C.free(unsafe.Pointer(mod.mod_type))

		mods[count] = &mod

		count++
	}

	var rv int

	if full_add {
		// API: int ldap_add (LDAP *ld, LDAP_CONST char *dn, LDAPMod **mods )
		rv = int(C._ldap_add(self.conn, _dn, &mods[0]))
	} else{
		// API: int ldap_modify (LDAP *ld, LDAP_CONST char *dn, LDAPMod **mods )
		rv = int(C._ldap_modify(self.conn, _dn, &mods[0]))
	}

	// FIXME: need to call ldap_mods_free(&mods) some where.
	// C._ldap_mods_free(&mods[0], 1) // not OK.
	return rv
}

func (self *Ldap) Modify(dn string, attrs map[string][]string) (error){

	changeType := C.LDAP_MOD_REPLACE
	full_add := false
	rv := self.doModify(dn, attrs, changeType, full_add)

	if rv != LDAP_OPT_SUCCESS {
		return errors.New(fmt.Sprintf("LDAP::Modify() error :  %d (%s)", rv, ErrorToString(rv)))
	}
	return nil
}

func (self *Ldap) ModifyDel(dn string, attrs map[string][]string) (error){

	changeType := C.LDAP_MOD_DELETE
	full_add := false
	rv := self.doModify(dn, attrs, changeType, full_add)

	if rv != LDAP_OPT_SUCCESS {
		return errors.New(fmt.Sprintf("LDAP::ModifyDel() error :  %d (%s)", rv, ErrorToString(rv)))
	}
	return nil
}

func (self *Ldap) ModifyAdd(dn string, attrs map[string][]string) (error){

	changeType := C.LDAP_MOD_ADD
	full_add := false
	rv := self.doModify(dn, attrs, changeType, full_add)
	if rv != LDAP_OPT_SUCCESS {
		return errors.New(fmt.Sprintf("LDAP::ModifyAdd() error :  %d (%s)", rv, ErrorToString(rv)))
	}
	return nil
}

func (self *Ldap) Add(dn string, attrs map[string][]string) (error){

	changeType := C.LDAP_MOD_ADD
	full_add := true
	rv := self.doModify(dn, attrs, changeType, full_add)
	if rv != LDAP_OPT_SUCCESS {
		return errors.New(fmt.Sprintf("LDAP::Add() error :  %d (%s)", rv, ErrorToString(rv)))
	}
	return nil
}

func (self *Ldap) Delete(dn string) (error){

	_dn := C.CString(dn)
	defer C.free(unsafe.Pointer(_dn))

	// API: int ldap_delete (LDAP *ld, LDAP_CONST char *dn)
	rv := C.ldap_delete_s(self.conn, _dn)

	if rv != LDAP_OPT_SUCCESS {
		return errors.New(fmt.Sprintf("LDAP::Delete() error :  %d (%s)", rv, ErrorToString(int(rv))))
	}

	return nil
}

// Rename() to rename LDAP entries.
//
// These  routines  are used to perform a LDAP rename operation.  The function changes the leaf compo-
// nent of an entry's distinguished name and  optionally moves the entry to a  new  parent  container.
// The  ldap_rename_s performs a rename operation synchronously.  The method takes dn, which points to
// the distinguished name of the entry whose attribute is being compared, newparent,the  distinguished
// name of the entry's new parent. If this parameter is NULL, only the RDN is changed.  The root DN is
// specified by passing a zero length string, "".  deleteoldrdn specifies whether the old  RDN  should
// be  retained  or  deleted.   Zero indicates that the old RDN should be retained. If you choose this
// option, the attribute will contain both names (the old and the new).  Non-zero indicates  that  the
// old  RDN should be deleted.  serverctrls points to an array of LDAPControl structures that list the
// client controls to use with this extended operation.  Use  NULL  to  specify  no  client  controls.
// clientctrls  points to an array of LDAPControl structures that list the client controls to use with
// the search.
// FIXME: support NULL and "" values for newSuperior parameter.
//
func (self *Ldap) Rename(dn string, newrdn string, newSuperior string, deleteOld bool) (error){

	_dn := C.CString(dn)
	defer C.free(unsafe.Pointer(_dn))

	_newrdn := C.CString(newrdn)
	defer C.free(unsafe.Pointer(_newrdn))

	_newSuperior := C.CString(newSuperior)
	defer C.free(unsafe.Pointer(_newSuperior))

	var _delete C.int = 0
	if deleteOld {
		_delete = 1
	}
 
	// API: int ldap_rename (LDAP *ld, char *newrdn, char *newSuperior, int deleteoldrdn)
	rv := C._ldap_rename(self.conn, _dn, _newrdn, _newSuperior, _delete)

	if rv != LDAP_OPT_SUCCESS {
		return errors.New(fmt.Sprintf("LDAP::Rename() error :  %d (%s)", rv, ErrorToString(int(rv))))
	}

	return nil
}
