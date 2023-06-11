package grpcurl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/golang/protobuf/jsonpb" //lint:ignore SA1019 we have to import this because it appears in exported API
	"github.com/golang/protobuf/proto"  //lint:ignore SA1019 we have to import this because it appears in exported API
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// InvocationEventHandler is a bag of callbacks for handling events that occur in the course
// of invoking an RPC. The handler also provides request data that is sent. The callbacks are
// generally called in the order they are listed below.
type InvocationEventHandler interface {
	// OnResolveMethod is called with a descriptor of the method that is being invoked.
	OnResolveMethod(*desc.MethodDescriptor)
	// OnSendHeaders is called with the request metadata that is being sent.
	OnSendHeaders(metadata.MD)
	// OnReceiveHeaders is called when response headers have been received.
	OnReceiveHeaders(metadata.MD)
	// OnReceiveResponse is called for each response message received.
	OnReceiveResponse(proto.Message)
	// OnReceiveTrailers is called when response trailers and final RPC status have been received.
	OnReceiveTrailers(*status.Status, metadata.MD)
}

// RequestMessageSupplier is a function that is called to retrieve request
// messages for a GRPC operation. This type is deprecated and will be removed in
// a future release.
//
// Deprecated: This is only used with the deprecated InvokeRpc. Instead, use
// RequestSupplier with InvokeRPC.
type RequestMessageSupplier func() ([]byte, error)

// InvokeRpc uses the given gRPC connection to invoke the given method. This function is deprecated
// and will be removed in a future release. It just delegates to the similarly named InvokeRPC
// method, whose signature is only slightly different.
//
// Deprecated: use InvokeRPC instead.
func InvokeRpc(ctx context.Context, source DescriptorSource, cc *grpc.ClientConn, methodName string,
	headers []string, handler InvocationEventHandler, requestData RequestMessageSupplier) error {

	return InvokeRPC(ctx, source, cc, methodName, headers, handler, func(m proto.Message) error {
		// New function is almost identical, but the request supplier function works differently.
		// So we adapt the logic here to maintain compatibility.
		data, err := requestData()
		if err != nil {
			return err
		}
		return jsonpb.Unmarshal(bytes.NewReader(data), m)
	})
}

// RequestSupplier is a function that is called to populate messages for a gRPC operation. The
// function should populate the given message or return a non-nil error. If the supplier has no
// more messages, it should return io.EOF. When it returns io.EOF, it should not in any way
// modify the given message argument.
type RequestSupplier func(proto.Message) error

