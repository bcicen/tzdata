[![GoDoc](https://godoc.org/github.com/bcicen/tzdata?status.svg)](https://godoc.org/github.com/bcicen/tzdata)

# tzdata

Embeddable timezone database for Go projects

## Build

```bash
go get -d github.com/bcicen/tzdata && \
cd ${GOPATH}/github.com/bcicen/tzdata && \
go generate
```

## Usage

Once the timezone data is built, it may be used in place of `time.LoadLocation()`:

```go
package main

import (
	"fmt"
	"time"

	"github.com/bcicen/tzdata"
)

func main() {
	loc, _ := tzdata.Load("America/New_York")
	fmt.Println(time.Now().In(loc))
}
```
