# Decoy API

## Mock Server API

The Decoy mock server serves dynamic API responses based on OpenAPI v3 specifications. When a spec is loaded at startup, the server validates incoming requests against it and returns mock responses with template-rendered examples.

Start the server with:

```bash
decoy server rest start -f spec.yaml
```

### `/mock` Endpoint

The `/mock` endpoint accepts any HTTP method and path. It strips the `/mock` prefix and matches the remainder against the routes defined in the loaded OpenAPI v3 spec. The request is validated against the spec before a response is returned.

#### Response selection

The server selects a response from the OpenAPI spec in this order:

1. **Response** — If `?decoy-response=` is provided, that response key is used (e.g. `200`, `4xx`). Otherwise the first response is used.
2. **Content type** — If `?decoy-content-type=` is provided, that media type is used. Falls back to the `Content-Type` request header, then the first content type in the response.
3. **Example** — If `?decoy-example=` is provided, that example key is used. Otherwise the first example is used.

#### Template rendering

If the selected example value is a string and the content type is `text/*` or `decoy-parse` is not explicitly disabled, the string is rendered through the Decoy template engine with the following context:

```json
{
  "Request": {
    "Method": "GET",
    "QueryParams": {"page": "1"},
    "Path": "/users",
    "Header": {"Authorization": "Bearer ..."},
    "Body": null,
    "Content-Type": "application/json"
  },
  "Response": {
    "ContentType": "application/json",
    "StatusCode": 200,
    "Example": "userResponse"
  }
}
```

This allows dynamic responses. For example, given an OpenAPI example value of:

```
{"id": {{randomInt 1 1000}}, "name": "{{randomName}} {{randomLastName}}", "page": {{.Request.QueryParams.page}}}
```

The response would be rendered as something like:

```json
{"id": 742, "name": "Priya Bell", "page": 1}
```

#### External values (`decoy://`)

If an example uses `externalValue`, the server supports the `decoy://` URL scheme to reference a stored template:

```yaml
examples:
  userResponse:
    externalValue: decoy://user-template
```

This loads the template named `user-template` from the decoy storage and uses its content as the example value.

#### Query parameter reference

| Parameter | Description |
|---|---|
| `decoy-response` | Select a specific response by status code key (e.g. `200`, `4xx`, `default`) |
| `decoy-content-type` | Select a specific response content type (e.g. `application/json`) |
| `decoy-example` | Select a specific example by name |
| `decoy-parse` | Set to `false` to disable template rendering on the example value |

#### Template context reference

| Key | Type | Description |
|---|---|---|
| `.Request.Method` | `string` | HTTP method of the incoming request |
| `.Request.QueryParams` | `map[string]any` | Query parameters (single values are strings, repeated params are arrays) |
| `.Request.Path` | `string` | Raw request path |
| `.Request.Header` | `map[string]any` | Request headers (single values are strings, repeated headers are arrays) |
| `.Request.Body` | `any` | Parsed JSON body, or raw string for non-JSON content types |
| `.Request.Content-Type` | `string` | Content-Type header from the request |
| `.Response.ContentType` | `string` | The selected response content type |
| `.Response.StatusCode` | `int` | The selected response status code |
| `.Response.Example` | `string` | The selected example name |

#### Complete example

Given this OpenAPI spec:

```yaml
openapi: "3.0.0"
info:
  title: Users API
  version: "1.0"
paths:
  /users/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        200:
          description: User details
          content:
            application/json:
              examples:
                default:
                  value: |
                    {"id": {{.Request.Path}}, "name": "{{randomName}} {{randomLastName}}", "requestedId": {{index .Request.QueryParams "id"}}}
```

A request to `GET /mock/users/42` would return:

```json
{"id": "/users/42", "name": "Aaron Rivera", "requestedId": "42"}
```
