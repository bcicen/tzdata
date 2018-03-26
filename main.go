//go:generate go run util/embd.go

package tzdata

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"time"
)

var (
	Built = ""
	// LocationNames contains all available timezone names
	LocationNames = allLocations()

	loaded = make(map[string]*time.Location)
)

// Uncompress and load all timezone data into memory for quicker access
func Preload() {
	for k, _ := range tzdata {
		loc, err := Load(k)
		if err != nil {
			panic(err)
		}
		loaded[k] = loc
	}
}

// Load a timezone Location by name from the embedded tz database
func Load(name string) (*time.Location, error) {
	name = resolveAlias(name)
	if _, ok := tzdata[name]; !ok {
		return nil, fmt.Errorf("unknown timezone: %s", name)
	}

	rbuf := bytes.NewBuffer(tzdata[name])
	gz, err := gzip.NewReader(rbuf)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	chunk := [20]byte{}
	var wbuf bytes.Buffer

gzloop:
	for {
		n, err := gz.Read(chunk[:])
		switch err {
		case io.ErrUnexpectedEOF:
			wbuf.Write(chunk[:n])
			fallthrough
		case io.EOF:
			break gzloop
		case nil:
			wbuf.Write(chunk[:])
		default:
			return nil, err
		}
	}

	return time.LoadLocationFromTZData(name, wbuf.Bytes())
}

func allLocations() (a []string) {
	for k, _ := range tzdata {
		a = append(a, k)
	}
	for k, _ := range aliases {
		a = append(a, k)
	}
	return a
}

func resolveAlias(name string) string {
	if aname, ok := aliases[name]; ok {
		return resolveAlias(aname)
	}
	return name
}
