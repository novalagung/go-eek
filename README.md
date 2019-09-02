# go-eek

Blazingly fast and secure go evaluation library, created on top of [Go plugin API](https://golang.org/pkg/plugin/).

On go-eek, evaluation expression is encapsulated into single function, stored in a go file, and then the particular file will be build into a plugin file (*.so file). Next, for every evaluation call, it will happen by consuming the plugin file. This is why go-eek is insanely fast.

go-eek accept standar Go expression.

## Example

```go
import "github.com/novalagung/go-eek"

// create new eek object and name it
obj := NewEek()
obj.SetName("simple operation")

// define variables (and default value of particular variable if available)
obj.DefineVariable(Var{Name: "VarA", Type: "int"})
obj.DefineVariable(Var{Name: "VarB", Type: "float64", DefaultValue: 10.5})

// specify the evaluation expression in go standard syntax
obj.PrepareEvalutation(`
    VarACasted := float64(VarA)
    VarC := VarACasted + VarB
    return VarC
`)

// build only need to happen once
err := obj.Build()
if err != nil {
    log.Fatal(err)
}

// evaluate!
output1, _ := obj.Evaluate(ExecVar{ "A": 9 })
fmt.Println("with A = 9, the result will be", output1)
output2, _ := obj.Evaluate(ExecVar{ "A": 12, "B": 12.4 })
fmt.Println("with A = 12 and B = 12.4, the result will be", output2)
```

More example available on the `*_test.go` file.

## Documentation

[Godoc documentation](http://godoc.org/github.com/novalagung/go-eek)

## Author

Noval Agung Prayog

## License

MIT License