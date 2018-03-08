package main

import (
	"fmt"
	"plugin"
	"time"
)

func main() {
	now := time.Now()

	p, err := plugin.Open("tzdata.so")
	if err != nil {
		panic(err)
	}

	f, err := p.Lookup("GetTZData")
	if err != nil {
		panic(err)
	}

	fn := f.(func(string) (string, *time.Location, error))
	name, loc, err := fn("new york")
	if err != nil {
		panic(err)
	}
	fmt.Println(name)
	fmt.Println(now.In(loc))
}
