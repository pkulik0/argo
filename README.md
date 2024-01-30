# Argo 

**Argo** is a user-friendly argument parser for Go.

It supports positional arguments, environment variables, default values, and flags.

Configuration is done through struct tags.


## Example

```go
package main
    
import "github.com/pkulik0/argo"

type myArgs struct {
	Address string `argo:"short,long=addr"`
	Port    int    `argo:"short,long,required"`
}

type myArgs2 struct {
	Username string `argo:"positional,default=admin"`
	Password int    `argo:"required,env"`
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
- `positional` - enables a positional argument at the next available position, can't be used with `short` or `long`
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
- `interface`
- Use `argo.RegisterSetter()` to register a custom setter for a type
