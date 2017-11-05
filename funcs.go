package main

import (
	"go/ast"
	"go/token"
)

func evalFuncCall(node *CallNode) ast.Expr {
	switch {
	case isUnaryOperator(node):
		return makeUnaryExpr(unaryOperatorMap[node.Callee.(*IdentNode).Ident], EvalExpr(node.Args[0]))

	case isCallableOperator(node):
		return makeNAryCallableExpr(node)

	case isLogicOperator(node):
		return makeNAryLogicExpr(node)

	case isLoop(node):
		return makeLoop(node)

	case isRecur(node):
		return makeRecur(node)

	case isAssert(node):
		return makeAssert(node)

	case isCoreFunc(node):
		return makeCoreCall(node)

	case checkLetArgs(node):
		return makeLetFun(node)

	case checkIfArgs(node):
		return makeIfStmtFunc(node)

	case checkFuncArgs(node):
		// TODO: In case of type annotations change the following
		returnField := []*ast.Field{makeField(nil, anyType)}
		results := makeFieldList(returnField)

		argIdents, ellipsis := getArgIdentsFromVector(node.Args[0].(*VectorNode))
		params := make([]*ast.Field, 0, len(argIdents))

		if len(argIdents) != 0 {
			params = append(params, makeField(argIdents, anyType))
		}

		if ellipsis != nil {
			params = append(params, makeField(I(ellipsis), makeEllipsis(anyType)))
		}

		fnType := makeFuncType(results, makeFieldList(params))
		body := makeFuncBody(EvalExprs(node.Args[1:]))

		return makeFuncLit(fnType, body)

	case checkDefArgs(node):
		panic("you can't have a def within an expression!")

	case checkNSArgs(node):
		panic("you can't define a namespace in an expression!")
	}

	callee := EvalExpr(node.Callee)
	if c, ok := callee.(*ast.Ident); ok {
		callee = makeIdomaticIdent(c.Name)
	}

	args := EvalExprs(node.Args)

	return makeFuncCall(callee, args)
}

func getArgIdentsFromVector(vect *VectorNode) ([]*ast.Ident, *ast.Ident) {
	args := vect.Nodes
	argIdents := make([]*ast.Ident, 0, len(vect.Nodes))

	var ident string
	var ellipsis *ast.Ident

	for i := 0; i < len(args); i++ {
		ident = args[i].(*IdentNode).Ident

		if ident == "&" {
			ellipsis = makeIdomaticIdent(args[i+1].(*IdentNode).Ident)
			break
		}

		argIdents = append(argIdents, makeIdomaticIdent(ident))
	}

	return argIdents, ellipsis
}

func makeFuncBody(exprs []ast.Expr) *ast.BlockStmt {
	wrapped := wrapExprsWithStmt(exprs)
	wrapped[len(wrapped)-1] = makeReturnStmt(E(wrapped[len(wrapped)-1].(*ast.ExprStmt).X))
	return makeBlockStmt(wrapped)
}

func makeFuncLit(typ *ast.FuncType, body *ast.BlockStmt) *ast.FuncLit {
	return &ast.FuncLit{
		Type: typ,
		Body: body,
	}
}

func makeFuncType(results, params *ast.FieldList) *ast.FuncType {
	return &ast.FuncType{
		Results: results,
		Params:  params,
	}
}

func makeFieldList(list []*ast.Field) *ast.FieldList {
	return &ast.FieldList{
		List: list,
	}
}

func makeField(names []*ast.Ident, typ ast.Expr) *ast.Field {
	return &ast.Field{
		Names: names,
		Type:  typ,
	}
}

func makeReturnStmt(exprs []ast.Expr) ast.Stmt {
	return &ast.ReturnStmt{
		Results: exprs,
	}
}

func makeFuncCall(callee ast.Expr, args []ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  callee,
		Args: args,
	}
}

// Fn type checks (let, fn, def, ns, etc.)

func checkIfArgs(node *CallNode) bool {
	if node.Callee.Type() != NodeIdent {
		return false
	}

	if callee := node.Callee.(*IdentNode); callee.Ident != "if" {
		return false
	}

	if len(node.Args) < 2 {
		return false
	}

	return true
}

// Only need this to check if "def" is in
// an expression, which is illegal
func checkDefArgs(node *CallNode) bool {
	if node.Callee.Type() != NodeIdent {
		return false
	}

	if callee := node.Callee.(*IdentNode); callee.Ident != "def" {
		return false
	}

	return true
}

