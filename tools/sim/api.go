package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"mime/multipart"
	"net/http"
	"time"
)

// client talks to the API as one user acting for one organisation.
type client struct {
	base  string // API_URL
	host  string // <slug>.localhost: the organisation the API resolves (X-Forwarded-Host)
	token string
	pin   string
}

var httpc = &http.Client{Timeout: 60 * time.Second}

// apiError is a refusal, kept so API frauds can record the status and message.
type apiError struct {
	Status int
	Msg    string
}

func (e *apiError) Error() string { return fmt.Sprintf("%d %s", e.Status, e.Msg) }

// do sends one request; body is JSON, or a *multipart.Writer's buffer with its content type.
func (c *client) do(method, path string, body io.Reader, contentType string, sign bool, out any) error {
	req, err := http.NewRequest(method, c.base+path, body)
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.host != "" {
		req.Header.Set("X-Forwarded-Host", c.host)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if sign {
		req.Header.Set("X-Signing-PIN", c.pin)
	}
	// The API trusts this as the visitor's address (no proxy in dev), so a fresh one
	// per request keeps the sim under the per-address sign-up and verify limits.
	req.Header.Set("X-Forwarded-For", fmt.Sprintf("10.%d.%d.%d", rand.IntN(256), rand.IntN(256), rand.IntN(256)))
	res, err := httpc.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	// Every answer is {status, message, data}.
	var env struct {
		Message string
		Data    json.RawMessage
	}
	_ = json.Unmarshal(raw, &env)
	if res.StatusCode >= 300 {
		return &apiError{res.StatusCode, env.Message}
	}
	if out != nil && len(env.Data) > 0 && string(env.Data) != "null" {
		return json.Unmarshal(env.Data, out)
	}
	return nil
}

func (c *client) json(method, path string, in any, sign bool, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	return c.do(method, path, body, "application/json", sign, out)
}

// form sends fields plus one PDF as multipart.
func (c *client) form(path string, fields map[string]string, file []byte, sign bool, out any) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			return err
		}
	}
	part, err := w.CreateFormFile("file", "document.pdf")
	if err != nil {
		return err
	}
	part.Write(file)
	w.Close()
	return c.do(http.MethodPost, path, &buf, w.FormDataContentType(), sign, out)
}

func login(base, host, email, password, pin string) (*client, error) {
	c := &client{base: base, host: host, pin: pin}
	var res struct{ Token string }
	if err := c.json(http.MethodPost, "/auth/login", map[string]string{"email": email, "password": password}, false, &res); err != nil {
		return nil, fmt.Errorf("sign in %s: %w", email, err)
	}
	c.token = res.Token
	return c, nil
}

// A unique PDF per document: the content is hashed and anchored, and every certificate gets its own.
func pdf(label string) []byte {
	return []byte("%PDF-1.4\n% " + label + "\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n")
}
