package handlers

// TraceHandler handles the trace command
type TraceHandler struct {
	BaseHandler
}

// NewTraceHandler creates a new trace handler
func NewTraceHandler() *TraceHandler {
	return &TraceHandler{
		BaseHandler: BaseHandler{
			Action: "Tracing",
		},
	}
}

// Handle executes the trace command
func (h *TraceHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
