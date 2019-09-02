# go-eek

Blazingly fast and secure go evaluation library, created on top of [Go plugin API](https://golang.org/pkg/plugin/).

On go-eek, evaluation expression is encapsulated into single function, stored in a go file, and then the particular file will be build into a plugin file (*.so file). Next, for every evaluation call, it will happen by consuming the plugin file. This is why go-eek is insanely fast.

go-eek accept standar Go expression.

## Example

```go
// import "github.com/novalagung/go-eek"

obj := NewEek()
obj.SetName("simple operation")

obj.DefineVariable(Var{Name: "VarA", Type: "int"})
obj.DefineVariable(Var{Name: "VarB", Type: "float64", DefaultValue: 10.5})

obj.PrepareEvalutation(`
    VarACasted := float64(VarA)
    VarC := VarACasted + VarB
    return VarC
`)

err := obj.Build()
if err != nil {
    log.Fatal(err)
}

output, err := obj.Evaluate(ExecVar{
    "A": 9,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println("result is", output)
```

## Author

Noval Agung Prayog

## License

MIT License