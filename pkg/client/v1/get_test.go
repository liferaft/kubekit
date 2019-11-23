package v1

import (
	"reflect"
	"sort"
	"testing"
)

func TestMapToSlice(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]string
		want []string
	}{
		{"nil", nil, []string{}},
		{"empty", map[string]string{}, []string{}},
		{"simple", map[string]string{"key": "value"}, []string{"key=value"}},
		{"empty key", map[string]string{"key": "value", "key1": ""}, []string{"key=value", "key1="}},
		{"orphan value", map[string]string{"key": "value", "": "value1"}, []string{"key=value"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapToSlice(tt.m); !arraySortedEqual(got, tt.want) {
				t.Errorf("MapToSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

// test if the string arrays are equal when sorted
// useful in this case because map iteration doesn't guarantee order
func arraySortedEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aCopy := make([]string, len(a))
	bCopy := make([]string, len(b))

	copy(aCopy, a)
	copy(bCopy, b)

	sort.Strings(aCopy)
	sort.Strings(bCopy)

	return reflect.DeepEqual(aCopy, bCopy)
}
