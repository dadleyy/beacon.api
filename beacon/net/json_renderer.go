package net

import "time"
import "net/http"
import "encoding/json"

type JsonRenderer struct {
	version string
}

type jsonResponse struct {
	Status  string     `json:"status"`
	Meta    Metadata   `json:"meta"`
	Errors  []string   `json:"errors"`
	Results ResultList `json:"results"`
}

func (js *JsonRenderer) Render(response http.ResponseWriter, result HandlerResult) error {
	headers := response.Header()
	headers["Content-Type"] = []string{"application/json"}

	errors := make([]string, 0, len(result.Errors))
	meta := Metadata{"time": time.Now(), "version": js.version}

	for _, e := range result.Errors {
		errors = append(errors, e.Error())
	}

	for key, value := range result.Metadata {
		meta[key] = value
	}

	out := jsonResponse{"SUCCESS", meta, errors, result.Results}
	writer := json.NewEncoder(response)

	if ec := len(result.Errors); ec >= 1 {
		response.WriteHeader(http.StatusBadRequest)
		out.Status = "ERRORED"
		return writer.Encode(out)
	}

	return writer.Encode(out)
}
