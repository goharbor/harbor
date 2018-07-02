//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

#define CK_PTR *
#ifndef NULL_PTR
#define NULL_PTR 0
#endif
#define CK_DEFINE_FUNCTION(returnType, name) returnType name
#define CK_DECLARE_FUNCTION(returnType, name) returnType name
#define CK_DECLARE_FUNCTION_POINTER(returnType, name) returnType (* name)
#define CK_CALLBACK_FUNCTION(returnType, name) returnType (* name)

#include <unistd.h>
#ifdef REPACK_STRUCTURES
# pragma pack(push, 1)
# include "pkcs11.h"
# pragma pack(pop)
#else
# include "pkcs11.h"
#endif

#ifdef REPACK_STRUCTURES

// Go doesn't support structures with non-default packing, but PKCS#11 requires
// pack(1) on Windows. Use structures with the same members as the CK_ ones but
// default packing, and copy data between the two.

typedef struct ckInfo {
	CK_VERSION cryptokiVersion;
	CK_UTF8CHAR manufacturerID[32];
	CK_FLAGS flags;
	CK_UTF8CHAR libraryDescription[32];
	CK_VERSION libraryVersion;
} ckInfo, *ckInfoPtr;

typedef struct ckAttr {
	CK_ATTRIBUTE_TYPE type;
	CK_VOID_PTR pValue;
	CK_ULONG ulValueLen;
} ckAttr, *ckAttrPtr;

typedef struct ckMech {
	CK_MECHANISM_TYPE mechanism;
	CK_VOID_PTR pParameter;
	CK_ULONG ulParameterLen;
} ckMech, *ckMechPtr;

CK_RV attrsToC(CK_ATTRIBUTE_PTR *attrOut, ckAttrPtr attrIn, CK_ULONG count);
void attrsFromC(ckAttrPtr attrOut, CK_ATTRIBUTE_PTR attrIn, CK_ULONG count);
void mechToC(CK_MECHANISM_PTR mechOut, ckMechPtr mechIn);

#define ATTR_TO_C(aout, ain, count, other) \
	CK_ATTRIBUTE_PTR aout; \
	{ \
		CK_RV e = attrsToC(&aout, ain, count); \
		if (e != CKR_OK ) { \
			if (other != NULL) free(other); \
			return e; \
		} \
	}
#define ATTR_FREE(aout) free(aout)
#define ATTR_FROM_C(aout, ain, count) attrsFromC(aout, ain, count)
#define MECH_TO_C(mout, min) \
	CK_MECHANISM mval, *mout = &mval; \
	if (min != NULL) { mechToC(mout, min); \
	} else { mout = NULL; }

#else // REPACK_STRUCTURES

// Dummy types and macros to avoid any unnecessary copying on UNIX

typedef CK_INFO ckInfo, *ckInfoPtr;
typedef CK_ATTRIBUTE ckAttr, *ckAttrPtr;
typedef CK_MECHANISM ckMech, *ckMechPtr;

#define ATTR_TO_C(aout, ain, count, other) CK_ATTRIBUTE_PTR aout = ain
#define ATTR_FREE(aout)
#define ATTR_FROM_C(aout, ain, count)
#define MECH_TO_C(mout, min) CK_MECHANISM_PTR mout = min

#endif // REPACK_STRUCTURES
