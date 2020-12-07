package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

func main() {

	if len(os.Args) != 4 {
		fmt.Println("Usage ifcmp <README.md> <interface.go> <interface>")
		os.Exit(1)
	}

	//doc := os.Args[1]
	interfaceSource := os.Args[2]
	interfaceName := os.Args[3]

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, interfaceSource, nil, 0)

	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}

	//fmt.Println(f.Name)
	//fmt.Println(f.Scope)
	//fmt.Printf("%+v", f)

	//fmt.Println()

	actual := &ast.Object{}
	found := false
	for _, obj := range f.Scope.Objects {
		if obj.Name == interfaceName && obj.Kind == ast.Typ {
			found = true
			actual = obj
		}
	}

	if !found {
		fmt.Println("interface %s not found in %s", interfaceName, interfaceSource)
		os.Exit(1)
	}

	fmt.Printf("%s\n", actual.Name) //GoCloak

	fmt.Printf("%T\n", actual.Decl) //*ast.TypeSpec

	d := actual.Decl

	fmt.Printf("%s\n", d.(*ast.TypeSpec).Name) //GoCloak

	fmt.Printf("%T\n", d.(*ast.TypeSpec).Type) //*ast.InterfaceType

	methods := (d.(*ast.TypeSpec).Type).(*ast.InterfaceType).Methods

	for _, m := range methods.List {
		// m is *ast.Field
		//
		fmt.Printf("%v\n", m.Names)
	}

	/* for _, decl := range actual.Decl {
		fmt.Println(decl)
	}

	for k, v := range f.Decls {
		fmt.Printf("%v %v\n", k, v)
	}

	ast.Print(fset, f)
	*/
	os.Exit(0)

}
