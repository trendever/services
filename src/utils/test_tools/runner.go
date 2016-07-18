package test_tools

import (
	"errors"
	"fmt"
	"reflect"
)

const (
	RuleDeep = 1 << iota
	RuleStr  = 2
)

type Test map[string]interface{}
type Tests []Test
type RunFunc func(Test) []interface{}
type Rule struct {
	Type int
	Key  string
}
type Runner struct {
	Tests  Tests
	Run    RunFunc
	Rules  []Rule
	Errors []interface{}
}

func NewRunner(tests Tests, f RunFunc, rules []Rule) *Runner {
	return &Runner{
		Tests: tests,
		Run:   f,
		Rules: rules,
	}
}

func (r *Runner) RunTests() {
	for j, test := range r.Tests {
		args := r.Run(test)
		for i, arg := range args {

			if i >= len(r.Rules) {
				r.AddError(errors.New(ExpectedButGot("count of args", len(r.Rules), len(args))))
				return
			}
			rule := r.Rules[i]
			var pass bool
			switch rule.Type {
			case RuleDeep:
				pass = reflect.DeepEqual(test[rule.Key], arg)
			case RuleStr:
				pass = ToStr(test[rule.Key]) == ToStr(arg)
			}
			if !pass {
				r.AddError(errors.New(ExpectedButGot(fmt.Sprintf("tests[%v] arg[%v]", j, rule.Key), test[rule.Key], arg)))
			}

		}
	}
}

func (r *Runner) AddError(e interface{}) {
	r.Errors = append(r.Errors, e)
}

func (r *Runner) HasErrors() bool {
	return len(r.Errors) > 0
}
