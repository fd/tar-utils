package tarbuild

import (
	"archive/tar"
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sabhiram/go-git-ignore"
)

func NewDir() *Dir {
	return &Dir{
		Name:  "/",
		Perm:  0755,
		User:  "root",
		Group: "root",
	}
}

func NewDirFromOS(root string) (*Dir, error) {
	rootDir := NewDir()

	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(root, func(name string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if root == name {
			return nil
		}
		name = strings.TrimPrefix(name, root)

		if fi.IsDir() {
			dir, err := rootDir.MkdirAll(name)
			if err != nil {
				return err
			}

			dir.Perm = fi.Mode().Perm()
		}

		if fi.Mode().IsRegular() {
			file, err := rootDir.AddFile(name, path.Join(root, name))
			if err != nil {
				return err
			}

			file.Perm = fi.Mode().Perm()
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	rootDir.BakeDeepEntries()
	return rootDir, nil
}

type Entry interface {
	isDir() bool
	name() string
	mode() os.FileMode
	bakeDeepEntries() []string
	applyIgnore(ignoreFileName string) error
	chown(user, group string, recursive bool)
	chmod(mask, mode uint32, recursive bool)
	writeToTar(path string, w *tar.Writer) error
}

type Dir struct {
	Name        string
	Perm        os.FileMode
	User        string
	Group       string
	Entries     []Entry
	DeepEntries []string
}

type File struct {
	Name         string
	Perm         os.FileMode
	User         string
	Group        string
	OriginalName string
}

func (d *Dir) Add(name string, entry Entry) (Entry, error) {
	name = path.Join(".", path.Join("/", name))

	dirName, fileName := path.Split(name)
	dirName = path.Join(".", path.Join("/", dirName))
	d, err := d.MkdirAll(dirName)
	if err != nil {
		return nil, err
	}

	e, err := d.GetEntry(fileName)
	if err == nil {
		if e.isDir() {
			return e.(*Dir).Add(entry.name(), entry)
		}
		if entry.isDir() {
			return nil, os.ErrExist
		}
		err := d.Remove(fileName)
		if err != nil {
			return nil, err
		}
		return d.Add(fileName, entry)
	}
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	if entry.isDir() {
		src := entry.(*Dir)
		dst := &Dir{}
		*dst = *src
		dst.Entries = nil
		dst.DeepEntries = nil
		dst.Entries = append(dst.Entries, src.Entries...)
		dst.Name = fileName
		d.Entries = append(d.Entries, dst)
		sort.Sort(d)
		return dst, nil
	}

	src := entry.(*File)
	dst := &File{}
	*dst = *src
	dst.Name = fileName
	d.Entries = append(d.Entries, dst)
	sort.Sort(d)
	return dst, nil
}

func (d *Dir) AddFile(name, original string) (*File, error) {
	name = path.Join(".", path.Join("/", name))

	dirName, fileName := path.Split(name)
	dirName = path.Join(".", path.Join("/", dirName))
	d, err := d.MkdirAll(dirName)
	if err != nil {
		return nil, err
	}

	_, err = d.GetEntry(fileName)
	if err == nil {
		err = os.ErrExist
	}
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	file := &File{
		Name:         fileName,
		Perm:         0644,
		User:         "root",
		Group:        "root",
		OriginalName: original,
	}

	d.Entries = append(d.Entries, file)
	sort.Sort(d)
	return file, nil
}

func (d *Dir) Mkdir(name string) (*Dir, error) {
	name = path.Join(".", path.Join("/", name))

	if name == "." {
		return nil, os.ErrExist
	}

	idx := strings.IndexRune(name, '/')
	if idx >= 0 {
		dirName, baseName := path.Split(name)
		dirName = path.Join(".", path.Join("/", dirName))
		dir, err := d.GetDir(dirName)
		if err != nil {
			return nil, err
		}
		return dir.Mkdir(baseName)
	}

	_, err := d.GetEntry(name)
	if err == nil {
		err = os.ErrExist
	}
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	dir := &Dir{
		Name:  name,
		Perm:  0755,
		User:  "root",
		Group: "root",
	}

	d.Entries = append(d.Entries, dir)
	sort.Sort(d)
	return dir, nil
}

func (d *Dir) MkdirAll(name string) (*Dir, error) {
	name = path.Join(".", path.Join("/", name))

	dir, err := d.GetDir(name)
	if err == nil {
		return dir, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	dirName, fileName := path.Split(name)
	dirName = path.Join(".", path.Join("/", dirName))
	d, err = d.MkdirAll(dirName)
	if err != nil {
		return nil, err
	}

	return d.Mkdir(fileName)
}

func (d *Dir) Remove(name string) error {
	name = path.Join(".", path.Join("/", name))

	dirName, eName := path.Split(name)
	dirName = path.Join(".", path.Join("/", dirName))
	if dirName != "." {

		dir, err := d.GetDir(dirName)
		if err != nil {
			return err
		}
		return dir.Remove(eName)
	}

	for i, e := range d.Entries {
		if e.name() != eName {
			continue
		}

		copy(d.Entries[i:], d.Entries[i+1:])
		d.Entries = d.Entries[:len(d.Entries)-1]
		return nil
	}

	return os.ErrNotExist
}

func (d *Dir) GetDir(name string) (*Dir, error) {
	e, err := d.GetEntry(name)
	if err != nil {
		return nil, err
	}
	d, ok := e.(*Dir)
	if !ok || d == nil {
		return nil, os.ErrPermission
	}
	return d, nil
}

func (d *Dir) GetFile(name string) (*File, error) {
	e, err := d.GetEntry(name)
	if err != nil {
		return nil, err
	}
	f, ok := e.(*File)
	if !ok || f == nil {
		return nil, os.ErrPermission
	}
	return f, nil
}

func (d *Dir) ReadFile(name string) ([]byte, error) {
	f, err := d.GetFile(name)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(f.OriginalName)
}

func (d *Dir) GetEntry(name string) (Entry, error) {
	name = path.Join(".", path.Join("/", name))

	if name == "." {
		return d, nil
	}

	idx := strings.IndexRune(name, '/')
	rem := ""
	if idx > 0 {
		rem = name[idx+1:]
		name = name[:idx]
	}

	for _, e := range d.Entries {
		if e.name() == name {
			if rem == "" {
				return e, nil
			}
			dir, ok := e.(*Dir)
			if !ok || dir == nil {
				return nil, os.ErrPermission
			}
			return dir.GetEntry(rem)
		}
	}

	return nil, os.ErrNotExist
}

func (d *Dir) Len() int {
	return len(d.Entries)
}

func (d *Dir) Less(i, j int) bool {
	a := d.Entries[i].name()
	b := d.Entries[j].name()
	return a < b
}

func (d *Dir) Swap(i, j int) {
	d.Entries[i], d.Entries[j] = d.Entries[j], d.Entries[i]
}

func (d *Dir) name() string {
	return d.Name
}

func (d *Dir) isDir() bool {
	return true
}

func (f *File) isDir() bool {
	return false
}

func (d *Dir) BakeDeepEntries() {
	d.bakeDeepEntries()
}

func (d *Dir) bakeDeepEntries() []string {
	var deep = d.DeepEntries[:0]

	for _, e := range d.Entries {
		// add e
		deep = append(deep, e.name())

		for _, c := range e.bakeDeepEntries() {
			deep = append(deep, path.Join(e.name(), c))
		}
	}

	sort.Strings(deep)

	d.DeepEntries = deep
	return deep
}

func (f *File) name() string {
	return f.Name
}

func (f *File) bakeDeepEntries() []string {
	return nil
}

func (d *Dir) ApplyIgnore(ignoreFileName string) error {
	err := d.applyIgnore(ignoreFileName)
	if err != nil {
		return err
	}

	d.bakeDeepEntries()
	return nil
}

func (f *File) applyIgnore(ignoreFileName string) error {
	return nil
}

func (d *Dir) applyIgnore(ignoreFileName string) error {
	data, err := d.ReadFile(ignoreFileName)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	ignoreRules, err := ignore.CompileIgnoreLines(lines...)
	if err != nil {
		return err
	}

	for _, n := range d.DeepEntries {
		if ignoreRules.MatchesPath(n) {
			err := d.Remove(n)
			if os.IsNotExist(err) {
				err = nil
			}
			if err != nil {
				return err
			}
			continue
		}
	}

	for _, e := range d.Entries {
		err := e.applyIgnore(ignoreFileName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Dir) chown(user, group string, recursive bool) {
	if user != "" {
		d.User = user
	}
	if group != "" {
		d.Group = group
	}
	if recursive {
		for _, e := range d.Entries {
			e.chown(user, group, recursive)
		}
	}
}

func (f *File) chown(user, group string, recursive bool) {
	if user != "" {
		f.User = user
	}
	if group != "" {
		f.Group = group
	}
}

func (d *Dir) chmod(mask, mode uint32, recursive bool) {
	d.Perm = os.FileMode((mask & mode) | (uint32(d.Perm) & (0xFFFFFFFF ^ mask)))
	if recursive {
		for _, e := range d.Entries {
			e.chmod(mask, mode, recursive)
		}
	}
}

func (f *File) chmod(mask, mode uint32, recursive bool) {
	f.Perm = os.FileMode((mask & mode) | (uint32(f.Perm) & (0xFFFFFFFF ^ mask)))
}

func (d *Dir) mode() os.FileMode  { return d.Perm }
func (f *File) mode() os.FileMode { return f.Perm }

func (d *Dir) writeEntriesToTar(path string, w *tar.Writer) error {
	for _, e := range d.Entries {
		err := e.writeToTar(filepath.Join(path, e.name()), w)
		if err != nil {
			return err
		}
	}
	return nil
}

var ftime = time.Date(1988, time.February, 1, 0, 0, 0, 0, time.UTC)

func (d *Dir) writeToTar(path string, w *tar.Writer) error {
	h := tar.Header{
		Typeflag:   tar.TypeDir,
		Mode:       int64(d.Perm | c_ISDIR),
		Name:       path + "/",
		Uname:      d.User,
		Gname:      d.Group,
		Size:       0,
		AccessTime: ftime,
		ChangeTime: ftime,
		ModTime:    ftime,
	}

	err := w.WriteHeader(&h)
	if err != nil {
		return err
	}

	return d.writeEntriesToTar(path, w)
}

func (f *File) writeToTar(path string, w *tar.Writer) error {
	data, err := ioutil.ReadFile(f.OriginalName)
	if err != nil {
		return err
	}

	h := tar.Header{
		Typeflag:   tar.TypeReg,
		Mode:       int64(f.Perm | c_ISREG),
		Name:       path,
		Uname:      f.User,
		Gname:      f.Group,
		Size:       int64(len(data)),
		AccessTime: ftime,
		ChangeTime: ftime,
		ModTime:    ftime,
	}

	err = w.WriteHeader(&h)
	if err != nil {
		return err
	}

	_, err = bytes.NewReader(data).WriteTo(w)
	return err
}

// Mode constants from the tar spec.
const (
	c_ISUID  = 04000   // Set uid
	c_ISGID  = 02000   // Set gid
	c_ISVTX  = 01000   // Save text (sticky bit)
	c_ISDIR  = 040000  // Directory
	c_ISFIFO = 010000  // FIFO
	c_ISREG  = 0100000 // Regular file
	c_ISLNK  = 0120000 // Symbolic link
	c_ISBLK  = 060000  // Block special file
	c_ISCHR  = 020000  // Character special file
	c_ISSOCK = 0140000 // Socket
)
