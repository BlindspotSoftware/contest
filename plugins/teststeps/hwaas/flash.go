package hwaas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/linuxboot/contest/pkg/event/testevent"
	"github.com/linuxboot/contest/pkg/target"
	"github.com/linuxboot/contest/pkg/xcontext"
)

const (
	read       = "read"
	write      = "write"
	stateReady = "ready"
	stateBusy  = "busy"
	stateError = "error"
)

// flashCmds is a helper function to call into the different flash commands
func (r *TargetRunner) flashCmds(ctx xcontext.Context, stdoutMsg, stderrMsg *strings.Builder, target *target.Target, args []string) error {
	if len(args) >= 2 {

		switch args[0] {

		case "write":
			if err := r.ts.flashWrite(ctx, stdoutMsg, stderrMsg, args[1], target, r.ev); err != nil {
				return err
			}

			return nil

		case "read":
			if err := r.ts.flashRead(ctx, stdoutMsg, stderrMsg, args[1], target, r.ev); err != nil {
				return err
			}

			return nil

		default:
			return fmt.Errorf("Failed to execute the flash command. The argument '%s' is not valid. Possible values are 'read /path/to/binary' and 'write /path/to/binary'.", args)
		}
	} else {
		return fmt.Errorf("Failed to execute the power command. Args is not valid. Possible values are 'read /path/to/binary' and 'write /path/to/binary'.")
	}
}

// flashWrite executes the flash write command.
func (ts *TestStep) flashWrite(ctx xcontext.Context, stdoutMsg, stderrMsg *strings.Builder, arg string, target *target.Target, ev testevent.Emitter) error {
	if arg == "" {
		return fmt.Errorf("No file was set to flash target.")
	}

	if err := ts.resetDUT(ctx); err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s:%d/contexts/%s/machines/%s/auxiliaries/%s/api/flash",
		ts.Parameter.Host, ts.Parameter.Port, ts.Parameter.ContextID, ts.Parameter.MachineID, ts.Parameter.DeviceID)

	targetInfo, err := getTargetState(ctx, endpoint)
	if err != nil {
		return err
	}
	if targetInfo.State == "busy" {
		return fmt.Errorf("Flashing DUT with %s failed: DUT is currently busy.\n", arg)
	}

	if err := postImage(ctx, endpoint, arg); err != nil {
		return fmt.Errorf("Flashing DUT with %s failed: %v\n", arg, err)
	}

	if err := flashTarget(ctx, endpoint); err != nil {
		return fmt.Errorf("Flashing DUT with %s failed: %v\n", arg, err)
	}

	if err := waitTarget(ctx, endpoint); err != nil {
		return err
	}

	if err := ts.unresetDUT(ctx); err != nil {
		return err
	}

	time.Sleep(time.Second)

	stdoutMsg.WriteString("DUT is flashed successfully.\n")

	return nil
}

// flashRead executes the flash read command.
func (ts *TestStep) flashRead(ctx xcontext.Context, stdoutMsg, stderrMsg *strings.Builder, arg string, target *target.Target, ev testevent.Emitter) error {
	if arg == "" {
		return fmt.Errorf("No file was set to read from target.")
	}

	if err := ts.resetDUT(ctx); err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s:%d/contexts/%s/machines/%s/auxiliaries/%s/api/flash",
		ts.Parameter.Host, ts.Parameter.Port, ts.Parameter.ContextID, ts.Parameter.MachineID, ts.Parameter.DeviceID)

	targetInfo, err := getTargetState(ctx, endpoint)
	if err != nil {
		return err
	}
	if targetInfo.State == "busy" {
		return fmt.Errorf("Reading image from DUT into %s failed: DUT is currently busy.\n", arg)
	}

	err = readTarget(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("Reading image from DUT into %s failed: %v\n", arg, err)
	}

	if err := waitTarget(ctx, endpoint); err != nil {
		return err
	}

	if err := pullImage(ctx, endpoint, arg); err != nil {
		return err
	}

	if err := ts.unresetDUT(ctx); err != nil {
		return err
	}

	time.Sleep(time.Second)

	stdoutMsg.WriteString("DUT flash was read successfully.\n")

	return nil
}

// this struct is the response for GET /flash
type getFlash struct {
	State string `json:"state"` // possible values: "ready", "busy" or "error"
	Error string `json:"error"`
}

