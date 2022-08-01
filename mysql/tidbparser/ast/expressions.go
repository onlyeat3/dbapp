// Copyright 2015 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package ast

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"virtdb/tidbparser/dependency/model"
	"virtdb/tidbparser/dependency/mysql"
	"virtdb/tidbparser/dependency/types"
	"virtdb/tidbparser/parser/opcode"
)

var (
	_ ExprNode = &BetweenExpr{}
	_ ExprNode = &BinaryOperationExpr{}
	_ ExprNode = &CaseExpr{}
	_ ExprNode = &ColumnNameExpr{}
	_ ExprNode = &CompareSubqueryExpr{}
	_ ExprNode = &DefaultExpr{}
	_ ExprNode = &ExistsSubqueryExpr{}
	_ ExprNode = &IsNullExpr{}
	_ ExprNode = &IsTruthExpr{}
	_ ExprNode = &ParamMarkerExpr{}
	_ ExprNode = &ParenthesesExpr{}
	_ ExprNode = &PatternInExpr{}
	_ ExprNode = &PatternLikeExpr{}
	_ ExprNode = &PatternRegexpExpr{}
	_ ExprNode = &PositionExpr{}
	_ ExprNode = &RowExpr{}
	_ ExprNode = &SubqueryExpr{}
	_ ExprNode = &UnaryOperationExpr{}
	_ ExprNode = &ValueExpr{}
	_ ExprNode = &ValuesExpr{}
	_ ExprNode = &VariableExpr{}

	_ Node = &ColumnName{}
	_ Node = &WhenClause{}
)

// ValueExpr is the simple value expression.
type ValueExpr struct {
	exprNode
	projectionOffset int
}

// Format the ExprNode into a Writer.
func (n *ValueExpr) Format(w io.Writer) {
	var s string
	switch n.Kind() {
	case types.KindNull:
		s = "NULL"
	case types.KindInt64:
		if n.Type.Flag&mysql.IsBooleanFlag != 0 {
			if n.GetInt64() > 0 {
				s = "TRUE"
			} else {
				s = "FALSE"
			}
		} else {
			s = strconv.FormatInt(n.GetInt64(), 10)
		}
	case types.KindUint64:
		s = strconv.FormatUint(n.GetUint64(), 10)
	case types.KindFloat32:
		s = strconv.FormatFloat(n.GetFloat64(), 'e', -1, 32)
	case types.KindFloat64:
		s = strconv.FormatFloat(n.GetFloat64(), 'e', -1, 64)
	case types.KindString, types.KindBytes:
		s = strconv.Quote(n.GetString())
	case types.KindMysqlDecimal:
		s = n.GetMysqlDecimal().String()
	case types.KindBinaryLiteral:
		if n.Type.Flag&mysql.UnsignedFlag != 0 {
			s = fmt.Sprintf("x'%x'", n.GetBytes())
		} else {
			s = n.GetBinaryLiteral().ToBitLiteralString(true)
		}
	default:
		panic("Can't format to string")
	}
	fmt.Fprint(w, s)
}

// NewValueExpr creates a ValueExpr with value, and sets default field type.
func NewValueExpr(value interface{}) *ValueExpr {
	if ve, ok := value.(*ValueExpr); ok {
		return ve
	}
	ve := &ValueExpr{}
	ve.SetValue(value)
	types.DefaultTypeForValue(value, &ve.Type)
	ve.projectionOffset = -1
	return ve
}

// SetProjectionOffset sets ValueExpr.projectionOffset for logical plan builder.
func (n *ValueExpr) SetProjectionOffset(offset int) {
	n.projectionOffset = offset
}

// GetProjectionOffset returns ValueExpr.projectionOffset.
func (n *ValueExpr) GetProjectionOffset() int {
	return n.projectionOffset
}

// Accept implements Node interface.
func (n *ValueExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*ValueExpr)
	return v.Leave(n)
}

