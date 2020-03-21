package pipeline

import "github.com/goreflect/gostructor/tags"

/*
TagInside - structure for getting information from tag valu and simplify code
*/
type TagInside struct {
	TokenType int
	Value     string
}

/*
getInfoFromTag - get information from tagvalue
if inside tag not specific information for getting, return error else return tagInside with value
*/
func getInfoFromTag(tagValue, tagType string) (TagInside, error) {
	switch tagType {
	case tags.TagHocon:
		return getInforFromHoconTagValue(tagValue)
	default:
		return TagInside{
			TokenType: 0,
			Value:     tagValue,
		}, nil
	}
}

func getInforFromHoconTagValue(tagValue string) (TagInside, error) {
	return TagInside{
		TokenType: 0,
		Value:     tagValue,
	}, nil
}