// getTargetState returns the flash state of the target.
// If an error occured, the field error is filled.
func getTargetState(ctx xcontext.Context, endpoint string) (getFlash, error) {
	resp, err := HTTPRequest(ctx, http.MethodGet, endpoint, bytes.NewBuffer(nil))
	if err != nil {
		return getFlash{}, fmt.Errorf("Failed to do HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return getFlash{}, fmt.Errorf("Could not extract response body: %v", err)
	}

	data := getFlash{}

	if err := json.Unmarshal(body, &data); err != nil {
		return getFlash{}, fmt.Errorf("Could not unmarshal response body: %v", err)
	}

	return data, nil
}

// pullImage downloads the binary from the target and stores it at 'filePath'.
func pullImage(ctx xcontext.Context, endpoint string, filePath string) error {
	endpoint = fmt.Sprintf("%s%s", endpoint, "/file")

	resp, err := HTTPRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("Failed to do HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to download binary. Statuscode: %d, Response Body: %v", resp.StatusCode, resp.Body)
	}

	// open/create file and copy the http response body into it
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("Failed to open/create file at the provided path '%s': %v", filePath, err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to copy binary to file: %v", err)
	}

	return nil
}

// postImage posts the binary to the target.
func postImage(ctx xcontext.Context, endpoint string, filePath string) error {
	// open the binary that shall be flashed
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Failed to open the file at the provided path: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	form, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("Failed to create the form-data header: %v", err)
	}

	if _, err := io.Copy(form, file); err != nil {
		return fmt.Errorf("Failed to copy file into form writer: %v", err)
	}

	writer.Close()

	// create the http request
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s%s", endpoint, "/file"), body)
	if err != nil {
		return fmt.Errorf("Failed to create the http request: %v", err)
	}
	// add the file to the header
	req.Header.Add("Content-Type", writer.FormDataContentType())

	// execute the http request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to do the http request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to upload binary. Statuscode: %d, Response Body: %v", resp.StatusCode, resp.Body)
	}

	return nil
}

type postFlash struct {
	Action string `json:"action"` // possible values: "read" or "write"
}

// readTarget reads the binary from the target into the flash buffer.
func readTarget(ctx xcontext.Context, endpoint string) error {
	postFlash := postFlash{
		Action: read,
	}

	flashBody, err := json.Marshal(postFlash)
	if err != nil {
		return fmt.Errorf("Failed to marshal body: %w", err)
	}

	resp, err := HTTPRequest(ctx, http.MethodPost, endpoint, bytes.NewBuffer(flashBody))
	if err != nil {
		return fmt.Errorf("Failed to do HTTP request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to read image from target. Statuscode: %d, Response Body: %v", resp.StatusCode, resp.Body)
	}

	return nil
}

// flashTarget flashes the target with the binary in the flash buffer.
func flashTarget(ctx xcontext.Context, endpoint string) error {
	postFlash := postFlash{
		Action: write,
	}

	flashBody, err := json.Marshal(postFlash)
	if err != nil {
		return fmt.Errorf("Failed to marshal body: %w", err)
	}

	resp, err := HTTPRequest(ctx, http.MethodPost, endpoint, bytes.NewBuffer(flashBody))
	if err != nil {
		return fmt.Errorf("Failed to do HTTP request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Failed to flash binary on target. Statuscode: %d, Response Body: %v", resp.StatusCode, resp.Body)
	}

	return nil
}

// waitTarget wait for the process that is running on the Flash endpoint, either its write or read.
func waitTarget(ctx xcontext.Context, endpoint string) error {
	timestamp := time.Now()

	for {
		targetInfo, err := getTargetState(ctx, endpoint)
		if err != nil {
			return err
		}
		if targetInfo.State == stateReady {
			break
		}
		if targetInfo.State == stateBusy {
			time.Sleep(time.Second)

			continue
		}
		if targetInfo.State == stateError {
			return fmt.Errorf("Error while flashing DUT: %s", targetInfo.Error)
		}
		if time.Since(timestamp) >= defaultTimeout {
			return fmt.Errorf("Flashing DUT failed: timeout")
		}
	}

	return nil
}

// HTTPRequest triggerers a http request and returns the response. The parameter that can be set are:
// method: can be every http method
// endpoint: api endpoint that shall be requested
// body: the body of the request
func HTTPRequest(ctx xcontext.Context, method string, endpoint string, body io.Reader) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
