package state

import (
	"testing"

	"github.com/hashicorp/terraform/states"
	"github.com/zclconf/go-cty/cty"
)

func TestValueAsString(t *testing.T) {
	tests := []struct {
		name    string
		value   cty.Value
		wantS   string
		wantErr bool
	}{
		{"Primitive", cty.StringVal("hello"), `hello`, false},
		{"Primitive", cty.StringVal("hello\nworld"), `hello
world`, false},
		{"Primitive", cty.StringVal(""), ``, false},
		{"Primitive", cty.StringVal("15"), `15`, false},
		{"Primitive", cty.StringVal("true"), `true`, false},
		{"Primitive", cty.NullVal(cty.String), `null`, false},
		{"Primitive", cty.NumberIntVal(2), `2`, false},
		{"Primitive", cty.NumberFloatVal(2.5), `2.5`, false},
		{"Primitive", cty.True, `true`, false},
		{"Primitive", cty.False, `false`, false},

		{"Lists", cty.ListVal([]cty.Value{cty.True, cty.False}), `[true,false]`, false},
		{"Lists", cty.ListValEmpty(cty.Bool), `[]`, false},

		{"Sets", cty.SetVal([]cty.Value{cty.True, cty.False}), `[false,true]`, false},
		{"Sets", cty.SetValEmpty(cty.Bool), `[]`, false},

		{"Tuples", cty.TupleVal([]cty.Value{cty.True, cty.NumberIntVal(5)}), `[true,5]`, false},
		{"Tuples", cty.EmptyTupleVal, `[]`, false},

		{"Maps", cty.MapValEmpty(cty.Bool), `{}`, false},
		{"Maps", cty.MapVal(map[string]cty.Value{"yes": cty.True, "no": cty.False}), `{"no":false,"yes":true}`, false},
		{"Maps", cty.NullVal(cty.Map(cty.Bool)), `null`, false},

		{"Objects", cty.EmptyObjectVal, `{}`, false},
		{"Objects", cty.ObjectVal(map[string]cty.Value{"bool": cty.True, "number": cty.Zero}), `{"bool":true,"number":0}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &states.OutputValue{
				Value:     tt.value,
				Sensitive: false,
			}
			gotS, err := ValueAsString(v)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValueAsString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotS != tt.wantS {
				t.Errorf("ValueAsString() = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}
