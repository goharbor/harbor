// Copyright 2013 Miek Gieben. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkcs11

/*
#include <stdlib.h>
#include <string.h>
#include "pkcs11go.h"
*/
import "C"
import "unsafe"

// GCMParams represents the parameters for the AES-GCM mechanism.
type GCMParams struct {
	arena
	params  *C.CK_GCM_PARAMS
	iv      []byte
	aad     []byte
	tagSize int
}

// NewGCMParams returns a pointer to AES-GCM parameters.
// This is a convenience function for passing GCM parameters to
// available mechanisms.
//
// *NOTE*
// Some HSMs, like CloudHSM, will ignore the IV you pass in and write their
// own. As a result, to support all libraries, memory is not freed
// automatically, so that after the EncryptInit/Encrypt operation the HSM's IV
// can be read back out. It is up to the caller to ensure that Free() is called
// on the GCMParams object at an appropriate time, which is after
// Encrypt/Decrypt. As an example:
//
// gcmParams := pkcs11.NewGCMParams(make([]byte, 12), nil, 128)
// p.ctx.EncryptInit(session, []*pkcs11.Mechanism{pkcs11.NewMechanism(pkcs11.CKM_AES_GCM, gcmParams)}, aesObjHandle)
// ct, _ := p.ctx.Encrypt(session, pt)
// iv := gcmParams.IV()
// gcmParams.Free()
func NewGCMParams(iv, aad []byte, tagSize int) *GCMParams {
	return &GCMParams{
		iv:      iv,
		aad:     aad,
		tagSize: tagSize,
	}
}

func cGCMParams(p *GCMParams) []byte {
	params := C.CK_GCM_PARAMS{
		ulTagBits: C.CK_ULONG(p.tagSize),
	}
	var arena arena
	if len(p.iv) > 0 {
		iv, ivLen := arena.Allocate(p.iv)
		params.pIv = C.CK_BYTE_PTR(iv)
		params.ulIvLen = ivLen
	}
	if len(p.aad) > 0 {
		aad, aadLen := arena.Allocate(p.aad)
		params.pAAD = C.CK_BYTE_PTR(aad)
		params.ulAADLen = aadLen
	}
	p.arena = arena
	p.params = &params
	return C.GoBytes(unsafe.Pointer(&params), C.int(unsafe.Sizeof(params)))
}

func (p *GCMParams) IV() []byte {
	if p == nil || p.params == nil {
		return nil
	}
	newIv := C.GoBytes(unsafe.Pointer(p.params.pIv), C.int(p.params.ulIvLen))
	iv := make([]byte, len(newIv))
	copy(iv, newIv)
	return iv
}

func (p *GCMParams) Free() {
	if p == nil || p.arena == nil {
		return
	}
	p.arena.Free()
	p.params = nil
	p.arena = nil
}
