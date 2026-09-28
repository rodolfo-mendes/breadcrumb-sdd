package cli

import (
	"reflect"
	"strings"
	"testing"
)

var flags = []Flag{
	{Short: 'a', Long: "all"},
	{Short: 'b'},
	{Short: 'o', Long: "output", Value: true},
	{Long: "html"},
}

func TestParse(t *testing.T) {
	for _, tc := range []struct {
		args     []string
		set      map[string]string
		operands []string
	}{
		{nil, map[string]string{}, nil},
		{[]string{"-a"}, map[string]string{"all": ""}, nil},
		{[]string{"-ab"}, map[string]string{"all": "", "b": ""}, nil},
		{[]string{"-oFILE"}, map[string]string{"output": "FILE"}, nil},
		{[]string{"-o", "FILE"}, map[string]string{"output": "FILE"}, nil},
		{[]string{"-abo", "FILE"}, map[string]string{"all": "", "b": "", "output": "FILE"}, nil},
		{[]string{"-aoFILE"}, map[string]string{"all": "", "output": "FILE"}, nil},
		{[]string{"-o", "-"}, map[string]string{"output": "-"}, nil},
		{[]string{"--output=FILE"}, map[string]string{"output": "FILE"}, nil},
		{[]string{"--output", "FILE"}, map[string]string{"output": "FILE"}, nil},
		{[]string{"--output="}, map[string]string{"output": ""}, nil},
		{[]string{"--html", "--all"}, map[string]string{"html": "", "all": ""}, nil},
		{[]string{"-o", "A", "-o", "B"}, map[string]string{"output": "B"}, nil},
		{[]string{"-a", "x", "-b"}, map[string]string{"all": ""}, []string{"x", "-b"}},
		{[]string{"-a", "--", "-b"}, map[string]string{"all": ""}, []string{"-b"}},
		{[]string{"-", "-a"}, map[string]string{}, []string{"-", "-a"}},
	} {
		set, operands, err := Parse(tc.args, flags)
		if err != nil {
			t.Errorf("%q: %v", tc.args, err)
			continue
		}
		if !reflect.DeepEqual(set, tc.set) || !reflect.DeepEqual(operands, tc.operands) {
			t.Errorf("%q: got %v %q, want %v %q", tc.args, set, operands, tc.set, tc.operands)
		}
	}
}

func TestParseErrors(t *testing.T) {
	for _, tc := range []struct {
		args []string
		err  string
	}{
		{[]string{"-x"}, "unknown flag -x"},
		{[]string{"-ax"}, "unknown flag -x"},
		{[]string{"--pdf"}, "unknown flag --pdf"},
		{[]string{"--out=FILE"}, "unknown flag --out"},
		{[]string{"-o"}, "flag -o needs a value"},
		{[]string{"--output"}, "flag --output needs a value"},
		{[]string{"--html=yes"}, "flag --html takes no value"},
		{[]string{"-h"}, "unknown flag -h"},
	} {
		if _, _, err := Parse(tc.args, flags); err == nil || err.Error() != tc.err {
			t.Errorf("%q: got error %v, want %q", tc.args, err, tc.err)
		}
	}
}

func TestIDsReadsTheFirstFieldOfEachLineForADash(t *testing.T) {
	stdin := strings.NewReader("TK-0001\tRefuted\n\nRQ-0002\n  TD-0003  \tx\ty\n")
	ids, err := IDs([]string{"IN-0001", "-", "IN-0002"}, stdin)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"IN-0001", "TK-0001", "RQ-0002", "TD-0003", "IN-0002"}
	if !reflect.DeepEqual(ids, want) {
		t.Errorf("got %q, want %q", ids, want)
	}
}
