// To use staticlint, build the tool using 'go build' within the cmd/staticlint directory. Once built, the tool
// can be run from the command line, with the target Go packages as arguments. For example:
//
//	./staticlint ./...
package main

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"strings"
)

// main is the entry point of the staticlint tool. It initializes and executes the multichecker
// with a predefined set of analyzers.
func main() { multiCheck() }

// multiCheck configures and executes the multichecker with a selection of analyzers.
// It includes standard analyzers like printf, shadow, and structtag + analyzers from
// staticcheck, simple, stylecheck, and quickfix packages. A custom analyzer ErrUsingOSExit is also added
// to the mix to prevent the use of os.Exit in the main function of the main package.
func multiCheck() {
	var checks []*analysis.Analyzer
	checks = addChecks(checks)

	checks = append(checks, printf.Analyzer, shadow.Analyzer, structtag.Analyzer, ErrUsingOSExit)

	multichecker.Main(
		checks...,
	)
}

// addChecks iterates over analyzers from various packages like staticcheck, stylecheck, simple, and quickfix,
// selectively adding them based on specific criteria such as their prefix or identifier.
func addChecks(checks []*analysis.Analyzer) []*analysis.Analyzer {
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			checks = append(checks, v.Analyzer)
		}
	}

	st := map[string]bool{
		"ST1005": true,
	}
	for _, v := range stylecheck.Analyzers {
		if st[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}

	qf := map[string]bool{
		"QF1003": true,
	}
	for _, v := range quickfix.Analyzers {
		if qf[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}

	s1 := map[string]bool{
		"S1006": true,
		"S1018": true,
	}
	for _, v := range simple.Analyzers {
		if s1[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}
	return checks
}

// ErrUsingOSExit is an analyzer that checks for occurrences of os.Exit within the main function of the main package.
var ErrUsingOSExit = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "check for os.Exit code in main",
	Run:  run,
}

// run is the execution function for the ErrUsingOSExit analyzer. It inspects Go files within the main package,
// looking for the use of os.Exit calls within the main function. After it reports them.
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if pass.Pkg.Name() == "main" {
			ast.Inspect(
				file, func(n ast.Node) bool {
					funcDecl, ok := n.(*ast.FuncDecl)
					if !ok {
						return true
					}
					if funcDecl.Name.Name != "main" {
						return true
					}

					ast.Inspect(
						funcDecl.Body, func(n ast.Node) bool {
							fn, ok := n.(*ast.CallExpr)
							if !ok {
								return true
							}

							if fun, ok := fn.Fun.(*ast.SelectorExpr); ok {
								if calledFunc := fmt.Sprintf(
									"%s.%s",
									fun.X,
									fun.Sel,
								); calledFunc == "os.Exit" {
									pass.Reportf(fun.Pos(), "os.Exit is used in main file")
								}
							}
							return true
						},
					)
					return true
				},
			)
		}
	}

	return nil, nil
}
