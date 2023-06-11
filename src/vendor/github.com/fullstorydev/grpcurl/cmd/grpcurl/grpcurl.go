// Command grpcurl makes gRPC requests (a la cURL, but HTTP/2). It can use a supplied descriptor
// file, protobuf sources, or service reflection to translate JSON or text request data into the
// appropriate protobuf messages and vice versa for presenting the response contents.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/descriptorpb"

	// Register gzip compressor so compressed responses will work
	_ "google.golang.org/grpc/encoding/gzip"
	// Register xds so xds and xds-experimental resolver schemes work
	_ "google.golang.org/grpc/xds"

	"github.com/fullstorydev/grpcurl"
)

// To avoid confusion between program error codes and the gRPC resonse
// status codes 'Cancelled' and 'Unknown', 1 and 2 respectively,
// the response status codes emitted use an offest of 64
const statusCodeOffset = 64

const no_version = "dev build <no version set>"

var version = no_version

var (
	exit = os.Exit

	isUnixSocket func() bool // nil when run on non-unix platform

	flags = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	help = flags.Bool("help", false, prettify(`
		Print usage instructions and exit.`))
	printVersion = flags.Bool("version", false, prettify(`
		Print version.`))
	plaintext = flags.Bool("plaintext", false, prettify(`
		Use plain-text HTTP/2 when connecting to server (no TLS).`))
	insecure = flags.Bool("insecure", false, prettify(`
		Skip server certificate and domain verification. (NOT SECURE!) Not
		valid with -plaintext option.`))
	cacert = flags.String("cacert", "", prettify(`
		File containing trusted root certificates for verifying the server.
		Ignored if -insecure is specified.`))
	cert = flags.String("cert", "", prettify(`
		File containing client certificate (public key), to present to the
		server. Not valid with -plaintext option. Must also provide -key option.`))
	key = flags.String("key", "", prettify(`
		File containing client private key, to present to the server. Not valid
		with -plaintext option. Must also provide -cert option.`))
	protoset      multiString
	protoFiles    multiString
	importPaths   multiString
	addlHeaders   multiString
	rpcHeaders    multiString
	reflHeaders   multiString
	expandHeaders = flags.Bool("expand-headers", false, prettify(`
		If set, headers may use '${NAME}' syntax to reference environment
		variables. These will be expanded to the actual environment variable
		value before sending to the server. For example, if there is an
		environment variable defined like FOO=bar, then a header of
		'key: ${FOO}' would expand to 'key: bar'. This applies to -H,
		-rpc-header, and -reflect-header options. No other expansion/escaping is
		performed. This can be used to supply credentials/secrets without having
		to put them in command-line arguments.`))
	authority = flags.String("authority", "", prettify(`
		The authoritative name of the remote server. This value is passed as the
		value of the ":authority" pseudo-header in the HTTP/2 protocol. When TLS
		is used, this will also be used as the server name when verifying the
		server's certificate. It defaults to the address that is provided in the
		positional arguments.`))
	userAgent = flags.String("user-agent", "", prettify(`
		If set, the specified value will be added to the User-Agent header set
		by the grpc-go library.
		`))
	data = flags.String("d", "", prettify(`
		Data for request contents. If the value is '@' then the request contents
		are read from stdin. For calls that accept a stream of requests, the
		contents should include all such request messages concatenated together
		(possibly delimited; see -format).`))
	format = flags.String("format", "json", prettify(`
		The format of request data. The allowed values are 'json' or 'text'. For
		'json', the input data must be in JSON format. Multiple request values
		may be concatenated (messages with a JSON representation other than
		object must be separated by whitespace, such as a newline). For 'text',
		the input data must be in the protobuf text format, in which case
		multiple request values must be separated by the "record separator"
		ASCII character: 0x1E. The stream should not end in a record separator.
		If it does, it will be interpreted as a final, blank message after the
		separator.`))
	allowUnknownFields = flags.Bool("allow-unknown-fields", false, prettify(`
		When true, the request contents, if 'json' format is used, allows
		unknown fields to be present. They will be ignored when parsing
		the request.`))
	connectTimeout = flags.Float64("connect-timeout", 0, prettify(`
		The maximum time, in seconds, to wait for connection to be established.
		Defaults to 10 seconds.`))
	formatError = flags.Bool("format-error", false, prettify(`
		When a non-zero status is returned, format the response using the
		value set by the -format flag .`))
	keepaliveTime = flags.Float64("keepalive-time", 0, prettify(`
		If present, the maximum idle time in seconds, after which a keepalive
		probe is sent. If the connection remains idle and no keepalive response
		is received for this same period then the connection is closed and the
		operation fails.`))
	maxTime = flags.Float64("max-time", 0, prettify(`
		The maximum total time the operation can take, in seconds. This is
		useful for preventing batch jobs that use grpcurl from hanging due to
		slow or bad network links or due to incorrect stream method usage.`))
	maxMsgSz = flags.Int("max-msg-sz", 0, prettify(`
		The maximum encoded size of a response message, in bytes, that grpcurl
		will accept. If not specified, defaults to 4,194,304 (4 megabytes).`))
	emitDefaults = flags.Bool("emit-defaults", false, prettify(`
		Emit default values for JSON-encoded responses.`))
	protosetOut = flags.String("protoset-out", "", prettify(`
		The name of a file to be written that will contain a FileDescriptorSet
		proto. With the list and describe verbs, the listed or described
		elements and their transitive dependencies will be written to the named
		file if this option is given. When invoking an RPC and this option is
		given, the method being invoked and its transitive dependencies will be
		included in the output file.`))
	msgTemplate = flags.Bool("msg-template", false, prettify(`
		When describing messages, show a template of input data.`))
	verbose = flags.Bool("v", false, prettify(`
		Enable verbose output.`))
	veryVerbose = flags.Bool("vv", false, prettify(`
		Enable very verbose output.`))
	serverName = flags.String("servername", "", prettify(`
		Override server name when validating TLS certificate. This flag is
		ignored if -plaintext or -insecure is used.
		NOTE: Prefer -authority. This flag may be removed in the future. It is
		an error to use both -authority and -servername (though this will be
		permitted if they are both set to the same value, to increase backwards
		compatibility with earlier releases that allowed both to be set).`))
	reflection = optionalBoolFlag{val: true}
)

