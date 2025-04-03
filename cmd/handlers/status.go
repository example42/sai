package handlers

// StatusHandler handles the status command
type StatusHandler struct {
	BaseHandler
}

// NewStatusHandler creates a new status handler
func NewStatusHandler() *StatusHandler {
	return &StatusHandler{
		BaseHandler: BaseHandler{
			Action: "status",
		},
	}
}

// Handle executes the status command
func (h *StatusHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
