package sqlbuilder

import (
	"reflect"
	"sync"
)

type structFields struct {
	fieldAlias      map[string]string
	taggedFields    map[string][]string
	quotedFields    map[string]struct{}
	omitEmptyFields map[string]omitEmptyTagMap
}

type structFieldsieldsParser func() *structFields

func makeDefaultFieldsParser(t reflect.Type) structFieldsieldsParser {
	return makeFieldsParser(t, nil, true)
}

func makeCustomFieldsParser(t reflect.Type, mapper FieldMapperFunc) structFieldsieldsParser {
	return makeFieldsParser(t, mapper, false)
}

func makeFieldsParser(t reflect.Type, mapper FieldMapperFunc, useDefault bool) structFieldsieldsParser {
	var once sync.Once
	sf := &structFields{
		fieldAlias:      map[string]string{},
		taggedFields:    map[string][]string{},
		quotedFields:    map[string]struct{}{},
		omitEmptyFields: map[string]omitEmptyTagMap{},
	}

	return func() *structFields {
		once.Do(func() {
			if useDefault {
				mapper = DefaultFieldMapper
			}

			sf.parse(t, mapper, "")
		})

		return sf
	}
}

func (sf *structFields) parse(t reflect.Type, mapper FieldMapperFunc, prefix string) {
	l := t.NumField()
	var anonymous []reflect.StructField

	for i := 0; i < l; i++ {
		field := t.Field(i)

		if field.Anonymous {
			ft := field.Type

			// If field is an anonymous struct or pointer to struct, parse it later.
			if k := ft.Kind(); k == reflect.Struct || (k == reflect.Ptr && ft.Elem().Kind() == reflect.Struct) {
				anonymous = append(anonymous, field)
				continue
			}
		}

		// Parse DBTag.
		dbtag := field.Tag.Get(DBTag)
		alias := dbtag

		if dbtag == "-" {
			continue
		}

		if dbtag == "" {
			if mapper == nil {
				alias = field.Name
			} else {
				alias = mapper(field.Name)
			}
		}

		// The alias name has been used by another field.
		// This field is shadowed.
		if _, ok := sf.fieldAlias[alias]; ok {
			continue
		}

		sf.fieldAlias[alias] = field.Name

		// Parse FieldTag.
		fieldtag := field.Tag.Get(FieldTag)
		tags := splitTokens(fieldtag)

		for _, t := range tags {
			if t != "" {
				sf.taggedFields[t] = append(sf.taggedFields[t], alias)
			}
		}

		sf.taggedFields[""] = append(sf.taggedFields[""], alias)

		// Parse FieldOpt.
		fieldopt := field.Tag.Get(FieldOpt)
		opts := optRegex.FindAllString(fieldopt, -1)

		for _, opt := range opts {
			optMap := getOptMatchedMap(opt)

			switch optMap[optName] {
			case fieldOptOmitEmpty:
				tags := getTagsFromOptParams(optMap[optParams])
				sf.appendOmitEmptyFieldsTags(alias, tags...)

			case fieldOptWithQuote:
				sf.quotedFields[alias] = struct{}{}
			}
		}
	}

	for _, field := range anonymous {
		ft := dereferencedType(field.Type)
		sf.parse(ft, mapper, prefix+field.Name+".")
	}
}

func (sf *structFields) appendOmitEmptyFieldsTags(alias string, tags ...string) {
	if sf.omitEmptyFields[alias] == nil {
		sf.omitEmptyFields[alias] = omitEmptyTagMap{}
	}

	for _, tag := range tags {
		sf.omitEmptyFields[alias][tag] = struct{}{}
	}
}

type omitEmptyTagMap map[string]struct{}

func (sm omitEmptyTagMap) containsAny(tags ...string) (res bool) {
	for _, tag := range tags {
		if _, res = sm[tag]; res {
			return
		}
	}

	return
}
