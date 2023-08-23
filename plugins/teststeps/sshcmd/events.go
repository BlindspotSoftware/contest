package sshcmd

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
func writeTestStep(step *TestStep, builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Input Parameter:\n")
		builder.WriteString("  Bin:\n")
		builder.WriteString(fmt.Sprintf("    Executable: %s\n", step.Bin.Executable))
		builder.WriteString(fmt.Sprintf("    Args: %v\n", step.Bin.Args))
		builder.WriteString(fmt.Sprintf("    WorkingDir: %s\n", step.Bin.WorkingDir))
		builder.WriteString("\n")

		builder.WriteString("  Transport:\n")
		builder.WriteString(fmt.Sprintf("    Protocol: %s\n", step.Transport.Proto))
		builder.WriteString("    Options: \n")
		optionsJSON, err := json.MarshalIndent(step.Transport.Options, "", "    ")
		if err != nil {
			builder.WriteString(fmt.Sprintf("%v\n", step.Transport.Options))
		} else {
			builder.WriteString(string(optionsJSON))
		}
		builder.WriteString("\n")

		builder.WriteString("  Options:\n")
		builder.WriteString(fmt.Sprintf("    Timeout: %s\n", time.Duration(step.Options.Timeout)))
		builder.WriteString("\n")

		builder.WriteString("Default Values:\n")
		builder.WriteString(fmt.Sprintf("  Timeout: %s\n", defaultTimeout))

		builder.WriteString("\n\n")
	}
}

// Function to format command information and append it to a string builder.
func writeCommand(command string, builders ...*strings.Builder) {
	for _, builder := range builders {
		builder.WriteString("Operation on DUT:\n")
		builder.WriteString(command)
		builder.WriteString("\n\n")

		builder.WriteString("Output:\n")
	}
}

// emitStderr emits the whole error message to Stderr and returns the error
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
