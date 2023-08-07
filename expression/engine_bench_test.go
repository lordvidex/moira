package expression

import (
	"testing"

	"github.com/moira-alert/moira"
)

var expressions = []getExpressionValuesTest{
	{
		values:        TriggerExpression{MainTargetValue: 10.0, WarnValue: ptr(60.0), ErrorValue: ptr(90.0), TriggerType: moira.RisingTrigger},
		expectedError: nil,
		expectedValue: moira.StateOK,
	},
	{
		values:        TriggerExpression{MainTargetValue: 60.0, WarnValue: ptr(60.0), ErrorValue: ptr(90.0), TriggerType: moira.RisingTrigger},
		expectedError: nil,
		expectedValue: moira.StateWARN,
	},
	{
		values:        TriggerExpression{MainTargetValue: 90.0, WarnValue: ptr(60.0), ErrorValue: ptr(90.0), TriggerType: moira.RisingTrigger},
		expectedError: nil,
		expectedValue: moira.StateERROR,
	},
}

func BenchmarkGovaluteEngine(b *testing.B) {
	engine := govaluteEngine{}
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(expressions); j++ {
			state, err := engine.Evaluate(&expressions[j].values)
			if err != nil {
				if err != expressions[j].expectedError {
					b.Errorf("expected %v, got %v", expressions[j].expectedValue, state)
				}
				continue
			}
			if state != expressions[j].expectedValue {
				b.Errorf("expected %v, got %v", expressions[j].expectedValue, state)
			}
		}
	}
}

func BenchmarkExprEngine(b *testing.B) {
	engine := exprEngine{}
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(expressions); j++ {
			state, err := engine.Evaluate(&expressions[j].values)
			if err != nil {
				if err != expressions[j].expectedError {
					b.Errorf("expected %v, got %v", expressions[j].expectedValue, state)
				}
				continue
			}
			if state != expressions[j].expectedValue {
				b.Errorf("expected %v, got %v", expressions[j].expectedValue, state)
			}
		}
	}
}
