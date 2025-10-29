package userconfig

import (
	"fmt"
	"strings"
)

// Validate checks if the Spec is valid
func (c *Spec) Validate() error {
	if err := c.EdgeCD.Validate(); err != nil {
		return fmt.Errorf("edgeCD validation failed: %w", err)
	}

	if err := c.Config.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Validate files if present
	for i, file := range c.Files {
		if err := file.Validate(); err != nil {
			return fmt.Errorf("file[%d] validation failed: %w", i, err)
		}
	}

	// Validate directories if present
	for i, dir := range c.Directories {
		if err := dir.Validate(); err != nil {
			return fmt.Errorf("directory[%d] validation failed: %w", i, err)
		}
	}

	return nil
}

// Validate checks if the EdgeCDSection is valid
func (e *EdgeCDSection) Validate() error {
	if err := e.Repo.Validate(); err != nil {
		return fmt.Errorf("repo validation failed: %w", err)
	}
	return nil
}

// Validate checks if the ConfigSection is valid
func (c *ConfigSection) Validate() error {
	if c.Path == "" {
		return fmt.Errorf("config.path is required")
	}

	if c.Spec == "" {
		return fmt.Errorf("config.spec is required")
	}

	if err := c.Repo.Validate(); err != nil {
		return fmt.Errorf("repo validation failed: %w", err)
	}

	return nil
}

// Validate checks if the RepoConfig is valid
func (r *RepoConfig) Validate() error {
	if r.URL == "" {
		return fmt.Errorf("repo.url is required")
	}

	if r.DestinationPath == "" {
		return fmt.Errorf("repo.destinationPath is required")
	}

	return nil
}

// Validate checks if the ConfigRepo is valid
func (r *ConfigRepo) Validate() error {
	if r.URL == "" {
		return fmt.Errorf("repo.url is required")
	}

	if r.DestPath == "" {
		return fmt.Errorf("repo.destPath is required")
	}

	return nil
}

// Validate checks if the FileSpec is valid
func (f *FileSpec) Validate() error {
	if f.Type == "" {
		return fmt.Errorf("file.type is required")
	}

	validTypes := []string{"file", "directory", "content"}
	isValidType := false
	for _, vt := range validTypes {
		if f.Type == vt {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("file.type must be one of: %s", strings.Join(validTypes, ", "))
	}

	if f.DestPath == "" {
		return fmt.Errorf("file.destPath is required")
	}

	// Type-specific validation
	switch f.Type {
	case "file", "directory":
		if f.SrcPath == "" {
			return fmt.Errorf("file.srcPath is required for type '%s'", f.Type)
		}
	case "content":
		if f.Content == "" {
			return fmt.Errorf("file.content is required for type 'content'")
		}
	}

	return nil
}

// Validate checks if the DirectorySpec is valid
func (d *DirectorySpec) Validate() error {
	if d.SourceDir == "" {
		return fmt.Errorf("directory.sourceDir is required")
	}

	if d.DestDir == "" {
		return fmt.Errorf("directory.destDir is required")
	}

	return nil
}

// SetDefaults sets default values for optional fields
func (c *Spec) SetDefaults() {
	// Set default spec file name if not provided
	if c.Config.Spec == "" {
		c.Config.Spec = "spec.yaml"
	}

	// Set default branch if not provided
	if c.EdgeCD.Repo.Branch == "" {
		c.EdgeCD.Repo.Branch = "main"
	}

	if c.Config.Repo.Branch == "" {
		c.Config.Repo.Branch = "main"
	}

	// Set default polling interval if not provided
	if c.PollingInterval == 0 {
		c.PollingInterval = 60 // Default to 60 seconds
	}

	// Set default file mode for files
	for i := range c.Files {
		if c.Files[i].FileMod == "" {
			c.Files[i].FileMod = "644"
		}
	}

	// Set default file mode for directories
	for i := range c.Directories {
		if c.Directories[i].FileMod == "" {
			c.Directories[i].FileMod = "755"
		}
	}
}
