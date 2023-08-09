package expression

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/moira-alert/moira"
)

type dump struct {
	expression TriggerExpression
	name       string
	id         string
}

var (
	dumpFile    = "dumps.txt"
	expressions []dump
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
	failReportFile, err := os.OpenFile("verification_report.toml", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer failReportFile.Close()
	writer := bufio.NewWriter(failReportFile)
	errorFormat := `[[error]]
id = "%s"
name = "%s"
expression = "%s"
err = '''%s'''`
	for _, expr := range expressions {
		expression := expr.expression
		t.Run(*expression.Expression, func(t *testing.T) {
			_, err := expression.Evaluate()
			if err != nil {
				failed++
				t.Errorf("error: %v\n", err)
				_, err = fmt.Fprintf(writer, errorFormat, expr.id, expr.name, *expression.Expression, err.Error())
				if err != nil {
					t.Fatal(err)
				}
				writer.WriteRune('\n')
			}
		})
	}
	writer.Flush()
	t.Logf("%d/%d failed tests\n", failed, len(expressions))
}

// reads the dump file line
func readDumpFile(file string) ([]dump, error) {
	r := regexp.MustCompile(`[t,T][0-9]+`)
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint
	scanner := bufio.NewScanner(f)
	expressions := make([]dump, 0)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ";")
		if len(fields) != 3 {
			return nil, fmt.Errorf("invalid dump file row %s", line)
		}
		matches := r.FindAllString(fields[2], -1)
		expression := TriggerExpression{
			Expression:              &fields[2],
			TriggerType:             moira.ExpressionTrigger,
			PreviousState:           moira.StateNODATA,
			MainTargetValue:         42,
			AdditionalTargetsValues: make(map[string]float64),
		}
		for _, match := range matches {
			match = strings.ToLower(match)
			expression.AdditionalTargetsValues[match] = 42
		}
		expressions = append(expressions, dump{
			expression: expression,
			name:       fields[1],
			id:         fields[0],
		})
	}
	return expressions, nil
}
