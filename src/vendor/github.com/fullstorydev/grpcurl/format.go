package grpcurl

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/golang/protobuf/jsonpb" //lint:ignore SA1019 we have to import this because it appears in exported API
	"github.com/golang/protobuf/proto"  //lint:ignore SA1019 we have to import this because it appears in exported API
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// RequestParser processes input into messages.
type RequestParser interface {
	// Next parses input data into the given request message. If called after
	// input is exhausted, it returns io.EOF. If the caller re-uses the same
	// instance in multiple calls to Next, it should call msg.Reset() in between
	// each call.
	Next(msg proto.Message) error
	// NumRequests returns the number of messages that have been parsed and
	// returned by a call to Next.
	NumRequests() int
}

type jsonRequestParser struct {
	dec          *json.Decoder
	unmarshaler  jsonpb.Unmarshaler
	requestCount int
}

// NewJSONRequestParser returns a RequestParser that reads data in JSON format
// from the given reader. The given resolver is used to assist with decoding of
// google.protobuf.Any messages.
//
// Input data that contains more than one message should just include all
// messages concatenated (though whitespace is necessary to separate some kinds
// of values in JSON).
//
// If the given reader has no data, the returned parser will return io.EOF on
// the very first call.
func NewJSONRequestParser(in io.Reader, resolver jsonpb.AnyResolver) RequestParser {
	return &jsonRequestParser{
		dec:         json.NewDecoder(in),
		unmarshaler: jsonpb.Unmarshaler{AnyResolver: resolver},
	}
}

// NewJSONRequestParserWithUnmarshaler is like NewJSONRequestParser but
// accepts a protobuf jsonpb.Unmarshaler instead of jsonpb.AnyResolver.
func NewJSONRequestParserWithUnmarshaler(in io.Reader, unmarshaler jsonpb.Unmarshaler) RequestParser {
	return &jsonRequestParser{
		dec:         json.NewDecoder(in),
		unmarshaler: unmarshaler,
	}
}

func (f *jsonRequestParser) Next(m proto.Message) error {
	var msg json.RawMessage
	if err := f.dec.Decode(&msg); err != nil {
		return err
	}
	f.requestCount++
	return f.unmarshaler.Unmarshal(bytes.NewReader(msg), m)
}

func (f *jsonRequestParser) NumRequests() int {
	return f.requestCount
}

const (
	textSeparatorChar = '\x1e'
)

type textRequestParser struct {
	r            *bufio.Reader
	err          error
	requestCount int
}

// NewTextRequestParser returns a RequestParser that reads data in the protobuf
// text format from the given reader.
//
// Input data that contains more than one message should include an ASCII
// 'Record Separator' character (0x1E) between each message.
//
// Empty text is a valid text format and represents an empty message. So if the
// given reader has no data, the returned parser will yield an empty message
// for the first call to Next and then return io.EOF thereafter. This also means
// that if the input data ends with a record separator, then a final empty
// message will be parsed *after* the separator.
func NewTextRequestParser(in io.Reader) RequestParser {
	return &textRequestParser{r: bufio.NewReader(in)}
}

func (f *textRequestParser) Next(m proto.Message) error {
	if f.err != nil {
		return f.err
	}

	var b []byte
	b, f.err = f.r.ReadBytes(textSeparatorChar)
	if f.err != nil && f.err != io.EOF {
		return f.err
	}
	// remove delimiter
	if len(b) > 0 && b[len(b)-1] == textSeparatorChar {
		b = b[:len(b)-1]
	}

	f.requestCount++

	return proto.UnmarshalText(string(b), m)
}

func (f *textRequestParser) NumRequests() int {
	return f.requestCount
}

// Formatter translates messages into string representations.
type Formatter func(proto.Message) (string, error)

// NewJSONFormatter returns a formatter that returns JSON strings. The JSON will
// include empty/default values (instead of just omitted them) if emitDefaults
// is true. The given resolver is used to assist with encoding of
// google.protobuf.Any messages.
func NewJSONFormatter(emitDefaults bool, resolver jsonpb.AnyResolver) Formatter {
	marshaler := jsonpb.Marshaler{
		EmitDefaults: emitDefaults,
		Indent:       "  ",
		AnyResolver:  resolver,
	}
	return marshaler.MarshalToString
}

// NewTextFormatter returns a formatter that returns strings in the protobuf
// text format. If includeSeparator is true then, when invoked to format
// multiple messages, all messages after the first one will be prefixed with the
// ASCII 'Record Separator' character (0x1E).
func NewTextFormatter(includeSeparator bool) Formatter {
	tf := textFormatter{useSeparator: includeSeparator}
	return tf.format
}