// BetweenExpr is for "between and" or "not between and" expression.
type BetweenExpr struct {
	exprNode
	// Expr is the expression to be checked.
	Expr ExprNode
	// Left is the expression for minimal value in the range.
	Left ExprNode
	// Right is the expression for maximum value in the range.
	Right ExprNode
	// Not is true, the expression is "not between and".
	Not bool
}

// Format the ExprNode into a Writer.
func (n *BetweenExpr) Format(w io.Writer) {
	n.Expr.Format(w)
	fmt.Fprint(w, " BETWEEN ")
	n.Left.Format(w)
	fmt.Fprint(w, " AND ")
	n.Right.Format(w)
}

// Accept implements Node interface.
func (n *BetweenExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}

	n = newNode.(*BetweenExpr)
	node, ok := n.Expr.Accept(v)
	if !ok {
		return n, false
	}
	n.Expr = node.(ExprNode)

	node, ok = n.Left.Accept(v)
	if !ok {
		return n, false
	}
	n.Left = node.(ExprNode)

	node, ok = n.Right.Accept(v)
	if !ok {
		return n, false
	}
	n.Right = node.(ExprNode)

	return v.Leave(n)
}

// BinaryOperationExpr is for binary operation like `1 + 1`, `1 - 1`, etc.
type BinaryOperationExpr struct {
	exprNode
	// Op is the operator code for BinaryOperation.
	Op opcode.Op
	// L is the left expression in BinaryOperation.
	L ExprNode
	// R is the right expression in BinaryOperation.
	R ExprNode
}

// Format the ExprNode into a Writer.
func (n *BinaryOperationExpr) Format(w io.Writer) {
	n.L.Format(w)
	fmt.Fprint(w, " ")
	n.Op.Format(w)
	fmt.Fprint(w, " ")
	n.R.Format(w)
}

// Accept implements Node interface.
func (n *BinaryOperationExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}

	n = newNode.(*BinaryOperationExpr)
	node, ok := n.L.Accept(v)
	if !ok {
		return n, false
	}
	n.L = node.(ExprNode)

	node, ok = n.R.Accept(v)
	if !ok {
		return n, false
	}
	n.R = node.(ExprNode)

	return v.Leave(n)
}

// WhenClause is the when clause in Case expression for "when condition then result".
type WhenClause struct {
	node
	// Expr is the condition expression in WhenClause.
	Expr ExprNode
	// Result is the result expression in WhenClause.
	Result ExprNode
}

// Accept implements Node Accept interface.
func (n *WhenClause) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}

	n = newNode.(*WhenClause)
	node, ok := n.Expr.Accept(v)
	if !ok {
		return n, false
	}
	n.Expr = node.(ExprNode)

	node, ok = n.Result.Accept(v)
	if !ok {
		return n, false
	}
	n.Result = node.(ExprNode)
	return v.Leave(n)
}

// CaseExpr is the case expression.
type CaseExpr struct {
	exprNode
	// Value is the compare value expression.
	Value ExprNode
	// WhenClauses is the condition check expression.
	WhenClauses []*WhenClause
	// ElseClause is the else result expression.
	ElseClause ExprNode
}

// Format the ExprNode into a Writer.
func (n *CaseExpr) Format(w io.Writer) {
	fmt.Fprint(w, "CASE ")
	n.Value.Format(w)
	fmt.Fprint(w, " ")
	for _, clause := range n.WhenClauses {
		fmt.Fprint(w, "WHEN ")
		clause.Expr.Format(w)
		fmt.Fprint(w, " THEN ")
		clause.Result.Format(w)
	}
	if n.ElseClause != nil {
		fmt.Fprint(w, " ELSE ")
		n.ElseClause.Format(w)
	}
	fmt.Fprint(w, " END")
}

