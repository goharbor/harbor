// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package semconv // import "go.opentelemetry.io/otel/semconv/v1.4.0"

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
<<<<<<< HEAD
	"go.opentelemetry.io/otel/semconv/internal"
=======
>>>>>>> 40ba15ca5a97e1a0c8cd3afebd03f2ab8596069c
	"go.opentelemetry.io/otel/trace"
)

// HTTP scheme attributes.
var (
	HTTPSchemeHTTP  = HTTPSchemeKey.String("http")
	HTTPSchemeHTTPS = HTTPSchemeKey.String("https")
)

var sc = &internal.SemanticConventions{
	EnduserIDKey:                EnduserIDKey,
	HTTPClientIPKey:             HTTPClientIPKey,
	HTTPFlavorKey:               HTTPFlavorKey,
	HTTPHostKey:                 HTTPHostKey,
	HTTPMethodKey:               HTTPMethodKey,
	HTTPRequestContentLengthKey: HTTPRequestContentLengthKey,
	HTTPRouteKey:                HTTPRouteKey,
	HTTPSchemeHTTP:              HTTPSchemeHTTP,
	HTTPSchemeHTTPS:             HTTPSchemeHTTPS,
	HTTPServerNameKey:           HTTPServerNameKey,
	HTTPStatusCodeKey:           HTTPStatusCodeKey,
	HTTPTargetKey:               HTTPTargetKey,
	HTTPURLKey:                  HTTPURLKey,
	HTTPUserAgentKey:            HTTPUserAgentKey,
	NetHostIPKey:                NetHostIPKey,
	NetHostNameKey:              NetHostNameKey,
	NetHostPortKey:              NetHostPortKey,
	NetPeerIPKey:                NetPeerIPKey,
	NetPeerNameKey:              NetPeerNameKey,
	NetPeerPortKey:              NetPeerPortKey,
	NetTransportIP:              NetTransportIP,
	NetTransportOther:           NetTransportOther,
	NetTransportTCP:             NetTransportTCP,
	NetTransportUDP:             NetTransportUDP,
	NetTransportUnix:            NetTransportUnix,
}

// NetAttributesFromHTTPRequest generates attributes of the net
// namespace as specified by the OpenTelemetry specification for a
// span.  The network parameter is a string that net.Dial function
// from standard library can understand.
func NetAttributesFromHTTPRequest(network string, request *http.Request) []attribute.KeyValue {
<<<<<<< HEAD
	return sc.NetAttributesFromHTTPRequest(network, request)
=======
	attrs := []attribute.KeyValue{}

	switch network {
	case "tcp", "tcp4", "tcp6":
		attrs = append(attrs, NetTransportTCP)
	case "udp", "udp4", "udp6":
		attrs = append(attrs, NetTransportUDP)
	case "ip", "ip4", "ip6":
		attrs = append(attrs, NetTransportIP)
	case "unix", "unixgram", "unixpacket":
		attrs = append(attrs, NetTransportUnix)
	default:
		attrs = append(attrs, NetTransportOther)
	}

	peerIP, peerName, peerPort := hostIPNamePort(request.RemoteAddr)
	if peerIP != "" {
		attrs = append(attrs, NetPeerIPKey.String(peerIP))
	}
	if peerName != "" {
		attrs = append(attrs, NetPeerNameKey.String(peerName))
	}
	if peerPort != 0 {
		attrs = append(attrs, NetPeerPortKey.Int(peerPort))
	}

	hostIP, hostName, hostPort := "", "", 0
	for _, someHost := range []string{request.Host, request.Header.Get("Host"), request.URL.Host} {
		hostIP, hostName, hostPort = hostIPNamePort(someHost)
		if hostIP != "" || hostName != "" || hostPort != 0 {
			break
		}
	}
	if hostIP != "" {
		attrs = append(attrs, NetHostIPKey.String(hostIP))
	}
	if hostName != "" {
		attrs = append(attrs, NetHostNameKey.String(hostName))
	}
	if hostPort != 0 {
		attrs = append(attrs, NetHostPortKey.Int(hostPort))
	}

	return attrs
>>>>>>> 40ba15ca5a97e1a0c8cd3afebd03f2ab8596069c
}

// hostIPNamePort extracts the IP address, name and (optional) port from hostWithPort.
// It handles both IPv4 and IPv6 addresses. If the host portion is not recognized
// as a valid IPv4 or IPv6 address, the `ip` result will be empty and the
// host portion will instead be returned in `name`.
func hostIPNamePort(hostWithPort string) (ip string, name string, port int) {
	var (
		hostPart, portPart string
		parsedPort         uint64
		err                error
	)
	if hostPart, portPart, err = net.SplitHostPort(hostWithPort); err != nil {
		hostPart, portPart = hostWithPort, ""
	}
	if parsedIP := net.ParseIP(hostPart); parsedIP != nil {
		ip = parsedIP.String()
	} else {
		name = hostPart
	}
	if parsedPort, err = strconv.ParseUint(portPart, 10, 16); err == nil {
		port = int(parsedPort)
	}
	return
}

