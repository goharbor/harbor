// Package grpcurl provides the core functionality exposed by the grpcurl command, for
// dynamically connecting to a server, using the reflection service to inspect the server,
// and invoking RPCs. The grpcurl command-line tool constructs a DescriptorSource, based
// on the command-line parameters, and supplies an InvocationEventHandler to supply request
// data (which can come from command-line args or the process's stdin) and to log the
// events (to the process's stdout).
package grpcurl

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/golang/protobuf/proto" //lint:ignore SA1019 we have to import this because it appears in exported API
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoprint"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

// ListServices uses the given descriptor source to return a sorted list of fully-qualified
// service names.
func ListServices(source DescriptorSource) ([]string, error) {
	svcs, err := source.ListServices()
	if err != nil {
		return nil, err
	}
	sort.Strings(svcs)
	return svcs, nil
}

type sourceWithFiles interface {
	GetAllFiles() ([]*desc.FileDescriptor, error)
}

var _ sourceWithFiles = (*fileSource)(nil)

// GetAllFiles uses the given descriptor source to return a list of file descriptors.
func GetAllFiles(source DescriptorSource) ([]*desc.FileDescriptor, error) {
	var files []*desc.FileDescriptor
	srcFiles, ok := source.(sourceWithFiles)

	// If an error occurs, we still try to load as many files as we can, so that
	// caller can decide whether to ignore error or not.
	var firstError error
	if ok {
		files, firstError = srcFiles.GetAllFiles()
	} else {
		// Source does not implement GetAllFiles method, so use ListServices
		// and grab files from there.
		svcNames, err := source.ListServices()
		if err != nil {
			firstError = err
		} else {
			allFiles := map[string]*desc.FileDescriptor{}
			for _, name := range svcNames {
				d, err := source.FindSymbol(name)
				if err != nil {
					if firstError == nil {
						firstError = err
					}
				} else {
					addAllFilesToSet(d.GetFile(), allFiles)
				}
			}
			files = make([]*desc.FileDescriptor, len(allFiles))
			i := 0
			for _, fd := range allFiles {
				files[i] = fd
				i++
			}
		}
	}

	sort.Sort(filesByName(files))
	return files, firstError
}

type filesByName []*desc.FileDescriptor

func (f filesByName) Len() int {
	return len(f)
}

func (f filesByName) Less(i, j int) bool {
	return f[i].GetName() < f[j].GetName()
}

func (f filesByName) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func addAllFilesToSet(fd *desc.FileDescriptor, all map[string]*desc.FileDescriptor) {
	if _, ok := all[fd.GetName()]; ok {
		// already added
		return
	}
	all[fd.GetName()] = fd
	for _, dep := range fd.GetDependencies() {
		addAllFilesToSet(dep, all)
	}
}

// ListMethods uses the given descriptor source to return a sorted list of method names
// for the specified fully-qualified service name.
func ListMethods(source DescriptorSource, serviceName string) ([]string, error) {
	dsc, err := source.FindSymbol(serviceName)
	if err != nil {
		return nil, err
	}
	if sd, ok := dsc.(*desc.ServiceDescriptor); !ok {
		return nil, notFound("Service", serviceName)
	} else {
		methods := make([]string, 0, len(sd.GetMethods()))
		for _, method := range sd.GetMethods() {
			methods = append(methods, method.GetFullyQualifiedName())
		}
		sort.Strings(methods)
		return methods, nil
	}
}

// MetadataFromHeaders converts a list of header strings (each string in
// "Header-Name: Header-Value" form) into metadata. If a string has a header
// name without a value (e.g. does not contain a colon), the value is assumed
// to be blank. Binary headers (those whose names end in "-bin") should be
// base64-encoded. But if they cannot be base64-decoded, they will be assumed to
// be in raw form and used as is.
func MetadataFromHeaders(headers []string) metadata.MD {
	md := make(metadata.MD)
	for _, part := range headers {
		if part != "" {
			pieces := strings.SplitN(part, ":", 2)
			if len(pieces) == 1 {
				pieces = append(pieces, "") // if no value was specified, just make it "" (maybe the header value doesn't matter)
			}
			headerName := strings.ToLower(strings.TrimSpace(pieces[0]))
			val := strings.TrimSpace(pieces[1])
			if strings.HasSuffix(headerName, "-bin") {
				if v, err := decode(val); err == nil {
					val = v
				}
			}
			md[headerName] = append(md[headerName], val)
		}
	}
	return md
}

