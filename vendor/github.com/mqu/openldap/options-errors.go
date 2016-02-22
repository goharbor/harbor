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
#include <ldap.h>

static inline char* to_charptr(const void* s) { return (char*)s; }

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

// FIXME : support all kind of option (int, int*, ...)
func (self *Ldap) SetOption(opt int, val int) error {

	// API: ldap_set_option (LDAP *ld,int option, LDAP_CONST void *invalue));
	rv := C.ldap_set_option(self.conn, C.int(opt), unsafe.Pointer(&val))

	if rv == LDAP_OPT_SUCCESS {
		return nil
	}

	return errors.New(fmt.Sprintf("LDAP::SetOption() error (%d) : %s", int(rv), ErrorToString(int(rv))))
}

// FIXME : support all kind of option (int, int*, ...) should take care of all return type for ldap_get_option
func (self *Ldap) GetOption(opt int) (val int, err error) {

	// API: int ldap_get_option (LDAP *ld,int option, LDAP_CONST void *invalue));
	rv := C.ldap_get_option(self.conn, C.int(opt), unsafe.Pointer(&val))

	if rv == LDAP_OPT_SUCCESS {
		return val, nil
	}

	return 0, errors.New(fmt.Sprintf("LDAP::GetOption() error (%d) : %s", rv, ErrorToString(int(rv))))
}

/*
** WORK IN PROGRESS!
**
** OpenLDAP reentrancy/thread-safeness should be dynamically
** checked using ldap_get_option().
**
** The -lldap implementation is not thread-safe.
**
** The -lldap_r implementation is:
**              LDAP_API_FEATURE_THREAD_SAFE (basic thread safety)
** but also be:
**              LDAP_API_FEATURE_SESSION_THREAD_SAFE
**              LDAP_API_FEATURE_OPERATION_THREAD_SAFE
**
** The preprocessor flag LDAP_API_FEATURE_X_OPENLDAP_THREAD_SAFE
** can be used to determine if -lldap_r is available at compile
** time.  You must define LDAP_THREAD_SAFE if and only if you
** link with -lldap_r.
**
** If you fail to define LDAP_THREAD_SAFE when linking with
** -lldap_r or define LDAP_THREAD_SAFE when linking with -lldap,
** provided header definations and declarations may be incorrect.
**
 */

func (self *Ldap) IsThreadSafe() bool {
	// fmt.Println("IsThreadSafe()")
	// opt, err := self.GetOption(LDAP_API_FEATURE_THREAD_SAFE) ; fmt.Println(opt, err)
	// opt, err = self.GetOption(LDAP_THREAD_SAFE) ; fmt.Println(opt, err)
	// opt, err = self.GetOption(LDAP_API_FEATURE_X_OPENLDAP_THREAD_SAFE) ; fmt.Println(opt, err)

	//FIXME: need to implement LDAP::GetOption(LDAP_OPT_API_FEATURE_INFO)
	return false
}

func ErrorToString(err int) string {

	// API: char * ldap_err2string (int err )
	result := C.GoString(C.to_charptr(unsafe.Pointer(C.ldap_err2string(C.int(err)))))
	return result
}

func (self *Ldap) Errno() int {
	opt, _ := self.GetOption(LDAP_OPT_ERROR_NUMBER)
	return opt
}
