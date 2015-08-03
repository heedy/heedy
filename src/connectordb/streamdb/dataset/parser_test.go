package dataset

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetString(t *testing.T) {
	testcases := []struct {
		in   string
		out  string
		outi string
		err  bool
	}{
		{"`blah`oh'lol'", "blah", "oh'lol'", false},
		{"\"testing\"", "testing", "", false},
		{"'who", "", "'who'", true},
	}
	for _, c := range testcases {
		a, b, e := getString(c.in)
		if c.err {
			require.Error(t, e)
		} else {
			require.Nil(t, e)
			require.Equal(t, c.out, a)
			require.Equal(t, c.outi, b)
		}

	}

}

func TestGetSymbol(t *testing.T) {
	testcases := []struct {
		in   string
		out  string
		outi string
		err  bool
	}{
		{"test:t2", "test", ":t2", false},
		{"'t(,)':t2", "t(,)", ":t2", false},
		{"t(", "t", "(", false},
		{"'who", "", "'who", true},
		{"'", "", "'", true},
		{"", "", "", false},
		{"test,test2", "test", ",test2", false},
	}
	for _, c := range testcases {
		a, b, e := getSymbol(c.in)
		if c.err {
			require.Error(t, e, fmt.Sprintf("%v", c))
		} else {
			require.Nil(t, e, fmt.Sprintf("%v", c))
			require.Equal(t, c.out, a, fmt.Sprintf("%v", c))
			require.Equal(t, c.outi, b, fmt.Sprintf("%v", c))
		}

	}

}

func TestEatSpace(t *testing.T) {
	testcases := []struct {
		in  string
		out string
	}{
		{"  \n\t blah blah ", "blah blah "},
		{"testing", "testing"},
		{"", ""},
	}
	for _, c := range testcases {
		require.Equal(t, c.out, eatSpace(c.in))
	}

}

func TestParsePipeliner(t *testing.T) {
	p, err := ParsePipeline("lt(5,'6.0,'):groot:lolol(pff)")
	require.NoError(t, err)
	require.Equal(t, 3, len(p))
	require.Equal(t, "lt", p[0].Symbol)
	require.Equal(t, "5", p[0].Args[0])
	require.Equal(t, "6.0,", p[0].Args[1])
	require.Equal(t, "groot", p[1].Symbol)
	require.Nil(t, p[1].Args)
	p, err = ParsePipeline("lt(5,'6.0,'):groot")
	require.NoError(t, err)
	require.Equal(t, 2, len(p))
	require.Equal(t, "lt", p[0].Symbol)
	require.Equal(t, "5", p[0].Args[0])
	require.Equal(t, "6.0,", p[0].Args[1])
	require.Equal(t, "groot", p[1].Symbol)
	require.Nil(t, p[1].Args)
	p, err = ParsePipeline("lt:")
	require.Error(t, err)
	p, err = ParsePipeline(":lt")
	require.Error(t, err)
	p, err = ParsePipeline("lt(a):")
	require.Error(t, err)
	p, err = ParsePipeline("lt(a):,")
	require.Error(t, err)
	p, err = ParsePipeline("lt(a):a,")
	require.Error(t, err)
	p, err = ParsePipeline("")
	require.NoError(t, err)
	require.Equal(t, 0, len(p))
	p, err = ParsePipeline("hi")
	require.NoError(t, err)
	require.Equal(t, 1, len(p))
}
