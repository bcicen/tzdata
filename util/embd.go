// adaption of https://github.com/akavel/embd-go/blob/master/embd.go

package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	d      = make(DataFiles)
	tzRoot = "/usr/share/zoneinfo/" // trailing slash is required
)

type DataFiles map[string]string // map of varname:path

type Contents struct {
	Args    DataFiles
	Pkg     string
	Files   map[string]File
	Aliases map[string]string
}

type File struct {
	VarName, Path string
	FileInfo      os.FileInfo
	DataFragments <-chan string
}

func visit(path string, f os.FileInfo, err error) error {
	if f.IsDir() {
		return nil
	}

	if strings.Contains(path, "/posix/") {
		return nil
	}

	if strings.Contains(path, "/right/") {
		return nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	if string(data[0:4]) != "TZif" {
		fmt.Printf("skipping non-tz file: %s\n", path)
		return nil
	}

	name := strings.Replace(path, tzRoot, "", -1)
	d[name] = path

	fmt.Printf("found tzdata path: %s\n", path)

	return nil
}

func main() {
	filepath.Walk(tzRoot, visit)
	err := Embed(d, "main", "tzdata.go")
	if err != nil {
		panic(err)
	}
}

func Embed(files DataFiles, pkg, outPath string) error {
	contents := Contents{
		Args:    files,
		Pkg:     pkg,
		Files:   map[string]File{},
		Aliases: map[string]string{},
	}

loop:
	for varname, path := range contents.Args {
		path := filepath.ToSlash(path)

		f, err := NewFile(varname, path)
		if err != nil {
			return err
		}

		for _, efile := range contents.Files {
			if os.SameFile(efile.FileInfo, f.FileInfo) {
				contents.Aliases[efile.VarName] = f.VarName
				fmt.Printf("added alias: %s = %s\n", efile.VarName, f.VarName)
				continue loop
			}
		}

		contents.Files[varname] = f
		fmt.Printf("added file: %s (%s)\n", f.VarName, f.Path)
	}

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	defer w.Flush()

	err = template.Must(template.New("Contents").Parse(Template)).Execute(w, contents)
	if err != nil {
		return err
	}

	return nil
}

func GoEscaped(buf []byte) string {
	s := fmt.Sprintf("%q", string(buf))
	return s[1 : len(s)-1]
}

func NewFile(varname, path string) (File, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return File{}, err
	}

	f, err := os.Open(path)
	if err != nil {
		return File{}, err
	}

	ch := make(chan string)
	go func() {
		defer f.Close()

		r := bufio.NewReader(f)

		buf := [20]byte{}
		for {
			n, err := io.ReadFull(r, buf[:])
			switch err {
			case io.ErrUnexpectedEOF:
				ch <- GoEscaped(buf[:n])
				fallthrough
			case io.EOF:
				close(ch)
				return
			case nil:
				ch <- GoEscaped(buf[:])
			default:
				panic(fmt.Errorf("%s: %s", path, err))
			}
		}
	}()
	return File{
		Path:          path,
		VarName:       varname,
		FileInfo:      stat,
		DataFragments: ch,
	}, nil
}

var Template = `
// DO NOT EDIT BY HAND
//

package {{.Pkg}}
var (
	tzdata = map[string][]byte {
{{range .Files}}
		// {{.VarName}} contains contents of "{{.Path}}" file.
		"{{.VarName}}": []byte("{{range .DataFragments}}{{.}}{{end}}"),
{{end}}`[1:] + `
	}
	aliases = map[string][]byte {
{{range $k, $v := .Aliases}}
		"{{$k}}": tzdata["{{$v}}"],
{{end}}
	}
)`
