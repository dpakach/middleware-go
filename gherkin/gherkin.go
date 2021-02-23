package gherkin
import (
	"fmt"

    "reflect"
    "errors"
)

type Table []map[string]string

type Token int

const (
    GIVEN Token = iota
    WHEN
    THEN
)

type StepDef struct {
	Token   Token
	Pattern string
	Action  reflect.Value
}

func (s *StepDef) Match(step Step) bool {
	if step.Token != s.Token {
		return false
	}
	fmt.Println(s)
	fmt.Println(step)

	if step.StepText == s.Pattern {
		dataLen := len(step.Data)
		if len(step.Table) != 0 {
			dataLen += 1
		}
		if dataLen != s.Action.Type().NumIn() {
			return false
		}

		for i, val := range step.Data {
			if s.Action.Type().In(i) != reflect.TypeOf(val) {
				return false
			}
		}

		return true
	}
	return false
}

func (s *StepDef) Run(args ...interface{}) error {
	callArgs := []reflect.Value{}
	for _, arg := range args {
		a := reflect.ValueOf(arg)
		callArgs = append(callArgs, a)
	}
	if numArgs := s.Action.Type().NumIn(); numArgs != len(callArgs) {
		return errors.New(fmt.Sprintf("Number of arguments mismatched, expected %v, got %v", len(callArgs), numArgs))
	}

	val := s.Action.Call(callArgs)
	fmt.Println(val)
	if len(val) > 0 {
		err, ok := val[0].Interface().(error)
		if ok {
			return err
		}
	}
	return nil
}

type Suite struct {
	steps   []*StepDef
}

func NewSuite() *Suite {
	return &Suite{
		[]*StepDef{},
	}
}

func (s *Suite) Given(pattern string, action interface{}) {
	s.addStep(GIVEN, pattern, action)
}

func (s *Suite) When(pattern string, action interface{}) {
	s.addStep(WHEN, pattern, action)
}

func (s *Suite) Then(pattern string, action interface{}) {
	s.addStep(THEN, pattern, action)
}

func (c *Suite) addStep(token Token, pattern string, action interface{}) error {
	v := reflect.ValueOf(action)
	typ := v.Type()
	if typ.Kind() != reflect.Func {
		panic(fmt.Sprintf("expected handler to be func, but got: %T", action))
	}
	for _, stepDef := range c.steps {
		if stepDef.Pattern == pattern {
			return errors.New("Step Definition already exists")
		}
	}
	c.steps = append(c.steps, &StepDef{token, pattern, v})
	return nil
}

func (c *Suite) GetMatch(step Step) (*StepDef, error) {
	for _, stepDef := range c.steps {
		fmt.Println(stepDef)
		if stepDef.Match(step) {
			return stepDef, nil
		}
	}
	return nil, errors.New("Could not find step definition")
}

func verifyReflectFunction(action interface{}) reflect.Value {
	v := reflect.ValueOf(action)
	typ := v.Type()
	if typ.Kind() != reflect.Func {
		panic(fmt.Sprintf("expected handler to be func, but got: %T", action))
	}
	return v
}

type Step struct {
    Token Token `json:"token"`
    StepText string `json:"step_text"`
    Data []interface{} `json:"data"`
    Table []map[string]string `json:"table"`
}