// InvokeRPC uses the given gRPC channel to invoke the given method. The given descriptor source
// is used to determine the type of method and the type of request and response message. The given
// headers are sent as request metadata. Methods on the given event handler are called as the
// invocation proceeds.
//
// The given requestData function supplies the actual data to send. It should return io.EOF when
// there is no more request data. If the method being invoked is a unary or server-streaming RPC
// (e.g. exactly one request message) and there is no request data (e.g. the first invocation of
// the function returns io.EOF), then an empty request message is sent.
//
// If the requestData function and the given event handler coordinate or share any state, they should
// be thread-safe. This is because the requestData function may be called from a different goroutine
// than the one invoking event callbacks. (This only happens for bi-directional streaming RPCs, where
// one goroutine sends request messages and another consumes the response messages).
func InvokeRPC(ctx context.Context, source DescriptorSource, ch grpcdynamic.Channel, methodName string,
	headers []string, handler InvocationEventHandler, requestData RequestSupplier) error {

	md := MetadataFromHeaders(headers)

	svc, mth := parseSymbol(methodName)
	if svc == "" || mth == "" {
		return fmt.Errorf("given method name %q is not in expected format: 'service/method' or 'service.method'", methodName)
	}

	dsc, err := source.FindSymbol(svc)
	if err != nil {
		// return a gRPC status error if hasStatus is true
		errStatus, hasStatus := status.FromError(err)
		switch {
		case hasStatus && isNotFoundError(err):
			return status.Errorf(errStatus.Code(), "target server does not expose service %q: %s", svc, errStatus.Message())
		case hasStatus:
			return status.Errorf(errStatus.Code(), "failed to query for service descriptor %q: %s", svc, errStatus.Message())
		case isNotFoundError(err):
			return fmt.Errorf("target server does not expose service %q", svc)
		}
		return fmt.Errorf("failed to query for service descriptor %q: %v", svc, err)
	}
	sd, ok := dsc.(*desc.ServiceDescriptor)
	if !ok {
		return fmt.Errorf("target server does not expose service %q", svc)
	}
	mtd := sd.FindMethodByName(mth)
	if mtd == nil {
		return fmt.Errorf("service %q does not include a method named %q", svc, mth)
	}

	handler.OnResolveMethod(mtd)

	// we also download any applicable extensions so we can provide full support for parsing user-provided data
	var ext dynamic.ExtensionRegistry
	alreadyFetched := map[string]bool{}
	if err = fetchAllExtensions(source, &ext, mtd.GetInputType(), alreadyFetched); err != nil {
		return fmt.Errorf("error resolving server extensions for message %s: %v", mtd.GetInputType().GetFullyQualifiedName(), err)
	}
	if err = fetchAllExtensions(source, &ext, mtd.GetOutputType(), alreadyFetched); err != nil {
		return fmt.Errorf("error resolving server extensions for message %s: %v", mtd.GetOutputType().GetFullyQualifiedName(), err)
	}

	msgFactory := dynamic.NewMessageFactoryWithExtensionRegistry(&ext)
	req := msgFactory.NewMessage(mtd.GetInputType())

	handler.OnSendHeaders(md)
	ctx = metadata.NewOutgoingContext(ctx, md)

	stub := grpcdynamic.NewStubWithMessageFactory(ch, msgFactory)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if mtd.IsClientStreaming() && mtd.IsServerStreaming() {
		return invokeBidi(ctx, stub, mtd, handler, requestData, req)
	} else if mtd.IsClientStreaming() {
		return invokeClientStream(ctx, stub, mtd, handler, requestData, req)
	} else if mtd.IsServerStreaming() {
		return invokeServerStream(ctx, stub, mtd, handler, requestData, req)
	} else {
		return invokeUnary(ctx, stub, mtd, handler, requestData, req)
	}
}

func invokeUnary(ctx context.Context, stub grpcdynamic.Stub, md *desc.MethodDescriptor, handler InvocationEventHandler,
	requestData RequestSupplier, req proto.Message) error {

	err := requestData(req)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error getting request data: %v", err)
	}
	if err != io.EOF {
		// verify there is no second message, which is a usage error
		err := requestData(req)
		if err == nil {
			return fmt.Errorf("method %q is a unary RPC, but request data contained more than 1 message", md.GetFullyQualifiedName())
		} else if err != io.EOF {
			return fmt.Errorf("error getting request data: %v", err)
		}
	}

	// Now we can actually invoke the RPC!
	var respHeaders metadata.MD
	var respTrailers metadata.MD
	resp, err := stub.InvokeRpc(ctx, md, req, grpc.Trailer(&respTrailers), grpc.Header(&respHeaders))

	stat, ok := status.FromError(err)
	if !ok {
		// Error codes sent from the server will get printed differently below.
		// So just bail for other kinds of errors here.
		return fmt.Errorf("grpc call for %q failed: %v", md.GetFullyQualifiedName(), err)
	}

	handler.OnReceiveHeaders(respHeaders)

	if stat.Code() == codes.OK {
		handler.OnReceiveResponse(resp)
	}

	handler.OnReceiveTrailers(stat, respTrailers)

	return nil
}

func invokeClientStream(ctx context.Context, stub grpcdynamic.Stub, md *desc.MethodDescriptor, handler InvocationEventHandler,
	requestData RequestSupplier, req proto.Message) error {

	// invoke the RPC!
	str, err := stub.InvokeRpcClientStream(ctx, md)

	// Upload each request message in the stream
	var resp proto.Message
	for err == nil {
		err = requestData(req)
		if err == io.EOF {
			resp, err = str.CloseAndReceive()
			break
		}
		if err != nil {
			return fmt.Errorf("error getting request data: %v", err)
		}

		err = str.SendMsg(req)
		if err == io.EOF {
			// We get EOF on send if the server says "go away"
			// We have to use CloseAndReceive to get the actual code
			resp, err = str.CloseAndReceive()
			break
		}

		req.Reset()
	}

	// finally, process response data
	stat, ok := status.FromError(err)
	if !ok {
		// Error codes sent from the server will get printed differently below.
		// So just bail for other kinds of errors here.
		return fmt.Errorf("grpc call for %q failed: %v", md.GetFullyQualifiedName(), err)
	}

	if str != nil {
		if respHeaders, err := str.Header(); err == nil {
			handler.OnReceiveHeaders(respHeaders)
		}
	}

	if stat.Code() == codes.OK {
		handler.OnReceiveResponse(resp)
	}

	if str != nil {
		handler.OnReceiveTrailers(stat, str.Trailer())
	}

	return nil
}