func init() {
	flags.Var(&addlHeaders, "H", prettify(`
		Additional headers in 'name: value' format. May specify more than one
		via multiple flags. These headers will also be included in reflection
		requests to a server.`))
	flags.Var(&rpcHeaders, "rpc-header", prettify(`
		Additional RPC headers in 'name: value' format. May specify more than
		one via multiple flags. These headers will *only* be used when invoking
		the requested RPC method. They are excluded from reflection requests.`))
	flags.Var(&reflHeaders, "reflect-header", prettify(`
		Additional reflection headers in 'name: value' format. May specify more
		than one via multiple flags. These headers will *only* be used during
		reflection requests and will be excluded when invoking the requested RPC
		method.`))
	flags.Var(&protoset, "protoset", prettify(`
		The name of a file containing an encoded FileDescriptorSet. This file's
		contents will be used to determine the RPC schema instead of querying
		for it from the remote server via the gRPC reflection API. When set: the
		'list' action lists the services found in the given descriptors (vs.
		those exposed by the remote server), and the 'describe' action describes
		symbols found in the given descriptors. May specify more than one via
		multiple -protoset flags. It is an error to use both -protoset and
		-proto flags.`))
	flags.Var(&protoFiles, "proto", prettify(`
		The name of a proto source file. Source files given will be used to
		determine the RPC schema instead of querying for it from the remote
		server via the gRPC reflection API. When set: the 'list' action lists
		the services found in the given files and their imports (vs. those
		exposed by the remote server), and the 'describe' action describes
		symbols found in the given files. May specify more than one via multiple
		-proto flags. Imports will be resolved using the given -import-path
		flags. Multiple proto files can be specified by specifying multiple
		-proto flags. It is an error to use both -protoset and -proto flags.`))
	flags.Var(&importPaths, "import-path", prettify(`
		The path to a directory from which proto sources can be imported, for
		use with -proto flags. Multiple import paths can be configured by
		specifying multiple -import-path flags. Paths will be searched in the
		order given. If no import paths are given, all files (including all
		imports) must be provided as -proto flags, and grpcurl will attempt to
		resolve all import statements from the set of file names given.`))
	flags.Var(&reflection, "use-reflection", prettify(`
		When true, server reflection will be used to determine the RPC schema.
		Defaults to true unless a -proto or -protoset option is provided. If
		-use-reflection is used in combination with a -proto or -protoset flag,
		the provided descriptor sources will be used in addition to server
		reflection to resolve messages and extensions.`))
}

