# mapenv

## Installation

```
$ go get github.com/h-celel/mapenv
```


## Usage

```go
var o struct{
	Var1 string `mpe:"SOME_ENVIRONMENTAL_VARIABLE"`
	Var2 int    `mpe:"ANOTHER_ENVIRONMENTAL_VARIABLE"`
}{}

err := mapenv.Decode(&o)
if err != nil {
    ...
}
```