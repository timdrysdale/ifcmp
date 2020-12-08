package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
)

/*
type Thing interface {
	ast.Expr
	ast.FieldList
	String()
}

//each method is essentially a func type...
type FuncType struct {
	Params  *FieldList // (incoming) parameters; non-nil
	Results *FieldList // (outgoing) results; or nil
}

type FieldList struct {
	ast.FieldList
}


type FieldList struct {
   216  	Opening token.Pos // position of opening parenthesis/brace, if any
   217  	List    []*Field  // field list; or nil
   218  	Closing token.Pos // position of closing parenthesis/brace, if any
   219  }


type Field struct {
   193  	Doc     *CommentGroup // associated documentation; or nil
   194  	Names   []*Ident      // field/method/parameter names; or nil
   195  	Type    Expr          // field/method/parameter type
   196  	Tag     *BasicLit     // field tag; or nil
   197  	Comment *CommentGroup // line comments; or nil
	198  }




type (

	// A ParenExpr node represents a parenthesized expression.
	ParenExpr struct {
		X Thing // parenthesized expression
	}

	// A SelectorExpr node represents an expression followed by a selector.
	SelectorExpr struct {
		X   Thing  // expression
		Sel string // field selector
	}

	// An IndexExpr node represents an expression followed by an index.
	IndexExpr struct {
		X     Thing // expression
		Index Thing // index expression
	}

	// A SliceExpr node represents an expression followed by slice indices.
	SliceExpr struct {
		X      Thing // expression
		Low    Thing // begin of slice range; or nil
		High   Thing // end of slice range; or nil
		Max    Thing // maximum capacity of slice; or nil
		Slice3 bool  // true if 3-index slice (2 colons present)
	}

	// A TypeAssertExpr node represents an expression followed by a
	// type assertion.
	//
	TypeAssertExpr struct {
		X    Thing // expression
		Type Thing // asserted type; nil means type switch X.(type)
	}

	// A CallExpr node represents an expression followed by an argument list.
	CallExpr struct {
		Fun  Thing   // function expression
		Args []Thing // function arguments; or nil
	}

	// A StarExpr node represents an expression of the form "*" Expression.
	// Semantically it could be a unary "*" expression, or a pointer type.
	//
	StarExpr struct {
		X Thing // operand
	}

	// A UnaryExpr node represents a unary expression.
	// Unary "*" expressions are represented via StarExpr nodes.
	//
	UnaryExpr struct {
		X Thing // operand
	}

	// A BinaryExpr node represents a binary expression.
	BinaryExpr struct {
		X Thing // left operand
		Y Thing // right operand
	}

	// A KeyValueExpr node represents (key : value) pairs
	// in composite literals.
	//
	KeyValueExpr struct {
		Key   Thing
		Value Thing
	}
	// An ArrayType node represents an array or slice type.
	ArrayType struct {
		Len Thing // Ellipsis node for [...]T array types, nil for slice types
		Elt Thing // element type
	}

	// A StructType node represents a struct type.
	StructType struct {
		Fields     *FieldList // list of field declarations
		Incomplete bool       // true if (source) fields are missing in the Fields list
	}

	// Pointer types are represented via StarExpr nodes.

	// A FuncType node represents a function type.
	FuncType struct {
		Func    token.Pos  // position of "func" keyword (token.NoPos if there is no "func")
		Params  *FieldList // (incoming) parameters; non-nil
		Results *FieldList // (outgoing) results; or nil
	}

	// An InterfaceType node represents an interface type.
	InterfaceType struct {
		Interface  token.Pos  // position of "interface" keyword
		Methods    *FieldList // list of methods
		Incomplete bool       // true if (source) methods are missing in the Methods list
	}

	// A MapType node represents a map type.
	MapType struct {
		Map   token.Pos // position of "map" keyword
		Key   Thing
		Value Thing
	}

	// A ChanType node represents a channel type.
	ChanType struct {
		Begin token.Pos // position of "chan" keyword or "<-" (whichever comes first)
		Arrow token.Pos // position of "<-" (token.NoPos if there is no "<-")
		Dir   ChanDir   // channel direction
		Value Thing     // value type
	}
)*/

type Param struct {
	Names []string
	Type  string
}

type Method struct {
	Name    string
	Params  []Param
	Results []string
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

func main() {

	actualMethods := make(map[string]Method)

	things := make(map[string]int)

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
		fmt.Printf("interface %s not found in %s", interfaceName, interfaceSource)
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
		methodName := m.Names[0].Name
		fmt.Printf("%T: %s", m.Type, methodName)

		switch m.Type.(type) {
		case *ast.FuncType:
			ft := m.Type.(*ast.FuncType)
			lp := 0
			lr := 0

			params := []Param{}

			if ft.Params != nil {
				if ft.Params.List != nil {
					lp = len(ft.Params.List)

					for _, item := range ft.Params.List {

						// overall stats on type
						v := reflect.ValueOf(item.Type)
						typeStr := v.String()
						if val, ok := things[typeStr]; !ok {
							things[typeStr] = 1
						} else {
							things[typeStr] = val + 1
						}
						//
						params = append(params, Param{
							Names: Names(item.Names),
							Type:  TypeString(item.Type),
						})
						fmt.Printf("%T", item.Type)
					}
				}
			}
			if ft.Results != nil {
				if ft.Results.List != nil {
					lr = len(ft.Results.List)
				}
			}
			fmt.Printf("%d/%d", lp, lr)
			actualMethods[methodName] = Method{
				Name:    methodName,
				Params:  params,
				Results: []string{},
			}
		}

		fmt.Printf("\n")

		//print params
		//m.Params

		//print returns
	}

	// params and results - check length

	/* for _, decl := range actual.Decl {
		fmt.Println(decl)
	}

	for k, v := range f.Decls {
		fmt.Printf("%v %v\n", k, v)
	}
	*/

	//ast.Print(fset, f)

	//for k, v := range things {
	//	fmt.Printf("%v:%d\n", k, v)
	//}

	//fmt.Println(actualMethods)

	for _, v := range actualMethods {
		fmt.Println(v.String())
	}

	os.Exit(0)

}

func (m *Method) String() string {

	str := m.Name + "("

	params := []string{}

	for _, p := range m.Params {
		params = append(params, strings.Join(p.Names, ", ")+" "+p.Type)
	}

	str = str + strings.Join(params, ", ") + ")"

	return str

}
