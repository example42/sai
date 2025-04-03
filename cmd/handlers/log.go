package handlers

// LogHandler handles the log command
type LogHandler struct {
	BaseHandler
}

// NewLogHandler creates a new log handler
func NewLogHandler() *LogHandler {
	return &LogHandler{
		BaseHandler: BaseHandler{
			Action: "Logging",
		},
	}
}

// Handle executes the log command
func (h *LogHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
