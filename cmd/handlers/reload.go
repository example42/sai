package handlers

// ReloadHandler handles the reload action
type ReloadHandler struct {
	BaseHandler
}

// NewReloadHandler creates a new ReloadHandler
func NewReloadHandler() *ReloadHandler {
	return &ReloadHandler{
		BaseHandler: BaseHandler{
			Action: "reload",
		},
	}
}

// Handle executes the reload action
func (h *ReloadHandler) Handle(software, provider string) {
	h.BaseHandler.Handle(software, provider)
}
