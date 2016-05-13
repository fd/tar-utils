package tarbuild

import (
	"fmt"
	"path/filepath"
	"strings"
)

func applyCHOWN(dst, src *Dir, op tarOp) error {
	var (
		args      = op.Args
		paths     []string
		recursive bool
		user      string
		group     string
	)

	if len(args) >= 1 && args[0] == "-R" {
		recursive = true
		args = args[1:]
	}

	if len(args) >= 1 {
		user = args[0]
		if idx := strings.IndexByte(user, ':'); idx >= 0 {
			group = user[idx+1:]
			user = user[:idx]
		}
		args = args[1:]
	}

	if len(args) == 0 {
		return fmt.Errorf("usage: CHOWN [-R] (<user> | <user>:<group> | :<group>) <glob>...")
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

				e.chown(user, group, recursive)
			}
		}
	}

	dst.BakeDeepEntries()
	return nil
}