// Accept implements Node Accept interface.
func (n *CaseExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}

	n = newNode.(*CaseExpr)
	if n.Value != nil {
		node, ok := n.Value.Accept(v)
		if !ok {
			return n, false
		}
		n.Value = node.(ExprNode)
	}
	for i, val := range n.WhenClauses {
		node, ok := val.Accept(v)
		if !ok {
			return n, false
		}
		n.WhenClauses[i] = node.(*WhenClause)
	}
	if n.ElseClause != nil {
		node, ok := n.ElseClause.Accept(v)
		if !ok {
			return n, false
		}
		n.ElseClause = node.(ExprNode)
	}
	return v.Leave(n)
}

// SubqueryExec represents a subquery executor interface.
// This interface is implemented in executor and used in plan/evaluator.
// It will execute the subselect and get the result.
type SubqueryExec interface {
	// ColumnCount returns column count for the sub query.
	ColumnCount() (int, error)
}

// SubqueryExpr represents a subquery.
type SubqueryExpr struct {
	exprNode
	// Query is the query SelectNode.
	Query        ResultSetNode
	SubqueryExec SubqueryExec
	Evaluated    bool
	Correlated   bool
	MultiRows    bool
	Exists       bool
}

// Format the ExprNode into a Writer.
func (n *SubqueryExpr) Format(w io.Writer) {
	panic("Not implemented")
}

// Accept implements Node Accept interface.
func (n *SubqueryExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*SubqueryExpr)
	if n.SubqueryExec != nil {
		return v.Leave(n)
	}
	node, ok := n.Query.Accept(v)
	if !ok {
		return n, false
	}
	n.Query = node.(ResultSetNode)
	return v.Leave(n)
}

// CompareSubqueryExpr is the expression for "expr cmp (select ...)".
// See https://dev.mysql.com/doc/refman/5.7/en/comparisons-using-subqueries.html
// See https://dev.mysql.com/doc/refman/5.7/en/any-in-some-subqueries.html
// See https://dev.mysql.com/doc/refman/5.7/en/all-subqueries.html
type CompareSubqueryExpr struct {
	exprNode
	// L is the left expression
	L ExprNode
	// Op is the comparison opcode.
	Op opcode.Op
	// R is the subquery for right expression, may be rewritten to other type of expression.
	R ExprNode
	// All is true, we should compare all records in subquery.
	All bool
}

// Format the ExprNode into a Writer.
func (n *CompareSubqueryExpr) Format(w io.Writer) {
	panic("Not implemented")
}

// Accept implements Node Accept interface.
func (n *CompareSubqueryExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*CompareSubqueryExpr)
	node, ok := n.L.Accept(v)
	if !ok {
		return n, false
	}
	n.L = node.(ExprNode)
	node, ok = n.R.Accept(v)
	if !ok {
		return n, false
	}
	n.R = node.(ExprNode)
	return v.Leave(n)
}

// ColumnName represents column name.
type ColumnName struct {
	node
	Schema model.CIStr
	Table  model.CIStr
	Name   model.CIStr
}

// Accept implements Node Accept interface.
func (n *ColumnName) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*ColumnName)
	return v.Leave(n)
}

// String implements Stringer interface.
func (n *ColumnName) String() string {
	result := n.Name.L
	if n.Table.L != "" {
		result = n.Table.L + "." + result
	}
	if n.Schema.L != "" {
		result = n.Schema.L + "." + result
	}
	return result
}

// OrigColName returns the full original column name.
func (n *ColumnName) OrigColName() (ret string) {
	ret = n.Name.O
	if n.Table.O == "" {
		return
	}
	ret = n.Table.O + "." + ret
	if n.Schema.O == "" {
		return
	}
	ret = n.Schema.O + "." + ret
	return
}

// ColumnNameExpr represents a column name expression.
type ColumnNameExpr struct {
	exprNode

	// Name is the referenced column name.
	Name *ColumnName

	// Refer is the result field the column name refers to.
	// The value of Refer.Expr is used as the value of the expression.
	Refer *ResultField
}

