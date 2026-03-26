package cli

import (
	"fmt"
	"strconv"
	"strings"
)

type flagType int

const (
	stringFlag flagType = iota
	boolFlag
	intFlag
)

type parsedArgs struct {
	values map[string]string
	args   []string
}

func parseCommandArgs(args []string, specs map[string]flagType) (*parsedArgs, error) {
	p := &parsedArgs{values: map[string]string{}}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			p.args = append(p.args, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "--") || arg == "--" {
			p.args = append(p.args, arg)
			continue
		}

		nameValue := strings.TrimPrefix(arg, "--")
		name, value, hasValue := strings.Cut(nameValue, "=")
		typ, ok := specs[name]
		if !ok {
			return nil, fmt.Errorf("unknown flag: --%s", name)
		}

		switch typ {
		case boolFlag:
			if !hasValue {
				p.values[name] = "true"
				continue
			}
			if value != "true" && value != "false" {
				return nil, fmt.Errorf("invalid boolean for --%s: %s", name, value)
			}
			p.values[name] = value
		case stringFlag, intFlag:
			if !hasValue {
				i++
				if i >= len(args) {
					return nil, fmt.Errorf("missing value for --%s", name)
				}
				value = args[i]
			}
			p.values[name] = value
		}
	}

	return p, nil
}

func (p *parsedArgs) String(name, def string) string {
	if v, ok := p.values[name]; ok {
		return v
	}
	return def
}

func (p *parsedArgs) Bool(name string) bool {
	return p.values[name] == "true"
}

func (p *parsedArgs) Int(name string, def int) (int, error) {
	v, ok := p.values[name]
	if !ok || v == "" {
		return def, nil
	}
	return strconv.Atoi(v)
}
