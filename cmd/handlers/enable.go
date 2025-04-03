package handlers

// EnableHandler handles the enable command
type EnableHandler struct {
	BaseHandler
}

// NewEnableHandler creates a new enable handler
func NewEnableHandler() *EnableHandler {
	return &EnableHandler{
		BaseHandler: BaseHandler{
			Action: "enable",
		},
	}
}

// Handle executes the enable command
func (h *EnableHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
