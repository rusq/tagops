package tagops

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_resize(t *testing.T) {
	type args struct {
		s  *[]any
		sz int
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
	}{
		{
			name: "grow slice",
			args: args{
				s:  &[]any{1, 2, 3},
				sz: 5,
			},
			wantLen: 5,
		},
		{
			name: "shrink slice",
			args: args{
				s:  &[]any{1, 2, 3},
				sz: 2,
			},
			wantLen: 2,
		},
		{
			name: "same size slice",
			args: args{
				s:  &[]any{1, 2, 3},
				sz: 3,
			},
			wantLen: 3,
		},
		{
			name: "empty slice",
			args: args{
				s:  &[]any{},
				sz: 3,
			},
			wantLen: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resize(tt.args.s, tt.args.sz)
			assert.Equal(t, tt.wantLen, len(*tt.args.s))
		})
	}
}

func TestToMap(t *testing.T) {
	type (
		Address struct {
			Street string `json:"street,omitempty"`
			City   string `json:"city,omitempty"`
			ZIP    int    `json:"zip"`
		}

		nestedStruct struct {
			Name    string  `json:"name,omitempty"`
			Age     int     `json:"age,omitempty"`
			Address Address `json:"address,omitempty"`
		}
	)

	type args struct {
		a         any
		tag       string
		omitempty bool
		flatten   bool
	}
	tests := []struct {
		name string
		args args
		want map[string]any
	}{
		{
			name: "empty struct",
			args: args{
				a:         struct{}{},
				tag:       "json",
				omitempty: true,
			},
			want: map[string]any{},
		},
		{
			name: "struct with omitempty",
			args: args{
				a: struct {
					Name string `json:"name,omitempty"`
					Age  int    `json:"age,omitempty"`
				}{
					Name: "John",
					Age:  0,
				},
				tag:       "json",
				omitempty: true,
			},
			want: map[string]any{
				"name": "John",
			},
		},
		{
			name: "struct without omitempty",
			args: args{
				a: struct {
					Name string `json:"name,omitempty"`
					Age  int    `json:"age,omitempty"`
				}{
					Name: "John",
					Age:  0,
				},
				tag:       "json",
				omitempty: false,
			},
			want: map[string]any{
				"name": "John",
				"age":  0,
			},
		},
		{
			name: "nested struct with omitempty",
			args: args{
				a: nestedStruct{
					Name: "John",
					Age:  0,
					Address: Address{
						Street: "123 Main St",
						City:   "Anytown",
						ZIP:    12345,
					},
				},
				tag:       "json",
				omitempty: true,
				flatten:   false,
			},
			want: map[string]any{
				"name":    "John",
				"address": map[string]any{"street": "123 Main St", "city": "Anytown", "zip": 12345},
			},
		},
		{
			name: "nested struct with omitempty, flatten",
			args: args{
				a: nestedStruct{
					Name: "John",
					Age:  0,
					Address: Address{
						Street: "123 Main St",
						City:   "Anytown",
						ZIP:    12345,
					},
				},
				tag:       "json",
				omitempty: true,
				flatten:   true,
			},
			want: map[string]any{
				"name":   "John",
				"street": "123 Main St",
				"city":   "Anytown",
				"zip":    12345,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToMap(tt.args.a, tt.args.tag, tt.args.omitempty, tt.args.flatten)
			assert.Equal(t, tt.want, got)
		})
	}
}

func ExampleToMap_flatten() {
	type Address struct {
		Street string `json:"street,omitempty"`
		City   string `json:"city,omitempty"`
		ZIP    int    `json:"zip"`
	}
	type Person struct {
		Name    string  `json:"name,omitempty"`
		Age     int     `json:"age,omitempty"`
		Address Address `json:"address,omitempty"`
	}

	p := Person{
		Name: "Alice",
		Age:  26,
		Address: Address{
			Street: "123 Main St",
			City:   "Anytown",
			ZIP:    0,
		},
	}

	mp := ToMap(p, "json", true, true)
	printjson(mp)
	// Output:
	// {
	//   "age": 26,
	//   "city": "Anytown",
	//   "name": "Alice",
	//   "street": "123 Main St",
	//   "zip": 0
	// }
}

func ExampleToMap_nested() {
	type Address struct {
		Street string `json:"street,omitempty"`
		City   string `json:"city,omitempty"`
		ZIP    int    `json:"zip"`
	}
	type Person struct {
		Name    string  `json:"name,omitempty"`
		Age     int     `json:"age,omitempty"`
		Address Address `json:"address,omitempty"`
	}

	p := Person{
		Name: "Alice",
		Age:  26,
		Address: Address{
			Street: "123 Main St",
			City:   "Anytown",
			ZIP:    0,
		},
	}

	mp := ToMap(p, "json", true, false)
	printjson(mp)
	// Output:
	// {
	//   "address": {
	//     "city": "Anytown",
	//     "street": "123 Main St",
	//     "zip": 0
	//   },
	//   "age": 26,
	//   "name": "Alice"
	// }
}

