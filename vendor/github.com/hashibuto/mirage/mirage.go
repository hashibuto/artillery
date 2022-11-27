package mirage

import (
	"fmt"
	"reflect"
	"strings"
)

type Info struct {
	Name      string
	TagKey    string
	IsPointer bool
	Kind      reflect.Kind
}

type Reflection struct {
	namesByIdx  map[int]*Info
	idxByName   map[string]int
	idxByTagKey map[string]int
	fieldByIdx  map[int]*reflect.StructField
	infoByIdx   map[int]*Info
	elem        reflect.Value
}

// Reflect produces a reflected version of "obj", including information about tag key names indicated by "tagName"
func Reflect(obj any, tagName string) *Reflection {
	idxByName := map[string]int{}
	idxByTagKey := map[string]int{}
	fieldByIdx := map[int]*reflect.StructField{}
	infoByIdx := map[int]*Info{}

	value := reflect.ValueOf(obj)
	elem := value.Elem()
	numFields := elem.NumField()
	elemType := elem.Type()
	for idx := 0; idx < numFields; idx++ {
		field := elemType.Field(idx)
		var tagKey string
		if tagName == "" {
			tagName = field.Name
			idxByTagKey[field.Name] = idx
		} else {
			tag := field.Tag.Get(tagName)
			if tag == "" {
				tagKey = field.Name

			} else {
				tagKey = strings.Split(tag, ",")[0]
			}
			idxByTagKey[tagKey] = idx
		}
		idxByName[field.Name] = idx
		fieldByIdx[idx] = &field

		kind := field.Type.Kind()
		isPointer := false
		if kind == reflect.Pointer {
			// Indirect to get the underlying kind
			kind = field.Type.Elem().Kind()
			isPointer = true
		}
		infoByIdx[idx] = &Info{
			Name:      field.Name,
			TagKey:    tagKey,
			IsPointer: isPointer,
			Kind:      kind,
		}
	}

	return &Reflection{
		idxByName:   idxByName,
		idxByTagKey: idxByTagKey,
		fieldByIdx:  fieldByIdx,
		infoByIdx:   infoByIdx,
		elem:        elem,
	}
}

// Keys returns an array of key names
func (r *Reflection) Keys() []string {
	keys := make([]string, len(r.fieldByIdx))
	for idx, field := range r.fieldByIdx {
		keys[idx] = field.Name
	}

	return keys
}

// TagKeys returns an array of keys garnered from a given tag name
func (r *Reflection) TagKeys() []string {
	tagKeys := make([]string, len(r.fieldByIdx))
	for tagKey, idx := range r.idxByTagKey {
		tagKeys[idx] = tagKey
	}

	return tagKeys
}

// Io returns a reflection io object for this instance of the reflected object
func (r *Reflection) Io() *ReflectionIo {
	valueByIdx := map[int]reflect.Value{}
	for idx := 0; idx < len(r.fieldByIdx); idx++ {
		valueByIdx[idx] = r.elem.Field(idx)
	}
	return &ReflectionIo{
		reflection: r,
		valueByIdx: valueByIdx,
	}
}

// Io returns a reflection io object for a new instance of the reflected object
func (r *Reflection) NewIo(obj any) *ReflectionIo {
	valueByIdx := map[int]reflect.Value{}
	elem := reflect.ValueOf(obj).Elem()
	for idx := 0; idx < len(r.fieldByIdx); idx++ {
		valueByIdx[idx] = elem.Field(idx)
	}
	return &ReflectionIo{
		reflection: r,
		valueByIdx: valueByIdx,
	}
}

// InfoByName returns the reflect kind for a given field by name
func (r *Reflection) InfoByName(fieldName string) (*Info, error) {
	idx, ok := r.idxByName[fieldName]
	if !ok {
		return nil, fmt.Errorf("Unknown field name \"%s\"", fieldName)
	}
	return r.infoByIdx[idx], nil
}

// InfoByTagKey returns the reflect kind for a given tag key
func (r *Reflection) InfoByTagKey(fieldName string) (*Info, error) {
	idx, ok := r.idxByTagKey[fieldName]
	if !ok {
		return nil, fmt.Errorf("Unknown tag key name \"%s\"", fieldName)
	}
	return r.infoByIdx[idx], nil
}

func (r *Reflection) InfoByIdx(idx int) (*Info, error) {
	if idx > len(r.infoByIdx) {
		return nil, fmt.Errorf("Index is out of bounds")
	}

	return r.infoByIdx[idx], nil
}

// FieldByName returns a field struct by field name
func (r *Reflection) FieldByName(fieldName string) (*reflect.StructField, error) {
	idx, ok := r.idxByTagKey[fieldName]
	if !ok {
		return nil, fmt.Errorf("Unknown field name \"%s\"", fieldName)
	}

	return r.fieldByIdx[idx], nil
}

// FieldByIdx returns a field struct by index
func (r *Reflection) FieldByIdx(idx int) (*reflect.StructField, error) {
	if idx > len(r.fieldByIdx) {
		return nil, fmt.Errorf("Index is out of bounds")
	}
	return r.fieldByIdx[idx], nil
}

// NumFields returns the number of fields on the structure
func (r *Reflection) NumFields() int {
	return len(r.fieldByIdx)
}

