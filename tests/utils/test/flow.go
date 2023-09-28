package test

import (
	"strings"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
)

type TestFlow struct {
	ctx     *TestContext
	skip    bool
	tests   []interface{}
	Results []*TestResult
}

func Flow(ctx *TestContext) *TestFlow {
	tf := TestFlow{ctx: ctx, tests: make([]interface{}, 0)}
	return &tf
}

func (t *TestFlow) Test(testCase TestCase) *TestFlow {
	if t.skip {
		return t
	}
	log.Printf("-------------- %s -------------- ", testCase.GetName())
	result := testCase.Execute(t.ctx)
	t.add(testCase)
	if result.IsFailed() {
		t.Results = append(t.Results, result)
		t.skip = true
	}
	return t
}

func (t *TestFlow) add(testCase TestCase) *TestFlow {
	t.tests = append(t.tests, testCase)
	return t
}

func (t *TestFlow) TearDown() *TestFlow {
	for i := len(t.tests) - 1; i >= 0; i-- {
		tc := t.tests[i]
		testCase, ok := tc.(TestCase)
		if ok {
			testCase.TearDown(t.ctx)
		} else {
			testFlow, ok := tc.(*TestFlow)
			if ok {
				testFlow.TearDown()
			}
		}
	}
	return t
}

func (t *TestFlow) Audit() error {
	var err error
	for _, res := range t.Results {
		if res.IsFailed() {
			if err == nil {
				err = errors.New("[Failed]")
			}
			err = errors.Wrap(err, strings.Join(res.Errors, " , "))
		}
	}

	return err
}