func ExampleToMap_anonymous() {
	// Anonymous structures are always flattened.
	type Person struct {
		Name string `json:"name,omitempty"`
		Age  int    `json:"age,omitempty"`
	}
	type Employee struct {
		Person
		Position string `json:"position,omitempty"`
	}

	p := Employee{
		Person: Person{
			Name: "Bob",
			Age:  30,
		},
		Position: "Manager",
	}

	mp := ToMap(p, "json", true, false) // flatten is ignored
	printjson(mp)
	// Output:
	// {
	//   "age": 30,
	//   "name": "Bob",
	//   "position": "Manager"
	// }
}

func printjson(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func Test_isExported(t *testing.T) {
	type args struct {
		fieldName string
	}
	tests := []struct {
		name      string
		args      args
		want      bool
		wantpanic bool
	}{
		{"exported", args{"Name"}, true, false},
		{"unexported", args{"name"}, false, false},
		{"unexported", args{"_name"}, false, false},
		{"unexported", args{"_Name"}, false, false},
		{"unexported", args{"_"}, false, false},
		{"empty", args{""}, false, true},
		{"funky name", args{"🥐"}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					assert.True(t, tt.wantpanic, "should not have panicked, but did")
				} else {
					assert.False(t, tt.wantpanic, "should have panicked, but did not")
				}
			}()
			if got := isExported(tt.args.fieldName); got != tt.want {
				t.Errorf("isExported() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeys(t *testing.T) {
	type args struct {
		m map[string]any
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty map",
			args: args{map[string]any{}},
			want: nil,
		},
		{
			name: "map with keys",
			args: args{map[string]any{"z": 26, "a": 1, "b": 2, "c": 3}},
			want: []string{"a", "b", "c", "z"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Keys(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Keys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapValues(t *testing.T) {
	type args struct {
		m     map[string]any
		order []string
	}
	tests := []struct {
		name       string
		args       args
		initialLen int
		wantOut    []any
		wantErr    bool
	}{
		{
			name: "empty map",
			args: args{
				m:     map[string]any{},
				order: []string{"a", "b", "c"},
			},
			initialLen: 0,
			wantOut:    []any{nil, nil, nil},
			wantErr:    false,
		},
		{
			name: "map with keys",
			args: args{
				m:     map[string]any{"z": 26, "a": 1, "b": 2, "c": 3},
				order: []string{"a", "b", "c", "z"},
			},
			initialLen: 0,
			wantOut:    []any{1, 2, 3, 26},
			wantErr:    false,
		},
		{
			name: "order is honored",
			args: args{
				m:     map[string]any{"z": 26, "a": 1, "b": 2, "c": 3},
				order: []string{"z", "b", "a", "c"},
			},
			initialLen: 0,
			wantOut:    []any{26, 2, 1, 3},
			wantErr:    false,
		},
		{
			name: "out slice is resized",
			args: args{
				m:     map[string]any{"z": 26, "a": 1, "b": 2, "c": 3},
				order: []string{"z", "b", "a", "c"},
			},
			initialLen: 2,
			wantOut:    []any{26, 2, 1, 3},
			wantErr:    false,
		},
		{
			name: "slice is shrunk",
			args: args{
				m:     map[string]any{"z": 26, "a": 1, "b": 2, "c": 3},
				order: []string{"z", "b"},
			},
			initialLen: 4,
			wantOut:    []any{26, 2},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out = make([]any, tt.initialLen)
			if err := MapValues(&out, tt.args.m, tt.args.order); (err != nil) != tt.wantErr {
				t.Errorf("MapValues() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantOut, out)
		})
	}
}

func TestValues(t *testing.T) {
	type args struct {
		a   any
		tag string
	}
	tests := []struct {
		name    string
		args    args
		want    []any
		wantErr bool
	}{
		{
			name: "empty struct",
			args: args{
				a:   struct{}{},
				tag: "json",
			},
			want:    []any{},
			wantErr: false,
		},
		{
			name: "struct with omitempty",
			args: args{
				a: struct {
					Name string `json:"name,omitempty"`
					Age  int    `json:"age,omitempty"`
				}{
					Name: "John",
					Age:  0,
				},
				tag: "json",
			},
			want:    []any{0, "John"},
			wantErr: false,
		},
		{
			name: "nested struct with omitempty",
			args: args{
				a: struct {
					Name    string `json:"name,omitempty"`
					Age     int    `json:"age,omitempty"`
					Address struct {
						Street string `json:"street,omitempty"`
						City   string `json:"city,omitempty"`
						ZIP    int    `json:"zip"`
					} `json:"address,omitempty"`
				}{
					Name: "John",
					Age:  0,
					Address: struct {
						Street string `json:"street,omitempty"`
						City   string `json:"city,omitempty"`
						ZIP    int    `json:"zip"`
					}{
						Street: "123 Main St",
						City:   "Anytown",
						ZIP:    12345,
					},
				},
				tag: "json",
			},
			want:    []any{0, "Anytown", "John", "123 Main St", 12345},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Values(tt.args.a, tt.args.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Values() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Values() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tagName(t *testing.T) {
	type (
		structWithTags struct {
			Name string `json:"name"`
		}
		structWithOmitempty struct {
			Name string `json:"name,omitempty"`
		}
		structWithNoTag struct {
			Name string
		}
		structWithSkip struct {
			Name string `json:"-"`
		}
		structWithEmptyTag struct {
			Name string `json:""`
		}
		structWithOmitemptyTag struct {
			Name string `json:",omitempty"`
		}
		structWithDashTagAndComma struct {
			Name string `json:"-,"`
		}
		structWithUnexportedFields struct {
			name string
		}
	)

	type args struct {
		fld       reflect.StructField
		val       reflect.Value
		tag       string
		omitempty bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "struct with tags",
			args: args{
				fld:       field(t, structWithTags{}, 0),
				val:       value(t, structWithTags{}, 0),
				tag:       "json",
				omitempty: false,
			},
			want:    "name",
			wantErr: false,
		},
		{
			name: "struct with omitempty, no value",
			args: args{
				fld:       field(t, structWithOmitempty{}, 0),
				val:       value(t, structWithOmitempty{}, 0),
				tag:       "json",
				omitempty: true,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "struct with no tag, with value",
			args: args{
				fld:       field(t, structWithOmitempty{Name: "John"}, 0),
				val:       value(t, structWithOmitempty{Name: "John"}, 0),
				tag:       "json",
				omitempty: true,
			},
			want:    "name",
			wantErr: false,
		},
		{
			name: "struct with no tag, no value",
			args: args{
				fld:       field(t, structWithNoTag{}, 0),
				val:       value(t, structWithNoTag{}, 0),
				tag:       "json",
				omitempty: true,
			},
			want:    "Name",
			wantErr: false,
		},
		{
			name: "struct with no tag, with value",
			args: args{
				fld:       field(t, structWithNoTag{Name: "John"}, 0),
				val:       value(t, structWithNoTag{Name: "John"}, 0),
				tag:       "json",
				omitempty: true,
			},
			want:    "Name",
			wantErr: false,
		},
		{
			name: "struct with skip",
			args: args{
				fld:       field(t, structWithSkip{}, 0),
				val:       value(t, structWithSkip{}, 0),
				tag:       "json",
				omitempty: false,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "struct with empty tag",
			args: args{
				fld:       field(t, structWithEmptyTag{}, 0),
				val:       value(t, structWithEmptyTag{}, 0),
				tag:       "json",
				omitempty: false,
			},
			want:    "Name",
			wantErr: false,
		},
		{
			name: "struct with omitempty tag, no value",
			args: args{
				fld:       field(t, structWithOmitemptyTag{}, 0),
				val:       value(t, structWithOmitemptyTag{}, 0),
				tag:       "json",
				omitempty: true,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "struct with omitempty tag, with value",
			args: args{
				fld:       field(t, structWithOmitemptyTag{Name: "John"}, 0),
				val:       value(t, structWithOmitemptyTag{Name: "John"}, 0),
				tag:       "json",
				omitempty: true,
			},
			want:    "Name",
			wantErr: false,
		},
		{
			name: "struct with dash tag and comma",
			args: args{
				fld:       field(t, structWithDashTagAndComma{}, 0),
				val:       value(t, structWithDashTagAndComma{}, 0),
				tag:       "json",
				omitempty: false,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "struct with unexported fields",
			args: args{
				fld:       field(t, structWithUnexportedFields{}, 0),
				val:       value(t, structWithUnexportedFields{}, 0),
				tag:       "json",
				omitempty: false,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tagName(tt.args.fld, tt.args.val, tt.args.tag, tt.args.omitempty)
			if (err != nil) != tt.wantErr {
				t.Errorf("tagName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("tagName() = %v, want %v", got, tt.want)
			}
		})
	}
}

var field = func(t *testing.T, a any, n int) reflect.StructField {
	t.Helper()
	f, _ := fieldValue(t, a, n)
	return f
}

var value = func(t *testing.T, a any, n int) reflect.Value {
	t.Helper()
	_, v := fieldValue(t, a, n)
	return v
}

func fieldValue(t *testing.T, a any, n int) (reflect.StructField, reflect.Value) {
	t.Helper()

	v := reflect.ValueOf(a)
	if v.Kind() != reflect.Struct {
		t.Fatalf("expected struct, got %v", v.Kind())
	}
	fld := v.Type().Field(n)
	val := v.Field(n)
	return fld, val
}
