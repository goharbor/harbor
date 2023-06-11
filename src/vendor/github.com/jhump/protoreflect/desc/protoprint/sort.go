package protoprint

import (
	"fmt"
	"strings"

	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"

	"github.com/jhump/protoreflect/desc"
)

// ElementKind is an enumeration of the types of elements in a protobuf
// file descriptor. This can be used by custom sort functions, for
// printing a file using a custom ordering of elements.
type ElementKind int

const (
	KindPackage = ElementKind(iota) + 1
	KindImport
	KindOption
	KindField
	KindMessage
	KindEnum
	KindService
	KindExtensionRange
	KindExtension
	KindReservedRange
	KindReservedName
	KindEnumValue
	KindMethod
)

// Element represents an element in a proto descriptor that can be
// printed. This interface is primarily used to allow users of this package to
// define custom sort orders for the printed output. The methods of this
// interface represent the values that can be used for ordering elements.
type Element interface {
	// Kind returns the kind of the element. The kind determines which other
	// methods are applicable.
	Kind() ElementKind
	// Name returns the element name. This is NOT applicable to syntax,
	// extension range, and reserved range kinds and will return the empty
	// string for these kinds. For custom options, this will be the
	// fully-qualified name of the corresponding extension.
	Name() string
	// Number returns the element number. This is only applicable to field,
	// extension, and enum value kinds and will return zero for all other kinds.
	Number() int32
	// NumberRange returns the range of numbers/tags for the element. This is
	// only applicable to extension ranges and reserved ranges and will return
	// (0, 0) for all other kinds.
	NumberRange() (int32, int32)
	// Extendee is the extended message for the extension element. Elements
	// other than extensions will return the empty string.
	Extendee() string
	// IsCustomOption returns true if the element is a custom option. If it is
	// not (including if the element kind is not option) then this method will
	// return false.
	IsCustomOption() bool
}

func asElement(v interface{}) Element {
	switch v := v.(type) {
	case pkg:
		return pkgElement(v)
	case imp:
		return impElement(v)
	case []option:
		return (*optionElement)(&v[0])
	case reservedRange:
		return resvdRangeElement(v)
	case string:
		return resvdNameElement(v)
	case *desc.FieldDescriptor:
		return (*fieldElement)(v)
	case *desc.MessageDescriptor:
		return (*msgElement)(v)
	case *desc.EnumDescriptor:
		return (*enumElement)(v)
	case *desc.EnumValueDescriptor:
		return (*enumValElement)(v)
	case *desc.ServiceDescriptor:
		return (*svcElement)(v)
	case *desc.MethodDescriptor:
		return (*methodElement)(v)
	case *dpb.DescriptorProto_ExtensionRange:
		return (*extRangeElement)(v)
	default:
		panic(fmt.Sprintf("unexpected type of element: %T", v))
	}
}

type pkgElement pkg

var _ Element = pkgElement("")

func (p pkgElement) Kind() ElementKind {
	return KindPackage
}

func (p pkgElement) Name() string {
	return string(p)
}

func (p pkgElement) Number() int32 {
	return 0
}

func (p pkgElement) NumberRange() (int32, int32) {
	return 0, 0
}

func (p pkgElement) Extendee() string {
	return ""
}

func (p pkgElement) IsCustomOption() bool {
	return false
}

type impElement imp

var _ Element = impElement("")

func (i impElement) Kind() ElementKind {
	return KindImport
}

func (i impElement) Name() string {
	return string(i)
}

func (i impElement) Number() int32 {
	return 0
}

func (i impElement) NumberRange() (int32, int32) {
	return 0, 0
}

func (i impElement) Extendee() string {
	return ""
}

func (i impElement) IsCustomOption() bool {
	return false
}

type optionElement option

var _ Element = (*optionElement)(nil)

func (o *optionElement) Kind() ElementKind {
	return KindOption
}

func (o *optionElement) Name() string {
	if strings.HasPrefix(o.name, "(") {
		// remove parentheses
		return o.name[1 : len(o.name)-1]
	}
	return o.name
}

func (o *optionElement) Number() int32 {
	return 0
}

func (o *optionElement) NumberRange() (int32, int32) {
	return 0, 0
}

func (o *optionElement) Extendee() string {
	return ""
}

func (o *optionElement) IsCustomOption() bool {
	return strings.HasPrefix(o.name, "(")
}

type resvdRangeElement reservedRange

var _ Element = resvdRangeElement{}

func (r resvdRangeElement) Kind() ElementKind {
	return KindReservedRange
}

func (r resvdRangeElement) Name() string {
	return ""
}

func (r resvdRangeElement) Number() int32 {
	return 0
}

func (r resvdRangeElement) NumberRange() (int32, int32) {
	return r.start, r.end
}

func (r resvdRangeElement) Extendee() string {
	return ""
}

func (r resvdRangeElement) IsCustomOption() bool {
	return false
}

