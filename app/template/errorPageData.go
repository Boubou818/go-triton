package template

// ErrorPageData contains information about an error.
type ErrorPageData struct {
	LocalizedTemplateData
	Error    error
	Message  string
	Expected bool
}
