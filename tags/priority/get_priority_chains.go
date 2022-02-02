package priority

/*GetPriorityChains - getting mostly priority chains by priority choosed
 */
func GetPriorityChains(ast *AST, priority string) []string {
	for _, value := range ast.Nodes {
		if value.Priority.Literal == priority {
			result := []string{}
			for _, valueTag := range value.TagOrders {
				result = append(result, valueTag.Literal)
			}
			return result
		}
	}
	return []string{}
}