type resvdNameElement string

var _ Element = resvdNameElement("")

func (r resvdNameElement) Kind() ElementKind {
	return KindReservedName
}

func (r resvdNameElement) Name() string {
	return string(r)
}

func (r resvdNameElement) Number() int32 {
	return 0
}

func (r resvdNameElement) NumberRange() (int32, int32) {
	return 0, 0
}

func (r resvdNameElement) Extendee() string {
	return ""
}

func (r resvdNameElement) IsCustomOption() bool {
	return false
}

type fieldElement desc.FieldDescriptor

var _ Element = (*fieldElement)(nil)

func (f *fieldElement) Kind() ElementKind {
	if (*desc.FieldDescriptor)(f).IsExtension() {
		return KindExtension
	}
	return KindField
}

func (f *fieldElement) Name() string {
	return (*desc.FieldDescriptor)(f).GetName()
}

func (f *fieldElement) Number() int32 {
	return (*desc.FieldDescriptor)(f).GetNumber()
}

func (f *fieldElement) NumberRange() (int32, int32) {
	return 0, 0
}

func (f *fieldElement) Extendee() string {
	fd := (*desc.FieldDescriptor)(f)
	if fd.IsExtension() {
		fd.GetOwner().GetFullyQualifiedName()
	}
	return ""
}

func (f *fieldElement) IsCustomOption() bool {
	return false
}

type msgElement desc.MessageDescriptor

var _ Element = (*msgElement)(nil)

func (m *msgElement) Kind() ElementKind {
	return KindMessage
}

func (m *msgElement) Name() string {
	return (*desc.MessageDescriptor)(m).GetName()
}

func (m *msgElement) Number() int32 {
	return 0
}

func (m *msgElement) NumberRange() (int32, int32) {
	return 0, 0
}

func (m *msgElement) Extendee() string {
	return ""
}

func (m *msgElement) IsCustomOption() bool {
	return false
}

type enumElement desc.EnumDescriptor

var _ Element = (*enumElement)(nil)

func (e *enumElement) Kind() ElementKind {
	return KindEnum
}

func (e *enumElement) Name() string {
	return (*desc.EnumDescriptor)(e).GetName()
}

func (e *enumElement) Number() int32 {
	return 0
}

func (e *enumElement) NumberRange() (int32, int32) {
	return 0, 0
}

func (e *enumElement) Extendee() string {
	return ""
}

func (e *enumElement) IsCustomOption() bool {
	return false
}

type enumValElement desc.EnumValueDescriptor

var _ Element = (*enumValElement)(nil)

func (e *enumValElement) Kind() ElementKind {
	return KindEnumValue
}

func (e *enumValElement) Name() string {
	return (*desc.EnumValueDescriptor)(e).GetName()
}

func (e *enumValElement) Number() int32 {
	return (*desc.EnumValueDescriptor)(e).GetNumber()
}

func (e *enumValElement) NumberRange() (int32, int32) {
	return 0, 0
}

func (e *enumValElement) Extendee() string {
	return ""
}

func (e *enumValElement) IsCustomOption() bool {
	return false
}

type svcElement desc.ServiceDescriptor

var _ Element = (*svcElement)(nil)

func (s *svcElement) Kind() ElementKind {
	return KindService
}

func (s *svcElement) Name() string {
	return (*desc.ServiceDescriptor)(s).GetName()
}

func (s *svcElement) Number() int32 {
	return 0
}

func (s *svcElement) NumberRange() (int32, int32) {
	return 0, 0
}

func (s *svcElement) Extendee() string {
	return ""
}

func (s *svcElement) IsCustomOption() bool {
	return false
}

type methodElement desc.MethodDescriptor

var _ Element = (*methodElement)(nil)

func (m *methodElement) Kind() ElementKind {
	return KindMethod
}

func (m *methodElement) Name() string {
	return (*desc.MethodDescriptor)(m).GetName()
}

func (m *methodElement) Number() int32 {
	return 0
}

func (m *methodElement) NumberRange() (int32, int32) {
	return 0, 0
}

func (m *methodElement) Extendee() string {
	return ""
}

func (m *methodElement) IsCustomOption() bool {
	return false
}

type extRangeElement dpb.DescriptorProto_ExtensionRange

var _ Element = (*extRangeElement)(nil)

func (e *extRangeElement) Kind() ElementKind {
	return KindExtensionRange
}

func (e *extRangeElement) Name() string {
	return ""
}

func (e *extRangeElement) Number() int32 {
	return 0
}

func (e *extRangeElement) NumberRange() (int32, int32) {
	ext := (*dpb.DescriptorProto_ExtensionRange)(e)
	return ext.GetStart(), ext.GetEnd()
}

func (e *extRangeElement) Extendee() string {
	return ""
}

func (e *extRangeElement) IsCustomOption() bool {
	return false
}
