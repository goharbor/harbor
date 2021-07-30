# uitable [![GoDoc](https://godoc.org/github.com/gosuri/uitable?status.svg)](https://godoc.org/github.com/gosuri/uitable) [![Build Status](https://travis-ci.org/gosuri/uitable.svg?branch=master)](https://travis-ci.org/gosuri/uitable)

uitable is a go library for representing data as tables for terminal applications. It provides primitives for sizing and wrapping columns to improve readability.

## Example Usage

Full source code for the example is available at [example/main.go](example/main.go)

```go
table := uitable.New()
table.MaxColWidth = 50

table.AddRow("NAME", "BIRTHDAY", "BIO")
for _, hacker := range hackers {
  table.AddRow(hacker.Name, hacker.Birthday, hacker.Bio)
}
fmt.Println(table)
```

Will render the data as:

```sh
NAME          BIRTHDAY          BIO
Ada Lovelace  December 10, 1815 Ada was a British mathematician and writer, chi...
Alan Turing   June 23, 1912     Alan was a British pioneering computer scientis...
```

For wrapping in two columns:

```go
table = uitable.New()
table.MaxColWidth = 80
table.Wrap = true // wrap columns

for _, hacker := range hackers {
  table.AddRow("Name:", hacker.Name)
  table.AddRow("Birthday:", hacker.Birthday)
  table.AddRow("Bio:", hacker.Bio)
  table.AddRow("") // blank
}
fmt.Println(table)
```

Will render the data as:

```
Name:     Ada Lovelace
Birthday: December 10, 1815
Bio:      Ada was a British mathematician and writer, chiefly known for her work on
          Charles Babbage's early mechanical general-purpose computer, the Analytical
          Engine

Name:     Alan Turing
Birthday: June 23, 1912
Bio:      Alan was a British pioneering computer scientist, mathematician, logician,
          cryptanalyst and theoretical biologist
```

## Installation

```
$ go get -v github.com/gosuri/uitable
```


[![Bitdeli Badge](https://d2weczhvl823v0.cloudfront.net/gosuri/uitable/trend.png)](https://bitdeli.com/free "Bitdeli Badge")

