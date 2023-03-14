/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

/*

Package binding defines interfaces for protocol bindings.

NOTE: Most applications that emit or consume events should use the ../client
package, which provides a simpler API to the underlying binding.

The interfaces in this package provide extra encoding and protocol information
to allow efficient forwarding and end-to-end reliable delivery between a
Receiver and a Sender belonging to different bindings. This is useful for
intermediary applications that route or forward events, but not necessary for
most "endpoint" applications that emit or consume events.

Protocol Bindings

A protocol binding usually implements a Message, a Sender and Receiver, a StructuredWriter and a BinaryWriter (depending on the supported encodings of the protocol) and an Write[ProtocolMessage] method.

Read and write events

The core of this package is the binding.Message interface.
Through binding.MessageReader It defines how to read a protocol specific message for an
encoded event in structured mode or binary mode.
The entity who receives a protocol specific data structure representing a message
(e.g. an HttpRequest) encapsulates it in a binding.Message implementation using a NewMessage method (e.g. http.NewMessage).
Then the entity that wants to send the binding.Message back on the wire,
translates it back to the protocol specific data structure (e.g. a Kafka ConsumerMessage), using
the writers BinaryWriter and StructuredWriter specific to that protocol.
Binding implementations exposes their writers
through a specific Write[ProtocolMessage] function (e.g. kafka.EncodeProducerMessage),
in order to simplify the encoding process.

The encoding process can be customized in order to mutate the final result with binding.TransformerFactory.
A bunch of these are provided directly by the binding/transformer module.

Usually binding.Message implementations can be encoded only one time, because the encoding process drain the message itself.
In order to consume a message several times, the binding/buffering package provides several APIs to buffer the Message.

A message can be converted to an event.Event using binding.ToEvent() method.
An event.Event can be used as Message casting it to binding.EventMessage.

In order to simplify the encoding process for each protocol, this package provide several utility methods like binding.Write and binding.DirectWrite.
The binding.Write method tries to preserve the structured/binary encoding, in order to be as much efficient as possible.

Messages can be eventually wrapped to change their behaviours and binding their lifecycle, like the binding.FinishMessage.
Every Message wrapper implements the MessageWrapper interface

Sender and Receiver

A Receiver receives protocol specific messages and wraps them to into binding.Message implementations.

A Sender converts arbitrary Message implementations to a protocol-specific form using the protocol specific Write method
and sends them.

Message and ExactlyOnceMessage provide methods to allow acknowledgments to
propagate when a reliable messages is forwarded from a Receiver to a Sender.
QoS 0 (unreliable), 1 (at-least-once) and 2 (exactly-once) are supported.

Transport

A binding implementation providing Sender and Receiver implementations can be used as a Transport through the BindingTransport adapter.

*/
package binding