var envVarRegex = regexp.MustCompile(`\${\w+}`)

// ExpandHeaders expands environment variables contained in the header string.
// If no corresponding environment variable is found an error is returned.
// TODO: Add escaping for `${`
func ExpandHeaders(headers []string) ([]string, error) {
	expandedHeaders := make([]string, len(headers))
	for idx, header := range headers {
		if header == "" {
			continue
		}
		results := envVarRegex.FindAllString(header, -1)
		if len(results) == 0 {
			expandedHeaders[idx] = headers[idx]
			continue
		}
		expandedHeader := header
		for _, result := range results {
			envVarName := result[2 : len(result)-1] // strip leading `${` and trailing `}`
			envVarValue, ok := os.LookupEnv(envVarName)
			if !ok {
				return nil, fmt.Errorf("header %q refers to missing environment variable %q", header, envVarName)
			}
			expandedHeader = strings.Replace(expandedHeader, result, envVarValue, -1)
		}
		expandedHeaders[idx] = expandedHeader
	}
	return expandedHeaders, nil
}

var base64Codecs = []*base64.Encoding{base64.StdEncoding, base64.URLEncoding, base64.RawStdEncoding, base64.RawURLEncoding}

func decode(val string) (string, error) {
	var firstErr error
	var b []byte
	// we are lenient and can accept any of the flavors of base64 encoding
	for _, d := range base64Codecs {
		var err error
		b, err = d.DecodeString(val)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		return string(b), nil
	}
	return "", firstErr
}

