package cmd

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/test"
	"github.com/linuxboot/contest/pkg/xcontext"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/options"
	"github.com/linuxboot/contest/plugins/teststeps/abstraction/transport"
)

const (
	ssh   = "ssh"
	local = "local"
)

type TargetRunner struct {
	ts *TestStep
	ev testevent.Emitter
}

func NewTargetRunner(ts *TestStep, ev testevent.Emitter) *TargetRunner {
	return &TargetRunner{
		ts: ts,
		ev: ev,
	}
}

func (r *TargetRunner) Run(ctx xcontext.Context, target *target.Target) error {
	var outputBuf strings.Builder

	ctx, cancel := options.NewOptions(ctx, defaultTimeout, r.ts.options.Timeout)
	defer cancel()

	pe := test.NewParamExpander(target)

	r.ts.writeTestStep(&outputBuf)

	transportProto, err := transport.NewTransport(r.ts.transport.Proto, []string{ssh, local}, r.ts.transport.Options, pe)
	if err != nil {
		err := fmt.Errorf("failed to create transport: %w", err)
		outputBuf.WriteString(fmt.Sprintf("%v", err))

		return emitStderr(ctx, outputBuf.String(), target, r.ev, err)
	}

	if err := r.ts.runCMD(ctx, &outputBuf, transportProto); err != nil {
		outputBuf.WriteString(fmt.Sprintf("%v\n", err))

		return emitStderr(ctx, outputBuf.String(), target, r.ev, err)
	}

	return emitStdout(ctx, outputBuf.String(), target, r.ev)
}

func (ts *TestStep) runCMD(ctx xcontext.Context, outputBuf *strings.Builder, transport transport.Transport,
) error {
	proc, err := transport.NewProcess(ctx, ts.Executable, ts.Args, ts.WorkingDir)
	if err != nil {
		err := fmt.Errorf("Failed to create proc: %w", err)
		outputBuf.WriteString(fmt.Sprintf("%v\n", err))

		return err
	}

	writeCommand(proc.String(), outputBuf)

	stdoutPipe, err := proc.StdoutPipe()
	if err != nil {
		err := fmt.Errorf("failed to pipe stdout: %v", err)
		outputBuf.WriteString(fmt.Sprintf("%v\n", err))

		return err
	}

	stderrPipe, err := proc.StderrPipe()
	if err != nil {
		err := fmt.Errorf("failed to pipe stderr: %v", err)
		outputBuf.WriteString(fmt.Sprintf("%v\n", err))

		return err
	}

	// try to start the process, if that succeeds then the outcome is the result of
	// waiting on the process for its result; this way there's a semantic difference
	// between "an error occured while launching" and "this was the outcome of the execution"
	outcome := proc.Start(ctx)
	if outcome == nil {
		outcome = proc.Wait(ctx)
	}

	stdout, stderr := getOutputFromReader(stdoutPipe, stderrPipe, outputBuf)

	outputBuf.WriteString(fmt.Sprintf("Command Stdout:\n%s\n", string(stdout)))
	outputBuf.WriteString(fmt.Sprintf("Command Stderr:\n%s\n", string(stderr)))

	if ts.ReportOnly {
		return nil
	}

	if outcome != nil {
		return fmt.Errorf("Error executing command: %v.\n", outcome)
	}

	if err := ts.parseOutput(outputBuf, stdout); err != nil {
		return err
	}

	return nil
}

// getOutputFromReader reads data from the provided io.Reader instances
// representing stdout and stderr, and returns the collected output as byte slices.
func getOutputFromReader(stdout, stderr io.Reader, outputBuf *strings.Builder) ([]byte, []byte) {
	// Read from the stdout and stderr pipe readers
	stdoutBuffer, err := readBuffer(stdout)
	if err != nil {
		outputBuf.WriteString(fmt.Sprintf("Failed to read from Stdout buffer: %v\n", err))
	}

	stderrBuffer, err := readBuffer(stderr)
	if err != nil {
		outputBuf.WriteString(fmt.Sprintf("Failed to read from Stderr buffer: %v\n", err))
	}

	return stdoutBuffer, stderrBuffer
}

// readBuffer reads data from the provided io.Reader and returns it as a byte slice.
// It dynamically accumulates the data using a bytes.Buffer.
func readBuffer(r io.Reader) ([]byte, error) {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, r)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (ts *TestStep) parseOutput(outputBuf *strings.Builder, stdout []byte) error {
	var errorString string

	for index, expect := range ts.Expect {
		re, err := regexp.Compile(expect.Regex)
		if err != nil {
			errorString += fmt.Sprintf("Failed to parse the regex for 'Expect%d': %v", index+1, err)
		}

		matches := re.FindAll(stdout, -1)
		if len(matches) > 0 {
			outputBuf.WriteString(fmt.Sprintf("Found the expected string for 'Expect%d' in Stdout: '%s'\n", index+1, expect))
		} else {
			errorString += fmt.Sprintf("Could not find the expected string '%s' for 'Expect%d' in Stdout.\n", expect, index+1)
		}
	}

	if errorString != "" {
		return fmt.Errorf("%s", errorString)
	}

	return nil
}
