package test_tools

import "testing"

func TestRunTests(t *testing.T) {
	tests := Tests{
		Test{
			"a":      1,
			"b":      2,
			"result": 4,
		},
		Test{
			"a":      1,
			"b":      2,
			"result": 5,
		},
		Test{
			"a":      1,
			"b":      2,
			"result": 3,
		},
	}
	runner := &Runner{
		Tests: tests,
		Run: func(test Test) []interface{} {
			return []interface{}{test["a"].(int) + test["b"].(int)}
		},
		Rules: []Rule{
			{RuleDeep, "result"},
		},
	}
	runner.RunTests()
	if len(runner.Errors) != 2 {
		t.Error(runner.Errors...)
	}
}

func TestNil(t *testing.T) {
	tests := Tests{
		Test{
			"result": nil,
		},
	}
	runner := &Runner{
		Tests: tests,
		Run: func(test Test) []interface{} {
			return []interface{}{nil, nil}
		},
		Rules: []Rule{
			{RuleDeep, "result"},
			{RuleStr, "result"},
		},
	}
	runner.RunTests()
	if runner.HasErrors() {
		t.Error(runner.Errors...)
	}
}
