// Package cli reads the command line of bcr: getopt-style flags
// (ADR-0006).
package cli

import (
	"fmt"
	"strings"
)

// Flag is a flag a command accepts.
type Flag struct {
	Short rune   // its letter, such as 'o'; 0 when it has none
	Long  string // its name, such as "output"; "" when it has none
	Value bool   // whether it takes a value
}

// name is how a flag is known in the result of Parse: its long name,
// or its letter when it has none.
func (f Flag) name() string {
	if f.Long != "" {
		return f.Long
	}
	return string(f.Short)
}

// Parse reads the flags in args and returns them with the operands
// that follow. Flags come before operands (ADR-0006): the first operand,
// or `--`, ends them. A flag given more than once keeps its last value;
// a flag with no value maps to "".
func Parse(args []string, flags []Flag) (map[string]string, []string, error) {
	set := map[string]string{}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--":
			return set, args[i+1:], nil
		case strings.HasPrefix(a, "--"):
			name, value, hasValue := strings.Cut(a[2:], "=")
			f, ok := findLong(flags, name)
			if !ok {
				return nil, nil, fmt.Errorf("unknown flag --%s", name)
			}
			switch {
			case !f.Value && hasValue:
				return nil, nil, fmt.Errorf("flag --%s takes no value", name)
			case f.Value && !hasValue:
				if i+1 == len(args) {
					return nil, nil, fmt.Errorf("flag --%s needs a value", name)
				}
				i++
				value = args[i]
			}
			set[f.name()] = value
		case strings.HasPrefix(a, "-") && a != "-":
			letters := []rune(a[1:])
			for j, c := range letters {
				f, ok := findShort(flags, c)
				if !ok {
					return nil, nil, fmt.Errorf("unknown flag -%c", c)
				}
				if !f.Value {
					set[f.name()] = ""
					continue
				}
				value := string(letters[j+1:])
				if value == "" {
					if i+1 == len(args) {
						return nil, nil, fmt.Errorf("flag -%c needs a value", c)
					}
					i++
					value = args[i]
				}
				set[f.name()] = value
				break
			}
		default:
			return set, args[i:], nil
		}
	}
	return set, nil, nil
}

func findLong(flags []Flag, name string) (Flag, bool) {
	for _, f := range flags {
		if f.Long != "" && f.Long == name {
			return f, true
		}
	}
	return Flag{}, false
}

func findShort(flags []Flag, c rune) (Flag, bool) {
	for _, f := range flags {
		if f.Short != 0 && f.Short == c {
			return f, true
		}
	}
	return Flag{}, false
}
