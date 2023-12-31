package sysbench

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/insomniacslk/xjson"
	"github.com/linuxboot/contest/pkg/event"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps"
)

// We need a default timeout to avoid endless running tests.
const (
	defaultTimeout = time.Minute
	in             = "input"
	exp            = "expect"
)

var (
	defaultArgs = []string{"cpu", "run"}
)

type inputStepParams struct {
	Transport struct {
		Proto   string          `json:"proto"`
		Options json.RawMessage `json:"options,omitempty"`
	} `json:"transport,omitempty"`

	Parameter struct {
		Args []string `json:"args"`
	} `json:"parameter"`

	Options struct {
		Timeout xjson.Duration `json:"timeout,omitempty"`
	} `json:"options,omitempty"`
}

type Expect struct {
	Option string `json:"option"`
	Value  string `json:"value"`
}

// Name is the name used to look this plugin up.
var Name = "Sysbench"

// TestStep implementation for this teststep plugin
type TestStep struct {
	inputStepParams
	expectStepParams []Expect
}

// Run executes the cmd step.
func (ts *TestStep) Run(ctx xcontext.Context, ch test.TestStepChannels, params test.TestStepParameters, ev testevent.Emitter, resumeState json.RawMessage) (json.RawMessage, error) {
	// Validate the parameter
	if err := ts.validateAndPopulate(params); err != nil {
		return nil, err
	}

	tr := NewTargetRunner(ts, ev)
	return teststeps.ForEachTarget(Name, ctx, ch, tr.Run)
}

func (ts *TestStep) validateAndPopulate(stepParams test.TestStepParameters) error {
	var (
		input        *test.Param
		tmp          Expect
		expectParams []Expect
	)

	if input = stepParams.GetOne(in); input.IsEmpty() {
		return fmt.Errorf("input parameter cannot be empty")
	}

	if err := json.Unmarshal(input.JSON(), &ts.inputStepParams); err != nil {
		return fmt.Errorf("failed to deserialize %q parameters: %v", in, err)
	}

	expect := stepParams.Get(exp)
	if len(expect) == 0 {
		return fmt.Errorf("expect parameter cannot be empty")
	}

	for _, item := range expect {
		if err := json.Unmarshal(item.JSON(), &tmp); err != nil {
			return fmt.Errorf("failed to deserialize %q parameter: %v", exp, err)
		}

		expectParams = append(expectParams, tmp)
	}

	ts.expectStepParams = expectParams

	return nil
}

// ValidateParameters validates the parameters associated to the TestStep
func (ts *TestStep) ValidateParameters(_ xcontext.Context, params test.TestStepParameters) error {
	return ts.validateAndPopulate(params)
}

// New initializes and returns a new HWaaS test step.
func New() test.TestStep {
	return &TestStep{}
}

// Load returns the name, factory and events which are needed to register the step.
func Load() (string, test.TestStepFactory, []event.Name) {
	return Name, New, Events
}

// Name returns the plugin name.
func (ts TestStep) Name() string {
	return Name
}