// HasTagKey returns true if the tag key exists
func (r *Reflection) HasTagKey(tagKeyName string) bool {
	_, ok := r.idxByTagKey[tagKeyName]
	return ok
}

// HasFieldName returns true if the field name exists
func (r *Reflection) HasFieldName(fieldName string) bool {
	_, ok := r.idxByName[fieldName]
	return ok
}

type ReflectionIo struct {
	reflection *Reflection
	valueByIdx map[int]reflect.Value
}

// ValueFromIdx returns the struct value referenced by the field index
func (r *ReflectionIo) ValueFromIdx(idx int) (any, error) {
	if idx > len(r.reflection.idxByName) {
		return nil, fmt.Errorf("Supplied index is out of range")
	}
	return r.valueByIdx[idx].Interface(), nil
}

// ValueFromName returns the struct value referenced by the field name
func (r *ReflectionIo) ValueFromName(name string) (any, error) {
	idx, ok := r.reflection.idxByName[name]
	if !ok {
		return nil, fmt.Errorf("Unable to locate field name %s in object", name)
	}

	return r.valueByIdx[idx].Interface(), nil
}

// ValueFromTagKey returns the struct value referenced by the tag key
func (r *ReflectionIo) ValueFromTagKey(tagKey string) (any, error) {
	idx, ok := r.reflection.idxByTagKey[tagKey]
	if !ok {
		return nil, fmt.Errorf("Unable to locate tag key %s in object", tagKey)
	}

	return r.valueByIdx[idx].Interface(), nil
}

// SetValueByIdx sets a value on the reflected object using the field index
func (r *ReflectionIo) SetValueByIdx(idx int, value any) error {
	if idx > len(r.reflection.idxByName) {
		return fmt.Errorf("Supplied index is out of range")
	}
	r.valueByIdx[idx].Set(reflect.ValueOf(value))
	return nil
}

// SetValueByName sets a value on the reflected object using the field name
func (r *ReflectionIo) SetValueByName(name string, value any) error {
	idx, ok := r.reflection.idxByName[name]
	if !ok {
		return fmt.Errorf("Unable to locate field name %s in object", name)
	}

	r.valueByIdx[idx].Set(reflect.ValueOf(value))
	return nil
}

// SetValueByTagKey sets a value on the reflected object using the tag key
func (r *ReflectionIo) SetValueByTagKey(tagKey string, value any) error {
	idx, ok := r.reflection.idxByTagKey[tagKey]
	if !ok {
		return fmt.Errorf("Unable to locate tag key %s in object", tagKey)
	}

	r.valueByIdx[idx].Set(reflect.ValueOf(value))
	return nil
}

// IsNilPointerByIdx returns true if the underlying value is a nil pointer
func (r *ReflectionIo) IsNilPointerByIdx(idx int) (bool, error) {
	if idx > len(r.valueByIdx) {
		return false, fmt.Errorf("Index is out of bounds")
	}
	info := r.reflection.infoByIdx[idx]
	if !info.IsPointer {
		return false, nil
	}

	return r.valueByIdx[idx].IsNil(), nil
}

// IsNilPointerByName returns true if the underlying value is a nil pointer
func (r *ReflectionIo) IsNilPointerByName(name string) (bool, error) {
	idx, ok := r.reflection.idxByName[name]
	if !ok {
		return false, fmt.Errorf("Field name \"%s\" does not exist", name)
	}
	if idx > len(r.valueByIdx) {
		return false, fmt.Errorf("Index is out of bounds")
	}
	info := r.reflection.infoByIdx[idx]
	if !info.IsPointer {
		return false, nil
	}

	return r.valueByIdx[idx].IsNil(), nil
}

// IsNilPointerByTagKey returns true if the underlying value is a nil pointer
func (r *ReflectionIo) IsNilPointerByTagKey(tagKey string) (bool, error) {
	idx, ok := r.reflection.idxByTagKey[tagKey]
	if !ok {
		return false, fmt.Errorf("Tag key \"%s\" does not exist", tagKey)
	}
	if idx > len(r.valueByIdx) {
		return false, fmt.Errorf("Index is out of bounds")
	}
	info := r.reflection.infoByIdx[idx]
	if !info.IsPointer {
		return false, nil
	}

	return r.valueByIdx[idx].IsNil(), nil
}

// InstantiateByIdx instantiates a new empty value of the field type and returns it
func (r *ReflectionIo) InstantiateByIdx(idx int) (any, error) {
	if idx > len(r.valueByIdx) {
		return nil, fmt.Errorf("Index is out of bounds")
	}

	field := r.reflection.fieldByIdx[idx]
	t := field.Type.Elem()
	v := reflect.New(t).Interface()
	return v, nil
}

// InstantiateByName instantiates a new empty value of the field type and returns it
func (r *ReflectionIo) InstantiateByName(name string) (any, error) {
	idx, ok := r.reflection.idxByName[name]
	if !ok {
		return nil, fmt.Errorf("Field name \"%s\" does not exist", name)
	}

	return r.InstantiateByIdx(idx)
}

// InstantiateByTagKey instantiates a new empty value of the field type and returns it
func (r *ReflectionIo) InstantiateByTagKey(tagKey string) (any, error) {
	idx, ok := r.reflection.idxByTagKey[tagKey]
	if !ok {
		return nil, fmt.Errorf("Tag key \"%s\" does not exist", tagKey)
	}

	return r.InstantiateByIdx(idx)
}
