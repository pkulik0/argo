# argo 

**Argo** is a user friendly argument parser for Go.
To use argo you need to define a struct with the arguments you want to parse.
Argo will use the struct tags to know how to parse the arguments.

Almost as effortless as it gets.

## Example

```go
    package main
    
    import "github.com/pkulik0/argo"

    type myArgs struct {
        Address string `argo:"short;long=addr"`
        Port    int    `argo:"short;long;required"`
    }

    func main() {
        args := &myArgs{}
        err := argo.Parse(args)
        if err != nil {
            panic(err)
        }
    }
```

## Supported attributes

- `short` - enables a single character flag 
- `long` - enables a multi character flag
- `required` - makes the argument required
- `env` - query the environment for the argument value
- `default` - provides a default value for the argument
- `help` - provides a help message for the argument

## Supported types

- `string`
- `intN`
- `uintN`
- `floatN`
- `bool`