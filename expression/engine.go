package expression

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/ast"
	"github.com/antonmedv/expr/vm"
	"github.com/patrickmn/go-cache"

	"github.com/moira-alert/moira"
)

type expressionEngine interface {
	Evaluate(expression TriggerExpression) (moira.State, error)
}

type govaluteEngine struct {
}

type exprEngine struct {
}

func (engine govaluteEngine) Evaluate(expression *TriggerExpression) (moira.State, error) {
	if expression == nil {
		return "", fmt.Errorf("empty expression")
	}
	expr, err := getExpression(expression)
	if err != nil {
		return "", ErrInvalidExpression{internalError: err}
	}
	result, err := expr.Eval(expression)
	if err != nil {
		return "", ErrInvalidExpression{internalError: err}
	}
	switch res := result.(type) {
	case moira.State:
		return res, nil
	default:
		return "", ErrInvalidExpression{internalError: fmt.Errorf("expression result must be state value")}
	}
}

func (exprEngine) Visit(node *ast.Node) {
	if n, ok := (*node).(*ast.IdentifierNode); ok {
		ast.Patch(node, &ast.CallNode{
			Arguments: []ast.Node{
				&ast.StringNode{Value: n.Value},
			},
			Callee: &ast.IdentifierNode{Value: "Get"},
		})
	}
}

func (engine exprEngine) Evaluate(expression *TriggerExpression) (moira.State, error) {
	if expression == nil {
		return "", fmt.Errorf("empty expression")
	}
	switch expression.TriggerType {
	case "":
		return "", fmt.Errorf("trigger_type is not set")
	case moira.ExpressionTrigger:
		if expression.Expression == nil || *expression.Expression == "" {
			return "", fmt.Errorf("trigger_type set to expression, but no expression provided")
		}
	case moira.FallingTrigger:
		if expression.ErrorValue != nil && expression.WarnValue != nil {
			expression.Expression = &warnErrorFalling
		} else if expression.ErrorValue != nil {
			expression.Expression = &errFalling
		} else {
			expression.Expression = &warnFalling
		}
	case moira.RisingTrigger:
		if expression.ErrorValue != nil && expression.WarnValue != nil {
			expression.Expression = &warnErrorRising
		} else if expression.ErrorValue != nil {
			expression.Expression = &errRising
		} else {
			expression.Expression = &warnRising
		}
	default:
		return "", fmt.Errorf("wrong set of parametres: warn_value - %v, error_value - %v, trigger_type: %v",
			expression.WarnValue, expression.ErrorValue, expression.TriggerType)
	}
	cacheKey := fmt.Sprintf("[VALIDATED]%s", *expression.Expression)
	pr, found := exprCache.Get(cacheKey)
	program, ok := pr.(*vm.Program)
	if !ok {
		found = false
		exprCache.Delete(cacheKey)
	}
	if !found {
		var err error
		program, err = expr.Compile(
			*expression.Expression,
			expr.Patch(engine),
			expr.Optimize(true),
		)
		if err != nil {
			return "", err
		}
		exprCache.Set(cacheKey, program, cache.NoExpiration)
	}
	result, err := expr.Run(program, map[string]interface{}{
		"Get": expression.Get,
	})
	if err != nil {
		return "", err
	}
	switch res := result.(type) {
	case moira.State:
		return res, nil
	default:
		return "", ErrInvalidExpression{internalError: fmt.Errorf("expression result must be state value")}
	}
}
