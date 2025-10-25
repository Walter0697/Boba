package installer

import (
	"fmt"
	"boba/internal/parser"
)

// DependencyResolver handles dependency resolution for tools and environments
type DependencyResolver struct{}

// NewDependencyResolver creates a new dependency resolver
func NewDependencyResolver() *DependencyResolver {
	return &DependencyResolver{}
}

// ResolveToolDependencies resolves tool dependencies and returns them in installation order
func (dr *DependencyResolver) ResolveToolDependencies(tools []parser.Tool) ([]parser.Tool, error) {
	// Create a map for quick lookup
	toolMap := make(map[string]parser.Tool)
	for _, tool := range tools {
		toolMap[tool.Name] = tool
	}
	
	// Track visited and visiting nodes for cycle detection
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	var result []parser.Tool
	
	// Depth-first search to resolve dependencies
	var visit func(toolName string) error
	visit = func(toolName string) error {
		if visited[toolName] {
			return nil // Already processed
		}
		
		if visiting[toolName] {
			return fmt.Errorf("circular dependency detected involving tool: %s", toolName)
		}
		
		tool, exists := toolMap[toolName]
		if !exists {
			return fmt.Errorf("dependency not found: %s", toolName)
		}
		
		visiting[toolName] = true
		
		// Visit all dependencies first
		for _, dep := range tool.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}
		
		visiting[toolName] = false
		visited[toolName] = true
		result = append(result, tool)
		
		return nil
	}
	
	// Process all tools
	for _, tool := range tools {
		if err := visit(tool.Name); err != nil {
			return nil, err
		}
	}
	
	return result, nil
}

// ResolveEnvironmentDependencies resolves environment dependencies and returns them in application order
func (dr *DependencyResolver) ResolveEnvironmentDependencies(environments []parser.Environment) ([]parser.Environment, error) {
	// Create a map for quick lookup
	envMap := make(map[string]parser.Environment)
	for _, env := range environments {
		envMap[env.Name] = env
	}
	
	// Track visited and visiting nodes for cycle detection
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	var result []parser.Environment
	
	// Depth-first search to resolve dependencies
	var visit func(envName string) error
	visit = func(envName string) error {
		if visited[envName] {
			return nil // Already processed
		}
		
		if visiting[envName] {
			return fmt.Errorf("circular dependency detected involving environment: %s", envName)
		}
		
		env, exists := envMap[envName]
		if !exists {
			return fmt.Errorf("environment dependency not found: %s", envName)
		}
		
		visiting[envName] = true
		
		// Visit all dependencies first
		for _, dep := range env.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}
		
		visiting[envName] = false
		visited[envName] = true
		result = append(result, env)
		
		return nil
	}
	
	// Process all environments
	for _, env := range environments {
		if err := visit(env.Name); err != nil {
			return nil, err
		}
	}
	
	return result, nil
}

// ValidateToolDependencies checks if all tool dependencies are available
func (dr *DependencyResolver) ValidateToolDependencies(tools []parser.Tool) error {
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}
	
	for _, tool := range tools {
		for _, dep := range tool.Dependencies {
			if !toolNames[dep] {
				return fmt.Errorf("tool %s depends on %s, but %s is not available", tool.Name, dep, dep)
			}
		}
	}
	
	return nil
}

// ValidateEnvironmentDependencies checks if all environment dependencies are available
func (dr *DependencyResolver) ValidateEnvironmentDependencies(environments []parser.Environment) error {
	envNames := make(map[string]bool)
	for _, env := range environments {
		envNames[env.Name] = true
	}
	
	for _, env := range environments {
		for _, dep := range env.Dependencies {
			if !envNames[dep] {
				return fmt.Errorf("environment %s depends on %s, but %s is not available", env.Name, dep, dep)
			}
		}
	}
	
	return nil
}

// GetInstallationOrder returns the correct installation order for tools and environments
func (dr *DependencyResolver) GetInstallationOrder(tools []parser.Tool, environments []parser.Environment) ([]parser.Tool, []parser.Environment, error) {
	// Resolve tool dependencies
	orderedTools, err := dr.ResolveToolDependencies(tools)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve tool dependencies: %w", err)
	}
	
	// Resolve environment dependencies
	orderedEnvironments, err := dr.ResolveEnvironmentDependencies(environments)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve environment dependencies: %w", err)
	}
	
	return orderedTools, orderedEnvironments, nil
}