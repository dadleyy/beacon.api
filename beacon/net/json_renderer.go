package net

import "time"
import "net/http"
import "encoding/json"

// JSONRenderer exposes a `Renderer` interface for rendering `HandlerResult`s in json
type JSONRenderer struct {
	version string
}

type jsonResponse struct {
	Status  string     `json:"status"`
	Meta    Metadata   `json:"meta"`
	Errors  []string   `json:"errors"`
	Results ResultList `json:"results"`
}

// Render uses a response writer and a `HandlerResult` to serialize the result in a json-api like format
func (js *JSONRenderer) Render(response http.ResponseWriter, result HandlerResult) error {
	headers := response.Header()
	headers.Set("Content-Type", "application/json")

	errors := make([]string, 0, len(result.Errors))
	meta := Metadata{"time": time.Now(), "version": js.version}

	for _, e := range result.Errors {
		errors = append(errors, e.Error())
	}

	for key, value := range result.Metadata {
		meta[key] = value
	}

	out := jsonResponse{
		Status:  "SUCCESS",
		Meta:    meta,
		Errors:  errors,
		Results: result.Results,
	}

	writer := json.NewEncoder(response)

	if ec := len(result.Errors); ec >= 1 {
		response.WriteHeader(http.StatusBadRequest)
		out.Status = "ERRORED"
		return writer.Encode(out)
	}

	return writer.Encode(out)
}