type multiString []string

func (s *multiString) String() string {
	return strings.Join(*s, ",")
}

func (s *multiString) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// Uses a file source as a fallback for resolving symbols and extensions, but
// only uses the reflection source for listing services
type compositeSource struct {
	reflection grpcurl.DescriptorSource
	file       grpcurl.DescriptorSource
}

func (cs compositeSource) ListServices() ([]string, error) {
	return cs.reflection.ListServices()
}

func (cs compositeSource) FindSymbol(fullyQualifiedName string) (desc.Descriptor, error) {
	d, err := cs.reflection.FindSymbol(fullyQualifiedName)
	if err == nil {
		return d, nil
	}
	return cs.file.FindSymbol(fullyQualifiedName)
}

func (cs compositeSource) AllExtensionsForType(typeName string) ([]*desc.FieldDescriptor, error) {
	exts, err := cs.reflection.AllExtensionsForType(typeName)
	if err != nil {
		// On error fall back to file source
		return cs.file.AllExtensionsForType(typeName)
	}
	// Track the tag numbers from the reflection source
	tags := make(map[int32]bool)
	for _, ext := range exts {
		tags[ext.GetNumber()] = true
	}
	fileExts, err := cs.file.AllExtensionsForType(typeName)
	if err != nil {
		return exts, nil
	}
	for _, ext := range fileExts {
		// Prioritize extensions found via reflection
		if !tags[ext.GetNumber()] {
			exts = append(exts, ext)
		}
	}
	return exts, nil
}

