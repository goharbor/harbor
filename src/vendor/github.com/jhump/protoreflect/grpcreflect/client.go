package grpcreflect

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"sync"

	"github.com/golang/protobuf/proto"
	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"google.golang.org/grpc/codes"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/internal"
)

// elementNotFoundError is the error returned by reflective operations where the
// server does not recognize a given file name, symbol name, or extension.
type elementNotFoundError struct {
	name    string
	kind    elementKind
	symType symbolType // only used when kind == elementKindSymbol
	tag     int32      // only used when kind == elementKindExtension

	// only errors with a kind of elementKindFile will have a cause, which means
	// the named file count not be resolved because of a dependency that could
	// not be found where cause describes the missing dependency
	cause *elementNotFoundError
}

type elementKind int

const (
	elementKindSymbol elementKind = iota
	elementKindFile
	elementKindExtension
)

type symbolType string

const (
	symbolTypeService = "Service"
	symbolTypeMessage = "Message"
	symbolTypeEnum    = "Enum"
	symbolTypeUnknown = "Symbol"
)

func symbolNotFound(symbol string, symType symbolType, cause *elementNotFoundError) error {
	return &elementNotFoundError{name: symbol, symType: symType, kind: elementKindSymbol, cause: cause}
}

func extensionNotFound(extendee string, tag int32, cause *elementNotFoundError) error {
	return &elementNotFoundError{name: extendee, tag: tag, kind: elementKindExtension, cause: cause}
}

func fileNotFound(file string, cause *elementNotFoundError) error {
	return &elementNotFoundError{name: file, kind: elementKindFile, cause: cause}
}

func (e *elementNotFoundError) Error() string {
	first := true
	var b bytes.Buffer
	for ; e != nil; e = e.cause {
		if first {
			first = false
		} else {
			fmt.Fprint(&b, "\ncaused by: ")
		}
		switch e.kind {
		case elementKindSymbol:
			fmt.Fprintf(&b, "%s not found: %s", e.symType, e.name)
		case elementKindExtension:
			fmt.Fprintf(&b, "Extension not found: tag %d for %s", e.tag, e.name)
		default:
			fmt.Fprintf(&b, "File not found: %s", e.name)
		}
	}
	return b.String()
}

// IsElementNotFoundError determines if the given error indicates that a file
// name, symbol name, or extension field was could not be found by the server.
func IsElementNotFoundError(err error) bool {
	_, ok := err.(*elementNotFoundError)
	return ok
}

// ProtocolError is an error returned when the server sends a response of the
// wrong type.
type ProtocolError struct {
	missingType reflect.Type
}

func (p ProtocolError) Error() string {
	return fmt.Sprintf("Protocol error: response was missing %v", p.missingType)
}

type extDesc struct {
	extendedMessageName string
	extensionNumber     int32
}

// Client is a client connection to a server for performing reflection calls
// and resolving remote symbols.
type Client struct {
	ctx  context.Context
	stub rpb.ServerReflectionClient

	connMu sync.Mutex
	cancel context.CancelFunc
	stream rpb.ServerReflection_ServerReflectionInfoClient

	cacheMu          sync.RWMutex
	protosByName     map[string]*dpb.FileDescriptorProto
	filesByName      map[string]*desc.FileDescriptor
	filesBySymbol    map[string]*desc.FileDescriptor
	filesByExtension map[extDesc]*desc.FileDescriptor
}

// NewClient creates a new Client with the given root context and using the
// given RPC stub for talking to the server.
func NewClient(ctx context.Context, stub rpb.ServerReflectionClient) *Client {
	cr := &Client{
		ctx:              ctx,
		stub:             stub,
		protosByName:     map[string]*dpb.FileDescriptorProto{},
		filesByName:      map[string]*desc.FileDescriptor{},
		filesBySymbol:    map[string]*desc.FileDescriptor{},
		filesByExtension: map[extDesc]*desc.FileDescriptor{},
	}
	// don't leak a grpc stream
	runtime.SetFinalizer(cr, (*Client).Reset)
	return cr
}

