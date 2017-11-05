package main

import (
	"go/ast"
	"go/token"
)

func getImports(node *CallNode) ast.Decl {
	if len(node.Args) < 2 {
		return nil
	}

	imports := node.Args[1:]
	specs := make([]ast.Spec, len(imports))

	for i, imp := range imports {
		if t := imp.Type(); t == NodeVector {
			specs[i] = makeImportSpecFromVector(imp.(*VectorNode))
		} else if t == NodeString {
			path := makeBasicLit(token.STRING, imp.(*StringNode).Value)
			specs[i] = makeImportSpec(path, nil)
		} else {
			panic("invalid import!")
		}
	}

	decl := makeGeneralDecl(token.IMPORT, specs)
	decl.Lparen = token.Pos(1) // Need this so we can have multiple imports
	return decl
}

func makeImportSpecFromVector(vect *VectorNode) *ast.ImportSpec {
	if len(vect.Nodes) < 3 {
		panic("invalid use of import!")
	}

	if vect.Nodes[0].Type() != NodeString {
		panic("invalid use of import!")
	}

	pathString := vect.Nodes[0].(*StringNode).Value
	path := makeBasicLit(token.STRING, pathString)

	if vect.Nodes[1].Type() != NodeIdent || vect.Nodes[1].(*IdentNode).Ident != ":as" {
		panic("invalid use of import! expecting: \":as\"!!!")
	}
	name := ast.NewIdent(vect.Nodes[2].(*IdentNode).Ident)

	return makeImportSpec(path, name)
}

func makeImportSpec(path *ast.BasicLit, name *ast.Ident) *ast.ImportSpec {
	return &ast.ImportSpec{
		Path: path,
		Name: name,
	}
}
