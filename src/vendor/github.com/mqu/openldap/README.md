OpenLDAP
====

this is Openldap binding in GO language. I don't work any more with golang, so, please fork this project.


Installation :
-----

Installation is easy and very quick, as you can see :

	# install openldap library and devel packages
	sudo apt-get install libldap libldap2-dev  # debian/ubuntu.
	sudo urpmi openldap-devel # fedora, RH, ...

	# install go
	go get github.com/mqu/openldap

	# verify you've got it :
	(cd $GOPATH ; go list ./...) | grep openldap

Usage
----

- Look a this [exemple](https://github.com/mqu/openldap/blob/master/_examples/test-openldap.go).
- a more complex example making  [LDAP search](https://github.com/mqu/openldap/blob/master/_examples/ldapsearch.go) that mimics ldapsearch command, printing out result on console.

Doc:
---
- run _go doc openldap_,
- will come soon, complete documentation in this [Wiki](https://github.com/mqu/openldap/wiki).
- look at [_examples/](https://github.com/mqu/openldap/blob/master/_examples/)*.go to see how to use this library.

Todo :
----

 - thread-safe test,
 - complete LDAP:GetOption() and LDAP:SetOption() method : now, they work only for integer values,
 - avoid using deprecated function (see LDAP_DEPRECATED flag and "// DEPRECATED" comments in *.go sources),
 - write some tests,
 - verify memory leaks (Valgrind),
 - support LDIF format (in, out),
 - add support for external commands (ldapadd, ldapdelete)
 - create an LDAP CLI (command line interface), like lftp, with commands like shell,
 - a nice GUI with GTK,
 - proxy, server,
 - what else ?


Links :
----

 - goc : http://code.google.com/p/go-wiki/wiki/cgo (how to bind native libraries to GO)
 - Openldap library (and server) : http://www.openldap.org/
 - Pure Go [LDAP](https://github.com/mmitton/ldap) library, with [ASN1](https://github.com/mmitton/asn1-ber) support.

Licence :
----

Copyright (C) 2012 - Marc Quinton.

Use of this source code is governed by the MIT Licence :
 http://opensource.org/licenses/mit-license.php

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
