package expression

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/moira-alert/moira"
)

var (
	dumpFile    = "dumps.txt"
	expressions []TriggerExpression
)

func TestMain(m *testing.M) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	expressions, err = readDumpFile(filepath.Join(wd, "../", dumpFile))
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(m.Run())
}

func TestValidate(t *testing.T) {
	failed := 0
	for _, expr := range expressions {
		t.Run(*expr.Expression, func(t *testing.T) {
			_, err := expr.Evaluate()
			if err != nil {
				failed++
				t.Errorf("error: %v\n", err)
			}
		})
	}
	t.Logf("%d/%d failed tests\n", failed, len(expressions))
}

// reads the dump file line
func readDumpFile(file string) ([]TriggerExpression, error) {
	r := regexp.MustCompile(`t[0-9]+`)
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint
	scanner := bufio.NewScanner(f)
	expressions := make([]TriggerExpression, 0)
	for scanner.Scan() {
		line := scanner.Text()
		matches := r.FindAllString(line, -1)
		expression := TriggerExpression{
			Expression:              &line,
			TriggerType:             moira.ExpressionTrigger,
			PreviousState:           moira.StateNODATA,
			MainTargetValue:         42,
			AdditionalTargetsValues: make(map[string]float64),
		}
		for _, match := range matches {
			expression.AdditionalTargetsValues[match] = 42
		}
		expressions = append(expressions, expression)
	}
	return expressions, nil
}
