package hwaas

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/linuxboot/contest/pkg/xcontext"
)

func (ts *TestStep) mountImage(ctx xcontext.Context, outputBuf *strings.Builder) error {
	hashSum, err := calcSHA256(ts.Parameter.Image)
	if err != nil {
		return err
	}

	if err := ts.checkMountImage(ctx, hashSum); err != nil {
		if err := ts.postMountImage(ctx); err != nil {
			return fmt.Errorf("failed to post image to api: %v", err)
		}
	}

	plugged, err := ts.checkUSBPlug(ctx)
	if err != nil {
		return fmt.Errorf("failed to check usb plug state: %v", err)
	}

	if plugged {
		if err := ts.plugUSB(ctx, unplug); err != nil {
			return fmt.Errorf("failed to unplug the usb device: %v", err)
		}
	}

	if err := ts.configureUSB(ctx, fmt.Sprintf("%x", hashSum)); err != nil {
		return fmt.Errorf("failed to configure usb device: %v", err)
	}

	if err := ts.plugUSB(ctx, plug); err != nil {
		return fmt.Errorf("failed to plug the usb device: %v", err)
	}

	outputBuf.WriteString("Image was mounted successfully.\n")

	return nil
}

func (ts *TestStep) checkMountImage(ctx xcontext.Context, hashSum []byte) error {
	endpoint := fmt.Sprintf("%s%s/images/%x", ts.Parameter.Host, ts.Parameter.Version, hashSum)

	resp, err := HTTPRequest(ctx, http.MethodGet, endpoint, bytes.NewBuffer(nil))
	if err != nil {
		return fmt.Errorf("failed to do HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not extract response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("image could not be checked. Statuscode: %d, Response Body: %v", resp.StatusCode, string(body))
	}

	return nil
}

func (ts *TestStep) postMountImage(ctx xcontext.Context) error {
	endpoint := fmt.Sprintf("%s%s/images/", ts.Parameter.Host, ts.Parameter.Version)

	file, err := os.Open(ts.Parameter.Image)
	if err != nil {
		return fmt.Errorf("failed to open the image at the provided path: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	form, err := writer.CreateFormFile("file", filepath.Base(ts.Parameter.Image))
	if err != nil {
		return fmt.Errorf("failed to create the form-data header: %v", err)
	}

	if _, err := io.Copy(form, file); err != nil {
		return fmt.Errorf("failed to copy file into form writer: %v", err)
	}

	writer.Close()

	// create the http request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create the http request: %v", err)
	}
	// add the file to the header
	req.Header.Add("Content-Type", writer.FormDataContentType())

	// execute the http request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do the http request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not extract response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload binary. Statuscode: %d, Response Body: %v", resp.StatusCode, string(respBody))
	}

	return nil
}

func (ts *TestStep) checkUSBPlug(ctx xcontext.Context) (bool, error) {
	endpoint := fmt.Sprintf("%s%s/contexts/%s/machines/%s/usb/plug",
		ts.Parameter.Host, ts.Parameter.Version, ts.Parameter.ContextID, ts.Parameter.MachineID)

	resp, err := HTTPRequest(ctx, http.MethodGet, endpoint, bytes.NewBuffer(nil))
	if err != nil {
		return false, fmt.Errorf("failed to do HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("could not extract response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("usb plug could not be checked. Statuscode: %d, Response Body: %v", resp.StatusCode, string(body))
	}

	switch string(body) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("failed to parse the response body: '%s'", string(body))
	}
}

const (
	plug   = true
	unplug = false
)

func (ts *TestStep) plugUSB(ctx xcontext.Context, plug bool) error {
	endpoint := fmt.Sprintf("%s%s/contexts/%s/machines/%s/usb/plug",
		ts.Parameter.Host, ts.Parameter.Version, ts.Parameter.ContextID, ts.Parameter.MachineID)

	var httpMethod string

	if plug {
		httpMethod = http.MethodPut
	} else {
		httpMethod = http.MethodDelete
	}

	resp, err := HTTPRequest(ctx, httpMethod, endpoint, bytes.NewBuffer(nil))
	if err != nil {
		return fmt.Errorf("failed to do HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not extract response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("usb device could not be plugged/unplugged. Statuscode: %d, Response Body: %v", resp.StatusCode, string(body))
	}

	return nil
}

type fileHashes []struct {
	FileHashes []string `json:"fileHashes"`
}

func (ts *TestStep) configureUSB(ctx xcontext.Context, hash string) error {
	endpoint := fmt.Sprintf("%s%s/contexts/%s/machines/%s/usb/functions",
		ts.Parameter.Host, ts.Parameter.Version, ts.Parameter.ContextID, ts.Parameter.MachineID)

	fileHashes := fileHashes{
		{
			FileHashes: []string{hash},
		},
	}

	imageHashBody, err := json.Marshal(fileHashes)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	resp, err := HTTPRequest(ctx, http.MethodPut, endpoint, bytes.NewBuffer(imageHashBody))
	if err != nil {
		return fmt.Errorf("failed to do HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not extract response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("usb device could not be configured. Statuscode: %d, Response Body: %v", resp.StatusCode, string(body))
	}

	return nil
}

func calcSHA256(path string) ([]byte, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("image not found: %v", err)
	}

	hash := sha256.New()
	hash.Write(file)

	return hash.Sum(nil), nil
}
