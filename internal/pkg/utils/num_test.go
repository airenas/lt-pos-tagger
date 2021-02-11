package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNumber(t *testing.T) {
	tests := []struct {
		v string
		e bool
		i string
	}{
		{v: "-10", e: true},
		{v: "-10.000", e: true},
		{v: "-10.999.000", e: true},
		{v: "10", e: true},
		{v: "5.12321e+10", e: true},
		{v: "-5.12321e-10", e: true},
		{v: "+5.12321e-10", e: true},
		{v: "1e+10", e: true},

		{v: "aaa5.12321e", e: false},
		{v: "ooo", e: false},
		{v: ",", e: false},
		{v: ".", e: false},
	}

	for i, tc := range tests {
		assert.Equal(t, tc.e, IsNumber(tc.v), "Fail %d - %s", i, tc.v)
	}
}
