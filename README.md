# Argo 

**Argo** is a user-friendly argument parser for Go.

It supports positional arguments, environment variables, default values, and flags.

Configuration is done through struct tags.


## Example

```go
package main
    
import "github.com/pkulik0/argo"

type example struct {
	Address string `argo:"short,long=addr"`
	Port    int    `argo:"short,long,required"`
}

type example2 struct {
	SecretNumber int8   `argo:"required,env"`
	Username     string `argo:"positional,default=admin"`
}

type example3 struct {
	Pi          float64 `argo:"short,long,default=3.14"`
	ApiKey      string  `argo:"env,required"`
	Verbose     bool    `argo:"short,long"`
	Name        string  `argo:"short,long,default=John"`
	Source      string  `argo:"positional"`
	Destination string  `argo:"positional,default=."`
}

func main() {
	args := &example{}
	err := argo.Parse(args)
	if err != nil {
		panic(err)
	}
}
```

## Supported attributes

- `short` - enables a single character flag 
- `long` - enables a multi character flag
- `positional` - enables a positional argument 
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
