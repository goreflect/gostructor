package pipeline

type IMiddleware interface {
	ActionConfigure(*structContext) error
}
