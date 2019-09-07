package eek

import (
	"crypto/md5"
	"encoding/hex"
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

var (
	regexFileName = regexp.MustCompile("[^A-Za-z0-9]+")
)

// Eek is main type used on the evaluation
type Eek struct {
	name              string
	functions         []Func
	variables         []Var
	packages          []string
	evaluationType    eekType
	evaluationFormula string
	baseBuildPath     string
	buildPath         string
	buildFilePath     string

	UseCachedBuildForSameFormula bool
}

// Func is reflect to a single typed reusable function
type Func struct {
	Name         string
	BodyFunction string
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

	eek.functions = make([]Func, 0)
	eek.variables = make([]Var, 0)
	eek.packages = make([]string, 0)
	eek.evaluationType = eekTypeSimple

	eek.UseCachedBuildForSameFormula = true

	eek.setDefaultBaseBuildPath()

	return eek
}

func (e *Eek) setDefaultBaseBuildPath() {
	basePath := ""
	tmpFolderName := "go-eek-plugins"

	switch runtime.GOOS {
	case "darwin", "freebsd", "linux":

		basePath = os.Getenv("TMPDIR")
		if basePath == "" {

			tempFolder := "/tmp"
			if tempBasePath := filepath.Join(tempFolder, tmpFolderName); e.isPathExists(tempBasePath) {
				basePath = tempFolder
			} else if err := os.MkdirAll(tempBasePath, os.ModePerm); err == nil {
				basePath = tempFolder
			}
		}
	case "windows":
		basePath = os.Getenv("TEMP")
	default:
	}

	if basePath == "" {
		basePath = "./"
	}

	defaultBaseBuildPath := filepath.Join(basePath, tmpFolderName)
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
func (e *Eek) ImportPackage(dependencies ...string) {
	e.packages = append(e.packages, dependencies...)
}

// DefineVariable used to define variables that will be used in the evaluation formula
func (e *Eek) DefineVariable(variable Var) {
	e.variables = append(e.variables, variable)
}

// DefineFunction used to define reusable functions that will be used in the evaluation formula
func (e *Eek) DefineFunction(fun Func) {
	e.functions = append(e.functions, fun)
}

// PrepareEvaluation prepare the layout of evaluation string
func (e *Eek) PrepareEvaluation(evaluationFormula string) {
	e.evaluationType = eekTypeSimple
	e.evaluationFormula = strings.TrimSpace(evaluationFormula)
}

// Build build the evaluation
func (e *Eek) Build() error {
	if e.name == "" {
		return fmt.Errorf("name is mandatory")
	} else if e.evaluationType != eekTypeSimple && e.evaluationType != eekTypeComplex {
		return fmt.Errorf("evaluationType is invalid")
	} else if e.evaluationFormula == "" {
		return fmt.Errorf("evaluation formula cannot be empty")
	}

	var code string
	var err error

	switch e.evaluationType {
	case eekTypeSimple:
		code, err = e.buildSimpleEvaluation()
	case eekTypeComplex:
		code, err = e.buildComplexEvaluation()
	}
	if err != nil {
		return err
	}

	// write code into temporary file, then build the code as go plugin file
	if err := e.writeToFileThenBuild(code); err != nil {
		return err
	}

	return nil
}

func (e *Eek) buildSimpleEvaluation() (string, error) {
	// code base code
	code := strings.TrimSpace(`
		package main

		$packages

		$functions

		$variables

		func Evaluate() interface{} {
			$evaluationFormula
		}
	`)

	// inject packages
	packageLayout := ""
	for _, each := range e.packages {
		if each == "" {
			continue
		}

		packageLayout = fmt.Sprintf("%s\n\"%s\"", packageLayout, each)
	}
	packageLayout = fmt.Sprintf(strings.TrimSpace(`import (%s)`), strings.TrimSpace(packageLayout))
	code = strings.Replace(code, "$packages", packageLayout, 1)

	// inject functions
	functionLayout := ""
	for _, each := range e.functions {
		bodyFunc := strings.TrimSpace(each.BodyFunction)
		if each.Name == "" || bodyFunc == "" {
			continue
		}

		functionLayout = fmt.Sprintf("%s\nvar %s = %s", functionLayout, each.Name, bodyFunc)
	}
	code = strings.Replace(code, "$functions", strings.TrimSpace(functionLayout), 1)

	// inject variables
	variableLayout := ""
	for _, each := range e.variables {
		if each.Name == "" || each.Type == "" {
			continue
		}

		if prefix := strings.ToUpper(string(each.Name[0])); prefix != string(each.Name[0]) {
			return "", fmt.Errorf("defined variable must be exported. %s must be %s%s", each.Name, prefix, each.Name[1:])
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
	code = strings.Replace(code, "$variables", variableLayout, 1)

	// inject evaluationFormula
	code = strings.Replace(code, "$evaluationFormula", e.evaluationFormula, 1)

	return code, nil
}

func (e *Eek) buildComplexEvaluation() (string, error) {
	return "", fmt.Errorf("currently complex evaluation is still not supported")
}

func (e *Eek) writeToFileThenBuild(code string) error {
	name := regexFileName.ReplaceAllString(e.name, "_")
	e.buildPath = filepath.Join(e.baseBuildPath, name)
	e.buildFilePath = filepath.Join(e.buildPath, fmt.Sprintf("%s_%s.so", name, e.md5(code)))

	if e.UseCachedBuildForSameFormula {
		if e.isPathExists(e.buildFilePath) {
			return nil
		}
	}

	os.RemoveAll(e.buildPath)
	err := os.MkdirAll(e.buildPath, os.ModePerm)
	if err != nil {
		return err
	}

	mainFilePath := filepath.Join(e.buildPath, "main.go")
	err = ioutil.WriteFile(mainFilePath, []byte(code), os.ModePerm)
	if err != nil {
		return err
	}

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
	if !e.isPathExists(e.buildFilePath) {
		return nil, fmt.Errorf("build file is not found. please try to rebuild the formula")
	}

	// open the build file path
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
			// recover panic error from reflect evaluationFormula
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

			// line underneath has a protential to generate a panic error
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

func (*Eek) md5(str string) string {
	hasher := md5.New()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (e *Eek) isPathExists(str string) bool {
	if _, err := os.Stat(str); err == nil {
		return true
	}

	return false
}
