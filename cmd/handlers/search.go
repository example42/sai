package handlers

// SearchHandler handles the search command
type SearchHandler struct {
	BaseHandler
}

// NewSearchHandler creates a new search handler
func NewSearchHandler() *SearchHandler {
	return &SearchHandler{
		BaseHandler: BaseHandler{
			Action: "search",
		},
	}
}

// Handle executes the search command
func (h *SearchHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
