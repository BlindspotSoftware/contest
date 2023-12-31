package secureboot

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/linuxboot/contest/pkg/event"
	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/xcontext"
)

// events that we may emit during the plugin's lifecycle
const (
	EventStdout = event.Name("Stdout")
	EventStderr = event.Name("Stderr")
)

// Events defines the events that a TestStep is allow to emit. Emitting an event
// that is not registered here will cause the plugin to terminate with an error.
var Events = []event.Name{
	EventStdout,
	EventStderr,
}

type eventPayload struct {
	Msg string
}

func emitEvent(ctx xcontext.Context, name event.Name, payload interface{}, tgt *target.Target, ev testevent.Emitter) error {
	payloadData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cannot marshal payload for event '%s': %w", name, err)
	}

	msg := json.RawMessage(payloadData)
	data := testevent.Data{
		EventName: name,
		Target:    tgt,
		Payload:   &msg,
	}

	if err := ev.Emit(ctx, data); err != nil {
		return fmt.Errorf("cannot emit event EventCmdStart: %w", err)
	}

	return nil
}

// Function to format teststep information and append it to a string builder.
func writeEnrollKeysTestStep(step *TestStep, builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Input Parameter:\n")
		builder.WriteString("  Transport:\n")
		builder.WriteString(fmt.Sprintf("    Protocol: %s\n", step.Transport.Proto))
		builder.WriteString("    Options: \n")
		optionsJSON, err := json.MarshalIndent(step.Transport.Options, "", "    ")
		if err != nil {
			builder.WriteString(fmt.Sprintf("%v", step.Transport.Options))
		} else {
			builder.WriteString(string(optionsJSON))
		}
		builder.WriteString("\n")

		builder.WriteString("  Parameter:\n")
		builder.WriteString(fmt.Sprintf("    ToolPath: %s\n", step.Parameter.ToolPath))
		builder.WriteString(fmt.Sprintf("    Hierarchy: %s\n", step.Parameter.Hierarchy))
		builder.WriteString(fmt.Sprintf("    Append: %t\n", step.Parameter.Append))
		builder.WriteString(fmt.Sprintf("    KeyFilePath: %s\n", step.Parameter.KeyFile))
		builder.WriteString(fmt.Sprintf("    CertFilePath: %s\n", step.Parameter.CertFile))
		builder.WriteString(fmt.Sprintf("    SigningKeyFilePath: %s\n", step.Parameter.SigningKeyFile))
		builder.WriteString(fmt.Sprintf("    SigningCertFilePath: %s\n", step.Parameter.SigningCertFile))
		builder.WriteString("\n")

		builder.WriteString("  Expect:\n")
		builder.WriteString(fmt.Sprintf("    ShouldFail: %t\n", step.expect.ShouldFail))
		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(step.Options.Timeout)))
		builder.WriteString("\n\n")
	}
}

// Function to format teststep information and append it to a string builder.
func writeRotateKeysTestStep(step *TestStep, builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Input Parameter:\n")
		builder.WriteString("  Transport:\n")
		builder.WriteString(fmt.Sprintf("    Protocol: %s\n", step.Transport.Proto))
		builder.WriteString("    Options: \n")
		optionsJSON, err := json.MarshalIndent(step.Transport.Options, "", "    ")
		if err != nil {
			builder.WriteString(fmt.Sprintf("%v", step.Transport.Options))
		} else {
			builder.WriteString(string(optionsJSON))
		}
		builder.WriteString("\n")

		builder.WriteString("  Parameter:\n")
		builder.WriteString(fmt.Sprintf("    ToolPath: %s\n", step.Parameter.ToolPath))
		builder.WriteString(fmt.Sprintf("    Hierarchy: %s\n", step.Parameter.Hierarchy))
		builder.WriteString(fmt.Sprintf("    KeyFilePath: %s\n", step.Parameter.KeyFile))
		builder.WriteString(fmt.Sprintf("    CertFilePath: %s\n", step.Parameter.CertFile))
		builder.WriteString("\n")

		builder.WriteString("  Expect:\n")
		builder.WriteString(fmt.Sprintf("    ShouldFail: %t\n", step.expect.ShouldFail))
		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(step.Options.Timeout)))
		builder.WriteString("\n\n")
	}
}

