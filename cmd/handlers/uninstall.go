package handlers

// UninstallHandler handles the uninstall command
type UninstallHandler struct {
	BaseHandler
}

// NewUninstallHandler creates a new uninstall handler
func NewUninstallHandler() *UninstallHandler {
	return &UninstallHandler{
		BaseHandler: BaseHandler{
			Action: "uninstall",
		},
	}
}

// Handle executes the uninstall command
func (h *UninstallHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
