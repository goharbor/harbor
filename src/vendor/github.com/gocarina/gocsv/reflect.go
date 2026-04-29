package gocsv

import (
	"reflect"
	"strings"
	"sync"
)

// --------------------------------------------------------------------------
// Reflection helpers

type structInfo struct {
	Fields []fieldInfo
}

// fieldInfo is a struct field that should be mapped to a CSV column, or vice-versa
// Each IndexChain element before the last is the index of an the embedded struct field
// that defines Key as a tag
type fieldInfo struct {
	keys         []string
	omitEmpty    bool
	IndexChain   []int
	defaultValue string
}

func (f fieldInfo) getFirstKey() string {
	return f.keys[0]
}

func (f fieldInfo) matchesKey(key string) bool {
	for _, k := range f.keys {
		if key == k || strings.TrimSpace(key) == k {
			return true
		}
	}
	return false
}

var structInfoCache sync.Map
var structMap = make(map[reflect.Type]*structInfo)
var structMapMutex sync.RWMutex

func getStructInfo(rType reflect.Type) *structInfo {
	stInfo, ok := structInfoCache.Load(rType)
	if ok {
		return stInfo.(*structInfo)
	}

	fieldsList := getFieldInfos(rType, []int{})
	stInfo = &structInfo{fieldsList}
	structInfoCache.Store(rType, stInfo)

	return stInfo.(*structInfo)
}

func getFieldInfos(rType reflect.Type, parentIndexChain []int) []fieldInfo {
	fieldsCount := rType.NumField()
	fieldsList := make([]fieldInfo, 0, fieldsCount)
	for i := 0; i < fieldsCount; i++ {
		field := rType.Field(i)
		if field.PkgPath != "" {
			continue
		}

		var cpy = make([]int, len(parentIndexChain))
		copy(cpy, parentIndexChain)
		indexChain := append(cpy, i)

		// if the field is a pointer to a struct, follow the pointer then create fieldinfo for each field
		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			// unless it implements marshalText or marshalCSV. Structs that implement this
			// should result in one value and not have their fields exposed
			if !(canMarshal(field.Type.Elem())) {
				fieldsList = append(fieldsList, getFieldInfos(field.Type.Elem(), indexChain)...)
			}
		}
		// if the field is a struct, create a fieldInfo for each of its fields
		if field.Type.Kind() == reflect.Struct {
			// unless it implements marshalText or marshalCSV. Structs that implement this
			// should result in one value and not have their fields exposed
			if !(canMarshal(field.Type)) {
				fieldsList = append(fieldsList, getFieldInfos(field.Type, indexChain)...)
			}
		}

		// if the field is an embedded struct, ignore the csv tag
		if field.Anonymous {
			continue
		}

		fieldInfo := fieldInfo{IndexChain: indexChain}
		fieldTag := field.Tag.Get(TagName)
		fieldTags := strings.Split(fieldTag, TagSeparator)
		filteredTags := []string{}
		for _, fieldTagEntry := range fieldTags {
			if fieldTagEntry == "omitempty" {
				fieldInfo.omitEmpty = true
			} else if strings.HasPrefix(fieldTagEntry, "default=") {
				fieldInfo.defaultValue = strings.TrimPrefix(fieldTagEntry, "default=")
			} else {
				filteredTags = append(filteredTags, normalizeName(fieldTagEntry))
			}
		}

		if len(filteredTags) == 1 && filteredTags[0] == "-" {
			continue
		} else if len(filteredTags) > 0 && filteredTags[0] != "" {
			fieldInfo.keys = filteredTags
		} else {
			fieldInfo.keys = []string{normalizeName(field.Name)}
		}
		fieldsList = append(fieldsList, fieldInfo)
	}
	return fieldsList
}

func getConcreteContainerInnerType(in reflect.Type) (inInnerWasPointer bool, inInnerType reflect.Type) {
	inInnerType = in.Elem()
	inInnerWasPointer = false
	if inInnerType.Kind() == reflect.Ptr {
		inInnerWasPointer = true
		inInnerType = inInnerType.Elem()
	}
	return inInnerWasPointer, inInnerType
}

func getConcreteReflectValueAndType(in interface{}) (reflect.Value, reflect.Type) {
	value := reflect.ValueOf(in)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	return value, value.Type()
}

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

func isErrorType(outType reflect.Type) bool {
	if outType.Kind() != reflect.Interface {
		return false
	}

	return outType.Implements(errorInterface)
}