// MetadataToString returns a string representation of the given metadata, for
// displaying to users.
func MetadataToString(md metadata.MD) string {
	if len(md) == 0 {
		return "(empty)"
	}

	keys := make([]string, 0, len(md))
	for k := range md {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b bytes.Buffer
	first := true
	for _, k := range keys {
		vs := md[k]
		for _, v := range vs {
			if first {
				first = false
			} else {
				b.WriteString("\n")
			}
			b.WriteString(k)
			b.WriteString(": ")
			if strings.HasSuffix(k, "-bin") {
				v = base64.StdEncoding.EncodeToString([]byte(v))
			}
			b.WriteString(v)
		}
	}
	return b.String()
}

var printer = &protoprint.Printer{
	Compact:                  true,
	OmitComments:             protoprint.CommentsNonDoc,
	SortElements:             true,
	ForceFullyQualifiedNames: true,
}

// GetDescriptorText returns a string representation of the given descriptor.
// This returns a snippet of proto source that describes the given element.
func GetDescriptorText(dsc desc.Descriptor, _ DescriptorSource) (string, error) {
	// Note: DescriptorSource is not used, but remains an argument for backwards
	// compatibility with previous implementation.
	txt, err := printer.PrintProtoToString(dsc)
	if err != nil {
		return "", err
	}
	// callers don't expect trailing newlines
	if txt[len(txt)-1] == '\n' {
		txt = txt[:len(txt)-1]
	}
	return txt, nil
}

// EnsureExtensions uses the given descriptor source to download extensions for
// the given message. It returns a copy of the given message, but as a dynamic
// message that knows about all extensions known to the given descriptor source.
func EnsureExtensions(source DescriptorSource, msg proto.Message) proto.Message {
	// load any server extensions so we can properly describe custom options
	dsc, err := desc.LoadMessageDescriptorForMessage(msg)
	if err != nil {
		return msg
	}

	var ext dynamic.ExtensionRegistry
	if err = fetchAllExtensions(source, &ext, dsc, map[string]bool{}); err != nil {
		return msg
	}

	// convert message into dynamic message that knows about applicable extensions
	// (that way we can show meaningful info for custom options instead of printing as unknown)
	msgFactory := dynamic.NewMessageFactoryWithExtensionRegistry(&ext)
	dm, err := fullyConvertToDynamic(msgFactory, msg)
	if err != nil {
		return msg
	}
	return dm
}

// fetchAllExtensions recursively fetches from the server extensions for the given message type as well as
// for all message types of nested fields. The extensions are added to the given dynamic registry of extensions
// so that all server-known extensions can be correctly parsed by grpcurl.
func fetchAllExtensions(source DescriptorSource, ext *dynamic.ExtensionRegistry, md *desc.MessageDescriptor, alreadyFetched map[string]bool) error {
	msgTypeName := md.GetFullyQualifiedName()
	if alreadyFetched[msgTypeName] {
		return nil
	}
	alreadyFetched[msgTypeName] = true
	if len(md.GetExtensionRanges()) > 0 {
		fds, err := source.AllExtensionsForType(msgTypeName)
		if err != nil {
			return fmt.Errorf("failed to query for extensions of type %s: %v", msgTypeName, err)
		}
		for _, fd := range fds {
			if err := ext.AddExtension(fd); err != nil {
				return fmt.Errorf("could not register extension %s of type %s: %v", fd.GetFullyQualifiedName(), msgTypeName, err)
			}
		}
	}
	// recursively fetch extensions for the types of any message fields
	for _, fd := range md.GetFields() {
		if fd.GetMessageType() != nil {
			err := fetchAllExtensions(source, ext, fd.GetMessageType(), alreadyFetched)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// fullConvertToDynamic attempts to convert the given message to a dynamic message as well
// as any nested messages it may contain as field values. If the given message factory has
// extensions registered that were not known when the given message was parsed, this effectively
// allows re-parsing to identify those extensions.
func fullyConvertToDynamic(msgFact *dynamic.MessageFactory, msg proto.Message) (proto.Message, error) {
	if _, ok := msg.(*dynamic.Message); ok {
		return msg, nil // already a dynamic message
	}
	md, err := desc.LoadMessageDescriptorForMessage(msg)
	if err != nil {
		return nil, err
	}
	newMsg := msgFact.NewMessage(md)
	dm, ok := newMsg.(*dynamic.Message)
	if !ok {
		// if message factory didn't produce a dynamic message, then we should leave msg as is
		return msg, nil
	}

	if err := dm.ConvertFrom(msg); err != nil {
		return nil, err
	}

	// recursively convert all field values, too
	for _, fd := range md.GetFields() {
		if fd.IsMap() {
			if fd.GetMapValueType().GetMessageType() != nil {
				m := dm.GetField(fd).(map[interface{}]interface{})
				for k, v := range m {
					// keys can't be nested messages; so we only need to recurse through map values, not keys
					newVal, err := fullyConvertToDynamic(msgFact, v.(proto.Message))
					if err != nil {
						return nil, err
					}
					dm.PutMapField(fd, k, newVal)
				}
			}
		} else if fd.IsRepeated() {
			if fd.GetMessageType() != nil {
				s := dm.GetField(fd).([]interface{})
				for i, e := range s {
					newVal, err := fullyConvertToDynamic(msgFact, e.(proto.Message))
					if err != nil {
						return nil, err
					}
					dm.SetRepeatedField(fd, i, newVal)
				}
			}
		} else {
			if fd.GetMessageType() != nil {
				v := dm.GetField(fd)
				newVal, err := fullyConvertToDynamic(msgFact, v.(proto.Message))
				if err != nil {
					return nil, err
				}
				dm.SetField(fd, newVal)
			}
		}
	}
	return dm, nil
}

// MakeTemplate returns a message instance for the given descriptor that is a
// suitable template for creating an instance of that message in JSON. In
// particular, it ensures that any repeated fields (which include map fields)
// are not empty, so they will render with a single element (to show the types
// and optionally nested fields). It also ensures that nested messages are not
// nil by setting them to a message that is also fleshed out as a template
// message.
func MakeTemplate(md *desc.MessageDescriptor) proto.Message {
	return makeTemplate(md, nil)
}

func makeTemplate(md *desc.MessageDescriptor, path []*desc.MessageDescriptor) proto.Message {
	switch md.GetFullyQualifiedName() {
	case "google.protobuf.Any":
		// empty type URL is not allowed by JSON representation
		// so we must give it a dummy type
		var any anypb.Any
		_ = anypb.MarshalFrom(&any, &emptypb.Empty{}, protov2.MarshalOptions{})
		return &any
	case "google.protobuf.Value":
		// unset kind is not allowed by JSON representation
		// so we must give it something
		return &structpb.Value{
			Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"google.protobuf.Value": {Kind: &structpb.Value_StringValue{
						StringValue: "supports arbitrary JSON",
					}},
				},
			}},
		}
	case "google.protobuf.ListValue":
		return &structpb.ListValue{
			Values: []*structpb.Value{
				{
					Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"google.protobuf.ListValue": {Kind: &structpb.Value_StringValue{
								StringValue: "is an array of arbitrary JSON values",
							}},
						},
					}},
				},
			},
		}
	case "google.protobuf.Struct":
		return &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"google.protobuf.Struct": {Kind: &structpb.Value_StringValue{
					StringValue: "supports arbitrary JSON objects",
				}},
			},
		}
	}

	dm := dynamic.NewMessage(md)

	// if the message is a recursive structure, we don't want to blow the stack
	for _, seen := range path {
		if seen == md {
			// already visited this type; avoid infinite recursion
			return dm
		}
	}
	path = append(path, dm.GetMessageDescriptor())

	// for repeated fields, add a single element with default value
	// and for message fields, add a message with all default fields
	// that also has non-nil message and non-empty repeated fields

	for _, fd := range dm.GetMessageDescriptor().GetFields() {
		if fd.IsRepeated() {
			switch fd.GetType() {
			case descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
				descriptorpb.FieldDescriptorProto_TYPE_UINT32:
				dm.AddRepeatedField(fd, uint32(0))

			case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
				descriptorpb.FieldDescriptorProto_TYPE_SINT32,
				descriptorpb.FieldDescriptorProto_TYPE_INT32,
				descriptorpb.FieldDescriptorProto_TYPE_ENUM:
				dm.AddRepeatedField(fd, int32(0))

			case descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
				descriptorpb.FieldDescriptorProto_TYPE_UINT64:
				dm.AddRepeatedField(fd, uint64(0))

			case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64,
				descriptorpb.FieldDescriptorProto_TYPE_SINT64,
				descriptorpb.FieldDescriptorProto_TYPE_INT64:
				dm.AddRepeatedField(fd, int64(0))

			case descriptorpb.FieldDescriptorProto_TYPE_STRING:
				dm.AddRepeatedField(fd, "")

			case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
				dm.AddRepeatedField(fd, []byte{})

			case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
				dm.AddRepeatedField(fd, false)

			case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
				dm.AddRepeatedField(fd, float32(0))

			case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
				dm.AddRepeatedField(fd, float64(0))

			case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE,
				descriptorpb.FieldDescriptorProto_TYPE_GROUP:
				dm.AddRepeatedField(fd, makeTemplate(fd.GetMessageType(), path))
			}
		} else if fd.GetMessageType() != nil {
			dm.SetField(fd, makeTemplate(fd.GetMessageType(), path))
		}
	}
	return dm
}

