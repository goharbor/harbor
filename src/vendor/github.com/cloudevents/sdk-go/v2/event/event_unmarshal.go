/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"

	jsoniter "github.com/json-iterator/go"

	"github.com/cloudevents/sdk-go/v2/types"
)

const specVersionV03Flag uint8 = 1 << 4
const specVersionV1Flag uint8 = 1 << 5
const dataBase64Flag uint8 = 1 << 6
const dataContentTypeFlag uint8 = 1 << 7

func checkFlag(state uint8, flag uint8) bool {
	return state&flag != 0
}

func appendFlag(state *uint8, flag uint8) {
	*state = (*state) | flag
}

var iterPool = sync.Pool{
	New: func() interface{} {
		return jsoniter.Parse(jsoniter.ConfigFastest, nil, 1024)
	},
}

func borrowIterator(reader io.Reader) *jsoniter.Iterator {
	iter := iterPool.Get().(*jsoniter.Iterator)
	iter.Reset(reader)
	return iter
}

func returnIterator(iter *jsoniter.Iterator) {
	iter.Error = nil
	iter.Attachment = nil
	iterPool.Put(iter)
}

func ReadJson(out *Event, reader io.Reader) error {
	iterator := borrowIterator(reader)
	defer returnIterator(iterator)

	return readJsonFromIterator(out, iterator)
}

