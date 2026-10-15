package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/aaron70/decoy/internal/services"
	"github.com/aaron70/decoy/pkg/decoy"
	"github.com/aaron70/goaty/errors"
	"github.com/aaron70/goaty/validations"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
)

func decoyResponse(w http.ResponseWriter, status int, message string, args ...any) {
	w.Header().Add("Content-Type", "application/json")
	res := struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}{
		Status:  status,
		Message: fmt.Sprintf(message, args...),
	}
	bytes, _ := json.MarshalIndent(res, "", "  ")
	w.WriteHeader(999)
	w.Write(bytes)
}

func (s RestServer) mockHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			decoyResponse(w, http.StatusInternalServerError, "The endpoint expects a well formed and complete OpenAPI v3 Specification. If the specification is missing something and the endpoint is trying to access it, the endpoint will panic. Please read carefully your server spec. Endpoint has panic with: %v", r)
		}
	}()

	if s.specRouter == nil {
		decoyResponse(w, http.StatusNoContent, "No OpenAPI spec was given, /mock endpoint is disabled")
		return
	}

	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/mock")
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}

	var err error
	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, err = io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			decoyResponse(w, http.StatusBadRequest, "Couldn't read request body: %v", err)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	selectedResponse := r.URL.Query().Get("decoy-response")
	selectedContentType := r.URL.Query().Get("decoy-content-type")
	if validations.StrIsBlank(selectedContentType) {
		selectedContentType = r.Header.Get("Content-Type")
	}
	selectedExample := r.URL.Query().Get("decoy-example")
	decoyParse := r.URL.Query().Get("decoy-parse")
	shouldParse := false
	if validations.StrIsBlank(decoyParse) || strings.HasPrefix(selectedContentType, "text/") {
		shouldParse = true
	}

	route, pathParams, err := s.specRouter.FindRoute(r)
	if err != nil {
		decoyResponse(w, http.StatusNotFound, "Unknown endpoint: %s", err)
		return
	}

	input := &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
	}
	if err := openapi3filter.ValidateRequest(r.Context(), input); err != nil {
		decoyResponse(w, http.StatusBadRequest, "Invalid request: %s", err)
		return
	}

	op := route.Operation

	responses := op.Responses.Map()

	var response *openapi3.ResponseRef
	selectedResponse, response = getOrAny(responses, selectedResponse)
	if response == nil {
		if validations.StrIsBlank(selectedResponse) {
			decoyResponse(w, http.StatusNoContent, "No responses for path %q in the given spec", r.URL.Path)
		} else {
			decoyResponse(w, http.StatusNotFound, "Response %q not found for path %s in the given spec", selectedResponse, r.URL.Path)
		}
		return
	}

	var contentType *openapi3.MediaType
	selectedContentType, contentType = getOrAny(response.Value.Content, selectedContentType)
	if contentType == nil {
		if validations.StrIsBlank(selectedResponse) {
			decoyResponse(w, http.StatusNoContent, "No content types for response %q in path %q", selectedResponse, r.URL.Path)
		} else {
			decoyResponse(w, http.StatusNotFound, "Content type %q not found in response %q for path %q", selectedContentType, selectedResponse, r.URL.Path)
		}
		return
	}

	var example *openapi3.ExampleRef
	selectedExample, example = getOrAny(contentType.Examples, selectedExample)
	if example == nil {
		if validations.StrIsBlank(selectedResponse) {
			decoyResponse(w, http.StatusNoContent, "No examples for content type %q in response %q in path %q", selectedContentType, selectedResponse, r.URL.Path)
		} else {
			decoyResponse(w, http.StatusNotFound, "Example %q not found for content type %q in response %q for path %q", selectedExample, selectedContentType, selectedResponse, r.URL.Path)
		}
		return
	}

	extraTemplates := make(map[string]string)
	data := make(map[string]any)
	value := example.Value.Value
	if !validations.StrIsBlank(example.Value.ExternalValue) {
		templateReference, err := resolveExternalValue(s.Decoy, example.Value.ExternalValue)
		if err != nil {
			decoyResponse(w, http.StatusNoContent, "Coudln't resolve the external value for example %s of content type %s from response %s in path %s: %s", selectedExample, selectedContentType, selectedResponse, r.URL.Path, err)
			return
		}
		value = templateReference.Tmpl
		extraTemplates = templateReference.ExtraTemplates
		data = templateReference.Data
	}

	statusCode := responseToStatusCode(selectedResponse)
	if shouldParse {
		requestContentType := r.Header.Get("Content-Type")
		var requestBody any
		if len(bodyBytes) > 0 {
			if strings.HasPrefix(requestContentType, "application/json") {
				if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
					requestBody = string(bodyBytes)
				}
			} else {
				requestBody = string(bodyBytes)
			}
		}

		maps.Insert(data, maps.All(map[string]any{
			"request": map[string]any{
				"method":       r.Method,
				"queryParams":  parseMapSliceString(r.URL.Query()),
				"path":         r.URL.Path,
				"pathParams":   pathParams,
				"header":       parseMapSliceString(r.Header),
				"body":         requestBody,
				"content-Type": requestContentType,
			},
			"response": map[string]any{
				"contentType": selectedContentType,
				"statusCode":  statusCode,
				"example":     selectedExample,
			},
		}))
		str, ok := value.(string)
		if !ok {
			decoyResponse(w, http.StatusInternalServerError, "couldn't parse template example: template example is not a string")
			return
		}

		tmpl, err := s.Decoy.Decoy.ParseTemplateString(str,
			decoy.WithData(data),
			decoy.WithExtraTemplates(extraTemplates),
		)
		if err != nil {
			decoyResponse(w, http.StatusInternalServerError, "couldn't parse template example: %s", err)
			return
		}
		w.WriteHeader(statusCode)
		w.Header().Set("Content-Type", selectedContentType)
		writeBody(w, selectedContentType, tmpl)
	} else {
		w.WriteHeader(statusCode)
		w.Header().Set("Content-Type", selectedContentType)
		writeBody(w, selectedContentType, value)
	}
}

