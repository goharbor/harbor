/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package binding

import "github.com/cloudevents/sdk-go/v2/binding/spec"

type finishMessage struct {
	Message
	finish func(error)
}

func (m *finishMessage) GetAttribute(k spec.Kind) (spec.Attribute, interface{}) {
	return m.Message.(MessageMetadataReader).GetAttribute(k)
}

func (m *finishMessage) GetExtension(s string) interface{} {
	return m.Message.(MessageMetadataReader).GetExtension(s)
}

func (m *finishMessage) GetWrappedMessage() Message {
	return m.Message
}

func (m *finishMessage) Finish(err error) error {
	err2 := m.Message.Finish(err) // Finish original message first
	if m.finish != nil {
		m.finish(err) // Notify callback
	}
	return err2
}

var _ MessageWrapper = (*finishMessage)(nil)

// WithFinish returns a wrapper for m that calls finish() and
// m.Finish() in its Finish().
// Allows code to be notified when a message is Finished.
func WithFinish(m Message, finish func(error)) Message {
	return &finishMessage{Message: m, finish: finish}
}
