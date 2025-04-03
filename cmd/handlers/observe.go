package handlers

// ObserveHandler handles the observe command
type ObserveHandler struct {
	BaseHandler
}

// NewObserveHandler creates a new observe handler
func NewObserveHandler() *ObserveHandler {
	return &ObserveHandler{
		BaseHandler: BaseHandler{
			Action: "Observing",
		},
	}
}

// Handle executes the observe command
func (h *ObserveHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
