package version

import (
	"reflect"
	"testing"
)

func TestNewSemVerAndString(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		args    args
		want    *SemVer
		wantErr bool
	}{
		{"empty version", args{""}, nil, true},
		{"incorrect version", args{"x.y.z"}, nil, true},
		{"partially incorrect version", args{"1.y.3"}, nil, true},
		{"missing number version", args{"1..3"}, nil, true},
		{"missing last number version", args{"1.2."}, nil, true},
		{"incomplete version", args{"1.2"}, &SemVer{1, 2, 0, []string{}, []int{1, 2}}, false},
		{"simple version", args{"1.2.3"}, &SemVer{1, 2, 3, []string{}, []int{1, 2, 3}}, false},
		{"version with labels", args{"1.2.3-4"}, &SemVer{1, 2, 3, []string{"4"}, []int{1, 2, 3}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSemVer(tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSemVer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSemVer() = %v, want %v", got, tt.want)
			}
			if err != nil {
				return
			}

			gotStr := got.String()
			if gotStr != tt.args.version {
				t.Errorf("String() = %v, want %v", gotStr, tt.args.version)
			}
		})
	}
}

func TestSemVer_Diff(t *testing.T) {
	type args struct {
		ver *SemVer
	}
	tests := []struct {
		name string
		args args
		want *SemVer
	}{
		{"differ label", args{&SemVer{1, 2, 3, []string{"fake", "dev"}, []int{1, 2, 3}}}, &SemVer{Major: 0, Minor: 0, Patch: 0, numbers: []int{0, 0, 0}, Labels: []string{"differ (dev != fake-dev)"}}},
		{"differ version", args{&SemVer{1, 2, 3, []string{"dev"}, []int{1, 2, 3}}}, &SemVer{Major: 0, Minor: 0, Patch: 0, numbers: []int{0, 0, 0}}},

		{"equal", args{&SemVer{1, 2, 3, []string{"dev"}, []int{1, 2, 3}}}, &SemVer{Major: 0, Minor: 0, Patch: 0, numbers: []int{0, 0, 0}}},

		{"less by major", args{&SemVer{0, 4, 0, []string{"dev"}, []int{0, 4, 0}}}, &SemVer{Major: 1, Minor: -2, Patch: 3, numbers: []int{1, -2, 3}}},
		{"great by major", args{&SemVer{2, 0, 4, []string{"dev"}, []int{2, 0, 4}}}, &SemVer{Major: -1, Minor: 2, Patch: -1, numbers: []int{-1, 2, -1}}},

		{"less by minor", args{&SemVer{1, 1, 4, []string{"dev"}, []int{1, 1, 4}}}, &SemVer{Major: 0, Minor: 1, Patch: -1, numbers: []int{0, 1, -1}}},
		{"great by minor", args{&SemVer{1, 3, 2, []string{"dev"}, []int{1, 3, 2}}}, &SemVer{Major: 0, Minor: -1, Patch: 1, numbers: []int{0, -1, 1}}},

		{"less by patch", args{&SemVer{1, 2, 1, []string{"dev"}, []int{1, 2, 1}}}, &SemVer{Major: 0, Minor: 0, Patch: 2, numbers: []int{0, 0, 2}}},
		{"great by patch", args{&SemVer{1, 2, 5, []string{"dev"}, []int{1, 2, 5}}}, &SemVer{Major: 0, Minor: 0, Patch: -2, numbers: []int{0, 0, -2}}},

		{"less by major & minor", args{&SemVer{0, 1, 4, []string{"dev"}, []int{0, 1, 4}}}, &SemVer{Major: 1, Minor: 1, Patch: -1, numbers: []int{1, 1, -1}}},
		{"great by major & patch", args{&SemVer{2, 1, 4, []string{"dev"}, []int{2, 1, 4}}}, &SemVer{Major: -1, Minor: 1, Patch: -1, numbers: []int{-1, 1, -1}}},
	}

	v := &SemVer{
		Major:   1,
		Minor:   2,
		Patch:   3,
		Labels:  []string{"dev"},
		numbers: []int{1, 2, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := v.Diff(tt.args.ver); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SemVer.Diff() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSemVer_Compare(t *testing.T) {
	type args struct {
		ver *SemVer
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"equal", args{&SemVer{1, 2, 3, []string{"dev"}, []int{1, 2, 3}}}, 0},

		{"less by major", args{&SemVer{0, 4, 0, []string{"dev"}, []int{0, 4, 0}}}, 1},
		{"great by major", args{&SemVer{2, 0, 4, []string{"dev"}, []int{2, 0, 4}}}, -1},

		{"less by minor", args{&SemVer{1, 1, 4, []string{"dev"}, []int{1, 1, 4}}}, 1},
		{"great by minor", args{&SemVer{1, 3, 2, []string{"dev"}, []int{1, 3, 2}}}, -1},

		{"less by patch", args{&SemVer{1, 2, 1, []string{"dev"}, []int{1, 2, 1}}}, 1},
		{"great by patch", args{&SemVer{1, 2, 5, []string{"dev"}, []int{1, 2, 5}}}, -1},

		{"less by major & minor", args{&SemVer{0, 1, 4, []string{"dev"}, []int{0, 1, 4}}}, 1},
		{"great by major & patch", args{&SemVer{2, 1, 4, []string{"dev"}, []int{2, 1, 4}}}, -1},
	}

	v := &SemVer{
		Major:   1,
		Minor:   2,
		Patch:   3,
		Labels:  []string{"dev"},
		numbers: []int{1, 2, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := v.Compare(tt.args.ver); got != tt.want {
				t.Errorf("SemVer.Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	type args struct {
		version1 string
		version2 string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"wrong version 1", args{"1.x.3-dev", "1.2.3-dev"}, 0, true},
		{"wrong version 2", args{"1.2.3-dev", "1..3-dev"}, 0, true},

		{"equal", args{"1.2.3-dev", "1.2.3-dev"}, 0, false},

		{"less by major", args{"1.2.3-dev", "0.4.0-dev"}, 1, false},
		{"great by major", args{"1.2.3-dev", "2.0.4-dev"}, -1, false},

		{"less by minor", args{"1.2.3-dev", "1.1.4-dev"}, 1, false},
		{"great by minor", args{"1.2.3-dev", "1.3.2-dev"}, -1, false},

		{"less by patch", args{"1.2.3-dev", "1.2.1-dev"}, 1, false},
		{"great by patch", args{"1.2.3-dev", "1.2.5-dev"}, -1, false},

		{"less by major & minor", args{"1.2.3-dev", "0.1.4-dev"}, 1, false},
		{"great by major & patch", args{"1.2.3-dev", "2.1.4-dev"}, -1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Compare(tt.args.version1, tt.args.version2)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemVer_Equal(t *testing.T) {
	type fields struct {
		Major   int
		Minor   int
		Patch   int
		Labels  []string
		numbers []int
	}
	type args struct {
		ver *SemVer
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"equal", fields{1, 2, 3, []string{"dev"}, []int{1, 2, 3}}, args{&SemVer{1, 2, 3, []string{"dev"}, []int{1, 2, 3}}}, true},
		{"diff major & minor", fields{1, 2, 3, []string{"dev"}, []int{1, 2, 3}}, args{&SemVer{0, 1, 3, []string{"dev"}, []int{0, 1, 3}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &SemVer{
				Major:   tt.fields.Major,
				Minor:   tt.fields.Minor,
				Patch:   tt.fields.Patch,
				Labels:  tt.fields.Labels,
				numbers: tt.fields.numbers,
			}
			if got := v.EQ(tt.args.ver); got != tt.want {
				t.Errorf("SemVer.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
