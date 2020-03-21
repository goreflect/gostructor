package middlewares

import "errors"

/*
ExecutorMiddlewaresByTagValue - entry point for all middlewares
TODO: increase by tagType
*/
func ExecutorMiddlewaresByTagValue(tagValue string, tagType string) error {
	if emptyTagValue(tagValue) {
		return errors.New("tagvalue can not be empty! ")
	}
	return nil
}
