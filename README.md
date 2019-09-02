# go-eek

Blazingly fast and secure go evaluation library, created on top of [Go `pkg/plugin` package](https://golang.org/pkg/plugin/).

On go-eek, the eval expression is encapsulated into a single function, and stored in a go file. The go file later on will be build into a plugin file (`*.so` file). And then next, for every evaluation call, it will happen in the plugin file. This is why go-eek is insanely fast.

go-eek accept standar Go syntax expression.

## Example

#### Simple Example

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

#### More Complex Example

```go
obj := eek.New("evaluation with 3rd party library")

obj.ImportPackage("fmt")
obj.ImportPackage("github.com/novalagung/gubrak")

obj.DefineVariable(eek.Var{Name: "MessageWin", Type: "string", DefaultValue: "Congrats! You win the lottery!"})
obj.DefineVariable(eek.Var{Name: "MessageLose", Type: "string", DefaultValue: "You lose"})
obj.DefineVariable(eek.Var{Name: "YourLotteryCode", Type: "int"})
obj.DefineVariable(eek.Var{Name: "RepeatUntil", Type: "int", DefaultValue: 5})

obj.PrepareEvalutation(`
    generateRandomNumber := func() int {
        return gubrak.RandomInt(0, 10)
    }

    i := 0
    for i < RepeatUntil {
        if generateRandomNumber() == YourLotteryCode {
            return fmt.Sprintf("%s after %d tried", MessageWin, i + 1)
        }

        i++
    }
    
    return MessageLose
`)

err := obj.Build()
if err != nil {
    log.Fatal(err)
}

output, _ = obj.Evaluate(eek.ExecVar{
    "YourLotteryCode": 3,
    "RepeatUntil":     10,
})
fmt.Println("output:", output)
```

#### Arithmethic expression Example

```go
obj := eek.New("aritmethic expressions")
obj.DefineVariable(eek.Var{Name: "N", Type: "int", DefaultValue: 34})
obj.PrepareEvalutation(`
    IF := func(cond bool, ok, nok string) string {
        if cond {
            return ok
        } else {
            return nok
        }
    }

    OR := func(cond1, cond2 bool) bool {
        return cond1 || cond2
    }

    NOT := func(cond bool) bool {
        return !cond
    }

    message := IF(N>20,IF(OR(N>40,N==40),IF(N>60,IF(NOT(N>80),"good",IF(N==90,"perfect","terrific")),"ok"),"ok, but still bad"),"bad")
    
    return message
`)

err := obj.Build()
if err != nil {
    log.Fatal(err)
}

output, _ := obj.Evaluate(eek.ExecVar{"N": 76})
fmt.Println(output)
```

More example available on the `*_test.go` file.

## Documentation

[Godoc documentation](http://godoc.org/github.com/novalagung/go-eek)

## Author

Noval Agung Prayog

## License

MIT License