// EndUserAttributesFromHTTPRequest generates attributes of the
// enduser namespace as specified by the OpenTelemetry specification
// for a span.
func EndUserAttributesFromHTTPRequest(request *http.Request) []attribute.KeyValue {
	return sc.EndUserAttributesFromHTTPRequest(request)
}

// HTTPClientAttributesFromHTTPRequest generates attributes of the
// http namespace as specified by the OpenTelemetry specification for
// a span on the client side.
func HTTPClientAttributesFromHTTPRequest(request *http.Request) []attribute.KeyValue {
	return sc.HTTPClientAttributesFromHTTPRequest(request)
}

// HTTPServerMetricAttributesFromHTTPRequest generates low-cardinality attributes
// to be used with server-side HTTP metrics.
func HTTPServerMetricAttributesFromHTTPRequest(serverName string, request *http.Request) []attribute.KeyValue {
	return sc.HTTPServerMetricAttributesFromHTTPRequest(serverName, request)
}

// HTTPServerAttributesFromHTTPRequest generates attributes of the
// http namespace as specified by the OpenTelemetry specification for
// a span on the server side. Currently, only basic authentication is
// supported.
func HTTPServerAttributesFromHTTPRequest(serverName, route string, request *http.Request) []attribute.KeyValue {
<<<<<<< HEAD
	return sc.HTTPServerAttributesFromHTTPRequest(serverName, route, request)
=======
	attrs := []attribute.KeyValue{
		HTTPMethodKey.String(request.Method),
		HTTPTargetKey.String(request.RequestURI),
	}

	if serverName != "" {
		attrs = append(attrs, HTTPServerNameKey.String(serverName))
	}
	if route != "" {
		attrs = append(attrs, HTTPRouteKey.String(route))
	}
	if values, ok := request.Header["X-Forwarded-For"]; ok && len(values) > 0 {
		if addresses := strings.SplitN(values[0], ",", 2); len(addresses) > 0 {
			attrs = append(attrs, HTTPClientIPKey.String(addresses[0]))
		}
	}

	return append(attrs, httpCommonAttributesFromHTTPRequest(request)...)
>>>>>>> 40ba15ca5a97e1a0c8cd3afebd03f2ab8596069c
}

// HTTPAttributesFromHTTPStatusCode generates attributes of the http
// namespace as specified by the OpenTelemetry specification for a
// span.
func HTTPAttributesFromHTTPStatusCode(code int) []attribute.KeyValue {
	return sc.HTTPAttributesFromHTTPStatusCode(code)
}

// SpanStatusFromHTTPStatusCode generates a status code and a message
// as specified by the OpenTelemetry specification for a span.
func SpanStatusFromHTTPStatusCode(code int) (codes.Code, string) {
	return internal.SpanStatusFromHTTPStatusCode(code)
}

// SpanStatusFromHTTPStatusCodeAndSpanKind generates a status code and a message
// as specified by the OpenTelemetry specification for a span.
// Exclude 4xx for SERVER to set the appropriate status.
func SpanStatusFromHTTPStatusCodeAndSpanKind(code int, spanKind trace.SpanKind) (codes.Code, string) {
<<<<<<< HEAD
	return internal.SpanStatusFromHTTPStatusCodeAndSpanKind(code, spanKind)
=======
	spanCode, valid := validateHTTPStatusCode(code)
	if !valid {
		return spanCode, fmt.Sprintf("Invalid HTTP status code %d", code)
	}
	category := code / 100
	if spanKind == trace.SpanKindServer && category == 4 {
		return codes.Unset, ""
	}
	return spanCode, ""
}

// Validates the HTTP status code and returns corresponding span status code.
// If the `code` is not a valid HTTP status code, returns span status Error
// and false.
func validateHTTPStatusCode(code int) (codes.Code, bool) {
	category := code / 100
	ranges, ok := validRangesPerCategory[category]
	if !ok {
		return codes.Error, false
	}
	ok = false
	for _, crange := range ranges {
		ok = crange.contains(code)
		if ok {
			break
		}
	}
	if !ok {
		return codes.Error, false
	}
	if category > 0 && category < 4 {
		return codes.Unset, true
	}
	return codes.Error, true
>>>>>>> 40ba15ca5a97e1a0c8cd3afebd03f2ab8596069c
}