type textFormatter struct {
	useSeparator bool
	numFormatted int
}

var protoTextMarshaler = proto.TextMarshaler{ExpandAny: true}

func (tf *textFormatter) format(m proto.Message) (string, error) {
	var buf bytes.Buffer
	if tf.useSeparator && tf.numFormatted > 0 {
		if err := buf.WriteByte(textSeparatorChar); err != nil {
			return "", err
		}
	}

	// If message implements MarshalText method (such as a *dynamic.Message),
	// it won't get details about whether or not to format to text compactly
	// or with indentation. So first see if the message also implements a
	// MarshalTextIndent method and use that instead if available.
	type indentMarshaler interface {
		MarshalTextIndent() ([]byte, error)
	}

	if indenter, ok := m.(indentMarshaler); ok {
		b, err := indenter.MarshalTextIndent()
		if err != nil {
			return "", err
		}
		if _, err := buf.Write(b); err != nil {
			return "", err
		}
	} else if err := protoTextMarshaler.Marshal(&buf, m); err != nil {
		return "", err
	}

	// no trailing newline needed
	str := buf.String()
	if len(str) > 0 && str[len(str)-1] == '\n' {
		str = str[:len(str)-1]
	}

	tf.numFormatted++

	return str, nil
}

type Format string

const (
	FormatJSON = Format("json")
	FormatText = Format("text")
)

// AnyResolverFromDescriptorSource returns an AnyResolver that will search for
// types using the given descriptor source.
func AnyResolverFromDescriptorSource(source DescriptorSource) jsonpb.AnyResolver {
	return &anyResolver{source: source}
}

// AnyResolverFromDescriptorSourceWithFallback returns an AnyResolver that will
// search for types using the given descriptor source and then fallback to a
// special message if the type is not found. The fallback type will render to
// JSON with a "@type" property, just like an Any message, but also with a
// custom "@value" property that includes the binary encoded payload.
func AnyResolverFromDescriptorSourceWithFallback(source DescriptorSource) jsonpb.AnyResolver {
	res := anyResolver{source: source}
	return &anyResolverWithFallback{AnyResolver: &res}
}

type anyResolver struct {
	source DescriptorSource

	er dynamic.ExtensionRegistry

	mu       sync.RWMutex
	mf       *dynamic.MessageFactory
	resolved map[string]func() proto.Message
}

func (r *anyResolver) Resolve(typeUrl string) (proto.Message, error) {
	mname := typeUrl
	if slash := strings.LastIndex(mname, "/"); slash >= 0 {
		mname = mname[slash+1:]
	}

	r.mu.RLock()
	factory := r.resolved[mname]
	r.mu.RUnlock()

	// already resolved?
	if factory != nil {
		return factory(), nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// double-check, in case we were racing with another goroutine
	// that resolved this one
	factory = r.resolved[mname]
	if factory != nil {
		return factory(), nil
	}

	// use descriptor source to resolve message type
	d, err := r.source.FindSymbol(mname)
	if err != nil {
		return nil, err
	}
	md, ok := d.(*desc.MessageDescriptor)
	if !ok {
		return nil, fmt.Errorf("unknown message: %s", typeUrl)
	}
	// populate any extensions for this message, too
	if exts, err := r.source.AllExtensionsForType(mname); err != nil {
		return nil, err
	} else if err := r.er.AddExtension(exts...); err != nil {
		return nil, err
	}

	if r.mf == nil {
		r.mf = dynamic.NewMessageFactoryWithExtensionRegistry(&r.er)
	}

	factory = func() proto.Message {
		return r.mf.NewMessage(md)
	}
	if r.resolved == nil {
		r.resolved = map[string]func() proto.Message{}
	}
	r.resolved[mname] = factory
	return factory(), nil
}

// anyResolverWithFallback can provide a fallback value for unknown
// messages that will format itself to JSON using an "@value" field
// that has the base64-encoded data for the unknown message value.
type anyResolverWithFallback struct {
	jsonpb.AnyResolver
}

func (r anyResolverWithFallback) Resolve(typeUrl string) (proto.Message, error) {
	msg, err := r.AnyResolver.Resolve(typeUrl)
	if err == nil {
		return msg, err
	}

	// Try "default" resolution logic. This mirrors the default behavior
	// of jsonpb, which checks to see if the given message name is registered
	// in the proto package.
	mname := typeUrl
	if slash := strings.LastIndex(mname, "/"); slash >= 0 {
		mname = mname[slash+1:]
	}
	//lint:ignore SA1019 new non-deprecated API requires other code changes; deferring...
	mt := proto.MessageType(mname)
	if mt != nil {
		return reflect.New(mt.Elem()).Interface().(proto.Message), nil
	}

	// finally, fallback to a special placeholder that can marshal itself
	// to JSON using a special "@value" property to show base64-encoded
	// data for the embedded message
	return &unknownAny{TypeUrl: typeUrl, Error: fmt.Sprintf("%s is not recognized; see @value for raw binary message data", mname)}, nil
}

type unknownAny struct {
	TypeUrl string `json:"@type"`
	Error   string `json:"@error"`
	Value   string `json:"@value"`
}

func (a *unknownAny) MarshalJSONPB(jsm *jsonpb.Marshaler) ([]byte, error) {
	if jsm.Indent != "" {
		return json.MarshalIndent(a, "", jsm.Indent)
	}
	return json.Marshal(a)
}

func (a *unknownAny) Unmarshal(b []byte) error {
	a.Value = base64.StdEncoding.EncodeToString(b)
	return nil
}

func (a *unknownAny) Reset() {
	a.Value = ""
}

func (a *unknownAny) String() string {
	b, err := a.MarshalJSONPB(&jsonpb.Marshaler{})
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err.Error())
	}
	return string(b)
}

