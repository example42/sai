package handlers

// DisableHandler handles the disable command
type DisableHandler struct {
	BaseHandler
}

// NewDisableHandler creates a new disable handler
func NewDisableHandler() *DisableHandler {
	return &DisableHandler{
		BaseHandler: BaseHandler{
			Action: "disable",
		},
	}
}

// Handle executes the disable command
func (h *DisableHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
