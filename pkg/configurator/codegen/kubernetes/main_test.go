package main

import (
	"testing"
)

func Test_toFilename(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "one", args: args{str: "this_is-a Test"}, want: "this-is-a-test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toFilename(tt.args.str); got != tt.want {
				t.Errorf("toFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toResource(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "one", args: args{str: "this_is-a Test"}, want: "this-is-a-test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toResource(tt.args.str); got != tt.want {
				t.Errorf("toResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toTemplate(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "one", args: args{str: "this_is-a Test"}, want: "thisIsATestTpl"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toTemplate(tt.args.str); got != tt.want {
				t.Errorf("toTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}