func checkFuncArgs(node *CallNode) bool {
	// Need an identifier for it to be "fn"
	if node.Callee.Type() != NodeIdent {
		return false
	}

	if callee := node.Callee.(*IdentNode); callee.Ident != "fn" {
		return false
	}

	// Need argument list and at least one expression
	if len(node.Args) < 2 {
		return false
	}

	// Parameters should be a vector
	params := node.Args[0]
	if params.Type() != NodeVector {
		return false
	}

	p := params.(*VectorNode)
	for _, param := range p.Nodes {
		// TODO: change this in case of variable unpacking
		if param.Type() != NodeIdent {
			return false
		}
	}

	return true
}

func checkLetArgs(node *CallNode) bool {
	// Need an identifier for it to be "let"
	if node.Callee.Type() != NodeIdent {
		return false
	}

	// Not a "let"
	if callee := node.Callee.(*IdentNode); callee.Ident != "let" {
		return false
	}

	// Need _at least_ the bindings & one expression
	if len(node.Args) < 2 {
		return false
	}

	// Bindings should be a vector
	bindings := node.Args[0]
	if bindings.Type() != NodeVector {
		return false
	}

	// The bindings should be also vectors
	b := bindings.(*VectorNode)
	for _, bind := range b.Nodes {
		if _, ok := bind.(*VectorNode); !ok {
			return false
		}
	}

	// The bound identifiers, should be identifiers
	for _, bind := range b.Nodes {
		bindingVect := bind.(*VectorNode)
		if bindingVect.Nodes[0].Type() != NodeIdent {
			return false
		}
	}

	return true
}

func isLoop(node *CallNode) bool {
	// Need an identifier for it to be "loop"
	if node.Callee.Type() != NodeIdent {
		return false
	}

	// Not a "loop"
	if callee := node.Callee.(*IdentNode); callee.Ident != "loop" {
		return false
	}

	// Bindings should be a vector
	bindings := node.Args[0]
	if bindings.Type() != NodeVector {
		return false
	}

	// The bindings should be also vectors
	b := bindings.(*VectorNode)
	for _, bind := range b.Nodes {
		if _, ok := bind.(*VectorNode); !ok {
			return false
		}
	}

	// The bound identifiers, should be identifiers
	for _, bind := range b.Nodes {
		bindingVect := bind.(*VectorNode)
		if bindingVect.Nodes[0].Type() != NodeIdent {
			return false
		}
	}

	if !searchForRecur(node.Args[1:]) {
		panic("no recur found in loop!")
	}

	return true
}

func isRecur(node *CallNode) bool {
	// Need an identifier for it to be "loop"
	if node.Callee.Type() != NodeIdent {
		return false
	}

	// Not a "loop"
	if callee := node.Callee.(*IdentNode); callee.Ident != "recur" {
		return false
	}

	// Bindings should be a vector
	bindings := node.Args[0]
	if bindings.Type() != NodeVector {
		return false
	}

	// The bindings should be also vectors
	b := bindings.(*VectorNode)
	for _, bind := range b.Nodes {
		if _, ok := bind.(*VectorNode); !ok {
			return false
		}
	}

	// The bound identifiers, should be identifiers
	for _, bind := range b.Nodes {
		bindingVect := bind.(*VectorNode)
		if bindingVect.Nodes[0].Type() != NodeIdent {
			return false
		}
	}

	return true
}

func searchForRecur(nodes []Node) bool {
	for _, node := range nodes {
		if node.Type() == NodeCall {
			n := node.(*CallNode)
			if ident, ok := n.Callee.(*IdentNode); ok && ident.Ident == "recur" {
				return true
			} else if searchForRecur(n.Args) {
				return true
			}
		}
	}

	return false
}

func addNewValuesToBindings(bindingsVector *VectorNode, vals []Node) *VectorNode {
	for i, _ := range bindingsVector.Nodes {
		bind := bindingsVector.Nodes[i].(*VectorNode).Nodes
		bind[len(bind)-1] = vals[i]
	}

	return bindingsVector
}

