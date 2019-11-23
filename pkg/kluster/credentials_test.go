package kluster

import "testing"

func Test_filter(t *testing.T) {
	type args struct {
		str       string
		maxLen    int
		sensitive bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Partial mask",
			args: args{
				str:       "ABCDEF123456",
				maxLen:    10,
				sensitive: false,
			},
			want: "ABCD...456",
		},
		{
			name: "Full mask",
			args: args{
				str:       "ABCDEF123456",
				maxLen:    10,
				sensitive: true,
			},
			want: "**********",
		},
		{
			name: "Partial mask less than 5 characters",
			args: args{
				str:       "ABCDE",
				maxLen:    3,
				sensitive: false,
			},
			want: "A...E",
		},
		{
			name: "Full mask less than 5 characters",
			args: args{
				str:       "ABCDE",
				maxLen:    3,
				sensitive: true,
			},
			want: "*****",
		},
		{
			name: "No data",
			args: args{
				str:       "",
				maxLen:    3,
				sensitive: true,
			},
			want: "(none)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filter(tt.args.str, tt.args.maxLen, tt.args.sensitive); got != tt.want {
				t.Errorf("filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
