package eek

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {
	Convey("Create Eek object with simple evaluation", t, func() {
		obj := NewEek()
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

			Convey("Test exec", func() {
				var output interface{}

				output, err = obj.Evaluate(ExecVar{
					"A": 9,
				})
				So(err, ShouldBeNil)
				fmt.Println("======", output)

				output, err = obj.Evaluate(ExecVar{
					"A": 1,
					"B": 2,
				})
				So(err, ShouldBeNil)
				fmt.Println("======", output)
			})
		})
	})
}