// FileByFilename asks the server for a file descriptor for the proto file with
// the given name.
func (cr *Client) FileByFilename(filename string) (*desc.FileDescriptor, error) {
	// hit the cache first
	cr.cacheMu.RLock()
	if fd, ok := cr.filesByName[filename]; ok {
		cr.cacheMu.RUnlock()
		return fd, nil
	}
	fdp, ok := cr.protosByName[filename]
	cr.cacheMu.RUnlock()
	// not there? see if we've downloaded the proto
	if ok {
		return cr.descriptorFromProto(fdp)
	}

	req := &rpb.ServerReflectionRequest{
		MessageRequest: &rpb.ServerReflectionRequest_FileByFilename{
			FileByFilename: filename,
		},
	}
	accept := func(fd *desc.FileDescriptor) bool {
		return fd.GetName() == filename
	}

	fd, err := cr.getAndCacheFileDescriptors(req, filename, "", accept)
	if isNotFound(err) {
		// file not found? see if we can look up via alternate name
		if alternate, ok := internal.StdFileAliases[filename]; ok {
			req := &rpb.ServerReflectionRequest{
				MessageRequest: &rpb.ServerReflectionRequest_FileByFilename{
					FileByFilename: alternate,
				},
			}
			fd, err = cr.getAndCacheFileDescriptors(req, alternate, filename, accept)
			if isNotFound(err) {
				err = fileNotFound(filename, nil)
			}
		} else {
			err = fileNotFound(filename, nil)
		}
	} else if e, ok := err.(*elementNotFoundError); ok {
		err = fileNotFound(filename, e)
	}
	return fd, err
}

// FileContainingSymbol asks the server for a file descriptor for the proto file
// that declares the given fully-qualified symbol.
func (cr *Client) FileContainingSymbol(symbol string) (*desc.FileDescriptor, error) {
	// hit the cache first
	cr.cacheMu.RLock()
	fd, ok := cr.filesBySymbol[symbol]
	cr.cacheMu.RUnlock()
	if ok {
		return fd, nil
	}

	req := &rpb.ServerReflectionRequest{
		MessageRequest: &rpb.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: symbol,
		},
	}
	accept := func(fd *desc.FileDescriptor) bool {
		return fd.FindSymbol(symbol) != nil
	}
	fd, err := cr.getAndCacheFileDescriptors(req, "", "", accept)
	if isNotFound(err) {
		err = symbolNotFound(symbol, symbolTypeUnknown, nil)
	} else if e, ok := err.(*elementNotFoundError); ok {
		err = symbolNotFound(symbol, symbolTypeUnknown, e)
	}
	return fd, err
}

// FileContainingExtension asks the server for a file descriptor for the proto
// file that declares an extension with the given number for the given
// fully-qualified message name.
func (cr *Client) FileContainingExtension(extendedMessageName string, extensionNumber int32) (*desc.FileDescriptor, error) {
	// hit the cache first
	cr.cacheMu.RLock()
	fd, ok := cr.filesByExtension[extDesc{extendedMessageName, extensionNumber}]
	cr.cacheMu.RUnlock()
	if ok {
		return fd, nil
	}

	req := &rpb.ServerReflectionRequest{
		MessageRequest: &rpb.ServerReflectionRequest_FileContainingExtension{
			FileContainingExtension: &rpb.ExtensionRequest{
				ContainingType:  extendedMessageName,
				ExtensionNumber: extensionNumber,
			},
		},
	}
	accept := func(fd *desc.FileDescriptor) bool {
		return fd.FindExtension(extendedMessageName, extensionNumber) != nil
	}
	fd, err := cr.getAndCacheFileDescriptors(req, "", "", accept)
	if isNotFound(err) {
		err = extensionNotFound(extendedMessageName, extensionNumber, nil)
	} else if e, ok := err.(*elementNotFoundError); ok {
		err = extensionNotFound(extendedMessageName, extensionNumber, e)
	}
	return fd, err
}

func (cr *Client) getAndCacheFileDescriptors(req *rpb.ServerReflectionRequest, expectedName, alias string, accept func(*desc.FileDescriptor) bool) (*desc.FileDescriptor, error) {
	resp, err := cr.send(req)
	if err != nil {
		return nil, err
	}

	fdResp := resp.GetFileDescriptorResponse()
	if fdResp == nil {
		return nil, &ProtocolError{reflect.TypeOf(fdResp).Elem()}
	}

	// Response can contain the result file descriptor, but also its transitive
	// deps. Furthermore, protocol states that subsequent requests do not need
	// to send transitive deps that have been sent in prior responses. So we
	// need to cache all file descriptors and then return the first one (which
	// should be the answer). If we're looking for a file by name, we can be
	// smarter and make sure to grab one by name instead of just grabbing the
	// first one.
	var fds []*dpb.FileDescriptorProto
	for _, fdBytes := range fdResp.FileDescriptorProto {
		fd := &dpb.FileDescriptorProto{}
		if err = proto.Unmarshal(fdBytes, fd); err != nil {
			return nil, err
		}

		if expectedName != "" && alias != "" && expectedName != alias && fd.GetName() == expectedName {
			// we found a file was aliased, so we need to update the proto to reflect that
			fd.Name = proto.String(alias)
		}

		cr.cacheMu.Lock()
		// store in cache of raw descriptor protos, but don't overwrite existing protos
		if existingFd, ok := cr.protosByName[fd.GetName()]; ok {
			fd = existingFd
		} else {
			cr.protosByName[fd.GetName()] = fd
		}
		cr.cacheMu.Unlock()

		fds = append(fds, fd)
	}

	// find the right result from the files returned
	for _, fd := range fds {
		result, err := cr.descriptorFromProto(fd)
		if err != nil {
			return nil, err
		}
		if accept(result) {
			return result, nil
		}
	}

	return nil, status.Errorf(codes.NotFound, "response does not include expected file")
}

