package tarbuild

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func applyCHMOD(dst, src *Dir, op tarOp) error {
	var (
		args      = op.Args
		paths     []string
		recursive bool
		mode      uint32
		mask      uint32
	)

	if len(args) >= 1 && args[0] == "-R" {
		recursive = true
		args = args[1:]
	}

	if len(args) >= 1 {
		re := regexp.MustCompile(`^(0[0-7]{3})|(([augw]+)([+=-])([rwx]+))$`)
		m := re.FindStringSubmatch(args[0])
		if m == nil {
			return fmt.Errorf("usage: CHMOD [-R] <mode> <glob>...")
		}
		args = args[1:]

		if len(m[1]) > 0 {
			mask = 0xFFFFFFFF
			if i, err := strconv.ParseUint(m[1], 8, 32); err != nil {
				return fmt.Errorf("usage: CHMOD [-R] <mode> <glob>...")
			} else {
				mode = uint32(i)
			}
		}

		if len(m[2]) > 0 {
			mask = 0
			mode = 0

			amask := uint32(0)
			if strings.IndexByte(m[3], 'a') >= 0 {
				amask = 0xFFFFFFFF
			}
			if strings.IndexByte(m[3], 'u') >= 0 {
				amask |= 0700
			}
			if strings.IndexByte(m[3], 'g') >= 0 {
				amask |= 0070
			}
			if strings.IndexByte(m[3], 'w') >= 0 {
				amask |= 0007
			}

			pmask := uint32(0)
			if strings.IndexByte(m[5], 'r') >= 0 {
				pmask |= 0444
			}
			if strings.IndexByte(m[5], 'w') >= 0 {
				pmask |= 0222
			}
			if strings.IndexByte(m[5], 'x') >= 0 {
				pmask |= 0111
			}

			if m[4] == "=" {
				mask = amask
				mode = pmask
			}
			if m[4] == "+" {
				mask = (amask & pmask)
				mode = 0xFFFFFFFF
			}
			if m[4] == "-" {
				mask = (amask & pmask)
				mode = 0x00000000
			}
		}
	}

	if len(args) == 0 {
		return fmt.Errorf("usage: CHMOD [-R] <mode> <glob>...")
	}

	paths = args

	for _, n := range dst.DeepEntries {
		for _, p := range paths {
			m, err := filepath.Match(p, n)
			if err != nil {
				return err
			}
			if m {
				e, err := dst.GetEntry(n)
				if err != nil {
					return err
				}

				e.chmod(mask, mode, recursive)
			}
		}
	}

	dst.BakeDeepEntries()
	return nil
}