func (a *unknownAny) ProtoMessage() {
}

var _ proto.Message = (*unknownAny)(nil)

// FormatOptions is a set of flags that are passed to a JSON or text formatter.
type FormatOptions struct {
	// EmitJSONDefaultFields flag, when true, includes empty/default values in the output.
	// FormatJSON only flag.
	EmitJSONDefaultFields bool

	// AllowUnknownFields is an option for the parser. When true,
	// it accepts input which includes unknown fields. These unknown fields
	// are skipped instead of returning an error.
	// FormatJSON only flag.
	AllowUnknownFields bool

	// IncludeTextSeparator is true then, when invoked to format multiple messages,
	// all messages after the first one will be prefixed with the
	// ASCII 'Record Separator' character (0x1E).
	// It might be useful when the output is piped to another grpcurl process.
	// FormatText only flag.
	IncludeTextSeparator bool
}

// RequestParserAndFormatter returns a request parser and formatter for the
// given format. The given descriptor source may be used for parsing message
// data (if needed by the format).
// It accepts a set of options. The field EmitJSONDefaultFields and IncludeTextSeparator
// are options for JSON and protobuf text formats, respectively. The AllowUnknownFields field
// is a JSON-only format flag.
// Requests will be parsed from the given in.
func RequestParserAndFormatter(format Format, descSource DescriptorSource, in io.Reader, opts FormatOptions) (RequestParser, Formatter, error) {
	switch format {
	case FormatJSON:
		resolver := AnyResolverFromDescriptorSource(descSource)
		unmarshaler := jsonpb.Unmarshaler{AnyResolver: resolver, AllowUnknownFields: opts.AllowUnknownFields}
		return NewJSONRequestParserWithUnmarshaler(in, unmarshaler), NewJSONFormatter(opts.EmitJSONDefaultFields, anyResolverWithFallback{AnyResolver: resolver}), nil
	case FormatText:
		return NewTextRequestParser(in), NewTextFormatter(opts.IncludeTextSeparator), nil
	default:
		return nil, nil, fmt.Errorf("unknown format: %s", format)
	}
}

// RequestParserAndFormatterFor returns a request parser and formatter for the
// given format. The given descriptor source may be used for parsing message
// data (if needed by the format). The flags emitJSONDefaultFields and
// includeTextSeparator are options for JSON and protobuf text formats,
// respectively. Requests will be parsed from the given in.
// This function is deprecated. Please use RequestParserAndFormatter instead.
// DEPRECATED
func RequestParserAndFormatterFor(format Format, descSource DescriptorSource, emitJSONDefaultFields, includeTextSeparator bool, in io.Reader) (RequestParser, Formatter, error) {
	return RequestParserAndFormatter(format, descSource, in, FormatOptions{
		EmitJSONDefaultFields: emitJSONDefaultFields,
		IncludeTextSeparator:  includeTextSeparator,
	})
}

