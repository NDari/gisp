package main

import (
	"go/ast"
	"go/token"
)

func makeIfStmtFunc(node *CallNode) ast.Expr {
	var elseBody ast.Stmt
	if len(node.Args) > 2 {
		elseBody = makeBlockStmt(S(makeReturnStmt(E(EvalExpr(node.Args[2])))))
	} else {
		elseBody = makeBlockStmt(S(makeReturnStmt(E(ast.NewIdent("nil")))))
	}

	cond := EvalExpr(node.Args[0])
	ifBody := makeBlockStmt(S(makeReturnStmt(E(EvalExpr(node.Args[1])))))

	ifStmt := makeIfStmt(cond, ifBody, elseBody)
	fnBody := makeBlockStmt(S(ifStmt))

	returnList := makeFieldList([]*ast.Field{makeField(nil, anyType)})
	fnType := makeFuncType(returnList, nil)

	fn := makeFuncLit(fnType, fnBody)

	return makeFuncCall(fn, EmptyE())
}

func makeLetFun(node *CallNode) ast.Expr {
	bindings := makeBindings(node.Args[0].(*VectorNode), token.DEFINE)

	body := append(bindings, wrapExprsWithStmt(EvalExprs(node.Args[1:]))...)
	body[len(body)-1] = makeReturnStmt(E(body[len(body)-1].(*ast.ExprStmt).X))

	fieldList := makeFieldList([]*ast.Field{makeField(nil, anyType)})
	typ := makeFuncType(fieldList, nil)
	fn := makeFuncLit(typ, makeBlockStmt(body))

	return makeFuncCall(fn, EmptyE())
}

func makeBindings(bindings *VectorNode, assignmentType token.Token) []ast.Stmt {
	assignments := make([]ast.Stmt, len(bindings.Nodes))

	for i, bind := range bindings.Nodes {
		b := bind.(*VectorNode)
		idents := b.Nodes[:len(b.Nodes)-1]

		vars := make([]ast.Expr, len(idents))

		for j, ident := range idents {
			vars[j] = makeIdomaticSelector(ident.(*IdentNode).Ident)
		}

		assignments[i] = makeAssignStmt(vars, E(EvalExpr(b.Nodes[len(b.Nodes)-1])), assignmentType)
	}

	return assignments
}

func mainable(fn *ast.FuncLit) {
	fn.Type.Results = nil

	returnStmt := fn.Body.List[len(fn.Body.List)-1].(*ast.ReturnStmt)
	fn.Body.List[len(fn.Body.List)-1] = makeExprStmt(returnStmt.Results[0])

	// return fn
}

// func makeTypeAssertFromArgList(expr ast.Expr, args []Node) *ast.TypeAssertExpr {

// 	argList

// 	// currently we only support HOF that were written in Gisp
// 	returnList := makeFieldList([]*ast.Field{makeField(nil, anyType)})
// 	fnType := makeFuncType(returnList, argList)

// 	return makeTypeAssertion(expr, fnType)
// }

//////////////////////////////////
// Checked makers from here on! //
//////////////////////////////////

func makeValueSpec(names []*ast.Ident, values []ast.Expr, typ ast.Expr) *ast.ValueSpec {
	return &ast.ValueSpec{
		Names:  names,
		Values: values,
		Type:   typ,
	}
}

func makeFunDeclFromFuncLit(name *ast.Ident, f *ast.FuncLit) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: name,
		Type: f.Type,
		Body: f.Body,
	}
}

func makeGeneralDecl(typ token.Token, specs []ast.Spec) *ast.GenDecl {
	return &ast.GenDecl{
		Tok:   typ,
		Specs: specs,
	}
}

func makeDeclStmt(decl ast.Decl) *ast.DeclStmt {
	return &ast.DeclStmt{
		Decl: decl,
	}
}

func makeTypeAssertion(expr, typ ast.Expr) *ast.TypeAssertExpr {
	return &ast.TypeAssertExpr{
		X:    expr,
		Type: typ,
	}
}

func makeEllipsis(typ ast.Expr) *ast.Ellipsis {
	return &ast.Ellipsis{
		Elt: typ,
	}
}
