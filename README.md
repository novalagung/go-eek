# go-eek

Blazingly fast and safe go evaluation library, created on top of [Go `pkg/plugin` package](https://golang.org/pkg/plugin/).

[![Go Report Card](https://goreportcard.com/badge/github.com/novalagung/go-eek?nocache=1)](https://goreportcard.com/report/github.com/novalagung/go-eek?nocache=1)
[![Build Status](https://travis-ci.org/novalagung/go-eek.svg?branch=master)](https://travis-ci.org/novalagung/go-eek)
[![Coverage Status](https://coveralls.io/repos/github/novalagung/go-eek/badge.svg?branch=master)](https://coveralls.io/github/novalagung/go-eek?branch=master)

On go-eek, the eval expression is encapsulated into a single function, and stored in a go file. The go file later on will be build into a plugin file (`*.so` file). And then next, for every evaluation call, it will happen in the plugin file. This is why go-eek is insanely fast.

go-eek accept standar Go syntax expression.

## Example

#### Simple Example

```go
import . "github.com/novalagung/go-eek"

// create new eek object and name it
obj := New()
obj.SetName("simple operation")

// define variables (and default value of particular variable if available)
obj.DefineVariable(Var{Name: "VarA", Type: "int"})
obj.DefineVariable(Var{Name: "VarB", Type: "float64", DefaultValue: 10.5})

// specify the evaluation expression in go standard syntax
obj.PrepareEvaluation(`
    castedVarA := float64(VarA)
    VarC := castedVarA + VarB
    return VarC
`)

// build only need to happen once
err := obj.Build()
if err != nil {
    log.Fatal(err)
}

// evaluate!
output1, _ := obj.Evaluate(ExecVar{ "VarA": 9 })
fmt.Println("with VarA = 9, the result will be", output1)
output2, _ := obj.Evaluate(ExecVar{ "VarA": 12, "VarB": 12.4 })
fmt.Println("with VarA = 12 and VarB = 12.4, the result will be", output2)
```

#### More Complex Example

```go
obj := eek.New("evaluation with 3rd party library")

obj.ImportPackage("fmt")
obj.ImportPackage("github.com/novalagung/gubrak")

obj.DefineVariable(eek.Var{Name: "VarMessageWin", Type: "string", DefaultValue: "Congrats! You win the lottery!"})
obj.DefineVariable(eek.Var{Name: "VarMessageLose", Type: "string", DefaultValue: "You lose"})
obj.DefineVariable(eek.Var{Name: "VarYourLotteryCode", Type: "int"})
obj.DefineVariable(eek.Var{Name: "VarRepeatUntil", Type: "int", DefaultValue: 5})

obj.PrepareEvaluation(`
    generateRandomNumber := func() int {
        return gubrak.RandomInt(0, 10)
    }

    i := 0
    for i < VarRepeatUntil {
        if generateRandomNumber() == VarYourLotteryCode {
            return fmt.Sprintf("%s after %d tried", VarMessageWin, i + 1)
        }

        i++
    }
    
    return VarMessageLose
`)

err := obj.Build()
if err != nil {
    log.Fatal(err)
}

output, _ = obj.Evaluate(eek.ExecVar{
    "VarYourLotteryCode": 3,
    "VarRepeatUntil":     10,
})
fmt.Println("output:", output)
```

#### Arithmethic expression Example

```go
obj := New("aritmethic expressions")
obj.DefineVariable(eek.Var{Name: "VarN", Type: "int"})
obj.DefineFunction(eek.Func{
    Name: "IF",
    BodyFunction: `
        func(cond bool, ok, nok string) string {
            if cond {
                return ok
            } else {
                return nok
            }
        }
    `,
})
obj.DefineFunction(eek.Func{
    Name: "OR",
    BodyFunction: `
        func(cond1, cond2 bool) bool {
            return cond1 || cond2
        }
    `,
})
obj.DefineFunction(eek.Func{
    Name:         "NOT",
    BodyFunction: `func(cond bool) bool { return !cond }`,
})
obj.PrepareEvaluation(`
    result := IF(VarN>20,IF(OR(VarN>40,N==40),IF(VarN>60,IF(NOT(VarN>80),"good",IF(VarN==90,"perfect","terrific")),"ok"),"ok, but still bad"),"bad")
    
    return result
`)

err := obj.Build()
if err != nil {
    log.Fatal(err)
}

output, _ := obj.Evaluate(eek.ExecVar{"VarN": 76})
fmt.Println(output)
```

More example available on the `*_test.go` file.

## Documentation

[Godoc documentation](http://godoc.org/github.com/novalagung/go-eek)

## Author

Noval Agung Prayog

## License

MIT License
