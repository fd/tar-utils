package tarbuild

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type tarSpec struct {
	Commands []tarOp
}

type tarOp struct {
	Name string
	Args []string
}

func (s *tarSpec) validate() error {
	for _, op := range s.Commands {
		if err := op.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (op *tarOp) validate() error {
	if op.Name == "COPY" {
		if len(op.Args) == 0 {
			return fmt.Errorf("invalid command: %q requires arguments", op.Name)
		}
		if len(op.Args) == 1 {
			op.Args = append(op.Args, op.Args[0])
		}
		return nil
	}

	if op.Name == "MKDIR" {
		if len(op.Args) == 0 {
			return fmt.Errorf("invalid command: %q requires arguments", op.Name)
		}
		return nil
	}

	if op.Name == "CHMOD" {
		if len(op.Args) == 0 {
			return fmt.Errorf("invalid command: %q requires arguments", op.Name)
		}
		return nil
	}

	if op.Name == "CHOWN" {
		if len(op.Args) == 0 {
			return fmt.Errorf("invalid command: %q requires arguments", op.Name)
		}
		return nil
	}

	return fmt.Errorf("invalid command %q", op.Name)
}

func parseConf(data []byte) (*tarSpec, error) {
	data = bytes.TrimSpace(data)

	if len(data) > 0 && data[0] == '{' {
		var spec *tarSpec
		err := json.Unmarshal(data, &spec)
		if err != nil {
			return nil, fmt.Errorf("invalid spec: %q (%v)", data, err)
		}
		return spec, nil
	}

	return parseTextConf(data)
}

func parseTextConf(data []byte) (*tarSpec, error) {
	var buf bytes.Buffer

	for inComment, idx := false, 0; idx < len(data); idx++ {
		c := data[idx]
		if inComment {
			if c == '\n' {
				inComment = false
			}
		} else {
			if c == '#' {
				inComment = true
			} else {
				buf.WriteByte(c)
			}
		}
	}
	data = append(data[:0], buf.Bytes()...)
	buf.Reset()

	data = bytes.Replace(data, []byte("\\\n"), nil, -1)

	for inArgJSON, idx := false, 0; idx < len(data); idx++ {
		c := data[idx]
		if inArgJSON {
			if c == '\n' {
				continue
			} else if c == ']' {
				inArgJSON = false
			}
		} else {
			if c == '[' {
				inArgJSON = true
			}
		}
		buf.WriteByte(c)
	}
	data = append(data[:0], buf.Bytes()...)
	buf.Reset()

	spec := &tarSpec{}

	lines := bytes.Split(data, []byte{'\n'})
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		idx := bytes.IndexByte(line, ' ')
		if idx < 0 {
			return nil, fmt.Errorf("invalid command: %q", line)
		}

		var (
			cmd     = line[:idx]
			args    = bytes.TrimSpace(line[idx:])
			argVals []string
		)
		if len(cmd) == 0 || len(args) == 0 {
			return nil, fmt.Errorf("invalid command: %q", line)
		}
		for _, c := range cmd {
			if 'A' <= c && c <= 'Z' {
				continue
			}
			return nil, fmt.Errorf("invalid command: %q", line)
		}
		if args[0] == '[' {
			err := json.Unmarshal(args, &argVals)
			if err != nil {
				return nil, fmt.Errorf("invalid command: %q (%v)", line, err)
			}
		} else {
			argParts := bytes.Split(args, []byte{' '})
			for _, arg := range argParts {
				arg = bytes.TrimSpace(arg)
				if len(arg) == 0 {
					continue
				}
				argVals = append(argVals, string(arg))
			}
		}

		spec.Commands = append(spec.Commands, tarOp{
			Name: string(cmd),
			Args: argVals,
		})
	}

	return spec, nil
}
