package eek

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEval(t *testing.T) {
	Convey("Create Eek object with simple evaluation", t, func() {
		obj := New()
		obj.SetName("simple operation")

		obj.DefineVariable(Var{Name: "A", Type: "int"})
		obj.DefineVariable(Var{Name: "B", Type: "float64", DefaultValue: 10.5})

		obj.PrepareEvalutation(`
			ACasted := float64(A)
			C := ACasted + B
			return C
		`)

		Convey("Build operation", func() {
			err := obj.Build()
			So(err, ShouldBeNil)

			Convey("Test exec 1", func() {
				var output interface{}

				output, err = obj.Evaluate(ExecVar{
					"A": 9,
				})
				So(err, ShouldBeNil)
				So(output.(float64), ShouldEqual, 19.5)
			})

			Convey("Test exec 2", func() {
				var output interface{}

				output, err = obj.Evaluate(ExecVar{
					"A": 1,
					"B": 2.1,
				})
				So(err, ShouldBeNil)
				So(output.(float64), ShouldEqual, 3.1)
			})

			Convey("Test exec error", func() {
				_, err = obj.Evaluate(ExecVar{
					"B": 2,
				})
				So(err, ShouldBeError)
				So(err.Error(), ShouldEqual, "Error on setting value of variable B (type int) with value 2 (type float64)")
			})
		})
	})
}

func TestComplexEval(t *testing.T) {
	Convey("Create Eek object with simple evaluation", t, func() {
		obj := New()
		obj.SetName("evaluation with 3rd party library")

		obj.ImportPackage("fmt")
		obj.ImportPackage("github.com/novalagung/gubrak")

		obj.DefineVariable(Var{Name: "MessageWin", Type: "string", DefaultValue: "Congrats! You win the lottery!"})
		obj.DefineVariable(Var{Name: "MessageLose", Type: "string", DefaultValue: "You lose"})
		obj.DefineVariable(Var{Name: "YourLotteryCode", Type: "int"})
		obj.DefineVariable(Var{Name: "RepeatUntil", Type: "int", DefaultValue: 5})

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

		Convey("Build operation", func() {
			err := obj.Build()
			So(err, ShouldBeNil)

			Convey("Test exec 1", func() {
				output, err := obj.Evaluate(ExecVar{
					"YourLotteryCode": 5,
				})
				So(err, ShouldBeNil)
				_ = output
			})

			Convey("Test exec 2", func() {
				output, err := obj.Evaluate(ExecVar{
					"YourLotteryCode": 3,
					"RepeatUntil":     10,
				})
				So(err, ShouldBeNil)
				_ = output
			})
		})
	})
}

func TestMathematicExpression(t *testing.T) {
	Convey("Create Eek object with simple evaluation", t, func() {
		obj := New("aritmethic expressions")
		obj.DefineVariable(Var{Name: "N", Type: "int", DefaultValue: 34})
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

		Convey("Build operation", func() {
			err := obj.Build()
			So(err, ShouldBeNil)

			Convey("Test exec 1", func() {
				output, err := obj.Evaluate(ExecVar{"N": 76})
				So(err, ShouldBeNil)
				So(output, ShouldEqual, "good")
			})
		})
	})
}
