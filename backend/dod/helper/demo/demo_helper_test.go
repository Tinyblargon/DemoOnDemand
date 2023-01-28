package demo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_CreateObject(t *testing.T) {
	testData := []struct {
		input  string
		output Demo
		err    bool
	}{
		{input: "", err: true},
		{input: "a", err: true},
		{input: "a_", err: true},
		{input: "a_b", err: true},
		{input: "a_b_", err: true},
		{input: "a_b_c", err: true},
		{input: "a_b_c_", err: true},
		{input: "a_b_c_d", err: true},
		{input: "a_1_c", output: Demo{Name: "c", User: "a", ID: 1}},
		{input: "a_2_c_d", output: Demo{Name: "c_d", User: "a", ID: 2}},
		{input: "a_3_c_d_e", output: Demo{Name: "c_d_e", User: "a", ID: 3}},
		{input: "a_4_c_d_e_f", output: Demo{Name: "c_d_e_f", User: "a", ID: 4}},
	}
	for _, e := range testData {
		output, err := CreateObject(e.input)
		if e.err {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
		require.Equal(t, e.output, output)
	}
}