func addRecurLabelAndBindings(label *IdentNode, bindingsVector *VectorNode, nodes []Node) {
	for _, node := range nodes {
		if node.Type() == NodeCall {
			n := node.(*CallNode)
			if ident, ok := n.Callee.(*IdentNode); ok && ident.Ident == "recur" {
				newValues := make([]Node, len(n.Args))
				copy(newValues, n.Args)

				n.Args = make([]Node, 2)
				n.Args[0] = addNewValuesToBindings(bindingsVector.Copy().(*VectorNode), newValues)
				n.Args[1] = label
			} else {
				addRecurLabelAndBindings(label, bindingsVector, n.Args)
			}
		}
	}
}

func makeLoop(node *CallNode) *ast.CallExpr {
	returnIdent := generateIdent()
	loopIdent := generateIdent()

	fnBody := EmptyS()

	bindingsVector := node.Args[0].(*VectorNode)

	addRecurLabelAndBindings(NewIdentNode(loopIdent.String()), bindingsVector.Copy().(*VectorNode), node.Args[1:])

	bindings := makeBindings(bindingsVector, token.DEFINE)
	returnIdentValueSpec := makeValueSpec(I(returnIdent), nil, anyType)
	returnIdentDecl := makeDeclStmt(makeGeneralDecl(token.VAR, []ast.Spec{returnIdentValueSpec}))

	fnBody = append(fnBody, bindings...)
	fnBody = append(fnBody, returnIdentDecl)

	init := makeAssignStmt(E(loopIdent), E(ast.NewIdent("true")), token.DEFINE)
	forBody := EmptyS()

	forBody = append(forBody, makeAssignStmt(E(loopIdent), E(ast.NewIdent("false")), token.ASSIGN))
	forBody = append(forBody, wrapExprsWithStmt(EvalExprs(node.Args[1:len(node.Args)-1]))...)
	forBody = append(forBody, makeAssignStmt(E(returnIdent), E(EvalExpr(node.Args[len(node.Args)-1])), token.ASSIGN))

	forStmt := makeForStmt(init, nil, loopIdent, makeBlockStmt(forBody))

	fnBody = append(fnBody, forStmt)
	fnBody = append(fnBody, makeReturnStmt(E(returnIdent)))

	results := makeFieldList([]*ast.Field{makeField(nil, anyType)})
	fnType := makeFuncType(results, nil)
	fn := makeFuncLit(fnType, makeBlockStmt(fnBody))

	return makeFuncCall(fn, EmptyE())
}

func makeRecur(node *CallNode) *ast.CallExpr {
	bindings := makeBindings(node.Args[0].(*VectorNode), token.ASSIGN)
	loopUpdate := makeAssignStmt(E(EvalExpr(node.Args[1])), E(ast.NewIdent("true")), token.ASSIGN)

	body := append(EmptyS(), bindings...)
	body = append(body, loopUpdate, makeReturnStmt(E(ast.NewIdent("nil"))))

	resultType := makeFieldList([]*ast.Field{makeField(nil, anyType)})
	fnType := makeFuncType(resultType, nil)
	fn := makeFuncLit(fnType, makeBlockStmt(body))
	return makeFuncCall(fn, EmptyE())
}

func isAssert(node *CallNode) bool {
	// Need an identifier for it to be "assert"
	if node.Callee.Type() != NodeIdent {
		return false
	}

	// Not a "loop"
	if callee := node.Callee.(*IdentNode); callee.Ident != "assert" {
		return false
	}

	if len(node.Args) != 2 {
		panic("assert needs 2 arguments")
	}

	if _, ok := node.Args[0].(*IdentNode); !ok {
		panic("assert's first argument needs to be a type")
	}

	return true
}

func makeAssert(node *CallNode) *ast.TypeAssertExpr {
	return makeTypeAssertion(EvalExpr(node.Args[1]), ast.NewIdent(node.Args[0].(*IdentNode).Ident))
}

var coreFuncs = []string{"get"}

func isCoreFunc(node *CallNode) bool {
	// Need an identifier for it to be a func
	if node.Callee.Type() != NodeIdent {
		return false
	}

	ident := node.Callee.(*IdentNode).Ident

	for _, v := range coreFuncs {
		if v == ident {
			return true
		}
	}

	return false
}

// TODO: just a quick and dirty implementation
func makeCoreCall(node *CallNode) ast.Expr {
	ident := node.Callee.(*IdentNode).Ident
	node.Callee.(*IdentNode).Ident = "core/" + ident
	return evalFuncCall(node)
}
