/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

/*
Package types implements the CloudEvents type system.

CloudEvents defines a set of abstract types for event context attributes. Each
type has a corresponding native Go type and a canonical string encoding.  The
native Go types used to represent the CloudEvents types are:
bool, int32, string, []byte, *url.URL, time.Time

 +----------------+----------------+-----------------------------------+
 |CloudEvents Type|Native Type     |Convertible From                   |
 +================+================+===================================+
 |Bool            |bool            |bool                               |
 +----------------+----------------+-----------------------------------+
 |Integer         |int32           |Any numeric type with value in     |
 |                |                |range of int32                     |
 +----------------+----------------+-----------------------------------+
 |String          |string          |string                             |
 +----------------+----------------+-----------------------------------+
 |Binary          |[]byte          |[]byte                             |
 +----------------+----------------+-----------------------------------+
 |URI-Reference   |*url.URL        |url.URL, types.URIRef, types.URI   |
 +----------------+----------------+-----------------------------------+
 |URI             |*url.URL        |url.URL, types.URIRef, types.URI   |
 |                |                |Must be an absolute URI.           |
 +----------------+----------------+-----------------------------------+
 |Timestamp       |time.Time       |time.Time, types.Timestamp         |
 +----------------+----------------+-----------------------------------+

Extension attributes may be stored as a native type or a canonical string.  The
To<Type> functions will convert to the desired <Type> from any convertible type
or from the canonical string form.

The Parse<Type> and Format<Type> functions convert native types to/from
canonical strings.

Note are no Parse or Format functions for URL or string. For URL use the
standard url.Parse() and url.URL.String(). The canonical string format of a
string is the string itself.

*/
package types
