package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

//go:embed stub.go.tmpl
var stubTemplate string

func main() {
	var outputFile string
	var disableFormatting bool
	flag.BoolVar(&disableFormatting, "no-fmt", false, "disable formatting of the output")

	flag.StringVar(&outputFile, "o", "", "output file name")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr,
			"Usage: %s [-no-fmt] -o <output.go> <input_directory> <interface>\n",
			os.Args[0])

		os.Exit(1)
	}

	inputDir := flag.Arg(0)
	interfaceName := flag.Arg(1)

	interfaceMethods, packageName, err := findInterface(inputDir, interfaceName)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding interface: %v\n", err)
		os.Exit(1)
	}

	if len(interfaceMethods) == 0 {
		fmt.Fprintf(os.Stderr, "Interface %s not found\n", interfaceName)
		os.Exit(1)
	}

	stubCode, err := generateStubCode(interfaceName,
		interfaceMethods,
		packageName,
		disableFormatting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating stub: %v\n", err)
		os.Exit(1)
	}

	if outputFile == "" {
		fmt.Println(stubCode)
	} else {
		err := os.WriteFile(outputFile, []byte(stubCode), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Stub generated in %s\n", outputFile)
	}
}

func findInterface(inputDir string, interfaceName string) ([]*ast.Field, string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo,
		Dir: inputDir,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, "", fmt.Errorf("load: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, "", fmt.Errorf("packages contain errors")
	}

	var interfaceMethods []*ast.Field
	var packageName string

	for _, pkg := range pkgs {
		packageName = pkg.Name
		for _, file := range pkg.Syntax {
			ast.Inspect(
				file, func(n ast.Node) bool {
					if ts, ok := n.(*ast.TypeSpec); ok && ts.Name.Name == interfaceName {
						if ift, ok := ts.Type.(*ast.InterfaceType); ok {
							interfaceMethods = ift.Methods.List
						}
					}
					return true
				})
		}
	}
	return interfaceMethods, packageName, nil
}

type methodData struct {
	Name        string
	Params      []string
	ParamNames  []string
	Results     []string
	ResultNames []string
}

func zip(a []string, b []string, fmtStr string) []string {
	if len(a) != len(b) {
		panic("unequal length")
	}
	var zipped []string
	for i := range a {
		zipped = append(zipped, fmt.Sprintf(fmtStr, a[i], b[i]))
	}
	return zipped
}

// joinl joins a list of strings with a separator, with arguments reversed
// compared to strings.Join.
func joinl(sep string, a []string) string {
	return strings.Join(a, sep)
}

func generateStubCode(interfaceName string,
	methods []*ast.Field,
	packageName string,
	disableFormatting bool) (string, error) {
	stubName := "Stub" + interfaceName

	funcMap := template.FuncMap{
		"join":  strings.Join,
		"zip":   zip,
		"joinl": joinl,
	}

	tmpl := template.Must(
		template.New("stub").
			Funcs(funcMap).
			Parse(stubTemplate))

	var methodsData []methodData

	for _, method := range methods {
		if len(method.Names) == 0 {
			continue
		}
		methodName := method.Names[0].Name
		funcType := method.Type.(*ast.FuncType)

		params := getFieldList(funcType.Params)
		paramNames := getFieldNames(funcType.Params)
		results := getFieldList(funcType.Results)
		resultNames := getResultNames(funcType.Results)

		methodsData = append(
			methodsData, methodData{
				Name:        methodName,
				Params:      params,
				ParamNames:  paramNames,
				Results:     results,
				ResultNames: resultNames,
			})
	}

	var buf strings.Builder
	fmt.Println(prettyPrint(methodsData))
	err := tmpl.Execute(
		&buf, struct {
			PackageName   string
			InterfaceName string
			StubName      string
			Methods       []methodData
		}{
			PackageName:   packageName,
			InterfaceName: interfaceName,
			StubName:      stubName,
			Methods:       methodsData,
		})

	if err != nil {
		return "", fmt.Errorf("error generating stub: %v", err)
	}

	if !disableFormatting {
		// Format the generated code
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, "", buf.String(), parser.ParseComments)
		if err != nil {
			return "", fmt.Errorf("error parsing generated code: %v", err)
		}

		var formattedBuf strings.Builder
		err = format.Node(&formattedBuf, fset, node)
		if err != nil {
			return "", fmt.Errorf("error formatting generated code: %v", err)
		}

		return formattedBuf.String(), nil
	} else {
		return buf.String(), nil
	}
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func getFieldList(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}
	var params []string
	for _, field := range fields.List {
		paramType := getTypeString(field.Type)
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				params = append(params, fmt.Sprintf("%s %s", name.Name, paramType))
			}
		} else {
			params = append(params, paramType)
		}
	}
	return params
}

func getFieldNames(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}
	var names []string
	for _, field := range fields.List {
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				names = append(names, name.Name)
			}
		} else {
			names = append(names, "_")
		}
	}
	return names
}

func getResultNames(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}
	var names []string
	for i, field := range fields.List {
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				names = append(names, name.Name)
			}
		} else {
			names = append(names, fmt.Sprintf("R%d", i))
		}
	}
	return names
}

func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", getTypeString(t.X), t.Sel.Name)
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.ArrayType:
		return "[]" + getTypeString(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", getTypeString(t.Key), getTypeString(t.Value))
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func(" + strings.Join(getFieldList(t.Params), ", "+
			"") + ") " + strings.Join(getFieldList(t.Results), ", ")
	default:
		return fmt.Sprintf("%T", expr)
	}
}
