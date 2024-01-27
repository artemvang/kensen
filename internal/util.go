package kensen

func topologicalSort(graph map[string][]string) []string {
	inDegree := make(map[string]int)
	linearOrder := []string{}

	for node := range graph {
		inDegree[node] = 0
	}

	for _, adjacentNodes := range graph {
		for _, v := range adjacentNodes {
			inDegree[v]++
		}
	}

	next := []string{}
	for node, inDeg := range inDegree {
		if inDeg == 0 {
			next = append(next, node)
		}
	}

	for len(next) > 0 {
		current := next[0]
		next = next[1:]

		linearOrder = append(linearOrder, current)

		for _, neighbor := range graph[current] {
			inDegree[neighbor]--

			if inDegree[neighbor] == 0 {
				next = append(next, neighbor)
			}
		}
	}

	return linearOrder
}
