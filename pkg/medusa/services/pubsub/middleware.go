package pubsub

// Middleware is a function that wraps a MessageHandler
type Middleware func(MessageHandler) MessageHandler

// MiddlewareChain applies multiple middleware in order
type MiddlewareChain []Middleware

// Apply applies middleware chain to a handler
func (mc MiddlewareChain) Apply(handler MessageHandler) MessageHandler {
	for i := len(mc) - 1; i >= 0; i-- {
		handler = mc[i](handler)
	}
	return handler
}
