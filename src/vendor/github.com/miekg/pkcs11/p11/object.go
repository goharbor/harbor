package p11

import (
	"errors"

	"github.com/miekg/pkcs11"
)

// Object represents a handle to a PKCS#11 object. It is attached to the
// session used to find it. Once that session is closed, operations on the
// Object will fail. Operations may also depend on the logged-in state of
// the application.
type Object struct {
	session      *sessionImpl
	objectHandle pkcs11.ObjectHandle
}

// Label returns the label of an object.
func (o Object) Label() (string, error) {
	labelBytes, err := o.Attribute(pkcs11.CKA_LABEL)
	if err != nil {
		return "", err
	}
	return string(labelBytes), nil
}

// Value returns an object's CKA_VALUE attribute, as bytes.
func (o Object) Value() ([]byte, error) {
	return o.Attribute(pkcs11.CKA_VALUE)
}

// Attribute gets exactly one attribute from a PKCS#11 object, returning
// an error if the attribute is not found, or if multiple attributes are
// returned. On success, it will return the value of that attribute as a slice
// of bytes. For attributes not present (i.e. CKR_ATTRIBUTE_TYPE_INVALID),
// Attribute returns a nil slice and nil error.
func (o Object) Attribute(attributeType uint) ([]byte, error) {
	o.session.Lock()
	defer o.session.Unlock()

	attrs, err := o.session.ctx.GetAttributeValue(o.session.handle, o.objectHandle,
		[]*pkcs11.Attribute{pkcs11.NewAttribute(attributeType, nil)})
	// The PKCS#11 spec states that C_GetAttributeValue may return
	// CKR_ATTRIBUTE_TYPE_INVALID if an object simply does not posses a given
	// attribute. We don't consider that an error, we just consider that
	// equivalent to an empty value.
	if err == pkcs11.Error(pkcs11.CKR_ATTRIBUTE_TYPE_INVALID) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if len(attrs) == 0 {
		return nil, errors.New("attribute not found")
	}
	if len(attrs) > 1 {
		return nil, errors.New("too many attributes found")
	}
	return attrs[0].Value, nil
}

// Set sets exactly one attribute on this object.
func (o Object) Set(attributeType uint, value []byte) error {
	o.session.Lock()
	defer o.session.Unlock()

	err := o.session.ctx.SetAttributeValue(o.session.handle, o.objectHandle,
		[]*pkcs11.Attribute{pkcs11.NewAttribute(attributeType, value)})
	if err != nil {
		return err
	}
	return nil
}

// Copy makes a copy of this object, with the attributes in template applied on
// top of it, if possible.
func (o Object) Copy(template []*pkcs11.Attribute) (Object, error) {
	s := o.session
	s.Lock()
	defer s.Unlock()
	newHandle, err := s.ctx.CopyObject(s.handle, o.objectHandle, template)
	if err != nil {
		return Object{}, err
	}
	return Object{
		session:      s,
		objectHandle: newHandle,
	}, nil
}

// Destroy destroys this object.
func (o Object) Destroy() error {
	s := o.session
	s.Lock()
	defer s.Unlock()
	return s.ctx.DestroyObject(s.handle, o.objectHandle)
}
