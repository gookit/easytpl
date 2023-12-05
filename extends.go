package easytpl

type ExtendsTpl struct {
	Renderer
}

// NewExtends create a new extends template instance
func NewExtends() *ExtendsTpl {
	return &ExtendsTpl{}
}
