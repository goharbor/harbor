package protoprint

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/internal"
	"github.com/jhump/protoreflect/dynamic"
)

// Printer knows how to format file descriptors as proto source code. Its fields
// provide some control over how the resulting source file is constructed and
// formatted.
type Printer struct {
	// If true, comments are rendered using "/*" style comments. Otherwise, they
	// are printed using "//" style line comments.
	PreferMultiLineStyleComments bool

	// If true, elements are sorted into a canonical order.
	//
	// The canonical order for elements in a file follows:
	//  1. Syntax
	//  2. Package
	//  3. Imports (sorted lexically)
	//  4. Options (sorted by name, standard options before custom options)
	//  5. Messages (sorted by name)
	//  6. Enums (sorted by name)
	//  7. Services (sorted by name)
	//  8. Extensions (grouped by extendee, sorted by extendee+tag)
	//
	// The canonical order of elements in a message follows:
	//  1. Options (sorted by name, standard options before custom options)
	//  2. Fields and One-Ofs (sorted by tag; one-ofs interleaved based on the
	//     minimum tag therein)
	//  3. Nested Messages (sorted by name)
	//  4. Nested Enums (sorted by name)
	//  5. Extension ranges (sorted by starting tag number)
	//  6. Nested Extensions (grouped by extendee, sorted by extendee+tag)
	//  7. Reserved ranges (sorted by starting tag number)
	//  8. Reserved names (sorted lexically)
	//
	// Methods are sorted within a service by name and appear after any service
	// options (which are sorted by name, standard options before custom ones).
	// Enum values are sorted within an enum, first by numeric value then by
	// name, and also appear after any enum options.
	//
	// Options for fields, enum values, and extension ranges are sorted by name,
	// standard options before custom ones.
	SortElements bool

	// The "less" function used to sort elements when printing. It is given two
	// elements, a and b, and should return true if a is "less than" b. In this
	// case, "less than" means that element a should appear earlier in the file
	// than element b.
	//
	// If this field is nil, no custom sorting is done and the SortElements
	// field is consulted to decide how to order the output. If this field is
	// non-nil, the SortElements field is ignored and this function is called to
	// order elements.
	CustomSortFunction func(a, b Element) bool

	// The indentation used. Any characters other than spaces or tabs will be
	// replaced with spaces. If unset/empty, two spaces will be used.
	Indent string

	// If true, detached comments (between elements) will be ignored.
	//
	// Deprecated: Use OmitComments bitmask instead.
	OmitDetachedComments bool

	// A bitmask of comment types to omit. If unset, all comments will be
	// included. Use CommentsAll to not print any comments.
	OmitComments CommentType

	// If true, trailing comments that typically appear on the same line as an
	// element (option, field, enum value, method) will be printed on a separate
	// line instead.
	//
	// So, with this set, you'll get output like so:
	//
	//    // leading comment for field
	//    repeated string names = 1;
	//    // trailing comment
	//
	// If left false, the printer will try to emit trailing comments on the same
	// line instead:
	//
	//    // leading comment for field
	//    repeated string names = 1; // trailing comment
	//
	// If the trailing comment has more than one line, it will automatically be
	// forced to the next line.
	TrailingCommentsOnSeparateLine bool

	// If true, the printed output will eschew any blank lines, which otherwise
	// appear between descriptor elements and comment blocks. Note that if
	// detached comments are being printed, this will cause them to be merged
	// into the subsequent leading comments. Similarly, any element trailing
	// comments will be merged into the subsequent leading comments.
	Compact bool

	// If true, all references to messages, extensions, and enums (such as in
	// options, field types, and method request and response types) will be
	// fully-qualified. When left unset, the referenced elements will contain
	// only as much qualifier as is required.
	//
	// For example, if a message is in the same package as the reference, the
	// simple name can be used. If a message shares some context with the
	// reference, only the unshared context needs to be included. For example:
	//
	//  message Foo {
	//    message Bar {
	//      enum Baz {
	//        ZERO = 0;
	//        ONE = 1;
	//      }
	//    }
	//
	//    // This field shares some context as the enum it references: they are
	//    // both inside of the namespace Foo:
	//    //    field is "Foo.my_baz"
	//    //     enum is "Foo.Bar.Baz"
	//    // So we only need to qualify the reference with the context that they
	//    // do NOT have in common:
	//    Bar.Baz my_baz = 1;
	//  }
	//
	// When printing fully-qualified names, they will be preceded by a dot, to
	// avoid any ambiguity that they might be relative vs. fully-qualified.
	ForceFullyQualifiedNames bool

	// The number of options that trigger short options expressions to be
	// rendered using multiple lines. Short options expressions are those
	// found on fields and enum values, that use brackets ("[" and "]") and
	// comma-separated options. If more options than this are present, they
	// will be expanded to multiple lines (one option per line).
	//
	// If unset (e.g. if zero), a default threshold of 3 is used.
	ShortOptionsExpansionThresholdCount int

	// The length of printed options that trigger short options expressions to
	// be rendered using multiple lines. If the short options contain more than
	// one option and their printed length is longer than this threshold, they
	// will be expanded to multiple lines (one option per line).
	//
	// If unset (e.g. if zero), a default threshold of 50 is used.
	ShortOptionsExpansionThresholdLength int

	// The length of a printed option value message literal that triggers the
	// message literal to be rendered using multiple lines instead of using a
	// compact single-line form. The message must include at least two fields
	// or contain a field that is a nested message to be expanded.
	//
	// If unset (e.g. if zero), a default threshold of 50 is used.
	MessageLiteralExpansionThresholdLength int
}

// CommentType is a kind of comments in a proto source file. This can be used
// as a bitmask.
type CommentType int

const (
	// CommentsDetached refers to comments that are not "attached" to any
	// source element. They are attributed to the subsequent element in the
	// file as "detached" comments.
	CommentsDetached CommentType = 1 << iota
	// CommentsTrailing refers to a comment block immediately following an
	// element in the source file. If another element immediately follows
	// the trailing comment, it is instead considered a leading comment for
	// that subsequent element.
	CommentsTrailing
	// CommentsLeading refers to a comment block immediately preceding an
	// element in the source file. For high-level elements (those that have
	// their own descriptor), these are used as doc comments for that element.
	CommentsLeading
	// CommentsTokens refers to any comments (leading, trailing, or detached)
	// on low-level elements in the file. "High-level" elements have their own
	// descriptors, e.g. messages, enums, fields, services, and methods. But
	// comments can appear anywhere (such as around identifiers and keywords,
	// sprinkled inside the declarations of a high-level element). This class
	// of comments are for those extra comments sprinkled into the file.
	CommentsTokens

	// CommentsNonDoc refers to comments that are *not* doc comments. This is a
	// bitwise union of everything other than CommentsLeading. If you configure
	// a printer to omit this, only doc comments on descriptor elements will be
	// included in the printed output.
	CommentsNonDoc = CommentsDetached | CommentsTrailing | CommentsTokens
	// CommentsAll indicates all kinds of comments. If you configure a printer
	// to omit this, no comments will appear in the printed output, even if the
	// input descriptors had source info and comments.
	CommentsAll = -1
)

// PrintProtoFiles prints all of the given file descriptors. The given open
// function is given a file name and is responsible for creating the outputs and
// returning the corresponding writer.
func (p *Printer) PrintProtoFiles(fds []*desc.FileDescriptor, open func(name string) (io.WriteCloser, error)) error {
	for _, fd := range fds {
		w, err := open(fd.GetName())
		if err != nil {
			return fmt.Errorf("failed to open %s: %v", fd.GetName(), err)
		}
		err = func() error {
			defer w.Close()
			return p.PrintProtoFile(fd, w)
		}()
		if err != nil {
			return fmt.Errorf("failed to write %s: %v", fd.GetName(), err)
		}
	}
	return nil
}

// PrintProtosToFileSystem prints all of the given file descriptors to files in
// the given directory. If file names in the given descriptors include path
// information, they will be relative to the given root.
func (p *Printer) PrintProtosToFileSystem(fds []*desc.FileDescriptor, rootDir string) error {
	return p.PrintProtoFiles(fds, func(name string) (io.WriteCloser, error) {
		fullPath := filepath.Join(rootDir, name)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, err
		}
		return os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	})
}

// pkg represents a package name
type pkg string

// imp represents an imported file name
type imp string

// ident represents an identifier
type ident string

// option represents a resolved descriptor option
type option struct {
	name string
	val  interface{}
}

// reservedRange represents a reserved range from a message or enum
type reservedRange struct {
	start, end int32
}

// PrintProtoFile prints the given single file descriptor to the given writer.
func (p *Printer) PrintProtoFile(fd *desc.FileDescriptor, out io.Writer) error {
	return p.printProto(fd, out)
}

