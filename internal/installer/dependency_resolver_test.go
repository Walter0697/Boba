package installer

import (
	"strings"
	"testing"
	"boba/internal/parser"
)

func TestResolveToolDependencies(t *testing.T) {
	resolver := NewDependencyResolver()
	
	// Test case 1: Simple dependency chain
	tools := []parser.Tool{
		{Name: "cli-a", Dependencies: []string{"aws-cli"}},
		{Name: "aws-cli", Dependencies: []string{}},
		{Name: "independent-tool", Dependencies: []string{}},
	}
	
	resolved, err := resolver.ResolveToolDependencies(tools)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// aws-cli should come before cli-a
	awsIndex := -1
	cliAIndex := -1
	for i, tool := range resolved {
		if tool.Name == "aws-cli" {
			awsIndex = i
		}
		if tool.Name == "cli-a" {
			cliAIndex = i
		}
	}
	
	if awsIndex == -1 || cliAIndex == -1 {
		t.Fatal("Both aws-cli and cli-a should be in resolved tools")
	}
	
	if awsIndex >= cliAIndex {
		t.Errorf("aws-cli (index %d) should come before cli-a (index %d)", awsIndex, cliAIndex)
	}
}

func TestResolveToolDependencies_CircularDependency(t *testing.T) {
	resolver := NewDependencyResolver()
	
	// Test case: Circular dependency
	tools := []parser.Tool{
		{Name: "tool-a", Dependencies: []string{"tool-b"}},
		{Name: "tool-b", Dependencies: []string{"tool-a"}},
	}
	
	_, err := resolver.ResolveToolDependencies(tools)
	if err == nil {
		t.Fatal("Expected circular dependency error, got nil")
	}
	
	if !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("Expected circular dependency error, got: %v", err)
	}
}

func TestResolveToolDependencies_MissingDependency(t *testing.T) {
	resolver := NewDependencyResolver()
	
	// Test case: Missing dependency
	tools := []parser.Tool{
		{Name: "cli-a", Dependencies: []string{"missing-tool"}},
	}
	
	_, err := resolver.ResolveToolDependencies(tools)
	if err == nil {
		t.Fatal("Expected missing dependency error, got nil")
	}
	
	if !strings.Contains(err.Error(), "dependency not found") {
		t.Errorf("Expected missing dependency error, got: %v", err)
	}
}

func TestResolveEnvironmentDependencies(t *testing.T) {
	resolver := NewDependencyResolver()
	
	// Test case: Environment dependency chain
	environments := []parser.Environment{
		{Name: "dev-env", Dependencies: []string{"base-env"}},
		{Name: "base-env", Dependencies: []string{}},
		{Name: "prod-env", Dependencies: []string{"base-env"}},
	}
	
	resolved, err := resolver.ResolveEnvironmentDependencies(environments)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// base-env should come before both dev-env and prod-env
	baseIndex := -1
	devIndex := -1
	prodIndex := -1
	
	for i, env := range resolved {
		switch env.Name {
		case "base-env":
			baseIndex = i
		case "dev-env":
			devIndex = i
		case "prod-env":
			prodIndex = i
		}
	}
	
	if baseIndex == -1 || devIndex == -1 || prodIndex == -1 {
		t.Fatal("All environments should be in resolved list")
	}
	
	if baseIndex >= devIndex {
		t.Errorf("base-env (index %d) should come before dev-env (index %d)", baseIndex, devIndex)
	}
	
	if baseIndex >= prodIndex {
		t.Errorf("base-env (index %d) should come before prod-env (index %d)", baseIndex, prodIndex)
	}
}

func TestGetInstallationOrder(t *testing.T) {
	resolver := NewDependencyResolver()
	
	tools := []parser.Tool{
		{Name: "cli-a", Dependencies: []string{"aws-cli"}},
		{Name: "aws-cli", Dependencies: []string{}},
	}
	
	environments := []parser.Environment{
		{Name: "dev-env", Dependencies: []string{"base-env"}},
		{Name: "base-env", Dependencies: []string{}},
	}
	
	orderedTools, orderedEnvs, err := resolver.GetInstallationOrder(tools, environments)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Verify tools are in correct order
	if len(orderedTools) != 2 {
		t.Fatalf("Expected 2 tools, got %d", len(orderedTools))
	}
	
	if orderedTools[0].Name != "aws-cli" || orderedTools[1].Name != "cli-a" {
		t.Errorf("Expected [aws-cli, cli-a], got [%s, %s]", orderedTools[0].Name, orderedTools[1].Name)
	}
	
	// Verify environments are in correct order
	if len(orderedEnvs) != 2 {
		t.Fatalf("Expected 2 environments, got %d", len(orderedEnvs))
	}
	
	if orderedEnvs[0].Name != "base-env" || orderedEnvs[1].Name != "dev-env" {
		t.Errorf("Expected [base-env, dev-env], got [%s, %s]", orderedEnvs[0].Name, orderedEnvs[1].Name)
	}
}