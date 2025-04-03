package handlers

// StopHandler handles the stop command
type StopHandler struct {
	BaseHandler
}

// NewStopHandler creates a new stop handler
func NewStopHandler() *StopHandler {
	return &StopHandler{
		BaseHandler: BaseHandler{
			Action: "stop",
		},
	}
}

// Handle executes the stop command
func (h *StopHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