// Function to format teststep information and append it to a string builder.
func writeResetTestStep(step *TestStep, builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Input Parameter:\n")
		builder.WriteString("  Transport:\n")
		builder.WriteString(fmt.Sprintf("    Protocol: %s\n", step.Transport.Proto))
		builder.WriteString("    Options: \n")
		optionsJSON, err := json.MarshalIndent(step.Transport.Options, "", "    ")
		if err != nil {
			builder.WriteString(fmt.Sprintf("%v", step.Transport.Options))
		} else {
			builder.WriteString(string(optionsJSON))
		}
		builder.WriteString("\n")

		builder.WriteString("  Parameter:\n")
		builder.WriteString(fmt.Sprintf("    ToolPath: %s\n", step.Parameter.ToolPath))
		builder.WriteString(fmt.Sprintf("    Hierarchy: %s\n", step.Parameter.Hierarchy))
		builder.WriteString(fmt.Sprintf("    SigningKeyFilePath: %s\n", step.Parameter.SigningKeyFile))
		builder.WriteString(fmt.Sprintf("    SigningCertFilePath: %s\n", step.Parameter.SigningCertFile))
		builder.WriteString(fmt.Sprintf("    CertFilePath: %s\n", step.Parameter.CertFile))
		builder.WriteString("\n")

		builder.WriteString("  Expect:\n")
		builder.WriteString(fmt.Sprintf("    ShouldFail: %t\n", step.expect.ShouldFail))
		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(step.Options.Timeout)))
		builder.WriteString("\n\n")
	}
}

// Function to format teststep information and append it to a string builder.
func writeCustomKeyTestStep(step *TestStep, builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Input Parameter:\n")
		builder.WriteString("  Transport:\n")
		builder.WriteString(fmt.Sprintf("    Protocol: %s\n", step.Transport.Proto))
		builder.WriteString("    Options: \n")
		optionsJSON, err := json.MarshalIndent(step.Transport.Options, "", "    ")
		if err != nil {
			builder.WriteString(fmt.Sprintf("%v", step.Transport.Options))
		} else {
			builder.WriteString(string(optionsJSON))
		}
		builder.WriteString("\n")

		builder.WriteString("  Parameter:\n")
		builder.WriteString(fmt.Sprintf("    ToolPath: %s\n", step.Parameter.ToolPath))
		builder.WriteString(fmt.Sprintf("    Hierarchy: %s\n", step.Parameter.Hierarchy))
		builder.WriteString(fmt.Sprintf("    CustomKeyFilePath: %s\n", step.Parameter.CustomKeyFile))
		builder.WriteString("\n")

		builder.WriteString("  Expect:\n")
		builder.WriteString(fmt.Sprintf("    ShouldFail: %t\n", step.expect.ShouldFail))
		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(step.Options.Timeout)))
		builder.WriteString("\n\n")
	}
}

// Function to format teststep information and append it to a string builder.
func writeStatusTestStep(step *TestStep, builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Input Parameter:\n")
		builder.WriteString("  Transport:\n")
		builder.WriteString(fmt.Sprintf("    Protocol: %s\n", step.Transport.Proto))
		builder.WriteString("    Options: \n")
		optionsJSON, err := json.MarshalIndent(step.Transport.Options, "", "    ")
		if err != nil {
			builder.WriteString(fmt.Sprintf("%v", step.Transport.Options))
		} else {
			builder.WriteString(string(optionsJSON))
		}
		builder.WriteString("\n")

		builder.WriteString("  Parameter:\n")
		builder.WriteString(fmt.Sprintf("    ToolPath: %s\n", step.Parameter.ToolPath))
		builder.WriteString("\n")

		builder.WriteString("  Expect:\n")
		builder.WriteString(fmt.Sprintf("      Secure Boot: %t\n", step.expect.SecureBoot))
		builder.WriteString(fmt.Sprintf("      Setup Mode: %t\n", step.expect.SetupMode))
		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(step.Options.Timeout)))
		builder.WriteString("\n\n")
	}
}

// Function to format command information and append it to a string builder.
func writeCommand(privileged bool, command string, args []string, builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Executing Command:\n")
		switch privileged {
		case false:
			builder.WriteString(fmt.Sprintf("%s %s", command, strings.Join(args, " ")))
		case true:
			builder.WriteString(fmt.Sprintf("sudo %s %s", command, strings.Join(args, " ")))

		}
		builder.WriteString("\n\n")
	}
}

// emitStderr emits the whole error message an returns the error
func emitStderr(ctx xcontext.Context, message string, tgt *target.Target, ev testevent.Emitter, err error) error {
	if err := emitEvent(ctx, EventStderr, eventPayload{Msg: message}, tgt, ev); err != nil {
		return fmt.Errorf("cannot emit event: %v", err)
	}

	return err
}

// emitStdout emits the whole message to Stdout
func emitStdout(ctx xcontext.Context, message string, tgt *target.Target, ev testevent.Emitter) error {
	if err := emitEvent(ctx, EventStdout, eventPayload{Msg: message}, tgt, ev); err != nil {
		return fmt.Errorf("cannot emit event: %v", err)
	}

	return nil
}
