package handlers

// MonitorHandler handles the monitor command
type MonitorHandler struct {
	BaseHandler
}

// NewMonitorHandler creates a new monitor handler
func NewMonitorHandler() *MonitorHandler {
	return &MonitorHandler{
		BaseHandler: BaseHandler{
			Action: "Monitoring",
		},
	}
}

// Handle executes the monitor command
func (h *MonitorHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
