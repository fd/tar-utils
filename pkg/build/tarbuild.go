package tarbuild

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Build(dst io.Writer, wd, conf string) error {
	wd, err := filepath.Abs(wd)
	if err != nil {
		return err
	}

	spec, err := loadTarSpec(conf)
	if err != nil {
		return err
	}

	err = spec.validate()
	if err != nil {
		return err
	}

	dstFS := NewDir()
	srcFS, err := NewDirFromOS(wd)
	if err != nil {
		return err
	}

	for _, op := range spec.Commands {
		err := applyOp(dstFS, srcFS, op)
		if err != nil {
			return err
		}
	}

	var (
		buf bytes.Buffer
		w   = tar.NewWriter(&buf)
	)

	err = dstFS.writeEntriesToTar("/", w)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	_, err = buf.WriteTo(dst)
	return err
}

func applyOp(dst, src *Dir, op tarOp) error {
	switch op.Name {
	case "MKDIR":
		return applyMKDIR(dst, src, op)
	case "COPY":
		return applyCOPY(dst, src, op)
	case "CHMOD":
		return applyCHMOD(dst, src, op)
	case "CHOWN":
		return applyCHOWN(dst, src, op)
	default:
		return fmt.Errorf("unsupported command %q", op.Name)
	}
}

func loadTarSpec(name string) (*tarSpec, error) {
	if name == "-" {
		return maybeParseTarspec(ioutil.ReadAll(os.Stdin))
	}
	return maybeParseTarspec(ioutil.ReadFile(name))
}

func maybeParseTarspec(data []byte, err error) (*tarSpec, error) {
	if err == nil {
		return parseConf(data)
	}
	return nil, err
}