// DefaultEventHandler logs events to a writer. This is not thread-safe, but is
// safe for use with InvokeRPC as long as NumResponses and Status are not read
// until the call to InvokeRPC completes.
type DefaultEventHandler struct {
	Out       io.Writer
	Formatter Formatter
	// 0 = default
	// 1 = verbose
	// 2 = very verbose
	VerbosityLevel int

	// NumResponses is the number of responses that have been received.
	NumResponses int
	// Status is the status that was received at the end of an RPC. It is
	// nil if the RPC is still in progress.
	Status *status.Status
}

// NewDefaultEventHandler returns an InvocationEventHandler that logs events to
// the given output. If verbose is true, all events are logged. Otherwise, only
// response messages are logged.
//
// Deprecated: NewDefaultEventHandler exists for compatibility.
// It doesn't allow fine control over the `VerbosityLevel`
// and provides only 0 and 1 options (which corresponds to the `verbose` argument).
// Use DefaultEventHandler{} initializer directly.
func NewDefaultEventHandler(out io.Writer, descSource DescriptorSource, formatter Formatter, verbose bool) *DefaultEventHandler {
	verbosityLevel := 0
	if verbose {
		verbosityLevel = 1
	}
	return &DefaultEventHandler{
		Out:            out,
		Formatter:      formatter,
		VerbosityLevel: verbosityLevel,
	}
}

var _ InvocationEventHandler = (*DefaultEventHandler)(nil)

func (h *DefaultEventHandler) OnResolveMethod(md *desc.MethodDescriptor) {
	if h.VerbosityLevel > 0 {
		txt, err := GetDescriptorText(md, nil)
		if err == nil {
			fmt.Fprintf(h.Out, "\nResolved method descriptor:\n%s\n", txt)
		}
	}
}

func (h *DefaultEventHandler) OnSendHeaders(md metadata.MD) {
	if h.VerbosityLevel > 0 {
		fmt.Fprintf(h.Out, "\nRequest metadata to send:\n%s\n", MetadataToString(md))
	}
}

func (h *DefaultEventHandler) OnReceiveHeaders(md metadata.MD) {
	if h.VerbosityLevel > 0 {
		fmt.Fprintf(h.Out, "\nResponse headers received:\n%s\n", MetadataToString(md))
	}
}

func (h *DefaultEventHandler) OnReceiveResponse(resp proto.Message) {
	h.NumResponses++
	if h.VerbosityLevel > 1 {
		fmt.Fprintf(h.Out, "\nEstimated response size: %d bytes\n", proto.Size(resp))
	}
	if h.VerbosityLevel > 0 {
		fmt.Fprint(h.Out, "\nResponse contents:\n")
	}
	if respStr, err := h.Formatter(resp); err != nil {
		fmt.Fprintf(h.Out, "Failed to format response message %d: %v\n", h.NumResponses, err)
	} else {
		fmt.Fprintln(h.Out, respStr)
	}
}

func (h *DefaultEventHandler) OnReceiveTrailers(stat *status.Status, md metadata.MD) {
	h.Status = stat
	if h.VerbosityLevel > 0 {
		fmt.Fprintf(h.Out, "\nResponse trailers received:\n%s\n", MetadataToString(md))
	}
}

// PrintStatus prints details about the given status to the given writer. The given
// formatter is used to print any detail messages that may be included in the status.
// If the given status has a code of OK, "OK" is printed and that is all. Otherwise,
// "ERROR:" is printed along with a line showing the code, one showing the message
// string, and each detail message if any are present. The detail messages will be
// printed as proto text format or JSON, depending on the given formatter.
func PrintStatus(w io.Writer, stat *status.Status, formatter Formatter) {
	if stat.Code() == codes.OK {
		fmt.Fprintln(w, "OK")
		return
	}
	fmt.Fprintf(w, "ERROR:\n  Code: %s\n  Message: %s\n", stat.Code().String(), stat.Message())

	statpb := stat.Proto()
	if len(statpb.Details) > 0 {
		fmt.Fprintf(w, "  Details:\n")
		for i, det := range statpb.Details {
			prefix := fmt.Sprintf("  %d)", i+1)
			fmt.Fprintf(w, "%s\t", prefix)
			prefix = strings.Repeat(" ", len(prefix)) + "\t"

			output, err := formatter(det)
			if err != nil {
				fmt.Fprintf(w, "Error parsing detail message: %v\n", err)
			} else {
				lines := strings.Split(output, "\n")
				for i, line := range lines {
					if i == 0 {
						// first line is already indented
						fmt.Fprintf(w, "%s\n", line)
					} else {
						fmt.Fprintf(w, "%s%s\n", prefix, line)
					}
				}
			}
		}
	}
}
