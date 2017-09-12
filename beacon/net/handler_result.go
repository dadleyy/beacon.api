package net

// HandlerResult public contract between routes and the server runtime for consistent rendering
type HandlerResult struct {
	Errors   []error
	Results  ResultList
	Metadata map[string]interface{}
	Redirect string
	NoRender bool
	Status   int
}
