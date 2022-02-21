package utils

import (
	"testing"
)

func TestIsRetryCode(t *testing.T) {
	type args struct {
		status int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "429", args: args{status: 429}, want: true},
		{name: "503", args: args{status: 503}, want: true},
		{name: "200", args: args{status: 200}, want: false},
		{name: "400", args: args{status: 400}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryCode(tt.args.status); got != tt.want {
				t.Errorf("IsRetryCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_randNum(t *testing.T) {
	type args struct {
		st int
	}
	tests := []struct {
		name string
		args args
		from float64
		to   float64
	}{
		{name: "1", args: args{st: 1}, from: 0.0, to: 1},
		{name: "1000", args: args{st: 1000}, from: 0, to: 1000},
		{name: "50", args: args{st: 50}, from: 0, to: 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := randNum(tt.args.st); got < tt.from || got > tt.to {
				t.Errorf("randNum() = %v, want in [%v, %v]", got, tt.from, tt.to)
			}
		})
	}
}
