package userconfig

// Spec represents the complete edge-cd configuration structure.
// This is the authoritative definition based on cmd/edge-cd/edge-cd script.
type Spec struct {
	EdgeCD          EdgeCDSection          `yaml:"edgeCD" json:"edgeCD"`
	Config          ConfigSection          `yaml:"config" json:"config"`
	PollingInterval int                    `yaml:"pollingIntervalSecond,omitempty" json:"pollingIntervalSecond,omitempty"`
	ExtraEnvs       []map[string]string    `yaml:"extraEnvs,omitempty" json:"extraEnvs,omitempty"`
	ServiceManager  ServiceManagerSection  `yaml:"serviceManager,omitempty" json:"serviceManager,omitempty"`
	PackageManager  PackageManagerSection  `yaml:"packageManager,omitempty" json:"packageManager,omitempty"`
	Files           []FileSpec             `yaml:"files,omitempty" json:"files,omitempty"`
	Directories     []DirectorySpec        `yaml:"directories,omitempty" json:"directories,omitempty"`
	Log             *LogSection            `yaml:"log,omitempty" json:"log,omitempty"`
}

// EdgeCDSection defines how edge-cd manages itself
type EdgeCDSection struct {
	Repo       RepoConfig         `yaml:"repo" json:"repo"`
	CommitPath string             `yaml:"commitPath,omitempty" json:"commitPath,omitempty"`
	AutoUpdate *AutoUpdateSection `yaml:"autoUpdate,omitempty" json:"autoUpdate,omitempty"`
}

// AutoUpdateSection controls edge-cd auto-update behavior
type AutoUpdateSection struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// ConfigSection defines user configuration repository settings
type ConfigSection struct {
	Spec       string     `yaml:"spec" json:"spec"`                       // Default: "spec.yaml"
	Path       string     `yaml:"path" json:"path"`                       // Required
	Repo       ConfigRepo `yaml:"repo" json:"repo"`
	CommitPath string     `yaml:"commitPath,omitempty" json:"commitPath,omitempty"`
}

// RepoConfig represents a git repository configuration for edge-cd itself
// Uses "destinationPath" field name
type RepoConfig struct {
	URL             string `yaml:"url" json:"url"`
	Branch          string `yaml:"branch,omitempty" json:"branch,omitempty"`
	DestinationPath string `yaml:"destinationPath" json:"destinationPath"`
}

// ConfigRepo represents a git repository configuration for user config
// Uses "destPath" field name (different from RepoConfig!)
type ConfigRepo struct {
	URL      string `yaml:"url" json:"url"`
	Branch   string `yaml:"branch,omitempty" json:"branch,omitempty"`
	DestPath string `yaml:"destPath" json:"destPath"` // NOTE: Different from RepoConfig!
}

// ServiceManagerSection defines the service manager to use
type ServiceManagerSection struct {
	Name string `yaml:"name" json:"name"`
}

// PackageManagerSection defines package management settings
type PackageManagerSection struct {
	Name             string   `yaml:"name" json:"name"`
	AutoUpgrade      bool     `yaml:"autoUpgrade,omitempty" json:"autoUpgrade,omitempty"`
	RequiredPackages []string `yaml:"requiredPackages,omitempty" json:"requiredPackages,omitempty"`
}

// FileSpec represents a single file to be managed
// Supports three types: "file", "directory", "content"
type FileSpec struct {
	Type         string        `yaml:"type" json:"type"`                                 // "file", "directory", "content"
	SrcPath      string        `yaml:"srcPath,omitempty" json:"srcPath,omitempty"`       // For type: file or directory
	DestPath     string        `yaml:"destPath" json:"destPath"`                         // Required
	Content      string        `yaml:"content,omitempty" json:"content,omitempty"`       // For type: content
	FileMod      string        `yaml:"fileMod,omitempty" json:"fileMod,omitempty"`       // Default: "644"
	SyncBehavior *SyncBehavior `yaml:"syncBehavior,omitempty" json:"syncBehavior,omitempty"`
}

// SyncBehavior defines actions to take when a file changes
type SyncBehavior struct {
	RestartServices []string `yaml:"restartServices,omitempty" json:"restartServices,omitempty"`
	Reboot          bool     `yaml:"reboot,omitempty" json:"reboot,omitempty"`
}

// DirectorySpec represents a directory to be managed
type DirectorySpec struct {
	SourceDir          string              `yaml:"sourceDir" json:"sourceDir"`
	DestDir            string              `yaml:"destDir" json:"destDir"`
	FileMod            string              `yaml:"fileMod,omitempty" json:"fileMod,omitempty"`
	RestartServicesMap map[string][]string `yaml:"restartServicesMap,omitempty" json:"restartServicesMap,omitempty"`
	RebootOnChange     []string            `yaml:"rebootOnChange,omitempty" json:"rebootOnChange,omitempty"`
}

// LogSection defines logging configuration
type LogSection struct {
	Format string `yaml:"format,omitempty" json:"format,omitempty"`
}