// Format the ExprNode into a Writer.
func (n *ColumnNameExpr) Format(w io.Writer) {
	name := strings.Replace(n.Name.String(), ".", "`.`", -1)
	fmt.Fprintf(w, "`%s`", name)
}

// Accept implements Node Accept interface.
func (n *ColumnNameExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*ColumnNameExpr)
	node, ok := n.Name.Accept(v)
	if !ok {
		return n, false
	}
	n.Name = node.(*ColumnName)
	return v.Leave(n)
}

// DefaultExpr is the default expression using default value for a column.
type DefaultExpr struct {
	exprNode
	// Name is the column name.
	Name *ColumnName
}

// Format the ExprNode into a Writer.
func (n *DefaultExpr) Format(w io.Writer) {
	panic("Not implemented")
}

// Accept implements Node Accept interface.
func (n *DefaultExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*DefaultExpr)
	if n.Name != nil {
		node, ok := n.Name.Accept(v)
		if !ok {
			return n, false
		}
		n.Name = node.(*ColumnName)
	}
	return v.Leave(n)
}

// ExistsSubqueryExpr is the expression for "exists (select ...)".
// See https://dev.mysql.com/doc/refman/5.7/en/exists-and-not-exists-subqueries.html
type ExistsSubqueryExpr struct {
	exprNode
	// Sel is the subquery, may be rewritten to other type of expression.
	Sel ExprNode
}

// Format the ExprNode into a Writer.
func (n *ExistsSubqueryExpr) Format(w io.Writer) {
	panic("Not implemented")
}

// Accept implements Node Accept interface.
func (n *ExistsSubqueryExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*ExistsSubqueryExpr)
	node, ok := n.Sel.Accept(v)
	if !ok {
		return n, false
	}
	n.Sel = node.(ExprNode)
	return v.Leave(n)
}

// PatternInExpr is the expression for in operator, like "expr in (1, 2, 3)" or "expr in (select c from t)".
type PatternInExpr struct {
	exprNode
	// Expr is the value expression to be compared.
	Expr ExprNode
	// List is the list expression in compare list.
	List []ExprNode
	// Not is true, the expression is "not in".
	Not bool
	// Sel is the subquery, may be rewritten to other type of expression.
	Sel ExprNode
}

// Format the ExprNode into a Writer.
func (n *PatternInExpr) Format(w io.Writer) {
	n.Expr.Format(w)
	fmt.Fprint(w, " IN (")
	for i, expr := range n.List {
		expr.Format(w)
		if i != len(n.List)-1 {
			fmt.Fprint(w, ",")
		}
	}
	fmt.Fprint(w, ")")
}

// Accept implements Node Accept interface.
func (n *PatternInExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*PatternInExpr)
	node, ok := n.Expr.Accept(v)
	if !ok {
		return n, false
	}
	n.Expr = node.(ExprNode)
	for i, val := range n.List {
		node, ok = val.Accept(v)
		if !ok {
			return n, false
		}
		n.List[i] = node.(ExprNode)
	}
	if n.Sel != nil {
		node, ok = n.Sel.Accept(v)
		if !ok {
			return n, false
		}
		n.Sel = node.(ExprNode)
	}
	return v.Leave(n)
}

// IsNullExpr is the expression for null check.
type IsNullExpr struct {
	exprNode
	// Expr is the expression to be checked.
	Expr ExprNode
	// Not is true, the expression is "is not null".
	Not bool
}

// Format the ExprNode into a Writer.
func (n *IsNullExpr) Format(w io.Writer) {
	n.Expr.Format(w)
	if n.Not {
		fmt.Fprint(w, " IS NOT NULL")
		return
	}
	fmt.Fprint(w, " IS NULL")
}

// Accept implements Node Accept interface.
func (n *IsNullExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*IsNullExpr)
	node, ok := n.Expr.Accept(v)
	if !ok {
		return n, false
	}
	n.Expr = node.(ExprNode)
	return v.Leave(n)
}

