package runners

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type HttpRunnerConfig struct {
	Method          string            `json:"method"`
	Url             string            `json:"url"`
	QueryParameters map[string]string `json:"queryParameters"`
	Headers         map[string]string `json:"headers"`
	Body            json.RawMessage   `json:"body"`
}

func (c *HttpRunnerConfig) Validate() error {
	if c.Method == "" {
		c.Method = "GET"
	} else {
		c.Method = strings.ToUpper(c.Method)
	}

	if c.Url == "" {
		return fmt.Errorf("Invalid configuration: Missing URL")
	}

	return nil
}

type HttpRunnerOutput struct {
	Response *http.Response
}

func (o HttpRunnerOutput) String() string {
	rawBody, _ := io.ReadAll(o.Response.Body)
	defer o.Response.Body.Close()
	if rawBody != nil {
		return fmt.Sprintf("Status Code: %d, Body: %s", o.Response.StatusCode, string(rawBody))
	} else {
		return fmt.Sprintf("Status Code: %d", o.Response.StatusCode)
	}
}

type HttpRunner struct {
	Client http.RoundTripper
}

func NewHttpRunner() Runner {
	return &runnerWrapper[*HttpRunnerConfig, *HttpRunnerOutput]{
		Runner: HttpRunner{
			Client: http.DefaultTransport,
		},
		ConfigDeserializer: func(config string) (*HttpRunnerConfig, error) {
			var c HttpRunnerConfig
			err := json.Unmarshal([]byte(config), &c)
			return &c, err
		},
	}
}

func (r HttpRunner) Run(ctx context.Context, config *HttpRunnerConfig) (*HttpRunnerOutput, error) {
	request, err := http.NewRequestWithContext(ctx, config.Method, config.Url, bytes.NewReader(config.Body))
	if err != nil {
		return nil, err
	}

	for key, val := range config.Headers {
		request.Header.Add(key, val)
	}

	res, err := r.Client.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	return &HttpRunnerOutput{res}, nil
}
