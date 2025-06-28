package pkg

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager == nil {
		t.Fatal("Manager should not be nil")
	}

	if manager.repoPath == "" {
		t.Fatal("Manager repoPath should not be empty")
	}
}

func TestValidatePackage(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test valid package
	pkg, err := manager.validatePackage("python")
	if err != nil {
		t.Fatalf("Failed to validate valid package: %v", err)
	}

	if pkg.Name != "python" {
		t.Fatalf("Expected package name 'python', got '%s'", pkg.Name)
	}

	// Test invalid package
	_, err = manager.validatePackage("nonexistent")
	if err == nil {
		t.Fatal("Expected error for invalid package")
	}
}

func TestIsPackageInstalled(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test with a package that should exist on most systems
	pkg := Package{
		Name:     "test",
		Commands: []string{"ls"}, // ls should exist on most systems
	}

	if !manager.isPackageInstalled(pkg) {
		t.Fatal("ls command should be available")
	}

	// Test with a package that shouldn't exist
	pkg = Package{
		Name:     "test",
		Commands: []string{"nonexistentcommand12345"},
	}

	if manager.isPackageInstalled(pkg) {
		t.Fatal("nonexistent command should not be available")
	}
}

func TestGetSystemVersion(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test with a package that might have a version
	version := manager.getSystemVersion("python")
	// Version might be empty if python is not installed, which is fine
	_ = version // Use version to avoid unused variable warning
}

func TestSanitizePackageList(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
		hasError bool
	}{
		{
			name:     "valid packages",
			input:    []string{"python", "node", "docker"},
			expected: []string{"python", "node", "docker"},
			hasError: false,
		},
		{
			name:     "duplicate packages",
			input:    []string{"python", "node", "python"},
			expected: []string{"python", "node"},
			hasError: false,
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: []string{},
			hasError: false,
		},
		{
			name:     "case insensitive",
			input:    []string{"Python", "NODE", "Docker"},
			expected: []string{"python", "node", "docker"},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizePackageList(tt.input)
			if tt.hasError && err == nil {
				t.Fatal("Expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d packages, got %d", len(tt.expected), len(result))
			}

			for i, pkg := range result {
				if pkg != tt.expected[i] {
					t.Fatalf("Expected package '%s', got '%s'", tt.expected[i], pkg)
				}
			}
		})
	}
}

func TestSupportsVersion(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		expected    bool
	}{
		{"python supports version", "python", true},
		{"node supports version", "node", true},
		{"docker does not support version", "docker", false},
		{"essentials does not support version", "essentials", false},
		{"nonexistent package", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SupportsVersion(tt.packageName)
			if result != tt.expected {
				t.Fatalf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		version     string
		hasError    bool
	}{
		{"valid python version", "python", "3.10", false},
		{"valid node version", "node", "18", false},
		{"invalid python version", "python", "1.0", true},
		{"invalid node version", "node", "99", true},
		{"empty version", "python", "", false},
		{"docker with version", "docker", "latest", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersion(tt.packageName, tt.version)
			if tt.hasError && err == nil {
				t.Fatal("Expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGetPackage(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		exists      bool
	}{
		{"existing package", "python", true},
		{"existing package", "node", true},
		{"existing package", "docker", true},
		{"nonexistent package", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg, exists := GetPackage(tt.packageName)
			if exists != tt.exists {
				t.Fatalf("Expected exists=%v, got %v", tt.exists, exists)
			}
			if tt.exists && pkg.Name == "" {
				t.Fatal("Package should have a name")
			}
		})
	}
}

func TestListPackages(t *testing.T) {
	packages := ListPackages()

	if len(packages) == 0 {
		t.Fatal("Should return at least some packages")
	}

	// Check that common packages exist
	expectedPackages := []string{"python", "node", "docker", "essentials"}
	for _, expected := range expectedPackages {
		if _, exists := packages[expected]; !exists {
			t.Fatalf("Expected package '%s' not found", expected)
		}
	}
}

func TestGetPackagesByCategory(t *testing.T) {
	packages := GetPackagesByCategory("development")

	if len(packages) == 0 {
		t.Fatal("Should return at least some development packages")
	}

	for _, pkg := range packages {
		if pkg.Category != "development" {
			t.Fatalf("Expected category 'development', got '%s'", pkg.Category)
		}
	}
}
