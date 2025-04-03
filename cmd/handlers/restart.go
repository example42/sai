package handlers

// RestartHandler handles the restart command
type RestartHandler struct {
	BaseHandler
}

// NewRestartHandler creates a new restart handler
func NewRestartHandler() *RestartHandler {
	return &RestartHandler{
		BaseHandler: BaseHandler{
			Action: "restart",
		},
	}
}

// Handle executes the restart command
func (h *RestartHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