// IsTruthExpr is the expression for true/false check.
type IsTruthExpr struct {
	exprNode
	// Expr is the expression to be checked.
	Expr ExprNode
	// Not is true, the expression is "is not true/false".
	Not bool
	// True indicates checking true or false.
	True int64
}

// Format the ExprNode into a Writer.
func (n *IsTruthExpr) Format(w io.Writer) {
	n.Expr.Format(w)
	if n.Not {
		fmt.Fprint(w, " IS NOT")
	} else {
		fmt.Fprint(w, " IS")
	}
	if n.True > 0 {
		fmt.Fprint(w, " TRUE")
	} else {
		fmt.Fprint(w, " FALSE")
	}
}

// Accept implements Node Accept interface.
func (n *IsTruthExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*IsTruthExpr)
	node, ok := n.Expr.Accept(v)
	if !ok {
		return n, false
	}
	n.Expr = node.(ExprNode)
	return v.Leave(n)
}

// PatternLikeExpr is the expression for like operator, e.g, expr like "%123%"
type PatternLikeExpr struct {
	exprNode
	// Expr is the expression to be checked.
	Expr ExprNode
	// Pattern is the like expression.
	Pattern ExprNode
	// Not is true, the expression is "not like".
	Not bool

	Escape byte

	PatChars []byte
	PatTypes []byte
}

// Format the ExprNode into a Writer.
func (n *PatternLikeExpr) Format(w io.Writer) {
	n.Expr.Format(w)
	fmt.Fprint(w, " LIKE ")
	n.Pattern.Format(w)
	if n.Escape != '\\' {
		fmt.Fprint(w, " ESCAPE ")
		fmt.Fprintf(w, "'%c'", n.Escape)
	}
}

// Accept implements Node Accept interface.
func (n *PatternLikeExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*PatternLikeExpr)
	if n.Expr != nil {
		node, ok := n.Expr.Accept(v)
		if !ok {
			return n, false
		}
		n.Expr = node.(ExprNode)
	}
	if n.Pattern != nil {
		node, ok := n.Pattern.Accept(v)
		if !ok {
			return n, false
		}
		n.Pattern = node.(ExprNode)
	}
	return v.Leave(n)
}

// ParamMarkerExpr expression holds a place for another expression.
// Used in parsing prepare statement.
type ParamMarkerExpr struct {
	exprNode
	Offset int
	Order  int
}

// Format the ExprNode into a Writer.
func (n *ParamMarkerExpr) Format(w io.Writer) {
	panic("Not implemented")
}

// Accept implements Node Accept interface.
func (n *ParamMarkerExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*ParamMarkerExpr)
	return v.Leave(n)
}

// ParenthesesExpr is the parentheses expression.
type ParenthesesExpr struct {
	exprNode
	// Expr is the expression in parentheses.
	Expr ExprNode
}

// Format the ExprNode into a Writer.
func (n *ParenthesesExpr) Format(w io.Writer) {
	fmt.Fprint(w, "(")
	n.Expr.Format(w)
	fmt.Fprint(w, ")")
}

// Accept implements Node Accept interface.
func (n *ParenthesesExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*ParenthesesExpr)
	if n.Expr != nil {
		node, ok := n.Expr.Accept(v)
		if !ok {
			return n, false
		}
		n.Expr = node.(ExprNode)
	}
	return v.Leave(n)
}

// PositionExpr is the expression for order by and group by position.
// MySQL use position expression started from 1, it looks a little confused inner.
// maybe later we will use 0 at first.
type PositionExpr struct {
	exprNode
	// N is the position, started from 1 now.
	N int
	// Refer is the result field the position refers to.
	Refer *ResultField
}

// Format the ExprNode into a Writer.
func (n *PositionExpr) Format(w io.Writer) {
	panic("Not implemented")
}

