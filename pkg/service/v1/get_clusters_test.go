package v1

import (
	"reflect"
	"testing"
)

func TestSliceToMap(t *testing.T) {
	tests := []struct {
		name    string
		slice   []string
		want    map[string]string
		wantErr bool
	}{
		{"nil", nil, map[string]string{}, false},
		{"empty", []string{}, map[string]string{}, false},
		{"simple", []string{"key=value"}, map[string]string{"key": "value"}, false},
		{"orphan key", []string{"key=value", "key1"}, nil, true},
		{"empty key", []string{"key=value", "key1="}, map[string]string{"key": "value", "key1": ""}, false},
		{"orphan value", []string{"key=value", "=value1"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SliceToMap(tt.slice)
			if (err != nil) != tt.wantErr {
				t.Errorf("SliceToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SliceToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