// ClientTransportCredentials is a helper function that constructs a TLS config with
// the given properties (see ClientTLSConfig) and then constructs and returns gRPC
// transport credentials using that config.
//
// Deprecated: Use grpcurl.ClientTLSConfig and credentials.NewTLS instead.
func ClientTransportCredentials(insecureSkipVerify bool, cacertFile, clientCertFile, clientKeyFile string) (credentials.TransportCredentials, error) {
	tlsConf, err := ClientTLSConfig(insecureSkipVerify, cacertFile, clientCertFile, clientKeyFile)
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(tlsConf), nil
}

// ClientTLSConfig builds transport-layer config for a gRPC client using the
// given properties. If cacertFile is blank, only standard trusted certs are used to
// verify the server certs. If clientCertFile is blank, the client will not use a client
// certificate. If clientCertFile is not blank then clientKeyFile must not be blank.
func ClientTLSConfig(insecureSkipVerify bool, cacertFile, clientCertFile, clientKeyFile string) (*tls.Config, error) {
	var tlsConf tls.Config

	if clientCertFile != "" {
		// Load the client certificates from disk
		certificate, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("could not load client key pair: %v", err)
		}
		tlsConf.Certificates = []tls.Certificate{certificate}
	}

	if insecureSkipVerify {
		tlsConf.InsecureSkipVerify = true
	} else if cacertFile != "" {
		// Create a certificate pool from the certificate authority
		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(cacertFile)
		if err != nil {
			return nil, fmt.Errorf("could not read ca certificate: %v", err)
		}

		// Append the certificates from the CA
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return nil, errors.New("failed to append ca certs")
		}

		tlsConf.RootCAs = certPool
	}

	return &tlsConf, nil
}

