package handlers

// TestHandler handles the test command
type TestHandler struct {
	BaseHandler
}

// NewTestHandler creates a new test handler
func NewTestHandler() *TestHandler {
	return &TestHandler{
		BaseHandler: BaseHandler{
			Action: "Testing",
		},
	}
}

// Handle executes the test command
func (h *TestHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