// Accept implements Node Accept interface.
func (n *PositionExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*PositionExpr)
	return v.Leave(n)
}

// PatternRegexpExpr is the pattern expression for pattern match.
type PatternRegexpExpr struct {
	exprNode
	// Expr is the expression to be checked.
	Expr ExprNode
	// Pattern is the expression for pattern.
	Pattern ExprNode
	// Not is true, the expression is "not rlike",
	Not bool

	// Re is the compiled regexp.
	Re *regexp.Regexp
	// Sexpr is the string for Expr expression.
	Sexpr *string
}

// Format the ExprNode into a Writer.
func (n *PatternRegexpExpr) Format(w io.Writer) {
	n.Expr.Format(w)
	if n.Not {
		fmt.Fprint(w, " NOT REGEXP ")
	} else {
		fmt.Fprint(w, " REGEXP ")
	}
	n.Pattern.Format(w)
}

// Accept implements Node Accept interface.
func (n *PatternRegexpExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*PatternRegexpExpr)
	node, ok := n.Expr.Accept(v)
	if !ok {
		return n, false
	}
	n.Expr = node.(ExprNode)
	node, ok = n.Pattern.Accept(v)
	if !ok {
		return n, false
	}
	n.Pattern = node.(ExprNode)
	return v.Leave(n)
}

// RowExpr is the expression for row constructor.
// See https://dev.mysql.com/doc/refman/5.7/en/row-subqueries.html
type RowExpr struct {
	exprNode

	Values []ExprNode
}

// Format the ExprNode into a Writer.
func (n *RowExpr) Format(w io.Writer) {
	panic("Not implemented")
}

// Accept implements Node Accept interface.
func (n *RowExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*RowExpr)
	for i, val := range n.Values {
		node, ok := val.Accept(v)
		if !ok {
			return n, false
		}
		n.Values[i] = node.(ExprNode)
	}
	return v.Leave(n)
}

// UnaryOperationExpr is the expression for unary operator.
type UnaryOperationExpr struct {
	exprNode
	// Op is the operator opcode.
	Op opcode.Op
	// V is the unary expression.
	V ExprNode
}

// Format the ExprNode into a Writer.
func (n *UnaryOperationExpr) Format(w io.Writer) {
	n.Op.Format(w)
	n.V.Format(w)
}

// Accept implements Node Accept interface.
func (n *UnaryOperationExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*UnaryOperationExpr)
	node, ok := n.V.Accept(v)
	if !ok {
		return n, false
	}
	n.V = node.(ExprNode)
	return v.Leave(n)
}

// ValuesExpr is the expression used in INSERT VALUES.
type ValuesExpr struct {
	exprNode
	// Column is column name.
	Column *ColumnNameExpr
}

// Format the ExprNode into a Writer.
func (n *ValuesExpr) Format(w io.Writer) {
	panic("Not implemented")
}

// Accept implements Node Accept interface.
func (n *ValuesExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*ValuesExpr)
	node, ok := n.Column.Accept(v)
	if !ok {
		return n, false
	}
	n.Column = node.(*ColumnNameExpr)
	return v.Leave(n)
}

// VariableExpr is the expression for variable.
type VariableExpr struct {
	exprNode
	// Name is the variable name.
	Name string
	// IsGlobal indicates whether this variable is global.
	IsGlobal bool
	// IsSystem indicates whether this variable is a system variable in current session.
	IsSystem bool
	// Value is the variable value.
	Value ExprNode
}

// Format the ExprNode into a Writer.
func (n *VariableExpr) Format(w io.Writer) {
	panic("Not implemented")
}

// Accept implements Node Accept interface.
func (n *VariableExpr) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*VariableExpr)
	if n.Value == nil {
		return v.Leave(n)
	}

	node, ok := n.Value.Accept(v)
	if !ok {
		return n, false
	}
	n.Value = node.(ExprNode)
	return v.Leave(n)
}
