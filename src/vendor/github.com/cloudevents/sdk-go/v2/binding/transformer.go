/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package binding

// Transformer is an interface that implements a transformation
// process while transferring the event from the Message
// implementation to the provided encoder
//
// When a write function (binding.Write, binding.ToEvent, buffering.CopyMessage, etc.)
// takes Transformer(s) as parameter, it eventually converts the message to a form
// which correctly implements MessageMetadataReader, in order to guarantee that transformation
// is applied
type Transformer interface {
	Transform(MessageMetadataReader, MessageMetadataWriter) error
}

// TransformerFunc is a type alias to implement a Transformer through a function pointer
type TransformerFunc func(MessageMetadataReader, MessageMetadataWriter) error

func (t TransformerFunc) Transform(r MessageMetadataReader, w MessageMetadataWriter) error {
	return t(r, w)
}

var _ Transformer = (TransformerFunc)(nil)

// Transformers is a utility alias to run several Transformer
type Transformers []Transformer

func (t Transformers) Transform(r MessageMetadataReader, w MessageMetadataWriter) error {
	for _, transformer := range t {
		err := transformer.Transform(r, w)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ Transformer = (Transformers)(nil)
