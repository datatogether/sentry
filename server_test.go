package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	cfg = &config{}
}

func TestServerRoutes(t *testing.T) {
	cases := []struct {
		method, endpoint string
		expectBody       bool
		body             []byte
		resStatus        int
	}{
		{"GET", "/", false, nil, 200},
		{"GET", "/healthcheck", false, nil, 200},
		{"PUT", "/healthcheck", false, nil, 200},
		{"POST", "/healthcheck", false, nil, 200},
		{"DELETE", "/healthcheck", false, nil, 200},
	}

	client := &http.Client{}
	server := httptest.NewServer(NewServerRoutes())

	for i, c := range cases {
		req, err := http.NewRequest(c.method, server.URL+c.endpoint, bytes.NewReader(c.body))
		if err != nil {
			t.Errorf("case %d error creating request: %s", i, err.Error())
			continue
		}

		res, err := client.Do(req)
		if err != nil {
			t.Errorf("case %d error performing request: %s", i, err.Error())
			continue
		}

		if res.StatusCode != c.resStatus {
			t.Errorf("case %d: %s - %s status code mismatch. expected: %d, got: %d", i, c.method, c.endpoint, c.resStatus, res.StatusCode)
			continue
		}

		if c.expectBody {
			env := &struct {
				Meta       map[string]interface{}
				Data       interface{}
				Pagination map[string]interface{}
			}{}

			if err := json.NewDecoder(res.Body).Decode(env); err != nil {
				t.Errorf("case %d: %s - %s error unmarshaling json envelope: %s", i, c.method, c.endpoint, err.Error())
				continue
			}

			if env.Meta == nil {
				t.Errorf("case %d: %s - %s doesn't have a meta field", i, c.method, c.endpoint)
				continue
			}
			if env.Data == nil {
				t.Errorf("case %d: %s - %s doesn't have a data field", i, c.method, c.endpoint)
				continue
			}
		}
	}
}