// ReadJson allows you to read the bytes reader as an event
func readJsonFromIterator(out *Event, iterator *jsoniter.Iterator) error {
	// Parsing dependency graph:
	//         SpecVersion
	//          ^     ^
	//          |     +--------------+
	//          +                    +
	//  All Attributes           datacontenttype (and datacontentencoding for v0.3)
	//  (except datacontenttype)     ^
	//                               |
	//                               |
	//                               +
	//                              Data

	var state uint8 = 0
	var cachedData []byte

	var (
		// Universally parseable fields.
		id              string
		typ             string
		source          types.URIRef
		subject         *string
		time            *types.Timestamp
		datacontenttype *string
		extensions      = make(map[string]interface{})

		// These fields require knowledge about the specversion to be parsed.
		schemaurl           jsoniter.Any
		datacontentencoding jsoniter.Any
		dataschema          jsoniter.Any
		dataBase64          jsoniter.Any
	)

	for key := iterator.ReadObject(); key != ""; key = iterator.ReadObject() {
		// Check if we have some error in our error cache
		if iterator.Error != nil {
			return iterator.Error
		}

		// We have a key, now we need to figure out what to do
		// depending on the parsing state

		// If it's a specversion, trigger state change
		if key == "specversion" {
			if checkFlag(state, specVersionV1Flag|specVersionV03Flag) {
				return fmt.Errorf("specversion was already provided")
			}
			sv := iterator.ReadString()

			// Check proper specversion
			switch sv {
			case CloudEventsVersionV1:
				con := &EventContextV1{
					ID:              id,
					Type:            typ,
					Source:          source,
					Subject:         subject,
					Time:            time,
					DataContentType: datacontenttype,
				}

				// Add the fields relevant for the version ...
				if dataschema != nil {
					var err error
					con.DataSchema, err = toUriPtr(dataschema)
					if err != nil {
						return err
					}
				}
				if dataBase64 != nil {
					stream := jsoniter.ConfigFastest.BorrowStream(nil)
					defer jsoniter.ConfigFastest.ReturnStream(stream)
					dataBase64.WriteTo(stream)
					cachedData = stream.Buffer()
					if stream.Error != nil {
						return stream.Error
					}
					appendFlag(&state, dataBase64Flag)
				}

				// ... add all remaining fields as extensions.
				if schemaurl != nil {
					extensions["schemaurl"] = schemaurl.GetInterface()
				}
				if datacontentencoding != nil {
					extensions["datacontentencoding"] = datacontentencoding.GetInterface()
				}

				out.Context = con
				appendFlag(&state, specVersionV1Flag)
			case CloudEventsVersionV03:
				con := &EventContextV03{
					ID:              id,
					Type:            typ,
					Source:          source,
					Subject:         subject,
					Time:            time,
					DataContentType: datacontenttype,
				}
				var err error
				// Add the fields relevant for the version ...
				if schemaurl != nil {
					con.SchemaURL, err = toUriRefPtr(schemaurl)
					if err != nil {
						return err
					}
				}
				if datacontentencoding != nil {
					con.DataContentEncoding, err = toStrPtr(datacontentencoding)
					if *con.DataContentEncoding != Base64 {
						err = ValidationError{"datacontentencoding": errors.New("invalid datacontentencoding value, the only allowed value is 'base64'")}
					}
					if err != nil {
						return err
					}
					appendFlag(&state, dataBase64Flag)
				}

				// ... add all remaining fields as extensions.
				if dataschema != nil {
					extensions["dataschema"] = dataschema.GetInterface()
				}
				if dataBase64 != nil {
					extensions["data_base64"] = dataBase64.GetInterface()
				}

				out.Context = con
				appendFlag(&state, specVersionV03Flag)
			default:
				return ValidationError{"specversion": errors.New("unknown value: " + sv)}
			}

			// Apply all extensions to the context object.
			for key, val := range extensions {
				if err := out.Context.SetExtension(key, val); err != nil {
					return err
				}
			}
			continue
		}

		// If no specversion ...
		if !checkFlag(state, specVersionV03Flag|specVersionV1Flag) {
			switch key {
			case "id":
				id = iterator.ReadString()
			case "type":
				typ = iterator.ReadString()
			case "source":
				source = readUriRef(iterator)
			case "subject":
				subject = readStrPtr(iterator)
			case "time":
				time = readTimestamp(iterator)
			case "datacontenttype":
				datacontenttype = readStrPtr(iterator)
				appendFlag(&state, dataContentTypeFlag)
			case "data":
				cachedData = iterator.SkipAndReturnBytes()
			case "data_base64":
				dataBase64 = iterator.ReadAny()
			case "dataschema":
				dataschema = iterator.ReadAny()
			case "schemaurl":
				schemaurl = iterator.ReadAny()
			case "datacontentencoding":
				datacontentencoding = iterator.ReadAny()
			default:
				extensions[key] = iterator.Read()
			}
			continue
		}

		// From this point downward -> we can assume the event has a context pointer non nil

		// If it's a datacontenttype, trigger state change
		if key == "datacontenttype" {
			if checkFlag(state, dataContentTypeFlag) {
				return fmt.Errorf("datacontenttype was already provided")
			}

			dct := iterator.ReadString()

			switch ctx := out.Context.(type) {
			case *EventContextV03:
				ctx.DataContentType = &dct
			case *EventContextV1:
				ctx.DataContentType = &dct
			}
			appendFlag(&state, dataContentTypeFlag)
			continue
		}

		// If it's a datacontentencoding and it's v0.3, trigger state change
		if checkFlag(state, specVersionV03Flag) && key == "datacontentencoding" {
			if checkFlag(state, dataBase64Flag) {
				return ValidationError{"datacontentencoding": errors.New("datacontentencoding was specified twice")}
			}

			dce := iterator.ReadString()

			if dce != Base64 {
				return ValidationError{"datacontentencoding": errors.New("invalid datacontentencoding value, the only allowed value is 'base64'")}
			}

			out.Context.(*EventContextV03).DataContentEncoding = &dce
			appendFlag(&state, dataBase64Flag)
			continue
		}

		// We can parse all attributes, except data.
		// If it's data or data_base64 and we don't have the attributes to process it, then we cache it
		// The expanded form of this condition is:
		// (checkFlag(state, specVersionV1Flag) && !checkFlag(state, dataContentTypeFlag) && (key == "data" || key == "data_base64")) ||
		// (checkFlag(state, specVersionV03Flag) && !(checkFlag(state, dataContentTypeFlag) && checkFlag(state, dataBase64Flag)) && key == "data")
		if (state&(specVersionV1Flag|dataContentTypeFlag) == specVersionV1Flag && (key == "data" || key == "data_base64")) ||
			((state&specVersionV03Flag == specVersionV03Flag) && (state&(dataContentTypeFlag|dataBase64Flag) != (dataContentTypeFlag | dataBase64Flag)) && key == "data") {
			if key == "data_base64" {
				appendFlag(&state, dataBase64Flag)
			}
			cachedData = iterator.SkipAndReturnBytes()
			continue
		}

		// At this point or this value is an attribute (excluding datacontenttype and datacontentencoding), or this value is data and this condition is valid:
		// (specVersionV1Flag & dataContentTypeFlag) || (specVersionV03Flag & dataContentTypeFlag & dataBase64Flag)
		switch eventContext := out.Context.(type) {
		case *EventContextV03:
			switch key {
			case "id":
				eventContext.ID = iterator.ReadString()
			case "type":
				eventContext.Type = iterator.ReadString()
			case "source":
				eventContext.Source = readUriRef(iterator)
			case "subject":
				eventContext.Subject = readStrPtr(iterator)
			case "time":
				eventContext.Time = readTimestamp(iterator)
			case "schemaurl":
				eventContext.SchemaURL = readUriRefPtr(iterator)
			case "data":
				iterator.Error = consumeData(out, checkFlag(state, dataBase64Flag), iterator)
			default:
				if eventContext.Extensions == nil {
					eventContext.Extensions = make(map[string]interface{}, 1)
				}
				iterator.Error = eventContext.SetExtension(key, iterator.Read())
			}
		case *EventContextV1:
			switch key {
			case "id":
				eventContext.ID = iterator.ReadString()
			case "type":
				eventContext.Type = iterator.ReadString()
			case "source":
				eventContext.Source = readUriRef(iterator)
			case "subject":
				eventContext.Subject = readStrPtr(iterator)
			case "time":
				eventContext.Time = readTimestamp(iterator)
			case "dataschema":
				eventContext.DataSchema = readUriPtr(iterator)
			case "data":
				iterator.Error = consumeData(out, false, iterator)
			case "data_base64":
				iterator.Error = consumeData(out, true, iterator)
			default:
				if eventContext.Extensions == nil {
					eventContext.Extensions = make(map[string]interface{}, 1)
				}
				iterator.Error = eventContext.SetExtension(key, iterator.Read())
			}
		}
	}

	if state&(specVersionV03Flag|specVersionV1Flag) == 0 {
		return ValidationError{"specversion": errors.New("no specversion")}
	}

	if iterator.Error != nil {
		return iterator.Error
	}

	// If there is a dataToken cached, we always defer at the end the processing
	// because nor datacontenttype or datacontentencoding are mandatory.
	if cachedData != nil {
		return consumeDataAsBytes(out, checkFlag(state, dataBase64Flag), cachedData)
	}
	return nil
}

