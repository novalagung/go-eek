package eek

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {
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