func (cr *Client) descriptorFromProto(fd *dpb.FileDescriptorProto) (*desc.FileDescriptor, error) {
	deps := make([]*desc.FileDescriptor, len(fd.GetDependency()))
	for i, depName := range fd.GetDependency() {
		if dep, err := cr.FileByFilename(depName); err != nil {
			return nil, err
		} else {
			deps[i] = dep
		}
	}
	d, err := desc.CreateFileDescriptor(fd, deps...)
	if err != nil {
		return nil, err
	}
	d = cr.cacheFile(d)
	return d, nil
}

func (cr *Client) cacheFile(fd *desc.FileDescriptor) *desc.FileDescriptor {
	cr.cacheMu.Lock()
	defer cr.cacheMu.Unlock()

	// cache file descriptor by name, but don't overwrite existing entry
	// (existing entry could come from concurrent caller)
	if existingFd, ok := cr.filesByName[fd.GetName()]; ok {
		return existingFd
	}
	cr.filesByName[fd.GetName()] = fd

	// also cache by symbols and extensions
	for _, m := range fd.GetMessageTypes() {
		cr.cacheMessageLocked(fd, m)
	}
	for _, e := range fd.GetEnumTypes() {
		cr.filesBySymbol[e.GetFullyQualifiedName()] = fd
		for _, v := range e.GetValues() {
			cr.filesBySymbol[v.GetFullyQualifiedName()] = fd
		}
	}
	for _, e := range fd.GetExtensions() {
		cr.filesBySymbol[e.GetFullyQualifiedName()] = fd
		cr.filesByExtension[extDesc{e.GetOwner().GetFullyQualifiedName(), e.GetNumber()}] = fd
	}
	for _, s := range fd.GetServices() {
		cr.filesBySymbol[s.GetFullyQualifiedName()] = fd
		for _, m := range s.GetMethods() {
			cr.filesBySymbol[m.GetFullyQualifiedName()] = fd
		}
	}

	return fd
}

func (cr *Client) cacheMessageLocked(fd *desc.FileDescriptor, md *desc.MessageDescriptor) {
	cr.filesBySymbol[md.GetFullyQualifiedName()] = fd
	for _, f := range md.GetFields() {
		cr.filesBySymbol[f.GetFullyQualifiedName()] = fd
	}
	for _, o := range md.GetOneOfs() {
		cr.filesBySymbol[o.GetFullyQualifiedName()] = fd
	}
	for _, e := range md.GetNestedEnumTypes() {
		cr.filesBySymbol[e.GetFullyQualifiedName()] = fd
		for _, v := range e.GetValues() {
			cr.filesBySymbol[v.GetFullyQualifiedName()] = fd
		}
	}
	for _, e := range md.GetNestedExtensions() {
		cr.filesBySymbol[e.GetFullyQualifiedName()] = fd
		cr.filesByExtension[extDesc{e.GetOwner().GetFullyQualifiedName(), e.GetNumber()}] = fd
	}
	for _, m := range md.GetNestedMessageTypes() {
		cr.cacheMessageLocked(fd, m) // recurse
	}
}

// AllExtensionNumbersForType asks the server for all known extension numbers
// for the given fully-qualified message name.
func (cr *Client) AllExtensionNumbersForType(extendedMessageName string) ([]int32, error) {
	req := &rpb.ServerReflectionRequest{
		MessageRequest: &rpb.ServerReflectionRequest_AllExtensionNumbersOfType{
			AllExtensionNumbersOfType: extendedMessageName,
		},
	}
	resp, err := cr.send(req)
	if err != nil {
		if isNotFound(err) {
			return nil, symbolNotFound(extendedMessageName, symbolTypeMessage, nil)
		}
		return nil, err
	}

	extResp := resp.GetAllExtensionNumbersResponse()
	if extResp == nil {
		return nil, &ProtocolError{reflect.TypeOf(extResp).Elem()}
	}
	return extResp.ExtensionNumber, nil
}

