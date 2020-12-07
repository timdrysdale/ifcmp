# ifcmp
Compare two golang interface definitions, e.g. check version in README matches actual package.

There may be differences in the way the an interface's list of methods is presented in the README and in the actual package, e.g. adding section headings, order of presention, commenting etc, which make a direct generation of this section of the README.md undesirable. This tool is intended to compare the two presentations and check for consistency.

## Method

Use Go's parser to create an AST and compare the AST for a given interface name. The ```README.md``` will need cleaning up before it can be presented to the parser. 

### Cleaning up the README.md for parsing

Proposed approach:

  - create an array containing verbatim code blocks
  - check each code block for the existence of an interface declaration for the intended interface
  - pass that code block for parsing
  

### What's in the AST?

Using the ```ast.Print()``` method, you can see the GoCloak interface as an ```ast.GenDecl``` (line 39), named on line 47, and identified as an ```*ast.InterfaceType```  on line 55. 

```
<snip>
    39  .  .  1: *ast.GenDecl {
    40  .  .  .  TokPos: ./testdata/gocloak.go:11:1
    41  .  .  .  Tok: type
    42  .  .  .  Lparen: -
    43  .  .  .  Specs: []ast.Spec (len = 1) {
    44  .  .  .  .  0: *ast.TypeSpec {
    45  .  .  .  .  .  Name: *ast.Ident {
    46  .  .  .  .  .  .  NamePos: ./testdata/gocloak.go:11:6
    47  .  .  .  .  .  .  Name: "GoCloak"
    48  .  .  .  .  .  .  Obj: *ast.Object {
    49  .  .  .  .  .  .  .  Kind: type
    50  .  .  .  .  .  .  .  Name: "GoCloak"
    51  .  .  .  .  .  .  .  Decl: *(obj @ 44)
    52  .  .  .  .  .  .  }
    53  .  .  .  .  .  }
    54  .  .  .  .  .  Assign: -
    55  .  .  .  .  .  Type: *ast.InterfaceType {
    56  .  .  .  .  .  .  Interface: ./testdata/gocloak.go:11:14
    57  .  .  .  .  .  .  Methods: *ast.FieldList {
    58  .  .  .  .  .  .  .  Opening: ./testdata/gocloak.go:11:24
    59  .  .  .  .  .  .  .  List: []*ast.Field (len = 189) {
    60  .  .  .  .  .  .  .  .  0: *ast.Field {
    61  .  .  .  .  .  .  .  .  .  Names: []*ast.Ident (len = 1) {
    62  .  .  .  .  .  .  .  .  .  .  0: *ast.Ident {
    63  .  .  .  .  .  .  .  .  .  .  .  NamePos: ./testdata/gocloak.go:13:2
    64  .  .  .  .  .  .  .  .  .  .  .  Name: "RestyClient"
    65  .  .  .  .  .  .  .  .  .  .  .  Obj: *ast.Object {
    66  .  .  .  .  .  .  .  .  .  .  .  .  Kind: func
    67  .  .  .  .  .  .  .  .  .  .  .  .  Name: "RestyClient"
    68  .  .  .  .  .  .  .  .  .  .  .  .  Decl: *(obj @ 60)
    69  .  .  .  .  .  .  .  .  .  .  .  }
    70  .  .  .  .  .  .  .  .  .  .  }
    71  .  .  .  .  .  .  .  .  .  }
 <snip>
```
Looking in ```go/ast/ast.go```, we find:

```
type File struct {
  	Doc        *CommentGroup   // associated documentation; or nil
  	Package    token.Pos       // position of "package" keyword
  	Name       *Ident          // package name
  	Decls      []Decl          // top-level declarations; or nil
  	Scope      *Scope          // package scope (this file only)
  	Imports    []*ImportSpec   // imports in this file
  	Unresolved []*Ident        // unresolved identifiers in this file
  	Comments   []*CommentGroup // list of all comments in the source file
}
```

The Name field is the package name, not the interface name, so we cannot use that.

in ```go/ast/scope.go``` we see
```
type Scope struct {
    20  	Outer   *Scope
    21  	Objects map[string]*Object
    22  }
```

We're looking here for a scope of Type, with the interface name.

The Objects map uses the object name as the key (which will be the interface name when the object is the interface)

```
type Object struct {
   	Kind ObjKind
   	Name string      // declared name
   	Decl interface{} // corresponding Field, XxxSpec, FuncDecl, LabeledStmt, AssignStmt, Scope; or nil
   	Data interface{} // object-specific data; or nil
   	Type interface{} // placeholder for type information; may be nil
   }

```

```
// ObjKind describes what an object represents.
  type ObjKind int
  
  // The list of possible Object kinds.
  const (
  	Bad ObjKind = iota // for error handling
  	Pkg                // package
  	Con                // constant
  	Typ                // type
  	Var                // variable
  	Fun                // function or method
  	Lbl                // label
  )
```   
   
The Decl field for an interface is of type ```*ast.TypeSpec```

