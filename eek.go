package eek

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

type eekType int

const (
	eekTypeSimple eekType = iota
	eekTypeComplex
)

// Eek is main type used on the evaluation
type Eek struct {
	name              string
	variables         map[string]Var
	packages          map[string]bool
	evaluationType    eekType
	operation         string
	baseBuildPath     string
	buildPath         string
	buildFilePath     string
	isBuildOnTempPath bool
}

// Var is reflect to a single typed variable with/without a default value
type Var struct {
	Name         string
	Type         string
	DefaultValue interface{}
}

// ExecVar is used on defining value in the evaluation
type ExecVar map[string]interface{}

// New used to create eek object. This function accept an optional variable that will be used as the evaluation name
func New(args ...string) *Eek {
	eek := new(Eek)

	if len(args) > 0 {
		eek.SetName(args[0])
	}

	eek.variables = make(map[string]Var)
	eek.packages = make(map[string]bool)
	eek.evaluationType = eekTypeSimple

	eek.setDefaultBaseBuildPath()

	return eek
}

func (e *Eek) setDefaultBaseBuildPath() {
	basePath := ""
	switch runtime.GOOS {
	case "darwin", "freebsd", "linux":
		basePath = os.Getenv("TMPDIR")
	case "windows":
		basePath = os.Getenv("TEMP")
	default:
	}

	if basePath == "" {
		basePath = "./"
	} else {
		isBuildOnTempPath = true // for later on
	}

	defaultBaseBuildPath := filepath.Join(basePath, "go-eek-plugins")
	e.SetBaseBuildPath(defaultBaseBuildPath)
}

// SetName set the evaluation name
func (e *Eek) SetName(name string) {
	e.name = name
}

// SetBaseBuildPath set the base build path. Every so file generated from build will be stored into <baseBuildPath>/<name>/<name>.so
func (e *Eek) SetBaseBuildPath(baseBuildPath string) {
	e.baseBuildPath = baseBuildPath
}

// ImportPackage specify which packages will be imported
func (e *Eek) ImportPackage(dependency ...string) {
	for _, each := range dependency {
		e.packages[each] = true
	}
}

// DefineVariable used to define variables that will be used in the operation
func (e *Eek) DefineVariable(variable Var) {
	e.variables[variable.Name] = variable
}

// PrepareEvalutation prepare the layout of evaluation string
func (e *Eek) PrepareEvalutation(operation string) {
	e.evaluationType = eekTypeSimple
	e.operation = strings.TrimSpace(operation)
}

// Build build the evaluation
func (e *Eek) Build() error {
	if e.name == "" {
		return fmt.Errorf("name is mandatory")
	} else if e.evaluationType != eekTypeSimple && e.evaluationType != eekTypeComplex {
		return fmt.Errorf("evaluationType is invalid")
	} else if e.operation == "" {
		return fmt.Errorf("evaluation cannot be empty")
	}

	switch e.evaluationType {
	case eekTypeSimple:
		return e.buildSimpleEvaluation()
	case eekTypeComplex:
		return e.buildComplexEvaluation()
	default:
		return nil
	}
}

func (e *Eek) buildSimpleEvaluation() error {
	layout := strings.TrimSpace(`
		package main

		$packages

		$variables

		func Evaluate() interface{} {
			$operation
		}
	`)

	packageLayout := ""
	for each := range e.packages {
		if each == "" {
			continue
		}

		packageLayout = fmt.Sprintf("%s\n\"%s\"", packageLayout, each)
	}
	packageLayout = fmt.Sprintf(strings.TrimSpace(`import (%s)`), strings.TrimSpace(packageLayout))
	layout = strings.Replace(layout, "$packages", packageLayout, 1)

	variableLayout := ""
	for _, each := range e.variables {
		if each.Name == "" {
			continue
		}

		if prefix := strings.ToUpper(string(each.Name[0])); prefix != string(each.Name[0]) {
			return fmt.Errorf("defined variable must be exported. %s must be %s%s", each.Name, prefix, each.Name[1:])
		}

		if each.DefaultValue == nil {
			variableLayout = fmt.Sprintf("%s\n%s %s", variableLayout, each.Name, each.Type)
		} else {
			switch each.DefaultValue.(type) {
			case string:
				variableLayout = fmt.Sprintf("%s\n%s %s = \"%v\"", variableLayout, each.Name, each.Type, each.DefaultValue)
			default:
				variableLayout = fmt.Sprintf("%s\n%s %s = %v", variableLayout, each.Name, each.Type, each.DefaultValue)
			}
		}
	}
	variableLayout = fmt.Sprintf(strings.TrimSpace(`var (%s)`), strings.TrimSpace(variableLayout))
	layout = strings.Replace(layout, "$variables", variableLayout, 1)

	layout = strings.Replace(layout, "$operation", e.operation, 1)

	err := e.writeToFile(layout)
	if err != nil {
		return err
	}

	err = e.buildPluginFile()
	if err != nil {
		return err
	}

	return nil
}

func (e *Eek) buildComplexEvaluation() error {
	return nil
}

func (e *Eek) writeToFile(code string) error {
	reg, err := regexp.Compile("[^A-Za-z0-9]+")
	if err != nil {
		return err
	}
	name := reg.ReplaceAllString(e.name, "_")

	e.buildPath = filepath.Join(e.baseBuildPath, name)
	err = os.MkdirAll(e.buildPath, os.ModePerm)
	if err != nil {
		return err
	}

	mainFilePath := filepath.Join(e.buildPath, "main.go")
	err = ioutil.WriteFile(mainFilePath, []byte(code), os.ModePerm)
	if err != nil {
		return err
	}

	e.buildFilePath = filepath.Join(e.buildPath, fmt.Sprintf("%s.so", name))

	return nil
}

func (e *Eek) buildPluginFile() error {
	op := fmt.Sprintf("cd %s && go build -buildmode=plugin -o %s", e.buildPath, filepath.Base(e.buildFilePath))

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin", "freebsd", "linux":
		cmd = exec.Command("bash", "-c", op)
	case "windows":
		cmd = exec.Command("cmd", "/C", op)
	default:
		return fmt.Errorf("unsupported operating system")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err.Error(), output)
	}

	return nil
}

// Evaluate execute using particular data
func (e *Eek) Evaluate(data ExecVar) (interface{}, error) {
	p, err := plugin.Open(e.buildFilePath)
	if err != nil {
		return nil, err
	}

	for varName, varValue := range data {
		lookedUpVar, err := p.Lookup(varName)
		if err != nil {
			return nil, err
		}

		(func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					recoveredInString := fmt.Sprintf("%v", recovered)
					recoveredInString = strings.Replace(recoveredInString, "reflect.Set: value of type ", "", -1)
					recoveredInString = strings.Replace(recoveredInString, "is not assignable to type ", "", -1)

					if parts := strings.Split(recoveredInString, " "); len(parts) == 2 {
						err = fmt.Errorf("Error on setting value of variable %s (type %s) with value %v (type %s)", varName, parts[0], varValue, parts[1])
					} else {
						err = fmt.Errorf("%v", recoveredInString)
					}
				}
			}()

			// reflect.Set: value of type int is not assignable to type float64
			reflect.ValueOf(lookedUpVar).Elem().Set(reflect.ValueOf(varValue))
		})()
		if err != nil {
			return nil, err
		}
	}

	lookedUpEvaluate, err := p.Lookup("Evaluate")
	if err != nil {
		return nil, err
	}

	result := lookedUpEvaluate.(func() interface{})()
	return result, nil
}