// ServerTransportCredentials builds transport credentials for a gRPC server using the
// given properties. If cacertFile is blank, the server will not request client certs
// unless requireClientCerts is true. When requireClientCerts is false and cacertFile is
// not blank, the server will verify client certs when presented, but will not require
// client certs. The serverCertFile and serverKeyFile must both not be blank.
func ServerTransportCredentials(cacertFile, serverCertFile, serverKeyFile string, requireClientCerts bool) (credentials.TransportCredentials, error) {
	var tlsConf tls.Config
	// TODO(jh): Remove this line once https://github.com/golang/go/issues/28779 is fixed
	// in Go tip. Until then, the recently merged TLS 1.3 support breaks the TLS tests.
	tlsConf.MaxVersion = tls.VersionTLS12

	// Load the server certificates from disk
	certificate, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
	if err != nil {
		return nil, fmt.Errorf("could not load key pair: %v", err)
	}
	tlsConf.Certificates = []tls.Certificate{certificate}

	if cacertFile != "" {
		// Create a certificate pool from the certificate authority
		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(cacertFile)
		if err != nil {
			return nil, fmt.Errorf("could not read ca certificate: %v", err)
		}

		// Append the certificates from the CA
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return nil, errors.New("failed to append ca certs")
		}

		tlsConf.ClientCAs = certPool
	}

	if requireClientCerts {
		tlsConf.ClientAuth = tls.RequireAndVerifyClientCert
	} else if cacertFile != "" {
		tlsConf.ClientAuth = tls.VerifyClientCertIfGiven
	} else {
		tlsConf.ClientAuth = tls.NoClientCert
	}

	return credentials.NewTLS(&tlsConf), nil
}

// BlockingDial is a helper method to dial the given address, using optional TLS credentials,
// and blocking until the returned connection is ready. If the given credentials are nil, the
// connection will be insecure (plain-text).
func BlockingDial(ctx context.Context, network, address string, creds credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// grpc.Dial doesn't provide any information on permanent connection errors (like
	// TLS handshake failures). So in order to provide good error messages, we need a
	// custom dialer that can provide that info. That means we manage the TLS handshake.
	result := make(chan interface{}, 1)

	writeResult := func(res interface{}) {
		// non-blocking write: we only need the first result
		select {
		case result <- res:
		default:
		}
	}

	// custom credentials and dialer will notify on error via the
	// writeResult function
	if creds != nil {
		creds = &errSignalingCreds{
			TransportCredentials: creds,
			writeResult:          writeResult,
		}
	}
	dialer := func(ctx context.Context, address string) (net.Conn, error) {
		// NB: We *could* handle the TLS handshake ourselves, in the custom
		// dialer (instead of customizing both the dialer and the credentials).
		// But that requires using insecure.NewCredentials() dial transport
		// option (so that the gRPC library doesn't *also* try to do a
		// handshake). And that would mean that the library would send the
		// wrong ":scheme" metaheader to servers: it would send "http" instead
		// of "https" because it is unaware that TLS is actually in use.
		conn, err := (&net.Dialer{}).DialContext(ctx, network, address)
		if err != nil {
			writeResult(err)
		}
		return conn, err
	}

	// Even with grpc.FailOnNonTempDialError, this call will usually timeout in
	// the face of TLS handshake errors. So we can't rely on grpc.WithBlock() to
	// know when we're done. So we run it in a goroutine and then use result
	// channel to either get the connection or fail-fast.
	go func() {
		// We put grpc.FailOnNonTempDialError *before* the explicitly provided
		// options so that it could be overridden.
		opts = append([]grpc.DialOption{grpc.FailOnNonTempDialError(true)}, opts...)
		// But we don't want caller to be able to override these two, so we put
		// them *after* the explicitly provided options.
		opts = append(opts, grpc.WithBlock(), grpc.WithContextDialer(dialer))

		if creds == nil {
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		} else {
			opts = append(opts, grpc.WithTransportCredentials(creds))
		}
		conn, err := grpc.DialContext(ctx, address, opts...)
		var res interface{}
		if err != nil {
			res = err
		} else {
			res = conn
		}
		writeResult(res)
	}()

	select {
	case res := <-result:
		if conn, ok := res.(*grpc.ClientConn); ok {
			return conn, nil
		}
		return nil, res.(error)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// errSignalingCreds is a wrapper around a TransportCredentials value, but
// it will use the writeResult function to notify on error.
type errSignalingCreds struct {
	credentials.TransportCredentials
	writeResult func(res interface{})
}

func (c *errSignalingCreds) ClientHandshake(ctx context.Context, addr string, rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn, auth, err := c.TransportCredentials.ClientHandshake(ctx, addr, rawConn)
	if err != nil {
		c.writeResult(err)
	}
	return conn, auth, err
}
