package tarbuild

import (
	"path"
	"path/filepath"
	"strings"
)

// COPY obeys the following rules:
//   * The <src> path must be inside the context of the build; you cannot
//     COPY ../something /something, because the first step of a docker build is
//     to send the context directory (and subdirectories) to the docker daemon.
//   * If <src> is a directory, the entire contents of the directory are copied,
//     including filesystem metadata.
//     Note: The directory itself is not copied, just its contents.
//   * If <src> is any other kind of file, it is copied individually along with
//     its metadata. In this case, if <dest> ends with a trailing slash /, it
//     will be considered a directory and the contents of <src> will be written
//     at <dest>/base(<src>).
//   * If multiple <src> resources are specified, either directly or due to the
//     use of a wildcard, then <dest> must be a directory, and it must end with
//     a slash /.
//   * If <dest> does not end with a trailing slash, it will be considered a
//     regular file and the contents of <src> will be written at <dest>.
//   * If <dest> doesn’t exist, it is created along with all missing directories
//     in its path.
func applyCOPY(dstFS, srcFS *Dir, op tarOp) error {
	dst := op.Args[len(op.Args)-1]
	src := op.Args[:len(op.Args)-1]

	var realSrc []Entry

	for _, n := range srcFS.DeepEntries {
		for _, p := range src {
			m, err := filepath.Match(p, n)
			if err != nil {
				return err
			}
			if m {
				e, err := srcFS.GetEntry(n)
				if err != nil {
					return err
				}

				realSrc = append(realSrc, e)
				break
			}
		}
	}

	if len(realSrc) > 1 && !strings.HasSuffix(dst, "/") {
		dst += "/"
	}

	for _, src := range realSrc {
		curDst := dst
		if strings.HasSuffix(curDst, "/") {
			curDst = path.Join(curDst, path.Base(src.name()))
		}

		_, err := dstFS.Add(curDst, src)
		if err != nil {
			return err
		}
	}

	dstFS.BakeDeepEntries()
	return nil
}
