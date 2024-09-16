// Package tagops is aimed to provide a set of functions to work with struct
// tags.
package tagops

const (
	tagsep     = ","         // tag separator
	fOmitEmpty = "omitempty" // omitempty tag value
)

// PrepareToMap returns a ToMap function with options set by opts.
func PrepareToMap(opts ...Option) func(any) map[string]any {
	var m Mapper = Mapper{
		Tag:       "json",
		Omitempty: false,
		Flatten:   false,
	}
	for _, o := range opts {
		o(&m)
	}
	return m.ToMap
}

// ToMap converts an argument a which should be some struct type, to a
// map[tag]value.  If omitempty is specified, fields having empty values and
// tag option "omitempty" are skipped.  If flatten is true, all nested
// non-anonymous structs are flattened into the parent map.
func ToMap(a any, tag string, omitempty bool, flatten bool) map[string]any {
	m := Mapper{
		Tag:       tag,
		Omitempty: omitempty,
		Flatten:   flatten,
	}
	return m.ToMap(a)
}

// Tags returns a sorted list of names in tags, given a struct object.  The
// empty fields are included and the map is flattened.
func Tags(a any, tag string) []string {
	m := Mapper{
		Tag:       tag,
		Omitempty: false,
		Flatten:   true,
	}
	return m.Tags(a)
}

// Values returns values for the struct object a, given a tag.  The empty
// fields are included and the map is flattened.  The values are returned in
// the alphabetical order of tags.
func Values(a any, tag string) ([]any, error) {
	m := Mapper{
		Tag:       tag,
		Omitempty: false,
		Flatten:   true,
	}
	return m.Values(a)
}
