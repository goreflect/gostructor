package pipeline

/*
IMiddleware - interface which need implement for executing yourselft writed middlewares
*/
type IMiddleware interface {
	ActionConfigure(*structContext) error
}