func invokeServerStream(ctx context.Context, stub grpcdynamic.Stub, md *desc.MethodDescriptor, handler InvocationEventHandler,
	requestData RequestSupplier, req proto.Message) error {

	err := requestData(req)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error getting request data: %v", err)
	}
	if err != io.EOF {
		// verify there is no second message, which is a usage error
		err := requestData(req)
		if err == nil {
			return fmt.Errorf("method %q is a server-streaming RPC, but request data contained more than 1 message", md.GetFullyQualifiedName())
		} else if err != io.EOF {
			return fmt.Errorf("error getting request data: %v", err)
		}
	}

	// Now we can actually invoke the RPC!
	str, err := stub.InvokeRpcServerStream(ctx, md, req)

	if respHeaders, err := str.Header(); err == nil {
		handler.OnReceiveHeaders(respHeaders)
	}

	// Download each response message
	for err == nil {
		var resp proto.Message
		resp, err = str.RecvMsg()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		handler.OnReceiveResponse(resp)
	}

	stat, ok := status.FromError(err)
	if !ok {
		// Error codes sent from the server will get printed differently below.
		// So just bail for other kinds of errors here.
		return fmt.Errorf("grpc call for %q failed: %v", md.GetFullyQualifiedName(), err)
	}

	handler.OnReceiveTrailers(stat, str.Trailer())

	return nil
}

func invokeBidi(ctx context.Context, stub grpcdynamic.Stub, md *desc.MethodDescriptor, handler InvocationEventHandler,
	requestData RequestSupplier, req proto.Message) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// invoke the RPC!
	str, err := stub.InvokeRpcBidiStream(ctx, md)

	var wg sync.WaitGroup
	var sendErr atomic.Value

	defer wg.Wait()

	if err == nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Concurrently upload each request message in the stream
			var err error
			for err == nil {
				err = requestData(req)

				if err == io.EOF {
					err = str.CloseSend()
					break
				}
				if err != nil {
					err = fmt.Errorf("error getting request data: %v", err)
					cancel()
					break
				}

				err = str.SendMsg(req)

				req.Reset()
			}

			if err != nil {
				sendErr.Store(err)
			}
		}()
	}

	if str != nil {
		if respHeaders, err := str.Header(); err == nil {
			handler.OnReceiveHeaders(respHeaders)
		}
	}

	// Download each response message
	for err == nil {
		var resp proto.Message
		resp, err = str.RecvMsg()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		handler.OnReceiveResponse(resp)
	}

	if se, ok := sendErr.Load().(error); ok && se != io.EOF {
		err = se
	}

	stat, ok := status.FromError(err)
	if !ok {
		// Error codes sent from the server will get printed differently below.
		// So just bail for other kinds of errors here.
		return fmt.Errorf("grpc call for %q failed: %v", md.GetFullyQualifiedName(), err)
	}

	if str != nil {
		handler.OnReceiveTrailers(stat, str.Trailer())
	}

	return nil
}

type notFoundError string

func notFound(kind, name string) error {
	return notFoundError(fmt.Sprintf("%s not found: %s", kind, name))
}

func (e notFoundError) Error() string {
	return string(e)
}

func isNotFoundError(err error) bool {
	if grpcreflect.IsElementNotFoundError(err) {
		return true
	}
	_, ok := err.(notFoundError)
	return ok
}

func parseSymbol(svcAndMethod string) (string, string) {
	pos := strings.LastIndex(svcAndMethod, "/")
	if pos < 0 {
		pos = strings.LastIndex(svcAndMethod, ".")
		if pos < 0 {
			return "", ""
		}
	}
	return svcAndMethod[:pos], svcAndMethod[pos+1:]
}
