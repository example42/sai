package handlers

// InstallHandler handles the install command
type InstallHandler struct {
	BaseHandler
}

// NewInstallHandler creates a new install handler
func NewInstallHandler() *InstallHandler {
	return &InstallHandler{
		BaseHandler: BaseHandler{
			Action: "install",
		},
	}
}

// Handle executes the install command
func (h *InstallHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
