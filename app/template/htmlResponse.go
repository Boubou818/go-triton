package template

import (
	"context"
	"net/http"
)

// HTMLResponse helps you create a HTTP response in HTML with MasterPageData.
type HTMLResponse struct {
	BaseResponse

	writer      http.ResponseWriter
	isCompleted bool
}

// NewHTMLResponse creates a new HTMLResponse.
func NewHTMLResponse(ctx context.Context, mgr *Manager, wr http.ResponseWriter) *HTMLResponse {
	return &HTMLResponse{
		BaseResponse: newBaseResponse(ctx, mgr),
		writer:       wr,
	}
}

// MustComplete finishes the response with the given MasterPageData, and panics if unexpected error happens.
func (h *HTMLResponse) MustComplete(d *MasterPageData) {
	h.checkCompletion()
	h.mgr.MustComplete(h.lang, d, h.writer)
}

// MustFail finishes the response with the given error object.
func (h *HTMLResponse) MustFail(err error) {
	h.checkCompletion()
	d := &ErrorPageData{Error: err}
	h.mgr.MustError(h.lang, d, h.writer)
}

func (h *HTMLResponse) checkCompletion() {
	if h.isCompleted {
		panic("Result has completed")
	}
	h.isCompleted = true
}
