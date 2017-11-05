package main

import (
	"go/ast"
	"go/token"
)

func EvalExprs(nodes []Node) []ast.Expr {
	out := make([]ast.Expr, len(nodes))

	for i, node := range nodes {
		out[i] = EvalExpr(node)
	}

	return out
}

func EvalExpr(node Node) ast.Expr {
	switch t := node.Type(); t {
	case NodeCall:
		node := node.(*CallNode)
		return evalFuncCall(node)

	case NodeVector:
		node := node.(*VectorNode)
		return makeVector(anyType, EvalExprs(node.Nodes))

	case NodeNumber:
		node := node.(*NumberNode)
		return makeBasicLit(node.NumberType, node.Value)

	case NodeString:
		node := node.(*StringNode)
		return makeBasicLit(token.STRING, node.Value)

	case NodeIdent:
		node := node.(*IdentNode)
		return makeIdomaticSelector(node.Ident)

	default:
		println(t)
		panic("not implemented yet!")
	}
}
