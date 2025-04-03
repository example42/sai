package handlers

// AskHandler handles the ask command
type AskHandler struct {
	BaseHandler
}

// NewAskHandler creates a new ask handler
func NewAskHandler() *AskHandler {
	return &AskHandler{
		BaseHandler: BaseHandler{
			Action: "Asking about",
		},
	}
}

// Handle executes the ask command
func (h *AskHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