// ListServices asks the server for the fully-qualified names of all exposed
// services.
func (cr *Client) ListServices() ([]string, error) {
	req := &rpb.ServerReflectionRequest{
		MessageRequest: &rpb.ServerReflectionRequest_ListServices{
			// proto doesn't indicate any purpose for this value and server impl
			// doesn't actually use it...
			ListServices: "*",
		},
	}
	resp, err := cr.send(req)
	if err != nil {
		return nil, err
	}

	listResp := resp.GetListServicesResponse()
	if listResp == nil {
		return nil, &ProtocolError{reflect.TypeOf(listResp).Elem()}
	}
	serviceNames := make([]string, len(listResp.Service))
	for i, s := range listResp.Service {
		serviceNames[i] = s.Name
	}
	return serviceNames, nil
}

func (cr *Client) send(req *rpb.ServerReflectionRequest) (*rpb.ServerReflectionResponse, error) {
	// we allow one immediate retry, in case we have a stale stream
	// (e.g. closed by server)
	resp, err := cr.doSend(true, req)
	if err != nil {
		return nil, err
	}

	// convert error response messages into errors
	errResp := resp.GetErrorResponse()
	if errResp != nil {
		return nil, status.Errorf(codes.Code(errResp.ErrorCode), "%s", errResp.ErrorMessage)
	}

	return resp, nil
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	s, ok := status.FromError(err)
	return ok && s.Code() == codes.NotFound
}

func (cr *Client) doSend(retry bool, req *rpb.ServerReflectionRequest) (*rpb.ServerReflectionResponse, error) {
	// TODO: Streams are thread-safe, so we shouldn't need to lock. But without locking, we'll need more machinery
	// (goroutines and channels) to ensure that responses are correctly correlated with their requests and thus
	// delivered in correct oder.
	cr.connMu.Lock()
	defer cr.connMu.Unlock()
	return cr.doSendLocked(retry, req)
}

func (cr *Client) doSendLocked(retry bool, req *rpb.ServerReflectionRequest) (*rpb.ServerReflectionResponse, error) {
	if err := cr.initStreamLocked(); err != nil {
		return nil, err
	}

	if err := cr.stream.Send(req); err != nil {
		if err == io.EOF {
			// if send returns EOF, must call Recv to get real underlying error
			_, err = cr.stream.Recv()
		}
		cr.resetLocked()
		if retry {
			return cr.doSendLocked(false, req)
		}
		return nil, err
	}

	if resp, err := cr.stream.Recv(); err != nil {
		cr.resetLocked()
		if retry {
			return cr.doSendLocked(false, req)
		}
		return nil, err
	} else {
		return resp, nil
	}
}

func (cr *Client) initStreamLocked() error {
	if cr.stream != nil {
		return nil
	}
	var newCtx context.Context
	newCtx, cr.cancel = context.WithCancel(cr.ctx)
	var err error
	cr.stream, err = cr.stub.ServerReflectionInfo(newCtx)
	return err
}

// Reset ensures that any active stream with the server is closed, releasing any
// resources.
func (cr *Client) Reset() {
	cr.connMu.Lock()
	defer cr.connMu.Unlock()
	cr.resetLocked()
}

func (cr *Client) resetLocked() {
	if cr.stream != nil {
		cr.stream.CloseSend()
		for {
			// drain the stream, this covers io.EOF too
			if _, err := cr.stream.Recv(); err != nil {
				break
			}
		}
		cr.stream = nil
	}
	if cr.cancel != nil {
		cr.cancel()
		cr.cancel = nil
	}
}

// ResolveService asks the server to resolve the given fully-qualified service
// name into a service descriptor.
func (cr *Client) ResolveService(serviceName string) (*desc.ServiceDescriptor, error) {
	file, err := cr.FileContainingSymbol(serviceName)
	if err != nil {
		return nil, setSymbolType(err, serviceName, symbolTypeService)
	}
	d := file.FindSymbol(serviceName)
	if d == nil {
		return nil, symbolNotFound(serviceName, symbolTypeService, nil)
	}
	if s, ok := d.(*desc.ServiceDescriptor); ok {
		return s, nil
	} else {
		return nil, symbolNotFound(serviceName, symbolTypeService, nil)
	}
}

