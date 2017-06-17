package net

type HandlerResult struct {
	Errors   []error
	Results  ResultList
	Metadata map[string]interface{}
	Redirect string
	NoRender bool
}