func writeBody(w http.ResponseWriter, contentType string, body any) {
	if body == nil {
		return
	}

	switch contentType {
	case "text/plain":
		fmt.Fprintf(w, "%s", body.(string))
	case "application/json":
		writeJsonBody(w, body)
	default:
		fmt.Fprintf(w, "%+v", body)
	}
}

func writeJsonBody(w http.ResponseWriter, body any) {
	str, ok := body.(string)
	if ok {
		w.Write([]byte(str))
		return
	}

	bytes, ok := body.([]byte)
	if ok {
		w.Write(bytes)
		return
	}

	bytes, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		decoyResponse(w, 500, "Invalid body, couldn't deserialize the example. Please fix the example in the spec and try again. Error: %s", err)
		return
	}
	w.Write(bytes)
}

func parseMapSliceString(queryParams map[string][]string) map[string]any {
	res := map[string]any{}
	for key, value := range queryParams {
		if len(value) == 0 {
			continue
		} else if len(value) == 1 {
			res[key] = value[0]
		} else {
			res[key] = value
		}
	}
	return res
}

func responseToStatusCode(response string) int {
	status, err := strconv.Atoi(response)
	if err != nil {
		switch strings.ToLower(response) {
		case "1xx":
			return 100
		case "2xx":
			return 200
		case "3xx":
			return 300
		case "4xx":
			return 400
		case "5xx":
			return 500
		default:
			return http.StatusMultiStatus
		}
	}
	return status
}

func resolveExternalValue(d *services.Decoy, rawUrl string) (decoy.TemplateURLReferenced, error) {
	url, err := url.Parse(rawUrl)
	if err != nil {
		return decoy.TemplateURLReferenced{}, errors.NewError(errors.ErrInvalidArgument, err, "Malformed external value: %s. Expected a valid URL format.", rawUrl)
	}

	switch url.Scheme {
	case "decoy":
		resolver := func(name string) (string, error) {
			tmpl, err := d.TemplateSvc.Get(name)
			if err != nil {
				return "", err
			}
			return tmpl.Tmpl, err
		}
		return decoy.Default.ResolveTemplateURL(resolver, rawUrl)
	default:
		return decoy.TemplateURLReferenced{}, errors.NewError(errors.ErrInvalidArgument, err, "Unsupported shceme %q, valid schemes are: decoy", url.Scheme)
	}
}

func getOrAny[T any](options map[string]T, key string) (string, T) {
	var value T
	if validations.StrIsBlank(key) {
		for key, value = range options {
			return key, value
		}
		return "", value
	}
	return key, options[key]
}
