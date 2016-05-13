package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/fd/tar-utils/pkg/build"

	"gopkg.in/alecthomas/kingpin.v2"
	"limbo.services/version"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		contextDir  string
		tarfileName string
		outputTar   string
	)

	app := kingpin.New("x-tar", "Tar utilities").Version(version.Get().String()).Author(version.Get().ReleasedBy)

	buildCmd := app.Command("build", "Make a new tar file")
	buildCmd.Arg("context-dir", "The context directory for the build").Default(".").ExistingDirVar(&contextDir)
	buildCmd.Flag("tarfile", "Tarfile location").Short('t').PlaceHolder("FILE").StringVar(&tarfileName)
	buildCmd.Flag("output", "Path to output Tar archive").Short('o').Default("-").PlaceHolder("FILE").StringVar(&outputTar)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case buildCmd.FullCommand():
		if tarfileName == "" {
			tarfileName = path.Join(contextDir, "Tarfile")
		}

		var buf bytes.Buffer

		err := tarbuild.Build(&buf, contextDir, tarfileName)
		if err != nil {
			return err
		}

		err = putStream(outputTar, &buf)
		if err != nil {
			return err
		}
	}

	return nil
}

const stdio = "-"

func openStream(name string) (io.Reader, error) {
	if name == stdio {
		return os.Stdin, nil
	}
	return os.Open(name)
}

func putStream(name string, buf *bytes.Buffer) error {
	if name == stdio {
		_, err := io.Copy(os.Stdout, buf)
		return err
	}
	return ioutil.WriteFile(name, buf.Bytes(), 0644)
}
