package handlers

// BuildHandler handles the build command
type BuildHandler struct {
	BaseHandler
}

// NewBuildHandler creates a new build handler
func NewBuildHandler() *BuildHandler {
	return &BuildHandler{
		BaseHandler: BaseHandler{
			Action: "Building",
		},
	}
}

// Handle executes the build command
func (h *BuildHandler) Handle(software string, provider string) {
	h.BaseHandler.Handle(software, provider)
}