func consumeDataAsBytes(e *Event, isBase64 bool, b []byte) error {
	if isBase64 {
		e.DataBase64 = true

		// Allocate payload byte buffer
		base64Encoded := b[1 : len(b)-1] // remove quotes
		e.DataEncoded = make([]byte, base64.StdEncoding.DecodedLen(len(base64Encoded)))
		length, err := base64.StdEncoding.Decode(e.DataEncoded, base64Encoded)
		if err != nil {
			return err
		}
		e.DataEncoded = e.DataEncoded[0:length]
		return nil
	}

	mt, _ := e.Context.GetDataMediaType()
	// Empty content type assumes json
	if mt != "" && mt != ApplicationJSON && mt != TextJSON {
		// If not json, then data is encoded as string
		iter := jsoniter.ParseBytes(jsoniter.ConfigFastest, b)
		src := iter.ReadString() // handles escaping
		e.DataEncoded = []byte(src)
		if iter.Error != nil {
			return fmt.Errorf("unexpected data payload for media type %q, expected a string: %w", mt, iter.Error)
		}
		return nil
	}

	e.DataEncoded = b
	return nil
}

func consumeData(e *Event, isBase64 bool, iter *jsoniter.Iterator) error {
	if isBase64 {
		e.DataBase64 = true

		// Allocate payload byte buffer
		base64Encoded := iter.ReadStringAsSlice()
		e.DataEncoded = make([]byte, base64.StdEncoding.DecodedLen(len(base64Encoded)))
		length, err := base64.StdEncoding.Decode(e.DataEncoded, base64Encoded)
		if err != nil {
			return err
		}
		e.DataEncoded = e.DataEncoded[0:length]
		return nil
	}

	mt, _ := e.Context.GetDataMediaType()
	if mt != ApplicationJSON && mt != TextJSON {
		// If not json, then data is encoded as string
		src := iter.ReadString() // handles escaping
		e.DataEncoded = []byte(src)
		if iter.Error != nil {
			return fmt.Errorf("unexpected data payload for media type %q, expected a string: %w", mt, iter.Error)
		}
		return nil
	}

	e.DataEncoded = iter.SkipAndReturnBytes()
	return nil
}

func readUriRef(iter *jsoniter.Iterator) types.URIRef {
	str := iter.ReadString()
	uriRef := types.ParseURIRef(str)
	if uriRef == nil {
		iter.Error = fmt.Errorf("cannot parse uri ref: %v", str)
		return types.URIRef{}
	}
	return *uriRef
}

func readStrPtr(iter *jsoniter.Iterator) *string {
	str := iter.ReadString()
	if str == "" {
		return nil
	}
	return &str
}

func readUriRefPtr(iter *jsoniter.Iterator) *types.URIRef {
	return types.ParseURIRef(iter.ReadString())
}

func readUriPtr(iter *jsoniter.Iterator) *types.URI {
	return types.ParseURI(iter.ReadString())
}

func readTimestamp(iter *jsoniter.Iterator) *types.Timestamp {
	t, err := types.ParseTimestamp(iter.ReadString())
	if err != nil {
		iter.Error = err
	}
	return t
}

func toStrPtr(val jsoniter.Any) (*string, error) {
	str := val.ToString()
	if val.LastError() != nil {
		return nil, val.LastError()
	}
	if str == "" {
		return nil, nil
	}
	return &str, nil
}

func toUriRefPtr(val jsoniter.Any) (*types.URIRef, error) {
	str := val.ToString()
	if val.LastError() != nil {
		return nil, val.LastError()
	}
	return types.ParseURIRef(str), nil
}

func toUriPtr(val jsoniter.Any) (*types.URI, error) {
	str := val.ToString()
	if val.LastError() != nil {
		return nil, val.LastError()
	}
	return types.ParseURI(str), nil
}

// UnmarshalJSON implements the json unmarshal method used when this type is
// unmarshaled using json.Unmarshal.
func (e *Event) UnmarshalJSON(b []byte) error {
	iterator := jsoniter.ConfigFastest.BorrowIterator(b)
	defer jsoniter.ConfigFastest.ReturnIterator(iterator)
	return readJsonFromIterator(e, iterator)
}
