package handlers

// CheckHandler handles the check command
type CheckHandler struct {
	BaseHandler
}

// NewCheckHandler creates a new check handler
func NewCheckHandler() *CheckHandler {
	return &CheckHandler{
		BaseHandler: BaseHandler{
			Action: "Checking",
		},
	}
}

// Handle executes the check command
func (h *CheckHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