func main() {
	flags.Usage = usage
	flags.Parse(os.Args[1:])
	if *help {
		usage()
		os.Exit(0)
	}
	if *printVersion {
		fmt.Fprintf(os.Stderr, "%s %s\n", filepath.Base(os.Args[0]), version)
		os.Exit(0)
	}

	// Do extra validation on arguments and figure out what user asked us to do.
	if *connectTimeout < 0 {
		fail(nil, "The -connect-timeout argument must not be negative.")
	}
	if *keepaliveTime < 0 {
		fail(nil, "The -keepalive-time argument must not be negative.")
	}
	if *maxTime < 0 {
		fail(nil, "The -max-time argument must not be negative.")
	}
	if *maxMsgSz < 0 {
		fail(nil, "The -max-msg-sz argument must not be negative.")
	}
	if *plaintext && *insecure {
		fail(nil, "The -plaintext and -insecure arguments are mutually exclusive.")
	}
	if *plaintext && *cert != "" {
		fail(nil, "The -plaintext and -cert arguments are mutually exclusive.")
	}
	if *plaintext && *key != "" {
		fail(nil, "The -plaintext and -key arguments are mutually exclusive.")
	}
	if (*key == "") != (*cert == "") {
		fail(nil, "The -cert and -key arguments must be used together and both be present.")
	}
	if *format != "json" && *format != "text" {
		fail(nil, "The -format option must be 'json' or 'text'.")
	}
	if *emitDefaults && *format != "json" {
		warn("The -emit-defaults is only used when using json format.")
	}

	args := flags.Args()

	if len(args) == 0 {
		fail(nil, "Too few arguments.")
	}
	var target string
	if args[0] != "list" && args[0] != "describe" {
		target = args[0]
		args = args[1:]
	}

	if len(args) == 0 {
		fail(nil, "Too few arguments.")
	}
	var list, describe, invoke bool
	if args[0] == "list" {
		list = true
		args = args[1:]
	} else if args[0] == "describe" {
		describe = true
		args = args[1:]
	} else {
		invoke = true
	}

	verbosityLevel := 0
	if *verbose {
		verbosityLevel = 1
	}
	if *veryVerbose {
		verbosityLevel = 2
	}

	var symbol string
	if invoke {
		if len(args) == 0 {
			fail(nil, "Too few arguments.")
		}
		symbol = args[0]
		args = args[1:]
	} else {
		if *data != "" {
			warn("The -d argument is not used with 'list' or 'describe' verb.")
		}
		if len(rpcHeaders) > 0 {
			warn("The -rpc-header argument is not used with 'list' or 'describe' verb.")
		}
		if len(args) > 0 {
			symbol = args[0]
			args = args[1:]
		}
	}

	if len(args) > 0 {
		fail(nil, "Too many arguments.")
	}
	if invoke && target == "" {
		fail(nil, "No host:port specified.")
	}
	if len(protoset) == 0 && len(protoFiles) == 0 && target == "" {
		fail(nil, "No host:port specified, no protoset specified, and no proto sources specified.")
	}
	if len(protoset) > 0 && len(reflHeaders) > 0 {
		warn("The -reflect-header argument is not used when -protoset files are used.")
	}
	if len(protoset) > 0 && len(protoFiles) > 0 {
		fail(nil, "Use either -protoset files or -proto files, but not both.")
	}
	if len(importPaths) > 0 && len(protoFiles) == 0 {
		warn("The -import-path argument is not used unless -proto files are used.")
	}
	if !reflection.val && len(protoset) == 0 && len(protoFiles) == 0 {
		fail(nil, "No protoset files or proto files specified and -use-reflection set to false.")
	}

	// Protoset or protofiles provided and -use-reflection unset
	if !reflection.set && (len(protoset) > 0 || len(protoFiles) > 0) {
		reflection.val = false
	}

	ctx := context.Background()
	if *maxTime > 0 {
		timeout := time.Duration(*maxTime * float64(time.Second))
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	dial := func() *grpc.ClientConn {
		dialTime := 10 * time.Second
		if *connectTimeout > 0 {
			dialTime = time.Duration(*connectTimeout * float64(time.Second))
		}
		ctx, cancel := context.WithTimeout(ctx, dialTime)
		defer cancel()
		var opts []grpc.DialOption
		if *keepaliveTime > 0 {
			timeout := time.Duration(*keepaliveTime * float64(time.Second))
			opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:    timeout,
				Timeout: timeout,
			}))
		}
		if *maxMsgSz > 0 {
			opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(*maxMsgSz)))
		}
		var creds credentials.TransportCredentials
		if !*plaintext {
			tlsConf, err := grpcurl.ClientTLSConfig(*insecure, *cacert, *cert, *key)
			if err != nil {
				fail(err, "Failed to create TLS config")
			}

			sslKeylogFile := os.Getenv("SSLKEYLOGFILE")
			if sslKeylogFile != "" {
				w, err := os.OpenFile(sslKeylogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
				if err != nil {
					fail(err, "Could not open SSLKEYLOGFILE %s", sslKeylogFile)
				}
				tlsConf.KeyLogWriter = w
			}

			creds = credentials.NewTLS(tlsConf)

			// can use either -servername or -authority; but not both
			if *serverName != "" && *authority != "" {
				if *serverName == *authority {
					warn("Both -servername and -authority are present; prefer only -authority.")
				} else {
					fail(nil, "Cannot specify different values for -servername and -authority.")
				}
			}
			overrideName := *serverName
			if overrideName == "" {
				overrideName = *authority
			}

			if overrideName != "" {
				opts = append(opts, grpc.WithAuthority(overrideName))
			}
		} else if *authority != "" {
			opts = append(opts, grpc.WithAuthority(*authority))
		}

		grpcurlUA := "grpcurl/" + version
		if version == no_version {
			grpcurlUA = "grpcurl/dev-build (no version set)"
		}
		if *userAgent != "" {
			grpcurlUA = *userAgent + " " + grpcurlUA
		}
		opts = append(opts, grpc.WithUserAgent(grpcurlUA))

		network := "tcp"
		if isUnixSocket != nil && isUnixSocket() {
			network = "unix"
		}
		cc, err := grpcurl.BlockingDial(ctx, network, target, creds, opts...)
		if err != nil {
			fail(err, "Failed to dial target host %q", target)
		}
		return cc
	}
	printFormattedStatus := func(w io.Writer, stat *status.Status, formatter grpcurl.Formatter) {
		formattedStatus, err := formatter(stat.Proto())
		if err != nil {
			fmt.Fprintf(w, "ERROR: %v", err.Error())
		}
		fmt.Fprint(w, formattedStatus)
	}

	if *expandHeaders {
		var err error
		addlHeaders, err = grpcurl.ExpandHeaders(addlHeaders)
		if err != nil {
			fail(err, "Failed to expand additional headers")
		}
		rpcHeaders, err = grpcurl.ExpandHeaders(rpcHeaders)
		if err != nil {
			fail(err, "Failed to expand rpc headers")
		}
		reflHeaders, err = grpcurl.ExpandHeaders(reflHeaders)
		if err != nil {
			fail(err, "Failed to expand reflection headers")
		}
	}

	var cc *grpc.ClientConn
	var descSource grpcurl.DescriptorSource
	var refClient *grpcreflect.Client
	var fileSource grpcurl.DescriptorSource
	if len(protoset) > 0 {
		var err error
		fileSource, err = grpcurl.DescriptorSourceFromProtoSets(protoset...)
		if err != nil {
			fail(err, "Failed to process proto descriptor sets.")
		}
	} else if len(protoFiles) > 0 {
		var err error
		fileSource, err = grpcurl.DescriptorSourceFromProtoFiles(importPaths, protoFiles...)
		if err != nil {
			fail(err, "Failed to process proto source files.")
		}
	}
	if reflection.val {
		md := grpcurl.MetadataFromHeaders(append(addlHeaders, reflHeaders...))
		refCtx := metadata.NewOutgoingContext(ctx, md)
		cc = dial()
		refClient = grpcreflect.NewClient(refCtx, reflectpb.NewServerReflectionClient(cc))
		reflSource := grpcurl.DescriptorSourceFromServer(ctx, refClient)
		if fileSource != nil {
			descSource = compositeSource{reflSource, fileSource}
		} else {
			descSource = reflSource
		}
	} else {
		descSource = fileSource
	}

	// arrange for the RPCs to be cleanly shutdown
	reset := func() {
		if refClient != nil {
			refClient.Reset()
			refClient = nil
		}
		if cc != nil {
			cc.Close()
			cc = nil
		}
	}
	defer reset()
	exit = func(code int) {
		// since defers aren't run by os.Exit...
		reset()
		os.Exit(code)
	}

	if list {
		if symbol == "" {
			svcs, err := grpcurl.ListServices(descSource)
			if err != nil {
				fail(err, "Failed to list services")
			}
			if len(svcs) == 0 {
				fmt.Println("(No services)")
			} else {
				for _, svc := range svcs {
					fmt.Printf("%s\n", svc)
				}
			}
			if err := writeProtoset(descSource, svcs...); err != nil {
				fail(err, "Failed to write protoset to %s", *protosetOut)
			}
		} else {
			methods, err := grpcurl.ListMethods(descSource, symbol)
			if err != nil {
				fail(err, "Failed to list methods for service %q", symbol)
			}
			if len(methods) == 0 {
				fmt.Println("(No methods)") // probably unlikely
			} else {
				for _, m := range methods {
					fmt.Printf("%s\n", m)
				}
			}
			if err := writeProtoset(descSource, symbol); err != nil {
				fail(err, "Failed to write protoset to %s", *protosetOut)
			}
		}

	} else if describe {
		var symbols []string
		if symbol != "" {
			symbols = []string{symbol}
		} else {
			// if no symbol given, describe all exposed services
			svcs, err := descSource.ListServices()
			if err != nil {
				fail(err, "Failed to list services")
			}
			if len(svcs) == 0 {
				fmt.Println("Server returned an empty list of exposed services")
			}
			symbols = svcs
		}
		for _, s := range symbols {
			if s[0] == '.' {
				s = s[1:]
			}

			dsc, err := descSource.FindSymbol(s)
			if err != nil {
				fail(err, "Failed to resolve symbol %q", s)
			}

			fqn := dsc.GetFullyQualifiedName()
			var elementType string
			switch d := dsc.(type) {
			case *desc.MessageDescriptor:
				elementType = "a message"
				parent, ok := d.GetParent().(*desc.MessageDescriptor)
				if ok {
					if d.IsMapEntry() {
						for _, f := range parent.GetFields() {
							if f.IsMap() && f.GetMessageType() == d {
								// found it: describe the map field instead
								elementType = "the entry type for a map field"
								dsc = f
								break
							}
						}
					} else {
						// see if it's a group
						for _, f := range parent.GetFields() {
							if f.GetType() == descriptorpb.FieldDescriptorProto_TYPE_GROUP && f.GetMessageType() == d {
								// found it: describe the map field instead
								elementType = "the type of a group field"
								dsc = f
								break
							}
						}
					}
				}
			case *desc.FieldDescriptor:
				elementType = "a field"
				if d.GetType() == descriptorpb.FieldDescriptorProto_TYPE_GROUP {
					elementType = "a group field"
				} else if d.IsExtension() {
					elementType = "an extension"
				}
			case *desc.OneOfDescriptor:
				elementType = "a one-of"
			case *desc.EnumDescriptor:
				elementType = "an enum"
			case *desc.EnumValueDescriptor:
				elementType = "an enum value"
			case *desc.ServiceDescriptor:
				elementType = "a service"
			case *desc.MethodDescriptor:
				elementType = "a method"
			default:
				err = fmt.Errorf("descriptor has unrecognized type %T", dsc)
				fail(err, "Failed to describe symbol %q", s)
			}

			txt, err := grpcurl.GetDescriptorText(dsc, descSource)
			if err != nil {
				fail(err, "Failed to describe symbol %q", s)
			}
			fmt.Printf("%s is %s:\n", fqn, elementType)
			fmt.Println(txt)

			if dsc, ok := dsc.(*desc.MessageDescriptor); ok && *msgTemplate {
				// for messages, also show a template in JSON, to make it easier to
				// create a request to invoke an RPC
				tmpl := grpcurl.MakeTemplate(dsc)
				options := grpcurl.FormatOptions{EmitJSONDefaultFields: true}
				_, formatter, err := grpcurl.RequestParserAndFormatter(grpcurl.Format(*format), descSource, nil, options)
				if err != nil {
					fail(err, "Failed to construct formatter for %q", *format)
				}
				str, err := formatter(tmpl)
				if err != nil {
					fail(err, "Failed to print template for message %s", s)
				}
				fmt.Println("\nMessage template:")
				fmt.Println(str)
			}
		}
		if err := writeProtoset(descSource, symbols...); err != nil {
			fail(err, "Failed to write protoset to %s", *protosetOut)
		}

	} else {
		// Invoke an RPC
		if cc == nil {
			cc = dial()
		}
		var in io.Reader
		if *data == "@" {
			in = os.Stdin
		} else {
			in = strings.NewReader(*data)
		}

		// if not verbose output, then also include record delimiters
		// between each message, so output could potentially be piped
		// to another grpcurl process
		includeSeparators := verbosityLevel == 0
		options := grpcurl.FormatOptions{
			EmitJSONDefaultFields: *emitDefaults,
			IncludeTextSeparator:  includeSeparators,
			AllowUnknownFields:    *allowUnknownFields,
		}
		rf, formatter, err := grpcurl.RequestParserAndFormatter(grpcurl.Format(*format), descSource, in, options)
		if err != nil {
			fail(err, "Failed to construct request parser and formatter for %q", *format)
		}
		h := &grpcurl.DefaultEventHandler{
			Out:            os.Stdout,
			Formatter:      formatter,
			VerbosityLevel: verbosityLevel,
		}

		err = grpcurl.InvokeRPC(ctx, descSource, cc, symbol, append(addlHeaders, rpcHeaders...), h, rf.Next)
		if err != nil {
			if errStatus, ok := status.FromError(err); ok && *formatError {
				h.Status = errStatus
			} else {
				fail(err, "Error invoking method %q", symbol)
			}
		}
		reqSuffix := ""
		respSuffix := ""
		reqCount := rf.NumRequests()
		if reqCount != 1 {
			reqSuffix = "s"
		}
		if h.NumResponses != 1 {
			respSuffix = "s"
		}
		if verbosityLevel > 0 {
			fmt.Printf("Sent %d request%s and received %d response%s\n", reqCount, reqSuffix, h.NumResponses, respSuffix)
		}
		if h.Status.Code() != codes.OK {
			if *formatError {
				printFormattedStatus(os.Stderr, h.Status, formatter)
			} else {
				grpcurl.PrintStatus(os.Stderr, h.Status, formatter)
			}
			exit(statusCodeOffset + int(h.Status.Code()))
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage:
	%s [flags] [address] [list|describe] [symbol]

The 'address' is only optional when used with 'list' or 'describe' and a
protoset or proto flag is provided.

If 'list' is indicated, the symbol (if present) should be a fully-qualified
service name. If present, all methods of that service are listed. If not
present, all exposed services are listed, or all services defined in protosets.

If 'describe' is indicated, the descriptor for the given symbol is shown. The
symbol should be a fully-qualified service, enum, or message name. If no symbol
is given then the descriptors for all exposed or known services are shown.

If neither verb is present, the symbol must be a fully-qualified method name in
'service/method' or 'service.method' format. In this case, the request body will
be used to invoke the named method. If no body is given but one is required
(i.e. the method is unary or server-streaming), an empty instance of the
method's request type will be sent.

The address will typically be in the form "host:port" where host can be an IP
address or a hostname and port is a numeric port or service name. If an IPv6
address is given, it must be surrounded by brackets, like "[2001:db8::1]". For
Unix variants, if a -unix=true flag is present, then the address must be the
path to the domain socket.

Available flags:
`, os.Args[0])
	flags.PrintDefaults()
}

func prettify(docString string) string {
	parts := strings.Split(docString, "\n")

	// cull empty lines and also remove trailing and leading spaces
	// from each line in the doc string
	j := 0
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		parts[j] = part
		j++
	}

	return strings.Join(parts[:j], "\n"+indent())
}

func warn(msg string, args ...interface{}) {
	msg = fmt.Sprintf("Warning: %s\n", msg)
	fmt.Fprintf(os.Stderr, msg, args...)
}

func fail(err error, msg string, args ...interface{}) {
	if err != nil {
		msg += ": %v"
		args = append(args, err)
	}
	fmt.Fprintf(os.Stderr, msg, args...)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		exit(1)
	} else {
		// nil error means it was CLI usage issue
		fmt.Fprintf(os.Stderr, "Try '%s -help' for more details.\n", os.Args[0])
		exit(2)
	}
}

func writeProtoset(descSource grpcurl.DescriptorSource, symbols ...string) error {
	if *protosetOut == "" {
		return nil
	}
	f, err := os.Create(*protosetOut)
	if err != nil {
		return err
	}
	defer f.Close()
	return grpcurl.WriteProtoset(f, descSource, symbols...)
}

type optionalBoolFlag struct {
	set, val bool
}

func (f *optionalBoolFlag) String() string {
	if !f.set {
		return "unset"
	}
	return strconv.FormatBool(f.val)
}

func (f *optionalBoolFlag) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	f.set = true
	f.val = v
	return nil
}

func (f *optionalBoolFlag) IsBoolFlag() bool {
	return true
}
