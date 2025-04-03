package handlers

// UpdateHandler handles the update command
type UpdateHandler struct {
	BaseHandler
}

// NewUpdateHandler creates a new update handler
func NewUpdateHandler() *UpdateHandler {
	return &UpdateHandler{
		BaseHandler: BaseHandler{
			Action: "Updating",
		},
	}
}

// Handle executes the update command
func (h *UpdateHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
