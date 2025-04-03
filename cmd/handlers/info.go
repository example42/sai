package handlers

// InfoHandler handles the info command
type InfoHandler struct {
	BaseHandler
}

// NewInfoHandler creates a new info handler
func NewInfoHandler() *InfoHandler {
	return &InfoHandler{
		BaseHandler: BaseHandler{
			Action: "info",
		},
	}
}

// Handle executes the info command
func (h *InfoHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