// ResolveMessage asks the server to resolve the given fully-qualified message
// name into a message descriptor.
func (cr *Client) ResolveMessage(messageName string) (*desc.MessageDescriptor, error) {
	file, err := cr.FileContainingSymbol(messageName)
	if err != nil {
		return nil, setSymbolType(err, messageName, symbolTypeMessage)
	}
	d := file.FindSymbol(messageName)
	if d == nil {
		return nil, symbolNotFound(messageName, symbolTypeMessage, nil)
	}
	if s, ok := d.(*desc.MessageDescriptor); ok {
		return s, nil
	} else {
		return nil, symbolNotFound(messageName, symbolTypeMessage, nil)
	}
}

// ResolveEnum asks the server to resolve the given fully-qualified enum name
// into an enum descriptor.
func (cr *Client) ResolveEnum(enumName string) (*desc.EnumDescriptor, error) {
	file, err := cr.FileContainingSymbol(enumName)
	if err != nil {
		return nil, setSymbolType(err, enumName, symbolTypeEnum)
	}
	d := file.FindSymbol(enumName)
	if d == nil {
		return nil, symbolNotFound(enumName, symbolTypeEnum, nil)
	}
	if s, ok := d.(*desc.EnumDescriptor); ok {
		return s, nil
	} else {
		return nil, symbolNotFound(enumName, symbolTypeEnum, nil)
	}
}

func setSymbolType(err error, name string, symType symbolType) error {
	if e, ok := err.(*elementNotFoundError); ok {
		if e.kind == elementKindSymbol && e.name == name && e.symType == symbolTypeUnknown {
			e.symType = symType
		}
	}
	return err
}

// ResolveEnumValues asks the server to resolve the given fully-qualified enum
// name into a map of names to numbers that represents the enum's values.
func (cr *Client) ResolveEnumValues(enumName string) (map[string]int32, error) {
	enumDesc, err := cr.ResolveEnum(enumName)
	if err != nil {
		return nil, err
	}
	vals := map[string]int32{}
	for _, valDesc := range enumDesc.GetValues() {
		vals[valDesc.GetName()] = valDesc.GetNumber()
	}
	return vals, nil
}

// ResolveExtension asks the server to resolve the given extension number and
// fully-qualified message name into a field descriptor.
func (cr *Client) ResolveExtension(extendedType string, extensionNumber int32) (*desc.FieldDescriptor, error) {
	file, err := cr.FileContainingExtension(extendedType, extensionNumber)
	if err != nil {
		return nil, err
	}
	d := findExtension(extendedType, extensionNumber, fileDescriptorExtensions{file})
	if d == nil {
		return nil, extensionNotFound(extendedType, extensionNumber, nil)
	} else {
		return d, nil
	}
}

func findExtension(extendedType string, extensionNumber int32, scope extensionScope) *desc.FieldDescriptor {
	// search extensions in this scope
	for _, ext := range scope.extensions() {
		if ext.GetNumber() == extensionNumber && ext.GetOwner().GetFullyQualifiedName() == extendedType {
			return ext
		}
	}

	// if not found, search nested scopes
	for _, nested := range scope.nestedScopes() {
		ext := findExtension(extendedType, extensionNumber, nested)
		if ext != nil {
			return ext
		}
	}

	return nil
}

type extensionScope interface {
	extensions() []*desc.FieldDescriptor
	nestedScopes() []extensionScope
}

// fileDescriptorExtensions implements extensionHolder interface on top of
// FileDescriptorProto
type fileDescriptorExtensions struct {
	proto *desc.FileDescriptor
}

func (fde fileDescriptorExtensions) extensions() []*desc.FieldDescriptor {
	return fde.proto.GetExtensions()
}

func (fde fileDescriptorExtensions) nestedScopes() []extensionScope {
	scopes := make([]extensionScope, len(fde.proto.GetMessageTypes()))
	for i, m := range fde.proto.GetMessageTypes() {
		scopes[i] = msgDescriptorExtensions{m}
	}
	return scopes
}

// msgDescriptorExtensions implements extensionHolder interface on top of
// DescriptorProto
type msgDescriptorExtensions struct {
	proto *desc.MessageDescriptor
}

func (mde msgDescriptorExtensions) extensions() []*desc.FieldDescriptor {
	return mde.proto.GetNestedExtensions()
}

func (mde msgDescriptorExtensions) nestedScopes() []extensionScope {
	scopes := make([]extensionScope, len(mde.proto.GetNestedMessageTypes()))
	for i, m := range mde.proto.GetNestedMessageTypes() {
		scopes[i] = msgDescriptorExtensions{m}
	}
	return scopes
}
