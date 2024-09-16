package tagops

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Mapper is a struct to map struct fields to map key/values.  The struct
// fields are mapped to map keys using the Tag. No tag value leads to undefined
// behavior.  The Mapper can be configured with options.
type Mapper struct {
	// Tag is the tag name.
	Tag string
	// Omitempty omits empty fields.
	Omitempty bool
	// Flatten flattens named nested structs (anonymous structs are always
	// flattened).
	Flatten bool
}

// New returns a new Mapper with options opts.
func New(opts ...Option) Mapper {
	m := Mapper{Tag: "json"}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// Option is a functional option for Mapper.
type Option func(*Mapper)

// Flatten returns an Option that sets the Flatten option to true.
func Flatten() Option {
	return func(o *Mapper) {
		o.Flatten = true
	}
}

// Tag returns an Option that sets the Tag to tag.
func Tag(tag string) Option {
	return func(o *Mapper) {
		o.Tag = tag
	}
}

// Omitempty sets the Omitempty option to true.
func Omitempty() Option {
	return func(o *Mapper) {
		o.Omitempty = true
	}
}

func (m Mapper) ToMap(a any) map[string]any {
	out := make(map[string]any)

	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	typ := v.Type()
	for i := range v.NumField() {
		field := typ.Field(i)

		if field.Type.Kind() == reflect.Struct && v.Field(i).Type() != reflect.TypeOf(time.Time{}) {
			nested := ToMap(v.Field(i).Interface(), m.Tag, m.Omitempty, m.Flatten)
			if field.Anonymous || m.Flatten {
				// flatten nested structs
				for key, val := range nested {
					out[key] = val
				}
			} else {
				// nested maps are not flattened
				key, err := tagName(field, v.Field(i), m.Tag, m.Omitempty)
				if errors.Is(err, errSkip) {
					continue
				}
				out[key] = nested
			}
		} else {
			key, err := tagName(field, v.Field(i), m.Tag, m.Omitempty)
			if errors.Is(err, errSkip) {
				continue
			}
			out[key] = v.Field(i).Interface()
		}
	}
	return out
}

// Tags returns a sorted list of names in tags, given a struct object.  The
// empty fields are included and the map is flattened.
func (m Mapper) Tags(a any) []string {
	return Keys(ToMap(a, m.Tag, m.Omitempty, m.Flatten))
}

// Values returns values for the struct object a, given a tag.  The empty
// fields are included and the map is flattened.  The values are returned in
// the alphabetical order of tags.
func (m Mapper) Values(a any) ([]any, error) {
	mp := ToMap(a, m.Tag, false, true)
	var ret = make([]any, 0, len(mp))
	if err := MapValues(&ret, mp, m.Tags(a)); err != nil {
		return nil, err
	}
	return ret, nil
}

// Keys returns a sorted list of keys for the map m.
func Keys(m map[string]any) []string {
	kk := slices.Collect(maps.Keys(m))
	sort.Strings(kk)
	return kk
}

// MapValues populates slice out with values from map m in the key order
// specified by order.  The size of out slice will be adjusted to order size
// to accomodate for all values.
func MapValues(out *[]any, m map[string]any, order []string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(error)
			if ok {
				err = e
			} else {
				err = fmt.Errorf("panic: %v", r)
			}
		}
	}()
	if len(*out) != len(order) {
		resize(out, len(order))
	}
	for i, col := range order {
		(*out)[i] = m[col]
	}
	return nil
}

// errSkip is returned by tagName to indicate that the field should be skipped.
var errSkip = errors.New("skip")

// tagName returns a tag name for the field, or an errSkip error if the field
// should be skipped.
func tagName(fld reflect.StructField, val reflect.Value, tag string, omitempty bool) (string, error) {
	if !isExported(fld.Name) {
		return "", errSkip
	}
	tagValue := strings.SplitN(fld.Tag.Get(tag), tagsep, 2)
	if len(tagValue) == 0 {
		return fld.Name, nil
	}
	if strings.EqualFold(tagValue[0], "-") {
		return "", errSkip
	}
	if tagValue[0] == "" {
		tagValue[0] = fld.Name
	}
	if omitempty {
		// if there's a tag option and that tag option is omitempty
		// and field is empty.
		if len(tagValue) > 1 && (tagValue[1] == fOmitEmpty && isEmpty(val)) {
			return "", errSkip
		}
	}
	return tagValue[0], nil
}

// isEmpty knows about some empty values.
func isEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == 0
	case reflect.Struct:
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return v.Interface().(time.Time).IsZero()
		}
		// fallthrough
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// isExported returns true if the field is exported.
func isExported(fieldName string) bool {
	firstRune, _ := utf8.DecodeRuneInString(fieldName)
	if firstRune == utf8.RuneError {
		panic(fmt.Sprintf("isExported: unsupported field: %q", fieldName))
	}
	return unicode.In(firstRune, unicode.Lu)
}

// resize resizes the slice to a requested size.  If slice is smaller, it is
// extended, if larger - truncated to the desired size.
func resize(s *[]any, sz int) {
	if s == nil {
		panic("resize: nil slice")
	}
	if len(*s) >= sz {
		// shrink
		*s = (*s)[:sz]
		return
	}
	// grow
	add := make([]any, sz-len(*s))
	*s = append(*s, add...)
}