// PrintProto prints the given descriptor and returns the resulting string. This
// can be used to print proto files, but it can also be used to get the proto
// "source form" for any kind of descriptor, which can be a more user-friendly
// way to present descriptors that are intended for human consumption.
func (p *Printer) PrintProtoToString(dsc desc.Descriptor) (string, error) {
	var buf bytes.Buffer
	if err := p.printProto(dsc, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (p *Printer) printProto(dsc desc.Descriptor, out io.Writer) error {
	w := newWriter(out)

	if p.Indent == "" {
		// default indent to two spaces
		p.Indent = "  "
	} else {
		// indent must be all spaces or tabs, so convert other chars to spaces
		ind := make([]rune, 0, len(p.Indent))
		for _, r := range p.Indent {
			if r == '\t' {
				ind = append(ind, r)
			} else {
				ind = append(ind, ' ')
			}
		}
		p.Indent = string(ind)
	}
	if p.OmitDetachedComments {
		p.OmitComments |= CommentsDetached
	}

	er := dynamic.ExtensionRegistry{}
	er.AddExtensionsFromFileRecursively(dsc.GetFile())
	mf := dynamic.NewMessageFactoryWithExtensionRegistry(&er)
	fdp := dsc.GetFile().AsFileDescriptorProto()
	sourceInfo := internal.CreateSourceInfoMap(fdp)
	extendOptionLocations(sourceInfo, fdp.GetSourceCodeInfo().GetLocation())

	path := findElement(dsc)
	switch d := dsc.(type) {
	case *desc.FileDescriptor:
		p.printFile(d, mf, w, sourceInfo)
	case *desc.MessageDescriptor:
		p.printMessage(d, mf, w, sourceInfo, path, 0)
	case *desc.FieldDescriptor:
		var scope string
		if md, ok := d.GetParent().(*desc.MessageDescriptor); ok {
			scope = md.GetFullyQualifiedName()
		} else {
			scope = d.GetFile().GetPackage()
		}
		if d.IsExtension() {
			fmt.Fprint(w, "extend ")
			extNameSi := sourceInfo.Get(append(path, internal.Field_extendeeTag))
			p.printElementString(extNameSi, w, 0, p.qualifyName(d.GetFile().GetPackage(), scope, d.GetOwner().GetFullyQualifiedName()))
			fmt.Fprintln(w, "{")

			p.printField(d, mf, w, sourceInfo, path, scope, 1)

			fmt.Fprintln(w, "}")
		} else {
			p.printField(d, mf, w, sourceInfo, path, scope, 0)
		}
	case *desc.OneOfDescriptor:
		md := d.GetOwner()
		elements := elementAddrs{dsc: md}
		for i := range md.GetFields() {
			elements.addrs = append(elements.addrs, elementAddr{elementType: internal.Message_fieldsTag, elementIndex: i})
		}
		p.printOneOf(d, elements, 0, mf, w, sourceInfo, path[:len(path)-1], 0, path[len(path)-1])
	case *desc.EnumDescriptor:
		p.printEnum(d, mf, w, sourceInfo, path, 0)
	case *desc.EnumValueDescriptor:
		p.printEnumValue(d, mf, w, sourceInfo, path, 0)
	case *desc.ServiceDescriptor:
		p.printService(d, mf, w, sourceInfo, path, 0)
	case *desc.MethodDescriptor:
		p.printMethod(d, mf, w, sourceInfo, path, 0)
	}

	return w.err
}

func findElement(dsc desc.Descriptor) []int32 {
	if dsc.GetParent() == nil {
		return nil
	}
	path := findElement(dsc.GetParent())
	switch d := dsc.(type) {
	case *desc.MessageDescriptor:
		if pm, ok := d.GetParent().(*desc.MessageDescriptor); ok {
			return append(path, internal.Message_nestedMessagesTag, getMessageIndex(d, pm.GetNestedMessageTypes()))
		}
		return append(path, internal.File_messagesTag, getMessageIndex(d, d.GetFile().GetMessageTypes()))

	case *desc.FieldDescriptor:
		if d.IsExtension() {
			if pm, ok := d.GetParent().(*desc.MessageDescriptor); ok {
				return append(path, internal.Message_extensionsTag, getFieldIndex(d, pm.GetNestedExtensions()))
			}
			return append(path, internal.File_extensionsTag, getFieldIndex(d, d.GetFile().GetExtensions()))
		}
		return append(path, internal.Message_fieldsTag, getFieldIndex(d, d.GetOwner().GetFields()))

	case *desc.OneOfDescriptor:
		return append(path, internal.Message_oneOfsTag, getOneOfIndex(d, d.GetOwner().GetOneOfs()))

	case *desc.EnumDescriptor:
		if pm, ok := d.GetParent().(*desc.MessageDescriptor); ok {
			return append(path, internal.Message_enumsTag, getEnumIndex(d, pm.GetNestedEnumTypes()))
		}
		return append(path, internal.File_enumsTag, getEnumIndex(d, d.GetFile().GetEnumTypes()))

	case *desc.EnumValueDescriptor:
		return append(path, internal.Enum_valuesTag, getEnumValueIndex(d, d.GetEnum().GetValues()))

	case *desc.ServiceDescriptor:
		return append(path, internal.File_servicesTag, getServiceIndex(d, d.GetFile().GetServices()))

	case *desc.MethodDescriptor:
		return append(path, internal.Service_methodsTag, getMethodIndex(d, d.GetService().GetMethods()))

	default:
		panic(fmt.Sprintf("unexpected descriptor type: %T", dsc))
	}
}

func getMessageIndex(md *desc.MessageDescriptor, list []*desc.MessageDescriptor) int32 {
	for i := range list {
		if md == list[i] {
			return int32(i)
		}
	}
	panic(fmt.Sprintf("unable to determine index of message %s", md.GetFullyQualifiedName()))
}

func getFieldIndex(fd *desc.FieldDescriptor, list []*desc.FieldDescriptor) int32 {
	for i := range list {
		if fd == list[i] {
			return int32(i)
		}
	}
	panic(fmt.Sprintf("unable to determine index of field %s", fd.GetFullyQualifiedName()))
}

func getOneOfIndex(ood *desc.OneOfDescriptor, list []*desc.OneOfDescriptor) int32 {
	for i := range list {
		if ood == list[i] {
			return int32(i)
		}
	}
	panic(fmt.Sprintf("unable to determine index of oneof %s", ood.GetFullyQualifiedName()))
}

func getEnumIndex(ed *desc.EnumDescriptor, list []*desc.EnumDescriptor) int32 {
	for i := range list {
		if ed == list[i] {
			return int32(i)
		}
	}
	panic(fmt.Sprintf("unable to determine index of enum %s", ed.GetFullyQualifiedName()))
}

func getEnumValueIndex(evd *desc.EnumValueDescriptor, list []*desc.EnumValueDescriptor) int32 {
	for i := range list {
		if evd == list[i] {
			return int32(i)
		}
	}
	panic(fmt.Sprintf("unable to determine index of enum value %s", evd.GetFullyQualifiedName()))
}

func getServiceIndex(sd *desc.ServiceDescriptor, list []*desc.ServiceDescriptor) int32 {
	for i := range list {
		if sd == list[i] {
			return int32(i)
		}
	}
	panic(fmt.Sprintf("unable to determine index of service %s", sd.GetFullyQualifiedName()))
}

func getMethodIndex(mtd *desc.MethodDescriptor, list []*desc.MethodDescriptor) int32 {
	for i := range list {
		if mtd == list[i] {
			return int32(i)
		}
	}
	panic(fmt.Sprintf("unable to determine index of method %s", mtd.GetFullyQualifiedName()))
}

func (p *Printer) newLine(w io.Writer) {
	if !p.Compact {
		fmt.Fprintln(w)
	}
}

func (p *Printer) printFile(fd *desc.FileDescriptor, mf *dynamic.MessageFactory, w *writer, sourceInfo internal.SourceInfoMap) {
	opts, err := p.extractOptions(fd, fd.GetOptions(), mf)
	if err != nil {
		return
	}

	fdp := fd.AsFileDescriptorProto()
	path := make([]int32, 1)

	path[0] = internal.File_packageTag
	sourceInfo.PutIfAbsent(append(path, 0), sourceInfo.Get(path))

	path[0] = internal.File_syntaxTag
	si := sourceInfo.Get(path)
	p.printElement(false, si, w, 0, func(w *writer) {
		syn := fdp.GetSyntax()
		if syn == "" {
			syn = "proto2"
		}
		fmt.Fprintf(w, "syntax = %q;", syn)
	})
	p.newLine(w)

	skip := map[interface{}]bool{}

	elements := elementAddrs{dsc: fd, opts: opts}
	if fdp.Package != nil {
		elements.addrs = append(elements.addrs, elementAddr{elementType: internal.File_packageTag, elementIndex: 0, order: -3})
	}
	for i := range fdp.GetDependency() {
		elements.addrs = append(elements.addrs, elementAddr{elementType: internal.File_dependencyTag, elementIndex: i, order: -2})
	}
	elements.addrs = append(elements.addrs, optionsAsElementAddrs(internal.File_optionsTag, -1, opts)...)
	for i := range fd.GetMessageTypes() {
		elements.addrs = append(elements.addrs, elementAddr{elementType: internal.File_messagesTag, elementIndex: i})
	}
	for i := range fd.GetEnumTypes() {
		elements.addrs = append(elements.addrs, elementAddr{elementType: internal.File_enumsTag, elementIndex: i})
	}
	for i := range fd.GetServices() {
		elements.addrs = append(elements.addrs, elementAddr{elementType: internal.File_servicesTag, elementIndex: i})
	}
	exts := p.computeExtensions(sourceInfo, fd.GetExtensions(), []int32{internal.File_extensionsTag})
	for i, extd := range fd.GetExtensions() {
		if extd.GetType() == descriptor.FieldDescriptorProto_TYPE_GROUP {
			// we don't emit nested messages for groups since
			// they get special treatment
			skip[extd.GetMessageType()] = true
		}
		elements.addrs = append(elements.addrs, elementAddr{elementType: internal.File_extensionsTag, elementIndex: i})
	}

	p.sort(elements, sourceInfo, nil)

	pkgName := fd.GetPackage()

	for i, el := range elements.addrs {
		d := elements.at(el)

		// skip[d] will panic if d is a slice (which it could be for []option),
		// so just ignore it since we don't try to skip options
		if reflect.TypeOf(d).Kind() != reflect.Slice && skip[d] {
			// skip this element
			continue
		}

		if i > 0 {
			p.newLine(w)
		}

		path = []int32{el.elementType, int32(el.elementIndex)}

		switch d := d.(type) {
		case pkg:
			si := sourceInfo.Get(path)
			p.printElement(false, si, w, 0, func(w *writer) {
				fmt.Fprintf(w, "package %s;", d)
			})
		case imp:
			si := sourceInfo.Get(path)
			var modifier string
			for _, idx := range fdp.PublicDependency {
				if fdp.Dependency[idx] == string(d) {
					modifier = "public "
					break
				}
			}
			if modifier == "" {
				for _, idx := range fdp.WeakDependency {
					if fdp.Dependency[idx] == string(d) {
						modifier = "weak "
						break
					}
				}
			}
			p.printElement(false, si, w, 0, func(w *writer) {
				fmt.Fprintf(w, "import %s%q;", modifier, d)
			})
		case []option:
			p.printOptionsLong(d, w, sourceInfo, path, 0)
		case *desc.MessageDescriptor:
			p.printMessage(d, mf, w, sourceInfo, path, 0)
		case *desc.EnumDescriptor:
			p.printEnum(d, mf, w, sourceInfo, path, 0)
		case *desc.ServiceDescriptor:
			p.printService(d, mf, w, sourceInfo, path, 0)
		case *desc.FieldDescriptor:
			extDecl := exts[d]
			p.printExtensions(extDecl, exts, elements, i, mf, w, sourceInfo, nil, internal.File_extensionsTag, pkgName, pkgName, 0)
			// we printed all extensions in the group, so we can skip the others
			for _, fld := range extDecl.fields {
				skip[fld] = true
			}
		}
	}
}

func findExtSi(fieldSi *descriptor.SourceCodeInfo_Location, extSis []*descriptor.SourceCodeInfo_Location) *descriptor.SourceCodeInfo_Location {
	if len(fieldSi.GetSpan()) == 0 {
		return nil
	}
	for _, extSi := range extSis {
		if isSpanWithin(fieldSi.Span, extSi.Span) {
			return extSi
		}
	}
	return nil
}

func isSpanWithin(span, enclosing []int32) bool {
	start := enclosing[0]
	var end int32
	if len(enclosing) == 3 {
		end = enclosing[0]
	} else {
		end = enclosing[2]
	}
	if span[0] < start || span[0] > end {
		return false
	}

	if span[0] == start {
		return span[1] >= enclosing[1]
	} else if span[0] == end {
		return span[1] <= enclosing[len(enclosing)-1]
	}
	return true
}

type extensionDecl struct {
	extendee   string
	sourceInfo *descriptor.SourceCodeInfo_Location
	fields     []*desc.FieldDescriptor
}

type extensions map[*desc.FieldDescriptor]*extensionDecl

func (p *Printer) computeExtensions(sourceInfo internal.SourceInfoMap, exts []*desc.FieldDescriptor, path []int32) extensions {
	extsMap := map[string]map[*descriptor.SourceCodeInfo_Location]*extensionDecl{}
	extSis := sourceInfo.GetAll(path)
	for _, extd := range exts {
		name := extd.GetOwner().GetFullyQualifiedName()
		extSi := findExtSi(extd.GetSourceInfo(), extSis)
		extsBySi := extsMap[name]
		if extsBySi == nil {
			extsBySi = map[*descriptor.SourceCodeInfo_Location]*extensionDecl{}
			extsMap[name] = extsBySi
		}
		extDecl := extsBySi[extSi]
		if extDecl == nil {
			extDecl = &extensionDecl{
				sourceInfo: extSi,
				extendee:   name,
			}
			extsBySi[extSi] = extDecl
		}
		extDecl.fields = append(extDecl.fields, extd)
	}

	ret := extensions{}
	for _, extsBySi := range extsMap {
		for _, extDecl := range extsBySi {
			for _, extd := range extDecl.fields {
				ret[extd] = extDecl
			}
		}
	}
	return ret
}

func (p *Printer) sort(elements elementAddrs, sourceInfo internal.SourceInfoMap, path []int32) {
	if p.CustomSortFunction != nil {
		sort.Stable(customSortOrder{elementAddrs: elements, less: p.CustomSortFunction})
	} else if p.SortElements {
		// canonical sorted order
		sort.Stable(elements)
	} else {
		// use source order (per location information in SourceCodeInfo); or
		// if that isn't present use declaration order, but grouped by type
		sort.Stable(elementSrcOrder{
			elementAddrs: elements,
			sourceInfo:   sourceInfo,
			prefix:       path,
		})
	}
}

func (p *Printer) qualifyName(pkg, scope string, fqn string) string {
	if p.ForceFullyQualifiedNames {
		// forcing fully-qualified names; make sure to include preceding dot
		if fqn[0] == '.' {
			return fqn
		}
		return fmt.Sprintf(".%s", fqn)
	}

	// compute relative name (so no leading dot)
	if fqn[0] == '.' {
		fqn = fqn[1:]
	}
	if len(scope) > 0 && scope[len(scope)-1] != '.' {
		scope = scope + "."
	}
	for scope != "" {
		if strings.HasPrefix(fqn, scope) {
			return fqn[len(scope):]
		}
		if scope == pkg+"." {
			break
		}
		pos := strings.LastIndex(scope[:len(scope)-1], ".")
		scope = scope[:pos+1]
	}
	return fqn
}

func (p *Printer) typeString(fld *desc.FieldDescriptor, scope string) string {
	if fld.IsMap() {
		return fmt.Sprintf("map<%s, %s>", p.typeString(fld.GetMapKeyType(), scope), p.typeString(fld.GetMapValueType(), scope))
	}
	switch fld.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		return "int32"
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		return "int64"
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		return "uint32"
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		return "uint64"
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		return "sint32"
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		return "sint64"
	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		return "fixed32"
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		return "fixed64"
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		return "sfixed32"
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		return "sfixed64"
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		return "float"
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return "double"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return "bool"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return "string"
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return "bytes"
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		return p.qualifyName(fld.GetFile().GetPackage(), scope, fld.GetEnumType().GetFullyQualifiedName())
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		return p.qualifyName(fld.GetFile().GetPackage(), scope, fld.GetMessageType().GetFullyQualifiedName())
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		return fld.GetMessageType().GetName()
	}
	panic(fmt.Sprintf("invalid type: %v", fld.GetType()))
}

func (p *Printer) printMessage(md *desc.MessageDescriptor, mf *dynamic.MessageFactory, w *writer, sourceInfo internal.SourceInfoMap, path []int32, indent int) {
	si := sourceInfo.Get(path)
	p.printBlockElement(true, si, w, indent, func(w *writer, trailer func(int, bool)) {
		p.indent(w, indent)

		fmt.Fprint(w, "message ")
		nameSi := sourceInfo.Get(append(path, internal.Message_nameTag))
		p.printElementString(nameSi, w, indent, md.GetName())
		fmt.Fprintln(w, "{")
		trailer(indent+1, true)

		p.printMessageBody(md, mf, w, sourceInfo, path, indent+1)
		p.indent(w, indent)
		fmt.Fprintln(w, "}")
	})
}

func (p *Printer) printMessageBody(md *desc.MessageDescriptor, mf *dynamic.MessageFactory, w *writer, sourceInfo internal.SourceInfoMap, path []int32, indent int) {
	opts, err := p.extractOptions(md, md.GetOptions(), mf)
	if err != nil {
		if w.err == nil {
			w.err = err
		}
		return
	}

	skip := map[interface{}]bool{}
	maxTag := internal.GetMaxTag(md.GetMessageOptions().GetMessageSetWireFormat())

	elements := elementAddrs{dsc: md, opts: opts}
	elements.addrs = append(elements.addrs, optionsAsElementAddrs(internal.Message_optionsTag, -1, opts)...)
	for i := range md.AsDescriptorProto().GetReservedRange() {
		elements.addrs = append(elements.addrs, elementAddr{elementType: internal.Message_reservedRangeTag, elementIndex: i})
	}
	for i := range md.AsDescriptorProto().GetReservedName() {
		elements.addrs = append(elements.addrs, elementAddr{elementType: internal.Message_reservedNameTag, elementIndex: i})
	}
	for i := range md.AsDescriptorProto().GetExtensionRange() {
		elements.addrs = append(elements.addrs, elementAddr{elementType: internal.Message_extensionRangeTag, elementIndex: i})
	}
	for i, fld := range md.GetFields() {
		if fld.IsMap() || fld.GetType() == descriptor.FieldDescriptorProto_TYPE_GROUP {
			// we don't emit nested messages for map types or groups since
			// they get special treatment
			skip[fld.GetMessageType()] = true
		}
		elements.addrs = append(elements.addrs, elementAddr{elementType: internal.Message_fieldsTag, elementIndex: i})
	}
	for i := range md.GetNestedMessageTypes() {
		elements.addrs = append(elements.addrs, elementAddr{elementType: internal.Message_nestedMessagesTag, elementIndex: i})
	}
	for i := range md.GetNestedEnumTypes() {
		elements.addrs = append(elements.addrs, elementAddr{elementType: internal.Message_enumsTag, elementIndex: i})
	}
	exts := p.computeExtensions(sourceInfo, md.GetNestedExtensions(), append(path, internal.Message_extensionsTag))
	for i, extd := range md.GetNestedExtensions() {
		if extd.GetType() == descriptor.FieldDescriptorProto_TYPE_GROUP {
			// we don't emit nested messages for groups since
			// they get special treatment
			skip[extd.GetMessageType()] = true
		}
		elements.addrs = append(elements.addrs, elementAddr{elementType: internal.Message_extensionsTag, elementIndex: i})
	}

	p.sort(elements, sourceInfo, path)

	pkg := md.GetFile().GetPackage()
	scope := md.GetFullyQualifiedName()

	for i, el := range elements.addrs {
		d := elements.at(el)

		// skip[d] will panic if d is a slice (which it could be for []option),
		// so just ignore it since we don't try to skip options
		if reflect.TypeOf(d).Kind() != reflect.Slice && skip[d] {
			// skip this element
			continue
		}

		if i > 0 {
			p.newLine(w)
		}

		childPath := append(path, el.elementType, int32(el.elementIndex))

		switch d := d.(type) {
		case []option:
			p.printOptionsLong(d, w, sourceInfo, childPath, indent)
		case *desc.FieldDescriptor:
			if d.IsExtension() {
				extDecl := exts[d]
				p.printExtensions(extDecl, exts, elements, i, mf, w, sourceInfo, path, internal.Message_extensionsTag, pkg, scope, indent)
				// we printed all extensions in the group, so we can skip the others
				for _, fld := range extDecl.fields {
					skip[fld] = true
				}
			} else {
				ood := d.GetOneOf()
				if ood == nil || ood.IsSynthetic() {
					p.printField(d, mf, w, sourceInfo, childPath, scope, indent)
				} else {
					// print the one-of, including all of its fields
					p.printOneOf(ood, elements, i, mf, w, sourceInfo, path, indent, d.AsFieldDescriptorProto().GetOneofIndex())
					for _, fld := range ood.GetChoices() {
						skip[fld] = true
					}
				}
			}
		case *desc.MessageDescriptor:
			p.printMessage(d, mf, w, sourceInfo, childPath, indent)
		case *desc.EnumDescriptor:
			p.printEnum(d, mf, w, sourceInfo, childPath, indent)
		case *descriptor.DescriptorProto_ExtensionRange:
			// collapse ranges into a single "extensions" block
			ranges := []*descriptor.DescriptorProto_ExtensionRange{d}
			addrs := []elementAddr{el}
			for idx := i + 1; idx < len(elements.addrs); idx++ {
				elnext := elements.addrs[idx]
				if elnext.elementType != el.elementType {
					break
				}
				extr := elements.at(elnext).(*descriptor.DescriptorProto_ExtensionRange)
				if !areEqual(d.Options, extr.Options, mf) {
					break
				}
				ranges = append(ranges, extr)
				addrs = append(addrs, elnext)
				skip[extr] = true
			}
			p.printExtensionRanges(md, ranges, maxTag, addrs, mf, w, sourceInfo, path, indent)
		case reservedRange:
			// collapse reserved ranges into a single "reserved" block
			ranges := []reservedRange{d}
			addrs := []elementAddr{el}
			for idx := i + 1; idx < len(elements.addrs); idx++ {
				elnext := elements.addrs[idx]
				if elnext.elementType != el.elementType {
					break
				}
				rr := elements.at(elnext).(reservedRange)
				ranges = append(ranges, rr)
				addrs = append(addrs, elnext)
				skip[rr] = true
			}
			p.printReservedRanges(ranges, maxTag, addrs, w, sourceInfo, path, indent)
		case string: // reserved name
			// collapse reserved names into a single "reserved" block
			names := []string{d}
			addrs := []elementAddr{el}
			for idx := i + 1; idx < len(elements.addrs); idx++ {
				elnext := elements.addrs[idx]
				if elnext.elementType != el.elementType {
					break
				}
				rn := elements.at(elnext).(string)
				names = append(names, rn)
				addrs = append(addrs, elnext)
				skip[rn] = true
			}
			p.printReservedNames(names, addrs, w, sourceInfo, path, indent)
		}
	}
}

func areEqual(a, b proto.Message, mf *dynamic.MessageFactory) bool {
	// proto.Equal doesn't handle unknown extensions very well :(
	// so we convert to a dynamic message (which should know about all extensions via
	// extension registry) and then compare
	return dynamic.MessagesEqual(asDynamicIfPossible(a, mf), asDynamicIfPossible(b, mf))
}

func asDynamicIfPossible(msg proto.Message, mf *dynamic.MessageFactory) proto.Message {
	if dm, ok := msg.(*dynamic.Message); ok {
		return dm
	} else {
		md, err := desc.LoadMessageDescriptorForMessage(msg)
		if err == nil {
			dm := mf.NewDynamicMessage(md)
			if dm.ConvertFrom(msg) == nil {
				return dm
			}
		}
	}
	return msg
}

func (p *Printer) printField(fld *desc.FieldDescriptor, mf *dynamic.MessageFactory, w *writer, sourceInfo internal.SourceInfoMap, path []int32, scope string, indent int) {
	var groupPath []int32
	var si *descriptor.SourceCodeInfo_Location

	group := isGroup(fld)

	if group {
		// compute path to group message type
		groupPath = make([]int32, len(path)-2)
		copy(groupPath, path)

		var candidates []*desc.MessageDescriptor
		var parentTag int32
		switch parent := fld.GetParent().(type) {
		case *desc.MessageDescriptor:
			// group in a message
			candidates = parent.GetNestedMessageTypes()
			parentTag = internal.Message_nestedMessagesTag
		case *desc.FileDescriptor:
			// group that is a top-level extension
			candidates = parent.GetMessageTypes()
			parentTag = internal.File_messagesTag
		}

		var groupMsgIndex int32
		for i, nmd := range candidates {
			if nmd == fld.GetMessageType() {
				// found it
				groupMsgIndex = int32(i)
				break
			}
		}
		groupPath = append(groupPath, parentTag, groupMsgIndex)

		// the group message is where the field's comments and position are stored
		si = sourceInfo.Get(groupPath)
	} else {
		si = sourceInfo.Get(path)
	}

	p.printBlockElement(true, si, w, indent, func(w *writer, trailer func(int, bool)) {
		p.indent(w, indent)
		if shouldEmitLabel(fld) {
			locSi := sourceInfo.Get(append(path, internal.Field_labelTag))
			p.printElementString(locSi, w, indent, labelString(fld.GetLabel()))
		}

		if group {
			fmt.Fprint(w, "group ")
		}

		typeSi := sourceInfo.Get(append(path, internal.Field_typeTag))
		p.printElementString(typeSi, w, indent, p.typeString(fld, scope))

		if !group {
			nameSi := sourceInfo.Get(append(path, internal.Field_nameTag))
			p.printElementString(nameSi, w, indent, fld.GetName())
		}

		fmt.Fprint(w, "= ")
		numSi := sourceInfo.Get(append(path, internal.Field_numberTag))
		p.printElementString(numSi, w, indent, fmt.Sprintf("%d", fld.GetNumber()))

		opts, err := p.extractOptions(fld, fld.GetOptions(), mf)
		if err != nil {
			if w.err == nil {
				w.err = err
			}
			return
		}

		// we use negative values for "extras" keys so they can't collide
		// with legit option tags

		if !fld.GetFile().IsProto3() && fld.AsFieldDescriptorProto().DefaultValue != nil {
			defVal := fld.GetDefaultValue()
			if fld.GetEnumType() != nil {
				defVal = fld.GetEnumType().FindValueByNumber(defVal.(int32))
			}
			opts[-internal.Field_defaultTag] = []option{{name: "default", val: defVal}}
		}

		jsn := fld.AsFieldDescriptorProto().GetJsonName()
		if jsn != "" && jsn != internal.JsonName(fld.GetName()) {
			opts[-internal.Field_jsonNameTag] = []option{{name: "json_name", val: jsn}}
		}

		p.printOptionsShort(fld, opts, internal.Field_optionsTag, w, sourceInfo, path, indent)

		if group {
			fmt.Fprintln(w, "{")
			trailer(indent+1, true)

			p.printMessageBody(fld.GetMessageType(), mf, w, sourceInfo, groupPath, indent+1)

			p.indent(w, indent)
			fmt.Fprintln(w, "}")

		} else {
			fmt.Fprint(w, ";")
			trailer(indent, false)
		}
	})
}

func shouldEmitLabel(fld *desc.FieldDescriptor) bool {
	return fld.IsProto3Optional() ||
		(!fld.IsMap() && fld.GetOneOf() == nil &&
			(fld.GetLabel() != descriptor.FieldDescriptorProto_LABEL_OPTIONAL || !fld.GetFile().IsProto3()))
}

func labelString(lbl descriptor.FieldDescriptorProto_Label) string {
	switch lbl {
	case descriptor.FieldDescriptorProto_LABEL_OPTIONAL:
		return "optional"
	case descriptor.FieldDescriptorProto_LABEL_REQUIRED:
		return "required"
	case descriptor.FieldDescriptorProto_LABEL_REPEATED:
		return "repeated"
	}
	panic(fmt.Sprintf("invalid label: %v", lbl))
}

func isGroup(fld *desc.FieldDescriptor) bool {
	return fld.GetType() == descriptor.FieldDescriptorProto_TYPE_GROUP
}

func (p *Printer) printOneOf(ood *desc.OneOfDescriptor, parentElements elementAddrs, startFieldIndex int, mf *dynamic.MessageFactory, w *writer, sourceInfo internal.SourceInfoMap, parentPath []int32, indent int, ooIndex int32) {
	oopath := append(parentPath, internal.Message_oneOfsTag, ooIndex)
	oosi := sourceInfo.Get(oopath)
	p.printBlockElement(true, oosi, w, indent, func(w *writer, trailer func(int, bool)) {
		p.indent(w, indent)
		fmt.Fprint(w, "oneof ")
		extNameSi := sourceInfo.Get(append(oopath, internal.OneOf_nameTag))
		p.printElementString(extNameSi, w, indent, ood.GetName())
		fmt.Fprintln(w, "{")
		indent++
		trailer(indent, true)

		opts, err := p.extractOptions(ood, ood.GetOptions(), mf)
		if err != nil {
			if w.err == nil {
				w.err = err
			}
			return
		}

		elements := elementAddrs{dsc: ood, opts: opts}
		elements.addrs = append(elements.addrs, optionsAsElementAddrs(internal.OneOf_optionsTag, -1, opts)...)

		count := len(ood.GetChoices())
		for idx := startFieldIndex; count > 0 && idx < len(parentElements.addrs); idx++ {
			el := parentElements.addrs[idx]
			if el.elementType != internal.Message_fieldsTag {
				continue
			}
			if parentElements.at(el).(*desc.FieldDescriptor).GetOneOf() == ood {
				// negative tag indicates that this element is actually a sibling, not a child
				elements.addrs = append(elements.addrs, elementAddr{elementType: -internal.Message_fieldsTag, elementIndex: el.elementIndex})
				count--
			}
		}

		// the fields are already sorted, but we have to re-sort in order to
		// interleave the options (in the event that we are using file location
		// order and the option locations are interleaved with the fields)
		p.sort(elements, sourceInfo, oopath)
		scope := ood.GetOwner().GetFullyQualifiedName()

		for i, el := range elements.addrs {
			if i > 0 {
				p.newLine(w)
			}

			switch d := elements.at(el).(type) {
			case []option:
				childPath := append(oopath, el.elementType, int32(el.elementIndex))
				p.printOptionsLong(d, w, sourceInfo, childPath, indent)
			case *desc.FieldDescriptor:
				childPath := append(parentPath, -el.elementType, int32(el.elementIndex))
				p.printField(d, mf, w, sourceInfo, childPath, scope, indent)
			}
		}

		p.indent(w, indent-1)
		fmt.Fprintln(w, "}")
	})
}

func (p *Printer) printExtensions(exts *extensionDecl, allExts extensions, parentElements elementAddrs, startFieldIndex int, mf *dynamic.MessageFactory, w *writer, sourceInfo internal.SourceInfoMap, parentPath []int32, extTag int32, pkg, scope string, indent int) {
	path := append(parentPath, extTag)
	p.printLeadingComments(exts.sourceInfo, w, indent)
	p.indent(w, indent)
	fmt.Fprint(w, "extend ")
	extNameSi := sourceInfo.Get(append(path, 0, internal.Field_extendeeTag))
	p.printElementString(extNameSi, w, indent, p.qualifyName(pkg, scope, exts.extendee))
	fmt.Fprintln(w, "{")

	if p.printTrailingComments(exts.sourceInfo, w, indent+1) && !p.Compact {
		// separator line between trailing comment and next element
		fmt.Fprintln(w)
	}

	count := len(exts.fields)
	first := true
	for idx := startFieldIndex; count > 0 && idx < len(parentElements.addrs); idx++ {
		el := parentElements.addrs[idx]
		if el.elementType != extTag {
			continue
		}
		fld := parentElements.at(el).(*desc.FieldDescriptor)
		if allExts[fld] == exts {
			if first {
				first = false
			} else {
				p.newLine(w)
			}
			childPath := append(path, int32(el.elementIndex))
			p.printField(fld, mf, w, sourceInfo, childPath, scope, indent+1)
			count--
		}
	}

	p.indent(w, indent)
	fmt.Fprintln(w, "}")
}

func (p *Printer) printExtensionRanges(parent *desc.MessageDescriptor, ranges []*descriptor.DescriptorProto_ExtensionRange, maxTag int32, addrs []elementAddr, mf *dynamic.MessageFactory, w *writer, sourceInfo internal.SourceInfoMap, parentPath []int32, indent int) {
	p.indent(w, indent)
	fmt.Fprint(w, "extensions ")

	var opts *descriptor.ExtensionRangeOptions
	var elPath []int32
	first := true
	for i, extr := range ranges {
		if first {
			first = false
		} else {
			fmt.Fprint(w, ", ")
		}
		opts = extr.Options
		el := addrs[i]
		elPath = append(parentPath, el.elementType, int32(el.elementIndex))
		si := sourceInfo.Get(elPath)
		p.printElement(true, si, w, inline(indent), func(w *writer) {
			if extr.GetStart() == extr.GetEnd()-1 {
				fmt.Fprintf(w, "%d ", extr.GetStart())
			} else if extr.GetEnd()-1 == maxTag {
				fmt.Fprintf(w, "%d to max ", extr.GetStart())
			} else {
				fmt.Fprintf(w, "%d to %d ", extr.GetStart(), extr.GetEnd()-1)
			}
		})
	}
	dsc := extensionRange{owner: parent, extRange: ranges[0]}
	p.extractAndPrintOptionsShort(dsc, opts, mf, internal.ExtensionRange_optionsTag, w, sourceInfo, elPath, indent)

	fmt.Fprintln(w, ";")
}

func (p *Printer) printReservedRanges(ranges []reservedRange, maxVal int32, addrs []elementAddr, w *writer, sourceInfo internal.SourceInfoMap, parentPath []int32, indent int) {
	p.indent(w, indent)
	fmt.Fprint(w, "reserved ")

	first := true
	for i, rr := range ranges {
		if first {
			first = false
		} else {
			fmt.Fprint(w, ", ")
		}
		el := addrs[i]
		si := sourceInfo.Get(append(parentPath, el.elementType, int32(el.elementIndex)))
		p.printElement(false, si, w, inline(indent), func(w *writer) {
			if rr.start == rr.end {
				fmt.Fprintf(w, "%d ", rr.start)
			} else if rr.end == maxVal {
				fmt.Fprintf(w, "%d to max ", rr.start)
			} else {
				fmt.Fprintf(w, "%d to %d ", rr.start, rr.end)
			}
		})
	}

	fmt.Fprintln(w, ";")
}

func (p *Printer) printReservedNames(names []string, addrs []elementAddr, w *writer, sourceInfo internal.SourceInfoMap, parentPath []int32, indent int) {
	p.indent(w, indent)
	fmt.Fprint(w, "reserved ")

	first := true
	for i, name := range names {
		if first {
			first = false
		} else {
			fmt.Fprint(w, ", ")
		}
		el := addrs[i]
		si := sourceInfo.Get(append(parentPath, el.elementType, int32(el.elementIndex)))
		p.printElementString(si, w, indent, quotedString(name))
	}

	fmt.Fprintln(w, ";")
}

func (p *Printer) printEnum(ed *desc.EnumDescriptor, mf *dynamic.MessageFactory, w *writer, sourceInfo internal.SourceInfoMap, path []int32, indent int) {
	si := sourceInfo.Get(path)
	p.printBlockElement(true, si, w, indent, func(w *writer, trailer func(int, bool)) {
		p.indent(w, indent)

		fmt.Fprint(w, "enum ")
		nameSi := sourceInfo.Get(append(path, internal.Enum_nameTag))
		p.printElementString(nameSi, w, indent, ed.GetName())
		fmt.Fprintln(w, "{")
		indent++
		trailer(indent, true)

		opts, err := p.extractOptions(ed, ed.GetOptions(), mf)
		if err != nil {
			if w.err == nil {
				w.err = err
			}
			return
		}

		skip := map[interface{}]bool{}

		elements := elementAddrs{dsc: ed, opts: opts}
		elements.addrs = append(elements.addrs, optionsAsElementAddrs(internal.Enum_optionsTag, -1, opts)...)
		for i := range ed.GetValues() {
			elements.addrs = append(elements.addrs, elementAddr{elementType: internal.Enum_valuesTag, elementIndex: i})
		}
		for i := range ed.AsEnumDescriptorProto().GetReservedRange() {
			elements.addrs = append(elements.addrs, elementAddr{elementType: internal.Enum_reservedRangeTag, elementIndex: i})
		}
		for i := range ed.AsEnumDescriptorProto().GetReservedName() {
			elements.addrs = append(elements.addrs, elementAddr{elementType: internal.Enum_reservedNameTag, elementIndex: i})
		}

		p.sort(elements, sourceInfo, path)

		for i, el := range elements.addrs {
			d := elements.at(el)

			// skip[d] will panic if d is a slice (which it could be for []option),
			// so just ignore it since we don't try to skip options
			if reflect.TypeOf(d).Kind() != reflect.Slice && skip[d] {
				// skip this element
				continue
			}

			if i > 0 {
				p.newLine(w)
			}

			childPath := append(path, el.elementType, int32(el.elementIndex))

			switch d := d.(type) {
			case []option:
				p.printOptionsLong(d, w, sourceInfo, childPath, indent)
			case *desc.EnumValueDescriptor:
				p.printEnumValue(d, mf, w, sourceInfo, childPath, indent)
			case reservedRange:
				// collapse reserved ranges into a single "reserved" block
				ranges := []reservedRange{d}
				addrs := []elementAddr{el}
				for idx := i + 1; idx < len(elements.addrs); idx++ {
					elnext := elements.addrs[idx]
					if elnext.elementType != el.elementType {
						break
					}
					rr := elements.at(elnext).(reservedRange)
					ranges = append(ranges, rr)
					addrs = append(addrs, elnext)
					skip[rr] = true
				}
				p.printReservedRanges(ranges, math.MaxInt32, addrs, w, sourceInfo, path, indent)
			case string: // reserved name
				// collapse reserved names into a single "reserved" block
				names := []string{d}
				addrs := []elementAddr{el}
				for idx := i + 1; idx < len(elements.addrs); idx++ {
					elnext := elements.addrs[idx]
					if elnext.elementType != el.elementType {
						break
					}
					rn := elements.at(elnext).(string)
					names = append(names, rn)
					addrs = append(addrs, elnext)
					skip[rn] = true
				}
				p.printReservedNames(names, addrs, w, sourceInfo, path, indent)
			}
		}

		p.indent(w, indent-1)
		fmt.Fprintln(w, "}")
	})
}

func (p *Printer) printEnumValue(evd *desc.EnumValueDescriptor, mf *dynamic.MessageFactory, w *writer, sourceInfo internal.SourceInfoMap, path []int32, indent int) {
	si := sourceInfo.Get(path)
	p.printElement(true, si, w, indent, func(w *writer) {
		p.indent(w, indent)

		nameSi := sourceInfo.Get(append(path, internal.EnumVal_nameTag))
		p.printElementString(nameSi, w, indent, evd.GetName())
		fmt.Fprint(w, "= ")

		numSi := sourceInfo.Get(append(path, internal.EnumVal_numberTag))
		p.printElementString(numSi, w, indent, fmt.Sprintf("%d", evd.GetNumber()))

		p.extractAndPrintOptionsShort(evd, evd.GetOptions(), mf, internal.EnumVal_optionsTag, w, sourceInfo, path, indent)

		fmt.Fprint(w, ";")
	})
}

func (p *Printer) printService(sd *desc.ServiceDescriptor, mf *dynamic.MessageFactory, w *writer, sourceInfo internal.SourceInfoMap, path []int32, indent int) {
	si := sourceInfo.Get(path)
	p.printBlockElement(true, si, w, indent, func(w *writer, trailer func(int, bool)) {
		p.indent(w, indent)

		fmt.Fprint(w, "service ")
		nameSi := sourceInfo.Get(append(path, internal.Service_nameTag))
		p.printElementString(nameSi, w, indent, sd.GetName())
		fmt.Fprintln(w, "{")
		indent++
		trailer(indent, true)

		opts, err := p.extractOptions(sd, sd.GetOptions(), mf)
		if err != nil {
			if w.err == nil {
				w.err = err
			}
			return
		}

		elements := elementAddrs{dsc: sd, opts: opts}
		elements.addrs = append(elements.addrs, optionsAsElementAddrs(internal.Service_optionsTag, -1, opts)...)
		for i := range sd.GetMethods() {
			elements.addrs = append(elements.addrs, elementAddr{elementType: internal.Service_methodsTag, elementIndex: i})
		}

		p.sort(elements, sourceInfo, path)

		for i, el := range elements.addrs {
			if i > 0 {
				p.newLine(w)
			}

			childPath := append(path, el.elementType, int32(el.elementIndex))

			switch d := elements.at(el).(type) {
			case []option:
				p.printOptionsLong(d, w, sourceInfo, childPath, indent)
			case *desc.MethodDescriptor:
				p.printMethod(d, mf, w, sourceInfo, childPath, indent)
			}
		}

		p.indent(w, indent-1)
		fmt.Fprintln(w, "}")
	})
}

func (p *Printer) printMethod(mtd *desc.MethodDescriptor, mf *dynamic.MessageFactory, w *writer, sourceInfo internal.SourceInfoMap, path []int32, indent int) {
	si := sourceInfo.Get(path)
	pkg := mtd.GetFile().GetPackage()
	p.printBlockElement(true, si, w, indent, func(w *writer, trailer func(int, bool)) {
		p.indent(w, indent)

		fmt.Fprint(w, "rpc ")
		nameSi := sourceInfo.Get(append(path, internal.Method_nameTag))
		p.printElementString(nameSi, w, indent, mtd.GetName())

		fmt.Fprint(w, "( ")
		inSi := sourceInfo.Get(append(path, internal.Method_inputTag))
		inName := p.qualifyName(pkg, pkg, mtd.GetInputType().GetFullyQualifiedName())
		if mtd.IsClientStreaming() {
			inName = "stream " + inName
		}
		p.printElementString(inSi, w, indent, inName)

		fmt.Fprint(w, ") returns ( ")

		outSi := sourceInfo.Get(append(path, internal.Method_outputTag))
		outName := p.qualifyName(pkg, pkg, mtd.GetOutputType().GetFullyQualifiedName())
		if mtd.IsServerStreaming() {
			outName = "stream " + outName
		}
		p.printElementString(outSi, w, indent, outName)
		fmt.Fprint(w, ") ")

		opts, err := p.extractOptions(mtd, mtd.GetOptions(), mf)
		if err != nil {
			if w.err == nil {
				w.err = err
			}
			return
		}

		if len(opts) > 0 {
			fmt.Fprintln(w, "{")
			indent++
			trailer(indent, true)

			elements := elementAddrs{dsc: mtd, opts: opts}
			elements.addrs = optionsAsElementAddrs(internal.Method_optionsTag, 0, opts)
			p.sort(elements, sourceInfo, path)

			for i, el := range elements.addrs {
				if i > 0 {
					p.newLine(w)
				}
				o := elements.at(el).([]option)
				childPath := append(path, el.elementType, int32(el.elementIndex))
				p.printOptionsLong(o, w, sourceInfo, childPath, indent)
			}

			p.indent(w, indent-1)
			fmt.Fprintln(w, "}")
		} else {
			fmt.Fprint(w, ";")
			trailer(indent, false)
		}
	})
}

func (p *Printer) printOptionsLong(opts []option, w *writer, sourceInfo internal.SourceInfoMap, path []int32, indent int) {
	p.printOptions(opts, w, indent,
		func(i int32) *descriptor.SourceCodeInfo_Location {
			return sourceInfo.Get(append(path, i))
		},
		func(w *writer, indent int, opt option, _ bool) {
			p.indent(w, indent)
			fmt.Fprint(w, "option ")
			p.printOption(opt.name, opt.val, w, indent)
			fmt.Fprint(w, ";")
		},
		false)
}

func (p *Printer) extractAndPrintOptionsShort(dsc interface{}, optsMsg proto.Message, mf *dynamic.MessageFactory, optsTag int32, w *writer, sourceInfo internal.SourceInfoMap, path []int32, indent int) {
	d, ok := dsc.(desc.Descriptor)
	if !ok {
		d = dsc.(extensionRange).owner
	}
	opts, err := p.extractOptions(d, optsMsg, mf)
	if err != nil {
		if w.err == nil {
			w.err = err
		}
		return
	}
	p.printOptionsShort(dsc, opts, optsTag, w, sourceInfo, path, indent)
}

func (p *Printer) printOptionsShort(dsc interface{}, opts map[int32][]option, optsTag int32, w *writer, sourceInfo internal.SourceInfoMap, path []int32, indent int) {
	elements := elementAddrs{dsc: dsc, opts: opts}
	elements.addrs = optionsAsElementAddrs(optsTag, 0, opts)
	if len(elements.addrs) == 0 {
		return
	}
	p.sort(elements, sourceInfo, path)

	// we render expanded form if there are many options
	count := 0
	for _, addr := range elements.addrs {
		opts := elements.at(addr).([]option)
		count += len(opts)
	}
	threshold := p.ShortOptionsExpansionThresholdCount
	if threshold <= 0 {
		threshold = 3
	}

	if count > threshold {
		p.printOptionElementsShort(elements, w, sourceInfo, path, indent, true)
	} else {
		var tmp bytes.Buffer
		tmpW := *w
		tmpW.Writer = &tmp
		p.printOptionElementsShort(elements, &tmpW, sourceInfo, path, indent, false)
		threshold := p.ShortOptionsExpansionThresholdLength
		if threshold <= 0 {
			threshold = 50
		}
		// we subtract 3 so we don't consider the leading " [" and trailing "]"
		if tmp.Len()-3 > threshold {
			p.printOptionElementsShort(elements, w, sourceInfo, path, indent, true)
		} else {
			// not too long: commit what we rendered
			b := tmp.Bytes()
			if w.space && len(b) > 0 && b[0] == ' ' {
				// don't write extra space
				b = b[1:]
			}
			w.Write(b)
			w.newline = tmpW.newline
			w.space = tmpW.space
		}
	}
}

func (p *Printer) printOptionElementsShort(addrs elementAddrs, w *writer, sourceInfo internal.SourceInfoMap, path []int32, indent int, expand bool) {
	if expand {
		fmt.Fprintln(w, "[")
		indent++
	} else {
		fmt.Fprint(w, "[")
	}
	for i, addr := range addrs.addrs {
		opts := addrs.at(addr).([]option)
		var childPath []int32
		if addr.elementIndex < 0 {
			// pseudo-option
			childPath = append(path, int32(-addr.elementIndex))
		} else {
			childPath = append(path, addr.elementType, int32(addr.elementIndex))
		}
		optIndent := indent
		if !expand {
			optIndent = inline(indent)
		}
		p.printOptions(opts, w, optIndent,
			func(i int32) *descriptor.SourceCodeInfo_Location {
				p := childPath
				if addr.elementIndex >= 0 {
					p = append(p, i)
				}
				return sourceInfo.Get(p)
			},
			func(w *writer, indent int, opt option, more bool) {
				if expand {
					p.indent(w, indent)
				}
				p.printOption(opt.name, opt.val, w, indent)
				if more {
					if expand {
						fmt.Fprintln(w, ",")
					} else {
						fmt.Fprint(w, ", ")
					}
				}
			},
			i < len(addrs.addrs)-1)
	}
	if expand {
		p.indent(w, indent-1)
	}
	fmt.Fprint(w, "] ")
}

func (p *Printer) printOptions(opts []option, w *writer, indent int, siFetch func(i int32) *descriptor.SourceCodeInfo_Location, fn func(w *writer, indent int, opt option, more bool), haveMore bool) {
	for i, opt := range opts {
		more := haveMore
		if !more {
			more = i < len(opts)-1
		}
		si := siFetch(int32(i))
		p.printElement(false, si, w, indent, func(w *writer) {
			fn(w, indent, opt, more)
		})
	}
}

func inline(indent int) int {
	if indent < 0 {
		// already inlined
		return indent
	}
	// negative indent means inline; indent 2 stops further in case value wraps
	return -indent - 2
}

func sortKeys(m map[interface{}]interface{}) []interface{} {
	res := make(sortedKeys, len(m))
	i := 0
	for k := range m {
		res[i] = k
		i++
	}
	sort.Sort(res)
	return ([]interface{})(res)
}

type sortedKeys []interface{}

func (k sortedKeys) Len() int {
	return len(k)
}

func (k sortedKeys) Swap(i, j int) {
	k[i], k[j] = k[j], k[i]
}

func (k sortedKeys) Less(i, j int) bool {
	switch i := k[i].(type) {
	case int32:
		return i < k[j].(int32)
	case uint32:
		return i < k[j].(uint32)
	case int64:
		return i < k[j].(int64)
	case uint64:
		return i < k[j].(uint64)
	case string:
		return i < k[j].(string)
	case bool:
		return !i && k[j].(bool)
	default:
		panic(fmt.Sprintf("invalid type for map key: %T", i))
	}
}

func (p *Printer) printOption(name string, optVal interface{}, w *writer, indent int) {
	fmt.Fprintf(w, "%s = ", name)

	switch optVal := optVal.(type) {
	case int32, uint32, int64, uint64:
		fmt.Fprintf(w, "%d", optVal)
	case float32, float64:
		fmt.Fprintf(w, "%f", optVal)
	case string:
		fmt.Fprintf(w, "%s", quotedString(optVal))
	case []byte:
		fmt.Fprintf(w, "%s", quotedBytes(string(optVal)))
	case bool:
		fmt.Fprintf(w, "%v", optVal)
	case ident:
		fmt.Fprintf(w, "%s", optVal)
	case *desc.EnumValueDescriptor:
		fmt.Fprintf(w, "%s", optVal.GetName())
	case proto.Message:
		// TODO: alternate approach so we can apply p.ForceFullyQualifiedNames
		// inside the resulting value?

		if indent < 0 {
			// if printing inline, always use compact form
			fmt.Fprintf(w, "{ %s }", proto.CompactTextString(optVal))
			return
		}
		m := proto.TextMarshaler{
			Compact:   true,
			ExpandAny: true,
		}
		str := strings.TrimSuffix(m.Text(optVal), " ")
		fieldCount := strings.Count(str, ":")
		nestedCount := strings.Count(str, "{") + strings.Count(str, "<")
		if fieldCount <= 1 && nestedCount == 0 {
			// can't expand
			fmt.Fprintf(w, "{ %s }", str)
			return
		}
		threshold := p.MessageLiteralExpansionThresholdLength
		if threshold == 0 {
			threshold = 50
		}
		if len(str) <= threshold {
			// no need to expand
			fmt.Fprintf(w, "{ %s }", str)
			return
		}

		// multi-line form
		m.Compact = false
		str = m.Text(optVal)
		fmt.Fprintln(w, "{")
		p.indentMessageLiteral(w, indent+1, str)
		p.indent(w, indent)
		fmt.Fprint(w, "}")
	default:
		panic(fmt.Sprintf("unknown type of value %T for field %s", optVal, name))
	}
}

func (p *Printer) indentMessageLiteral(w *writer, indent int, val string) {
	lines := strings.Split(val, "\n")
	for _, l := range lines {
		if l == "" {
			continue
		}
		if p.Indent != "  " {
			var prefix int
			for i := 0; i < len(l); i++ {
				if l[i] != ' ' {
					prefix = i
					break
				}
			}
			// replace text marshaller indent (2 spaces) with p.Indent
			prefixStr := strings.ReplaceAll(l[:prefix], "  ", p.Indent)
			l = prefixStr + l[prefix:]
		}
		p.indent(w, indent)
		fmt.Fprintln(w, l)
	}
}

type edgeKind int

const (
	edgeKindOption edgeKind = iota
	edgeKindFile
	edgeKindMessage
	edgeKindField
	edgeKindOneOf
	edgeKindExtensionRange
	edgeKindEnum
	edgeKindEnumVal
	edgeKindService
	edgeKindMethod
)

// edges in simple state machine for matching options paths
// whose prefix should be included in source info to handle
// the way options are printed (which cannot always include
// the full path from original source)
var edges = map[edgeKind]map[int32]edgeKind{
	edgeKindFile: {
		internal.File_optionsTag:    edgeKindOption,
		internal.File_messagesTag:   edgeKindMessage,
		internal.File_enumsTag:      edgeKindEnum,
		internal.File_extensionsTag: edgeKindField,
		internal.File_servicesTag:   edgeKindService,
	},
	edgeKindMessage: {
		internal.Message_optionsTag:        edgeKindOption,
		internal.Message_fieldsTag:         edgeKindField,
		internal.Message_oneOfsTag:         edgeKindOneOf,
		internal.Message_nestedMessagesTag: edgeKindMessage,
		internal.Message_enumsTag:          edgeKindEnum,
		internal.Message_extensionsTag:     edgeKindField,
		internal.Message_extensionRangeTag: edgeKindExtensionRange,
		// TODO: reserved range tag
	},
	edgeKindField: {
		internal.Field_optionsTag: edgeKindOption,
	},
	edgeKindOneOf: {
		internal.OneOf_optionsTag: edgeKindOption,
	},
	edgeKindExtensionRange: {
		internal.ExtensionRange_optionsTag: edgeKindOption,
	},
	edgeKindEnum: {
		internal.Enum_optionsTag: edgeKindOption,
		internal.Enum_valuesTag:  edgeKindEnumVal,
	},
	edgeKindEnumVal: {
		internal.EnumVal_optionsTag: edgeKindOption,
	},
	edgeKindService: {
		internal.Service_optionsTag: edgeKindOption,
		internal.Service_methodsTag: edgeKindMethod,
	},
	edgeKindMethod: {
		internal.Method_optionsTag: edgeKindOption,
	},
}

func extendOptionLocations(sc internal.SourceInfoMap, locs []*descriptor.SourceCodeInfo_Location) {
	// we iterate in the order that locations appear in descriptor
	// for determinism (if we ranged over the map, order and thus
	// potentially results are non-deterministic)
	for _, loc := range locs {
		allowed := edges[edgeKindFile]
		for i := 0; i+1 < len(loc.Path); i += 2 {
			nextKind, ok := allowed[loc.Path[i]]
			if !ok {
				break
			}
			if nextKind == edgeKindOption {
				// We've found an option entry. This could be arbitrarily
				// deep (for options that nested messages) or it could end
				// abruptly (for non-repeated fields). But we need a path
				// that is exactly the path-so-far plus two: the option tag
				// and an optional index for repeated option fields (zero
				// for non-repeated option fields). This is used for
				// querying source info when printing options.
				// for sorting elements
				newPath := make([]int32, i+3)
				copy(newPath, loc.Path)
				sc.PutIfAbsent(newPath, loc)
				// we do another path of path-so-far plus two, but with
				// explicit zero index -- just in case this actual path has
				// an extra path element, but it's not an index (e.g the
				// option field is not repeated, but the source info we are
				// looking at indicates a tag of a nested field)
				newPath[len(newPath)-1] = 0
				sc.PutIfAbsent(newPath, loc)
				// finally, we need the path-so-far plus one, just the option
				// tag, for sorting option groups
				newPath = newPath[:len(newPath)-1]
				sc.PutIfAbsent(newPath, loc)

				break
			} else {
				allowed = edges[nextKind]
			}
		}
	}
}

func (p *Printer) extractOptions(dsc desc.Descriptor, opts proto.Message, mf *dynamic.MessageFactory) (map[int32][]option, error) {
	md, err := desc.LoadMessageDescriptorForMessage(opts)
	if err != nil {
		return nil, err
	}
	dm := mf.NewDynamicMessage(md)
	if err = dm.ConvertFrom(opts); err != nil {
		return nil, fmt.Errorf("failed convert %s to dynamic message: %v", md.GetFullyQualifiedName(), err)
	}

	pkg := dsc.GetFile().GetPackage()
	var scope string
	if _, ok := dsc.(*desc.FileDescriptor); ok {
		scope = pkg
	} else {
		scope = dsc.GetFullyQualifiedName()
	}

	options := map[int32][]option{}
	var uninterpreted []interface{}
	for _, fldset := range [][]*desc.FieldDescriptor{md.GetFields(), mf.GetExtensionRegistry().AllExtensionsForType(md.GetFullyQualifiedName())} {
		for _, fld := range fldset {
			if dm.HasField(fld) {
				val := dm.GetField(fld)
				var opts []option
				var name string
				if fld.IsExtension() {
					name = fmt.Sprintf("(%s)", p.qualifyName(pkg, scope, fld.GetFullyQualifiedName()))
				} else {
					name = fld.GetName()
				}
				switch val := val.(type) {
				case []interface{}:
					if fld.GetNumber() == internal.UninterpretedOptionsTag {
						// we handle uninterpreted options differently
						uninterpreted = val
						continue
					}

					for _, e := range val {
						if fld.GetType() == descriptor.FieldDescriptorProto_TYPE_ENUM {
							ev := fld.GetEnumType().FindValueByNumber(e.(int32))
							if ev == nil {
								// have to skip unknown enum values :(
								continue
							}
							e = ev
						}
						opts = append(opts, option{name: name, val: e})
					}
				case map[interface{}]interface{}:
					for k := range sortKeys(val) {
						v := val[k]
						vf := fld.GetMapValueType()
						if vf.GetType() == descriptor.FieldDescriptorProto_TYPE_ENUM {
							ev := vf.GetEnumType().FindValueByNumber(v.(int32))
							if ev == nil {
								// have to skip unknown enum values :(
								continue
							}
							v = ev
						}
						entry := mf.NewDynamicMessage(fld.GetMessageType())
						entry.SetFieldByNumber(1, k)
						entry.SetFieldByNumber(2, v)
						opts = append(opts, option{name: name, val: entry})
					}
				default:
					if fld.GetType() == descriptor.FieldDescriptorProto_TYPE_ENUM {
						ev := fld.GetEnumType().FindValueByNumber(val.(int32))
						if ev == nil {
							// have to skip unknown enum values :(
							continue
						}
						val = ev
					}
					opts = append(opts, option{name: name, val: val})
				}
				if len(opts) > 0 {
					options[fld.GetNumber()] = opts
				}
			}
		}
	}

	// if there are uninterpreted options, add those too
	if len(uninterpreted) > 0 {
		opts := make([]option, len(uninterpreted))
		for i, u := range uninterpreted {
			var unint *descriptor.UninterpretedOption
			if un, ok := u.(*descriptor.UninterpretedOption); ok {
				unint = un
			} else {
				dm := u.(*dynamic.Message)
				unint = &descriptor.UninterpretedOption{}
				if err := dm.ConvertTo(unint); err != nil {
					return nil, err
				}
			}

			var buf bytes.Buffer
			for ni, n := range unint.Name {
				if ni > 0 {
					buf.WriteByte('.')
				}
				if n.GetIsExtension() {
					fmt.Fprintf(&buf, "(%s)", n.GetNamePart())
				} else {
					buf.WriteString(n.GetNamePart())
				}
			}

			var v interface{}
			switch {
			case unint.IdentifierValue != nil:
				v = ident(unint.GetIdentifierValue())
			case unint.StringValue != nil:
				v = string(unint.GetStringValue())
			case unint.DoubleValue != nil:
				v = unint.GetDoubleValue()
			case unint.PositiveIntValue != nil:
				v = unint.GetPositiveIntValue()
			case unint.NegativeIntValue != nil:
				v = unint.GetNegativeIntValue()
			case unint.AggregateValue != nil:
				v = ident(unint.GetAggregateValue())
			}

			opts[i] = option{name: buf.String(), val: v}
		}
		options[internal.UninterpretedOptionsTag] = opts
	}

	return options, nil
}

func optionsAsElementAddrs(optionsTag int32, order int, opts map[int32][]option) []elementAddr {
	var optAddrs []elementAddr
	for tag := range opts {
		optAddrs = append(optAddrs, elementAddr{elementType: optionsTag, elementIndex: int(tag), order: order})
	}
	return optAddrs
}

// quotedBytes implements the text format for string literals for protocol
// buffers. Since the underlying data is a bytes field, this encodes all
// bytes outside the 7-bit ASCII printable range. To preserve unicode strings
// without byte escapes, use quotedString.
func quotedBytes(s string) string {
	var b bytes.Buffer
	b.WriteByte('"')
	// Loop over the bytes, not the runes.
	for i := 0; i < len(s); i++ {
		// Divergence from C++: we don't escape apostrophes.
		// There's no need to escape them, and the C++ parser
		// copes with a naked apostrophe.
		switch c := s[i]; c {
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		case '"':
			b.WriteString("\\")
		case '\\':
			b.WriteString("\\\\")
		default:
			if c >= 0x20 && c < 0x7f {
				b.WriteByte(c)
			} else {
				fmt.Fprintf(&b, "\\%03o", c)
			}
		}
	}
	b.WriteByte('"')

	return b.String()
}

// quotedString implements the text format for string literals for protocol
// buffers. This form is also acceptable for string literals in option values
// by the protocol buffer compiler, protoc.
func quotedString(s string) string {
	var b bytes.Buffer
	b.WriteByte('"')
	// Loop over the bytes, not the runes.
	for {
		r, n := utf8.DecodeRuneInString(s)
		if n == 0 {
			break // end of string
		}
		if r == utf8.RuneError && n == 1 {
			// Invalid UTF8! Use an octal byte escape to encode the bad byte.
			fmt.Fprintf(&b, "\\%03o", s[0])
			s = s[1:]
			continue
		}

		// Divergence from C++: we don't escape apostrophes.
		// There's no need to escape them, and the C++ parser
		// copes with a naked apostrophe.
		switch r {
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		case '"':
			b.WriteString("\\")
		case '\\':
			b.WriteString("\\\\")
		default:
			if unicode.IsPrint(r) {
				b.WriteRune(r)
			} else {
				// if it's not printable, use a unicode escape
				if r > 0xffff {
					fmt.Fprintf(&b, "\\U%08X", r)
				} else if r > 0x7F {
					fmt.Fprintf(&b, "\\u%04X", r)
				} else {
					fmt.Fprintf(&b, "\\%03o", byte(r))
				}
			}
		}

		s = s[n:]
	}

	b.WriteByte('"')

	return b.String()
}

type elementAddr struct {
	elementType  int32
	elementIndex int
	order        int
}

type elementAddrs struct {
	addrs []elementAddr
	dsc   interface{}
	opts  map[int32][]option
}

func (a elementAddrs) Len() int {
	return len(a.addrs)
}

func (a elementAddrs) Less(i, j int) bool {
	// explicit order is considered first
	if a.addrs[i].order < a.addrs[j].order {
		return true
	} else if a.addrs[i].order > a.addrs[j].order {
		return false
	}
	// if order is equal, sort by element type
	if a.addrs[i].elementType < a.addrs[j].elementType {
		return true
	} else if a.addrs[i].elementType > a.addrs[j].elementType {
		return false
	}

	di := a.at(a.addrs[i])
	dj := a.at(a.addrs[j])

	switch vi := di.(type) {
	case *desc.FieldDescriptor:
		// fields are ordered by tag number
		vj := dj.(*desc.FieldDescriptor)
		// regular fields before extensions; extensions grouped by extendee
		if !vi.IsExtension() && vj.IsExtension() {
			return true
		} else if vi.IsExtension() && !vj.IsExtension() {
			return false
		} else if vi.IsExtension() && vj.IsExtension() {
			if vi.GetOwner() != vj.GetOwner() {
				return vi.GetOwner().GetFullyQualifiedName() < vj.GetOwner().GetFullyQualifiedName()
			}
		}
		return vi.GetNumber() < vj.GetNumber()

	case *desc.EnumValueDescriptor:
		// enum values ordered by number then name
		vj := dj.(*desc.EnumValueDescriptor)
		if vi.GetNumber() == vj.GetNumber() {
			return vi.GetName() < vj.GetName()
		}
		return vi.GetNumber() < vj.GetNumber()

	case *descriptor.DescriptorProto_ExtensionRange:
		// extension ranges ordered by tag
		return vi.GetStart() < dj.(*descriptor.DescriptorProto_ExtensionRange).GetStart()

	case reservedRange:
		// reserved ranges ordered by tag, too
		return vi.start < dj.(reservedRange).start

	case string:
		// reserved names lexically sorted
		return vi < dj.(string)

	case pkg:
		// reserved names lexically sorted
		return vi < dj.(pkg)

	case imp:
		// reserved names lexically sorted
		return vi < dj.(imp)

	case []option:
		// options sorted by name, extensions last
		return optionLess(vi, dj.([]option))

	default:
		// all other descriptors ordered by name
		return di.(desc.Descriptor).GetName() < dj.(desc.Descriptor).GetName()
	}
}

func (a elementAddrs) Swap(i, j int) {
	a.addrs[i], a.addrs[j] = a.addrs[j], a.addrs[i]
}

func (a elementAddrs) at(addr elementAddr) interface{} {
	switch dsc := a.dsc.(type) {
	case *desc.FileDescriptor:
		switch addr.elementType {
		case internal.File_packageTag:
			return pkg(dsc.GetPackage())
		case internal.File_dependencyTag:
			return imp(dsc.AsFileDescriptorProto().GetDependency()[addr.elementIndex])
		case internal.File_optionsTag:
			return a.opts[int32(addr.elementIndex)]
		case internal.File_messagesTag:
			return dsc.GetMessageTypes()[addr.elementIndex]
		case internal.File_enumsTag:
			return dsc.GetEnumTypes()[addr.elementIndex]
		case internal.File_servicesTag:
			return dsc.GetServices()[addr.elementIndex]
		case internal.File_extensionsTag:
			return dsc.GetExtensions()[addr.elementIndex]
		}
	case *desc.MessageDescriptor:
		switch addr.elementType {
		case internal.Message_optionsTag:
			return a.opts[int32(addr.elementIndex)]
		case internal.Message_fieldsTag:
			return dsc.GetFields()[addr.elementIndex]
		case internal.Message_nestedMessagesTag:
			return dsc.GetNestedMessageTypes()[addr.elementIndex]
		case internal.Message_enumsTag:
			return dsc.GetNestedEnumTypes()[addr.elementIndex]
		case internal.Message_extensionsTag:
			return dsc.GetNestedExtensions()[addr.elementIndex]
		case internal.Message_extensionRangeTag:
			return dsc.AsDescriptorProto().GetExtensionRange()[addr.elementIndex]
		case internal.Message_reservedRangeTag:
			rng := dsc.AsDescriptorProto().GetReservedRange()[addr.elementIndex]
			return reservedRange{start: rng.GetStart(), end: rng.GetEnd() - 1}
		case internal.Message_reservedNameTag:
			return dsc.AsDescriptorProto().GetReservedName()[addr.elementIndex]
		}
	case *desc.FieldDescriptor:
		if addr.elementType == internal.Field_optionsTag {
			return a.opts[int32(addr.elementIndex)]
		}
	case *desc.OneOfDescriptor:
		switch addr.elementType {
		case internal.OneOf_optionsTag:
			return a.opts[int32(addr.elementIndex)]
		case -internal.Message_fieldsTag:
			return dsc.GetOwner().GetFields()[addr.elementIndex]
		}
	case *desc.EnumDescriptor:
		switch addr.elementType {
		case internal.Enum_optionsTag:
			return a.opts[int32(addr.elementIndex)]
		case internal.Enum_valuesTag:
			return dsc.GetValues()[addr.elementIndex]
		case internal.Enum_reservedRangeTag:
			rng := dsc.AsEnumDescriptorProto().GetReservedRange()[addr.elementIndex]
			return reservedRange{start: rng.GetStart(), end: rng.GetEnd()}
		case internal.Enum_reservedNameTag:
			return dsc.AsEnumDescriptorProto().GetReservedName()[addr.elementIndex]
		}
	case *desc.EnumValueDescriptor:
		if addr.elementType == internal.EnumVal_optionsTag {
			return a.opts[int32(addr.elementIndex)]
		}
	case *desc.ServiceDescriptor:
		switch addr.elementType {
		case internal.Service_optionsTag:
			return a.opts[int32(addr.elementIndex)]
		case internal.Service_methodsTag:
			return dsc.GetMethods()[addr.elementIndex]
		}
	case *desc.MethodDescriptor:
		if addr.elementType == internal.Method_optionsTag {
			return a.opts[int32(addr.elementIndex)]
		}
	case extensionRange:
		if addr.elementType == internal.ExtensionRange_optionsTag {
			return a.opts[int32(addr.elementIndex)]
		}
	}

	panic(fmt.Sprintf("location for unknown field %d of %T", addr.elementType, a.dsc))
}

type extensionRange struct {
	owner    *desc.MessageDescriptor
	extRange *descriptor.DescriptorProto_ExtensionRange
}

type elementSrcOrder struct {
	elementAddrs
	sourceInfo internal.SourceInfoMap
	prefix     []int32
}

func (a elementSrcOrder) Less(i, j int) bool {
	ti := a.addrs[i].elementType
	ei := a.addrs[i].elementIndex

	tj := a.addrs[j].elementType
	ej := a.addrs[j].elementIndex

	var si, sj *descriptor.SourceCodeInfo_Location
	if ei < 0 {
		si = a.sourceInfo.Get(append(a.prefix, -int32(ei)))
	} else if ti < 0 {
		p := make([]int32, len(a.prefix)-2)
		copy(p, a.prefix)
		si = a.sourceInfo.Get(append(p, ti, int32(ei)))
	} else {
		si = a.sourceInfo.Get(append(a.prefix, ti, int32(ei)))
	}
	if ej < 0 {
		sj = a.sourceInfo.Get(append(a.prefix, -int32(ej)))
	} else if tj < 0 {
		p := make([]int32, len(a.prefix)-2)
		copy(p, a.prefix)
		sj = a.sourceInfo.Get(append(p, tj, int32(ej)))
	} else {
		sj = a.sourceInfo.Get(append(a.prefix, tj, int32(ej)))
	}

	if (si == nil) != (sj == nil) {
		// generally, we put unknown elements after known ones;
		// except package, imports, and option elements go first

		// i will be unknown and j will be known
		swapped := false
		if si != nil {
			ti, tj = tj, ti
			swapped = true
		}
		switch a.dsc.(type) {
		case *desc.FileDescriptor:
			// NB: These comparisons are *trying* to get things ordered so that
			// 1) If the package element has no source info, it appears _first_.
			// 2) If any import element has no source info, it appears _after_
			//    the package element but _before_ any other element.
			// 3) If any option element has no source info, it appears _after_
			//    the package and import elements but _before_ any other element.
			// If the package, imports, and options are all missing source info,
			// this will sort them all to the top in expected order. But if they
			// are mixed (some _do_ have source info, some do not), and elements
			// with source info have spans that positions them _after_ other
			// elements in the file, then this Less function will be unstable
			// since the above dual objectives for imports and options ("before
			// this but after that") may be in conflict with one another. This
			// should not cause any problems, other than elements being possibly
			// sorted in a confusing order.
			//
			// Well-formed descriptors should instead have consistent source
			// info: either all elements have source info or none do. So this
			// should not be an issue in practice.
			if ti == internal.File_packageTag {
				return !swapped
			}
			if ti == internal.File_dependencyTag {
				if tj == internal.File_packageTag {
					// imports will come *after* package
					return swapped
				}
				return !swapped
			}
			if ti == internal.File_optionsTag {
				if tj == internal.File_packageTag || tj == internal.File_dependencyTag {
					// options will come *after* package and imports
					return swapped
				}
				return !swapped
			}
		case *desc.MessageDescriptor:
			if ti == internal.Message_optionsTag {
				return !swapped
			}
		case *desc.EnumDescriptor:
			if ti == internal.Enum_optionsTag {
				return !swapped
			}
		case *desc.ServiceDescriptor:
			if ti == internal.Service_optionsTag {
				return !swapped
			}
		}
		return swapped

	} else if si == nil || sj == nil {
		// let stable sort keep unknown elements in same relative order
		return false
	}

	for idx := 0; idx < len(sj.Span); idx++ {
		if idx >= len(si.Span) {
			return true
		}
		if si.Span[idx] < sj.Span[idx] {
			return true
		}
		if si.Span[idx] > sj.Span[idx] {
			return false
		}
	}
	return false
}

type customSortOrder struct {
	elementAddrs
	less func(a, b Element) bool
}

func (cso customSortOrder) Less(i, j int) bool {
	ei := asElement(cso.at(cso.addrs[i]))
	ej := asElement(cso.at(cso.addrs[j]))
	return cso.less(ei, ej)
}

func optionLess(i, j []option) bool {
	ni := i[0].name
	nj := j[0].name
	if ni[0] != '(' && nj[0] == '(' {
		return true
	} else if ni[0] == '(' && nj[0] != '(' {
		return false
	}
	return ni < nj
}

func (p *Printer) printBlockElement(isDecriptor bool, si *descriptor.SourceCodeInfo_Location, w *writer, indent int, el func(w *writer, trailer func(indent int, wantTrailingNewline bool))) {
	includeComments := isDecriptor || p.includeCommentType(CommentsTokens)

	if includeComments && si != nil {
		p.printLeadingComments(si, w, indent)
	}
	el(w, func(indent int, wantTrailingNewline bool) {
		if includeComments && si != nil {
			if p.printTrailingComments(si, w, indent) && wantTrailingNewline && !p.Compact {
				// separator line between trailing comment and next element
				fmt.Fprintln(w)
			}
		}
	})
	if indent >= 0 && !w.newline {
		// if we're not printing inline but element did not have trailing newline, add one now
		fmt.Fprintln(w)
	}
}

func (p *Printer) printElement(isDecriptor bool, si *descriptor.SourceCodeInfo_Location, w *writer, indent int, el func(*writer)) {
	includeComments := isDecriptor || p.includeCommentType(CommentsTokens)

	if includeComments && si != nil {
		p.printLeadingComments(si, w, indent)
	}
	el(w)
	if includeComments && si != nil {
		p.printTrailingComments(si, w, indent)
	}
	if indent >= 0 && !w.newline {
		// if we're not printing inline but element did not have trailing newline, add one now
		fmt.Fprintln(w)
	}
}

func (p *Printer) printElementString(si *descriptor.SourceCodeInfo_Location, w *writer, indent int, str string) {
	p.printElement(false, si, w, inline(indent), func(w *writer) {
		fmt.Fprintf(w, "%s ", str)
	})
}

func (p *Printer) includeCommentType(c CommentType) bool {
	return (p.OmitComments & c) == 0
}

func (p *Printer) printLeadingComments(si *descriptor.SourceCodeInfo_Location, w *writer, indent int) bool {
	endsInNewLine := false

	if p.includeCommentType(CommentsDetached) {
		for _, c := range si.GetLeadingDetachedComments() {
			if p.printComment(c, w, indent, true) {
				// if comment ended in newline, add another newline to separate
				// this comment from the next
				p.newLine(w)
				endsInNewLine = true
			} else if indent < 0 {
				// comment did not end in newline and we are trying to inline?
				// just add a space to separate this comment from what follows
				fmt.Fprint(w, " ")
				endsInNewLine = false
			} else {
				// comment did not end in newline and we are *not* trying to inline?
				// add newline to end of comment and add another to separate this
				// comment from what follows
				fmt.Fprintln(w) // needed to end comment, regardless of p.Compact
				p.newLine(w)
				endsInNewLine = true
			}
		}
	}

	if p.includeCommentType(CommentsLeading) && si.GetLeadingComments() != "" {
		endsInNewLine = p.printComment(si.GetLeadingComments(), w, indent, true)
		if !endsInNewLine {
			if indent >= 0 {
				// leading comment didn't end with newline but needs one
				// (because we're *not* inlining)
				fmt.Fprintln(w) // needed to end comment, regardless of p.Compact
				endsInNewLine = true
			} else {
				// space between comment and following element when inlined
				fmt.Fprint(w, " ")
			}
		}
	}

	return endsInNewLine
}

func (p *Printer) printTrailingComments(si *descriptor.SourceCodeInfo_Location, w *writer, indent int) bool {
	if p.includeCommentType(CommentsTrailing) && si.GetTrailingComments() != "" {
		if !p.printComment(si.GetTrailingComments(), w, indent, p.TrailingCommentsOnSeparateLine) && indent >= 0 {
			// trailing comment didn't end with newline but needs one
			// (because we're *not* inlining)
			fmt.Fprintln(w) // needed to end comment, regardless of p.Compact
		} else if indent < 0 {
			fmt.Fprint(w, " ")
		}
		return true
	}

	return false
}

func (p *Printer) printComment(comments string, w *writer, indent int, forceNextLine bool) bool {
	if comments == "" {
		return false
	}

	var multiLine bool
	if indent < 0 {
		// use multi-line style when inlining
		multiLine = true
	} else {
		multiLine = p.PreferMultiLineStyleComments
	}
	if multiLine && strings.Contains(comments, "*/") {
		// can't emit '*/' in a multi-line style comment
		multiLine = false
	}

	lines := strings.Split(comments, "\n")

	// first, remove leading and trailing blank lines
	if lines[0] == "" {
		lines = lines[1:]
	}
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return false
	}

	if indent >= 0 && !w.newline {
		// last element did not have trailing newline, so we
		// either need to tack on newline or, if comment is
		// just one line, inline it on the end
		if forceNextLine || len(lines) > 1 {
			fmt.Fprintln(w)
		} else {
			if !w.space {
				fmt.Fprint(w, " ")
			}
			indent = inline(indent)
		}
	}

	if len(lines) == 1 && multiLine {
		p.indent(w, indent)
		line := lines[0]
		if line[0] == ' ' && line[len(line)-1] != ' ' {
			// add trailing space for symmetry
			line += " "
		}
		fmt.Fprintf(w, "/*%s*/", line)
		if indent >= 0 {
			fmt.Fprintln(w)
			return true
		}
		return false
	}

	if multiLine {
		// multi-line style comments that actually span multiple lines
		// get a blank line before and after so that comment renders nicely
		lines = append(lines, "", "")
		copy(lines[1:], lines)
		lines[0] = ""
	}

	for i, l := range lines {
		p.maybeIndent(w, indent, i > 0)
		if multiLine {
			if i == 0 {
				// first line
				fmt.Fprintf(w, "/*%s\n", strings.TrimRight(l, " \t"))
			} else if i == len(lines)-1 {
				// last line
				if l == "" {
					fmt.Fprint(w, " */")
				} else {
					fmt.Fprintf(w, " *%s*/", l)
				}
				if indent >= 0 {
					fmt.Fprintln(w)
				}
			} else {
				fmt.Fprintf(w, " *%s\n", strings.TrimRight(l, " \t"))
			}
		} else {
			fmt.Fprintf(w, "//%s\n", strings.TrimRight(l, " \t"))
		}
	}

	// single-line comments always end in newline; multi-line comments only
	// end in newline for non-negative (e.g. non-inlined) indentation
	return !multiLine || indent >= 0
}

func (p *Printer) indent(w io.Writer, indent int) {
	for i := 0; i < indent; i++ {
		fmt.Fprint(w, p.Indent)
	}
}

func (p *Printer) maybeIndent(w io.Writer, indent int, requireIndent bool) {
	if indent < 0 && requireIndent {
		p.indent(w, -indent)
	} else {
		p.indent(w, indent)
	}
}

type writer struct {
	io.Writer
	err     error
	space   bool
	newline bool
}

func newWriter(w io.Writer) *writer {
	return &writer{Writer: w, newline: true}
}

func (w *writer) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	w.newline = false

	if w.space {
		// skip any trailing space if the following
		// character is semicolon, comma, or close bracket
		if p[0] != ';' && p[0] != ',' {
			_, err := w.Writer.Write([]byte{' '})
			if err != nil {
				w.err = err
				return 0, err
			}
		}
		w.space = false
	}

	if p[len(p)-1] == ' ' {
		w.space = true
		p = p[:len(p)-1]
	}
	if len(p) > 0 && p[len(p)-1] == '\n' {
		w.newline = true
	}

	num, err := w.Writer.Write(p)
	if err != nil {
		w.err = err
	} else if w.space {
		// pretend space was written
		num++
	}
	return num, err
}
