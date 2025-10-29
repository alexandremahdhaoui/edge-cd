package userconfig

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSpec_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Spec
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Spec{
				EdgeCD: EdgeCDSection{
					Repo: RepoConfig{
						URL:             "https://github.com/example/edge-cd.git",
						Branch:          "main",
						DestinationPath: "/usr/local/src/edge-cd",
					},
				},
				Config: ConfigSection{
					Spec: "spec.yaml",
					Path: "./devices/${HOSTNAME}",
					Repo: ConfigRepo{
						URL:      "https://github.com/example/config.git",
						Branch:   "main",
						DestPath: "/usr/local/src/config",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing config.path",
			config: &Spec{
				EdgeCD: EdgeCDSection{
					Repo: RepoConfig{
						URL:             "https://github.com/example/edge-cd.git",
						DestinationPath: "/usr/local/src/edge-cd",
					},
				},
				Config: ConfigSection{
					Spec: "spec.yaml",
					Path: "", // Missing
					Repo: ConfigRepo{
						URL:      "https://github.com/example/config.git",
						DestPath: "/usr/local/src/config",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing edgeCD.repo.url",
			config: &Spec{
				EdgeCD: EdgeCDSection{
					Repo: RepoConfig{
						URL:             "", // Missing
						DestinationPath: "/usr/local/src/edge-cd",
					},
				},
				Config: ConfigSection{
					Spec: "spec.yaml",
					Path: "./devices/${HOSTNAME}",
					Repo: ConfigRepo{
						URL:      "https://github.com/example/config.git",
						DestPath: "/usr/local/src/config",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing config.repo.destPath",
			config: &Spec{
				EdgeCD: EdgeCDSection{
					Repo: RepoConfig{
						URL:             "https://github.com/example/edge-cd.git",
						DestinationPath: "/usr/local/src/edge-cd",
					},
				},
				Config: ConfigSection{
					Spec: "spec.yaml",
					Path: "./devices/${HOSTNAME}",
					Repo: ConfigRepo{
						URL:      "https://github.com/example/config.git",
						DestPath: "", // Missing
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileSpec_Validate(t *testing.T) {
	tests := []struct {
		name    string
		file    FileSpec
		wantErr bool
	}{
		{
			name: "valid file type",
			file: FileSpec{
				Type:     "file",
				SrcPath:  "/src/file.txt",
				DestPath: "/dest/file.txt",
			},
			wantErr: false,
		},
		{
			name: "valid content type",
			file: FileSpec{
				Type:     "content",
				Content:  "some content",
				DestPath: "/dest/file.txt",
			},
			wantErr: false,
		},
		{
			name: "missing type",
			file: FileSpec{
				SrcPath:  "/src/file.txt",
				DestPath: "/dest/file.txt",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			file: FileSpec{
				Type:     "invalid",
				DestPath: "/dest/file.txt",
			},
			wantErr: true,
		},
		{
			name: "file type missing srcPath",
			file: FileSpec{
				Type:     "file",
				DestPath: "/dest/file.txt",
			},
			wantErr: true,
		},
		{
			name: "content type missing content",
			file: FileSpec{
				Type:     "content",
				DestPath: "/dest/file.txt",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.file.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSpec_SetDefaults(t *testing.T) {
	config := &Spec{
		EdgeCD: EdgeCDSection{
			Repo: RepoConfig{
				URL:             "https://github.com/example/edge-cd.git",
				DestinationPath: "/usr/local/src/edge-cd",
			},
		},
		Config: ConfigSection{
			Path: "./devices/${HOSTNAME}",
			Repo: ConfigRepo{
				URL:      "https://github.com/example/config.git",
				DestPath: "/usr/local/src/config",
			},
		},
		Files: []FileSpec{
			{
				Type:     "file",
				SrcPath:  "/src/file.txt",
				DestPath: "/dest/file.txt",
			},
		},
	}

	config.SetDefaults()

	// Check defaults were set
	if config.Config.Spec != "spec.yaml" {
		t.Errorf("Expected spec to be 'spec.yaml', got '%s'", config.Config.Spec)
	}

	if config.EdgeCD.Repo.Branch != "main" {
		t.Errorf("Expected edgeCD.repo.branch to be 'main', got '%s'", config.EdgeCD.Repo.Branch)
	}

	if config.Config.Repo.Branch != "main" {
		t.Errorf("Expected config.repo.branch to be 'main', got '%s'", config.Config.Repo.Branch)
	}

	if config.PollingInterval != 60 {
		t.Errorf("Expected pollingInterval to be 60, got %d", config.PollingInterval)
	}

	if config.Files[0].FileMod != "644" {
		t.Errorf("Expected file.fileMod to be '644', got '%s'", config.Files[0].FileMod)
	}
}

func TestSpec_YAMLMarshaling(t *testing.T) {
	yamlData := `edgeCD:
  repo:
    url: https://github.com/example/edge-cd.git
    branch: main
    destinationPath: /usr/local/src/edge-cd
config:
  spec: spec.yaml
  path: ./devices/${HOSTNAME}
  repo:
    url: https://github.com/example/config.git
    branch: main
    destPath: /usr/local/src/config
pollingIntervalSecond: 60
serviceManager:
  name: systemd
packageManager:
  name: apt
  autoUpgrade: true
  requiredPackages:
    - git
    - curl
files:
  - type: file
    srcPath: /src/file.txt
    destPath: /dest/file.txt
    fileMod: "644"
    syncBehavior:
      restartServices:
        - nginx
      reboot: false
`

	var config Spec
	err := yaml.Unmarshal([]byte(yamlData), &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// Verify critical fields
	if config.EdgeCD.Repo.URL != "https://github.com/example/edge-cd.git" {
		t.Errorf("Expected edgeCD.repo.url to be 'https://github.com/example/edge-cd.git', got '%s'", config.EdgeCD.Repo.URL)
	}

	if config.EdgeCD.Repo.DestinationPath != "/usr/local/src/edge-cd" {
		t.Errorf("Expected edgeCD.repo.destinationPath to be '/usr/local/src/edge-cd', got '%s'", config.EdgeCD.Repo.DestinationPath)
	}

	if config.Config.Repo.DestPath != "/usr/local/src/config" {
		t.Errorf("Expected config.repo.destPath to be '/usr/local/src/config', got '%s'", config.Config.Repo.DestPath)
	}

	if config.Config.Spec != "spec.yaml" {
		t.Errorf("Expected config.spec to be 'spec.yaml', got '%s'", config.Config.Spec)
	}

	if len(config.Files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(config.Files))
	}

	if config.Files[0].SyncBehavior == nil {
		t.Fatal("Expected syncBehavior to be set")
	}

	if len(config.Files[0].SyncBehavior.RestartServices) != 1 || config.Files[0].SyncBehavior.RestartServices[0] != "nginx" {
		t.Errorf("Expected restartServices to contain 'nginx', got %v", config.Files[0].SyncBehavior.RestartServices)
	}

	// Test marshaling back to YAML
	out, err := yaml.Marshal(&config)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}

	// Verify we can unmarshal again
	var config2 Spec
	err = yaml.Unmarshal(out, &config2)
	if err != nil {
		t.Fatalf("Failed to unmarshal re-marshaled YAML: %v", err)
	}

	// Verify critical fields match
	if config2.EdgeCD.Repo.URL != config.EdgeCD.Repo.URL {
		t.Errorf("Marshaling round-trip failed for edgeCD.repo.url")
	}

	if config2.Config.Repo.DestPath != config.Config.Repo.DestPath {
		t.Errorf("Marshaling round-trip failed for config.repo.destPath")
	}
}

func TestConfigRepo_vs_RepoConfig_FieldNames(t *testing.T) {
	// This test verifies the intentional difference between ConfigRepo and RepoConfig
	// ConfigRepo uses "destPath" while RepoConfig uses "destinationPath"

	edgeCDYAML := `url: https://github.com/example/edge-cd.git
branch: main
destinationPath: /usr/local/src/edge-cd
`

	var edgeCDRepo RepoConfig
	err := yaml.Unmarshal([]byte(edgeCDYAML), &edgeCDRepo)
	if err != nil {
		t.Fatalf("Failed to unmarshal RepoConfig: %v", err)
	}

	if edgeCDRepo.DestinationPath != "/usr/local/src/edge-cd" {
		t.Errorf("RepoConfig should use 'destinationPath', got '%s'", edgeCDRepo.DestinationPath)
	}

	configYAML := `url: https://github.com/example/config.git
branch: main
destPath: /usr/local/src/config
`

	var configRepo ConfigRepo
	err = yaml.Unmarshal([]byte(configYAML), &configRepo)
	if err != nil {
		t.Fatalf("Failed to unmarshal ConfigRepo: %v", err)
	}

	if configRepo.DestPath != "/usr/local/src/config" {
		t.Errorf("ConfigRepo should use 'destPath', got '%s'", configRepo.DestPath)
	}
}
