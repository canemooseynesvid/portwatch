package envelope

import "fmt"

// Handler processes a single Envelope.
type Handler func(Envelope) error

// Router dispatches envelopes to registered handlers based on tag matching.
// If an envelope carries no matching tag, the default handler is used when set.
type Router struct {
	routes         map[string]Handler
	defaultHandler Handler
}

// NewRouter returns an empty Router.
func NewRouter() *Router {
	return &Router{routes: make(map[string]Handler)}
}

// Register associates a handler with the given tag.
// The last registration for a tag wins.
func (r *Router) Register(tag string, h Handler) {
	r.routes[tag] = h
}

// SetDefault registers a fallback handler used when no tag matches.
func (r *Router) SetDefault(h Handler) {
	r.defaultHandler = h
}

// Dispatch sends env to the first handler whose tag appears in the envelope.
// If multiple tags match, only the first registered match is invoked.
// Returns an error if no handler is found and no default is set.
func (r *Router) Dispatch(env Envelope) error {
	for tag, h := range r.routes {
		if env.HasTag(tag) {
			return h(env)
		}
	}
	if r.defaultHandler != nil {
		return r.defaultHandler(env)
	}
	return fmt.Errorf("envelope %s: no matching route and no default handler", env.ID[:8])
}
