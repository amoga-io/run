package pkg

import (
	"fmt"
)

// DependencyNode represents a node in the dependency graph
type DependencyNode struct {
	Name         string
	Dependencies []string
	Visited      bool
	InStack      bool
}

// DependencyGraph represents the package dependency graph
type DependencyGraph struct {
	Nodes map[string]*DependencyNode
}

// NewDependencyGraph creates a new dependency graph from the package registry
func NewDependencyGraph() *DependencyGraph {
	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
	}

	// Build graph from package registry
	for name, pkg := range PackageRegistry {
		graph.Nodes[name] = &DependencyNode{
			Name:         name,
			Dependencies: pkg.Dependencies,
			Visited:      false,
			InStack:      false,
		}
	}

	return graph
}

// DetectCircularDependencies detects circular dependencies in the graph
func (dg *DependencyGraph) DetectCircularDependencies() ([]string, error) {
	var cycles []string

	// Reset visited flags
	for _, node := range dg.Nodes {
		node.Visited = false
		node.InStack = false
	}

	// Check each node for cycles
	for _, node := range dg.Nodes {
		if !node.Visited {
			if cycle := dg.detectCycleDFS(node); len(cycle) > 0 {
				cycles = append(cycles, cycle...)
			}
		}
	}

	if len(cycles) > 0 {
		return cycles, fmt.Errorf("circular dependencies detected: %v", cycles)
	}

	return nil, nil
}

// detectCycleDFS performs depth-first search to detect cycles
func (dg *DependencyGraph) detectCycleDFS(node *DependencyNode) []string {
	node.Visited = true
	node.InStack = true

	for _, depName := range node.Dependencies {
		// Skip system dependencies (not in our registry)
		if _, exists := dg.Nodes[depName]; !exists {
			continue
		}

		depNode := dg.Nodes[depName]

		if !depNode.Visited {
			if cycle := dg.detectCycleDFS(depNode); len(cycle) > 0 {
				return append([]string{node.Name}, cycle...)
			}
		} else if depNode.InStack {
			// Found a cycle
			return []string{node.Name, depName}
		}
	}

	node.InStack = false
	return nil
}

// GetInstallationOrder returns the order in which packages should be installed
func (dg *DependencyGraph) GetInstallationOrder(packages []string) ([]string, error) {
	// Validate package names
	for _, pkg := range packages {
		if err := ValidatePackageName(pkg); err != nil {
			return nil, fmt.Errorf("invalid package name %s: %w", pkg, err)
		}
	}

	// Build dependency graph for requested packages
	subGraph := dg.buildSubGraph(packages)

	// Detect cycles
	if cycles, err := subGraph.DetectCircularDependencies(); err != nil {
		return nil, err
	} else if len(cycles) > 0 {
		return nil, fmt.Errorf("circular dependencies in requested packages: %v", cycles)
	}

	// Perform topological sort
	return subGraph.topologicalSort(), nil
}

// buildSubGraph builds a subgraph containing only the requested packages and their dependencies
func (dg *DependencyGraph) buildSubGraph(packages []string) *DependencyGraph {
	subGraph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
	}

	// Add requested packages
	for _, pkg := range packages {
		if node, exists := dg.Nodes[pkg]; exists {
			subGraph.Nodes[pkg] = &DependencyNode{
				Name:         node.Name,
				Dependencies: node.Dependencies,
				Visited:      false,
				InStack:      false,
			}
		}
	}

	// Add dependencies recursively
	added := make(map[string]bool)
	for _, pkg := range packages {
		dg.addDependenciesRecursive(subGraph, pkg, added)
	}

	return subGraph
}

// addDependenciesRecursive adds dependencies recursively to the subgraph
func (dg *DependencyGraph) addDependenciesRecursive(subGraph *DependencyGraph, pkgName string, added map[string]bool) {
	if added[pkgName] {
		return
	}

	node, exists := dg.Nodes[pkgName]
	if !exists {
		return // System dependency, skip
	}

	added[pkgName] = true
	subGraph.Nodes[pkgName] = &DependencyNode{
		Name:         node.Name,
		Dependencies: node.Dependencies,
		Visited:      false,
		InStack:      false,
	}

	// Add dependencies
	for _, dep := range node.Dependencies {
		dg.addDependenciesRecursive(subGraph, dep, added)
	}
}

// topologicalSort performs topological sorting of the dependency graph
func (dg *DependencyGraph) topologicalSort() []string {
	var result []string
	visited := make(map[string]bool)

	// Reset visited flags
	for _, node := range dg.Nodes {
		node.Visited = false
	}

	// Visit all nodes
	for _, node := range dg.Nodes {
		if !visited[node.Name] {
			dg.topologicalSortDFS(node, visited, &result)
		}
	}

	// Reverse the result to get correct order
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// topologicalSortDFS performs depth-first search for topological sorting
func (dg *DependencyGraph) topologicalSortDFS(node *DependencyNode, visited map[string]bool, result *[]string) {
	visited[node.Name] = true

	for _, depName := range node.Dependencies {
		if depNode, exists := dg.Nodes[depName]; exists && !visited[depName] {
			dg.topologicalSortDFS(depNode, visited, result)
		}
	}

	*result = append(*result, node.Name)
}

// ValidateDependencies validates all dependencies in the registry
func ValidateDependencies() error {
	graph := NewDependencyGraph()

	// Check for circular dependencies
	if cycles, err := graph.DetectCircularDependencies(); err != nil {
		return fmt.Errorf("dependency validation failed: %w", err)
	} else if len(cycles) > 0 {
		return fmt.Errorf("circular dependencies found: %v", cycles)
	}

	// Check for missing dependencies
	for _, node := range graph.Nodes {
		for _, dep := range node.Dependencies {
			if _, exists := PackageRegistry[dep]; !exists {
				// This is a system dependency, which is fine
				continue
			}
		}
	}

	return nil
}

// GetDependencyTree returns a tree representation of dependencies for a package
func GetDependencyTree(packageName string) (map[string]interface{}, error) {
	if err := ValidatePackageName(packageName); err != nil {
		return nil, fmt.Errorf("invalid package name: %w", err)
	}

	pkg, exists := GetPackage(packageName)
	if !exists {
		return nil, fmt.Errorf("package '%s' not found", packageName)
	}

	tree := make(map[string]interface{})
	tree["name"] = pkg.Name
	tree["description"] = pkg.Description

	if len(pkg.Dependencies) > 0 {
		deps := make([]map[string]interface{}, 0)
		for _, dep := range pkg.Dependencies {
			depInfo := map[string]interface{}{
				"name": dep,
			}

			// Check if it's a package in our registry
			if depPkg, exists := GetPackage(dep); exists {
				depInfo["type"] = "package"
				depInfo["description"] = depPkg.Description
			} else {
				depInfo["type"] = "system"
			}

			deps = append(deps, depInfo)
		}
		tree["dependencies"] = deps
	}

	return tree, nil
}
