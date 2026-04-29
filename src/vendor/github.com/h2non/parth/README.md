# parth

    go get -u github.com/h2non/parth

Package parth provides path parsing for segment unmarshaling and slicing. In
other words, parth provides simple and flexible access to (URL) path parameters.

Along with string, all basic non-alias types are supported. An interface is
available for implementation by user-defined types. When handling an int, uint,
or float of any size, the first valid value within the specified segment will be
used.

## Usage

```go
Variables
func Segment(path string, i int, v interface{}) error
func Sequent(path, key string, v interface{}) error
func Span(path string, i, j int) (string, error)
func SubSeg(path, key string, i int, v interface{}) error
func SubSpan(path, key string, i, j int) (string, error)
type Parth
    func New(path string) *Parth
    func NewBySpan(path string, i, j int) *Parth
    func NewBySubSpan(path, key string, i, j int) *Parth
    func (p *Parth) Err() error
    func (p *Parth) Segment(i int, v interface{})
    func (p *Parth) Sequent(key string, v interface{})
    func (p *Parth) Span(i, j int) string
    func (p *Parth) SubSeg(key string, i int, v interface{})
    func (p *Parth) SubSpan(key string, i, j int) string
type Unmarshaler
```

### Setup ("By Index")

```go
import (
    "fmt"

    "github.com/codemodus/parth"
)

func handler(w http.ResponseWriter, r *http.Request) {
    var s string
    if err := parth.Segment(r.URL.Path, 4, &s); err != nil {
        fmt.Fprintln(os.Stderr, err)
    }

    fmt.Println(r.URL.Path)
    fmt.Printf("%v (%T)\n", s, s)

    // Output:
    // /some/path/things/42/others/3
    // others (string)
}
```

### Setup ("By Key")

```go
import (
    "fmt"

    "github.com/codemodus/parth"
)

func handler(w http.ResponseWriter, r *http.Request) {
    var i int64
    if err := parth.Sequent(r.URL.Path, "things", &i); err != nil {
        fmt.Fprintln(os.Stderr, err)
    }

    fmt.Println(r.URL.Path)
    fmt.Printf("%v (%T)\n", i, i)

    // Output:
    // /some/path/things/42/others/3
    // 42 (int64)
}
```

### Setup (Parth Type)

```go
import (
    "fmt"

    "github.com/codemodus/parth"
)

func handler(w http.ResponseWriter, r *http.Request) {
    var s string
    var f float32

    p := parth.New(r.URL.Path)
    p.Segment(2, &s)
    p.SubSeg("key", 1, &f)
    if err := p.Err(); err != nil {
        fmt.Fprintln(os.Stderr, err)
    }

    fmt.Println(r.URL.Path)
    fmt.Printf("%v (%T)\n", s, s)
    fmt.Printf("%v (%T)\n", f, f)

    // Output:
    // /zero/one/two/key/four/5.5/six
    // two (string)
    // 5.5 (float32)
}
```

### Setup (Unmarshaler)

```go
import (
    "fmt"

    "github.com/codemodus/parth"
)

func handler(w http.ResponseWriter, r *http.Request) {
    /*
        type mytype []byte

        func (m *mytype) UnmarshalSegment(seg string) error {
            *m = []byte(seg)
        }
    */

    var m mytype
    if err := parth.Segment(r.URL.Path, 4, &m); err != nil {
        fmt.Fprintln(os.Stderr, err)
    }

    fmt.Println(r.URL.Path)
    fmt.Printf("%v == %q (%T)\n", m, m, m)

    // Output:
    // /zero/one/two/key/four/5.5/six
    // [102 111 117 114] == "four" (mypkg.mytype)
}
```

## More Info

### Keep Using http.HandlerFunc And Minimize context.Context Usage

The most obvious use case for parth is when working with any URL path such as
the one found at http.Request.URL.Path. parth is fast enough that it can be used
multiple times in place of a single use of similar router-parameter schemes or
even context.Context. There is no need to use an alternate http handler function
definition in order to pass data that is already being passed. The http.Request
type already holds URL data and parth is great at handling it. Additionally,
parth takes care of parsing selected path segments into the types actually
needed. Parth not only does more, it's usually faster and less intrusive than
the alternatives.

### Indexes

If an index is negative, the negative count begins with the last segment.
Providing a 0 for the second index is a special case which acts as an alias for
the end of the path. An error is returned if: 1. Any index is out of range of
the path; 2. When there are two indexes, the first index does not precede the
second index.

### Keys

If a key is involved, functions will only handle the portion of the path
subsequent to the provided key. An error is returned if the key cannot be found
in the path.

### First Whole, First Decimal (Restated - Important!)

When handling an int, uint, or float of any size, the first valid value within
the specified segment will be used.

## Documentation

View the [GoDoc](http://godoc.org/github.com/codemodus/parth)

## Benchmarks

    Go 1.11
    benchmark                             iter       time/iter   bytes alloc        allocs
    ---------                             ----       ---------   -----------        ------
    BenchmarkSegmentString-8          30000000     39.60 ns/op        0 B/op   0 allocs/op
    BenchmarkSegmentInt-8             20000000     65.60 ns/op        0 B/op   0 allocs/op
    BenchmarkSegmentIntNegIndex-8     20000000     86.60 ns/op        0 B/op   0 allocs/op
    BenchmarkSpan-8                  100000000     18.20 ns/op        0 B/op   0 allocs/op
    BenchmarkStdlibSegmentString-8     5000000    454.00 ns/op       50 B/op   2 allocs/op
    BenchmarkStdlibSegmentInt-8        3000000    526.00 ns/op       50 B/op   2 allocs/op
    BenchmarkStdlibSpan-8              3000000    518.00 ns/op       69 B/op   2 allocs/op
    BenchmarkContextLookupSetGet-8     1000000   1984.00 ns/op      480 B/op   6 allocs/op

