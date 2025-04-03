package handlers

// StartHandler handles the start command
type StartHandler struct {
	BaseHandler
}

// NewStartHandler creates a new start handler
func NewStartHandler() *StartHandler {
	return &StartHandler{
		BaseHandler: BaseHandler{
			Action: "start",
		},
	}
}

// Handle executes the start command
func (h *StartHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
