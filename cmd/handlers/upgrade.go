package handlers

// UpgradeHandler handles the upgrade command
type UpgradeHandler struct {
	BaseHandler
}

// NewUpgradeHandler creates a new upgrade handler
func NewUpgradeHandler() *UpgradeHandler {
	return &UpgradeHandler{
		BaseHandler: BaseHandler{
			Action: "upgrade",
		},
	}
}

// Handle executes the upgrade command
func (h *UpgradeHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
