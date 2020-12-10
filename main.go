package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"

	mdast "github.com/gomarkdown/markdown/ast"
	mdparser "github.com/gomarkdown/markdown/parser"
)

type Param struct {
	Names []string
	Type  string
}

type Method struct {
	Name    string
	Params  []Param
	Results []string
	Idx     int
}

func TypeString(expr ast.Expr) string {

	switch expr.(type) {
	case *ast.Ident:
		return expr.(*ast.Ident).Name
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.ArrayType:
		return "[]" + TypeString(expr.(*ast.ArrayType).Elt)
	case *ast.StarExpr:
		return "*" + TypeString(expr.(*ast.StarExpr).X)
	case *ast.SelectorExpr:
		return TypeString(expr.(*ast.SelectorExpr).X) + "." + expr.(*ast.SelectorExpr).Sel.Name
	}

	return ""

}

func Names(idents []*ast.Ident) []string {

	names := []string{}
	for _, f := range idents {
		if f != nil {
			names = append(names, f.Name)
		}
	}
	return names
}

func GetMethodMap(f *ast.File, interfaceName string) (map[string]Method, error) {

	debug := false

	methodMap := make(map[string]Method)

	actual := &ast.Object{}
	found := false
	for _, obj := range f.Scope.Objects {
		if obj.Name == interfaceName && obj.Kind == ast.Typ {
			found = true
			actual = obj
		}
	}

	if !found {
		return methodMap, errors.New("interface not found")
	}

	d := actual.Decl

	if debug {
		fmt.Printf("%s\n", actual.Name)            //GoCloak
		fmt.Printf("%T\n", actual.Decl)            //*ast.TypeSpec
		fmt.Printf("%s\n", d.(*ast.TypeSpec).Name) //GoCloak
		fmt.Printf("%T\n", d.(*ast.TypeSpec).Type) //*ast.InterfaceType
	}

	methods := (d.(*ast.TypeSpec).Type).(*ast.InterfaceType).Methods

	for idx, m := range methods.List {
		// m is *ast.Field
		methodName := m.Names[0].Name

		switch m.Type.(type) {
		case *ast.FuncType:
			ft := m.Type.(*ast.FuncType)
			params := []Param{}
			results := []string{}

			if ft.Params != nil {
				if ft.Params.List != nil {
					for _, item := range ft.Params.List {
						params = append(params, Param{
							Names: Names(item.Names),
							Type:  TypeString(item.Type),
						})
					}
				}
			}

			if ft.Results != nil {
				if ft.Results.List != nil {
					for _, item := range ft.Results.List {
						results = append(results, TypeString(item.Type))
					}
				}
			}

			methodMap[methodName] = Method{
				Name:    methodName,
				Params:  params,
				Results: results,
				Idx:     idx,
			}
		}
	}

	return methodMap, nil
}

func main() {

	debug := false

	if len(os.Args) != 4 {
		fmt.Println("Usage ifcmp <README.md> <interface.go> <interface>")
		os.Exit(1)
	}

	doc := os.Args[1]
	interfaceSource := os.Args[2]
	interfaceName := os.Args[3]

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, interfaceSource, nil, 0)

	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}

	actualMethods, err := GetMethodMap(f, interfaceName)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	docContent, err := ioutil.ReadFile(doc)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	mdp := mdparser.New()

	docAst := mdp.Parse(docContent)
	var buf bytes.Buffer

	buf.WriteString("package main\n")

	mdast.WalkFunc(docAst, func(node mdast.Node, entering bool) mdast.WalkStatus {

		if _, ok := node.(*mdast.CodeBlock); ok {

			cb := node.(*mdast.CodeBlock)
			if bytes.Compare(cb.Info, []byte("go")) == 0 {
				searchStr := fmt.Sprintf("type %s interface", interfaceName)
				//searchStr := interfaceName
				idx := bytes.Index(cb.Literal, []byte(searchStr))
				if idx > -1 {
					_, err := buf.Write(cb.Literal)
					if err != nil {
						panic(err)
					}
				}
			}
		}
		return mdast.GoToNext
	})

	fsetdoc := token.NewFileSet()

	fdoc, err := parser.ParseFile(fsetdoc, interfaceName, &buf, 0)

	if err != nil {
		fmt.Printf("ParserError: %s\n", err.Error())
		os.Exit(1)
	}

	docMethods, err := GetMethodMap(fdoc, interfaceName)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	foundErrors := false

	for k, v := range actualMethods {
		if vd, ok := docMethods[k]; ok {

			if !(reflect.DeepEqual(v.Params, vd.Params) && reflect.DeepEqual(v.Results, vd.Results)) {
				foundErrors = true
				fmt.Printf("Actual: %s\nReadme: %s\n\n", v.String(), vd.String())
			}
		} else {
			foundErrors = true
			fmt.Printf("Actual: %s\nReadme: %s\n\n", v.String(), "")
		}
	}

	// check for extra methods in doc which are not in the interface
	for k, vd := range docMethods {
		if _, ok := actualMethods[k]; !ok {
			foundErrors = true
			fmt.Printf("Actual: %s\nReadme: %s\n\n", "", vd.String())
		}
	}

	if debug {
		sortedMethods := SortMethods(actualMethods)

		for _, v := range sortedMethods {
			fmt.Println(v.String())
		}

		fmt.Println(&buf)
	}

	if foundErrors {
		os.Exit(1)
	}
	os.Exit(0)

}

// String prettyprints the method
func (m *Method) String() string {

	str := m.Name + "("

	params := []string{}

	for _, p := range m.Params {
		params = append(params, strings.Join(p.Names, ", ")+" "+p.Type)
	}

	str = str + strings.Join(params, ", ") + ")"

	if len(m.Results) > 0 {
		str = str + " "
		if len(m.Results) > 1 {
			str = str + "(" + strings.Join(m.Results, ", ") + ")"
		} else {
			str = str + m.Results[0]
		}
	}
	return str

}

type MethodSlice []*Method

//https://stackoverflow.com/questions/19946992/sorting-a-map-of-structs-golang
func (m MethodSlice) Len() int {
	return len(m)
}

func (m MethodSlice) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m MethodSlice) Less(i, j int) bool {
	return m[i].Idx < m[j].Idx
}

func SortMethods(methodMap map[string]Method) MethodSlice {

	methods := make(MethodSlice, 0, len(methodMap))

	for _, v := range methodMap {
		method := v
		methods = append(methods, &method)
	}

	sort.Sort(methods)

	return methods

}
