//go:generate go run util/embd.go

package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/dexyk/stringosim"
)

type TZData struct {
	Name     string
	Location *time.Location
}

func main() {
	now := time.Now()
	name, loc, err := GetTZData(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println(name)
	fmt.Println(now.In(loc))
}

func GetTZData(search string) (name string, loc *time.Location, err error) {
	var curDistance int
	var curMatchTz string

	for s, tzname := range tzmap {
		dist := stringosim.Levenshtein([]rune(search), []rune(s),
			stringosim.LevenshteinSimilarityOptions{
				InsertCost:      1,
				DeleteCost:      3,
				SubstituteCost:  5,
				CaseInsensitive: true,
			})

		if name == "" || dist < curDistance {
			name = s
			curDistance = dist
			curMatchTz = tzname
		}
	}

	if curDistance > 8 {
		return name, loc, fmt.Errorf("location not found")
	}

	loc, err = load(curMatchTz)

	return name, loc, err
}

func resolveAlias(name string) string {
	if aname, ok := aliases[name]; ok {
		fmt.Printf("%s->%s\n", name, aname)
		return resolveAlias(aname)
	}
	return name
}

func load(name string) (*time.Location, error) {
	name = resolveAlias(name)
	if data, ok := tzdata[name]; ok {
		return time.LoadLocationFromTZData(name, data)
	}
	return nil, fmt.Errorf("unknown timezone: %s", name)
}
