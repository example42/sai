package handlers

// ListHandler handles the list command
type ListHandler struct {
	BaseHandler
}

// NewListHandler creates a new list handler
func NewListHandler() *ListHandler {
	return &ListHandler{
		BaseHandler: BaseHandler{
			Action: "list",
		},
	}
}

// Handle executes the list command
func (h *ListHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
