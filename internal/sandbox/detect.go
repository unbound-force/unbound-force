package sandbox

import (
	"runtime"
	"strings"
)

// PlatformConfig captures host platform properties that
// influence container flags. Detected once at sandbox start
// and passed to configuration builders.
type PlatformConfig struct {
	// OS is the host operating system ("darwin" or "linux").
	OS string

	// Arch is the host CPU architecture ("arm64" or "amd64").
	Arch string

	// SELinux is true when SELinux is in enforcing mode.
	// Always false on macOS. When true, volume mounts
	// require the :Z relabeling flag.
	SELinux bool
}

// DetectPlatform detects the host platform configuration
// (architecture and SELinux status) for container flag
// selection. SELinux detection uses the injectable ExecCmd
// and ReadFile on Options for testability.
func DetectPlatform(opts Options) PlatformConfig {
	p := PlatformConfig{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	// SELinux only exists on Linux. Skip detection on macOS
	// and other platforms.
	if p.OS != "linux" {
		return p
	}

	// Check /etc/selinux/config for SELINUX=enforcing.
	// This is the persistent configuration; getenforce gives
	// the runtime state. We check both for accuracy.
	if data, err := opts.ReadFile("/etc/selinux/config"); err == nil {
		if strings.Contains(string(data), "SELINUX=enforcing") {
			// Verify with getenforce for runtime state.
			if out, err := opts.ExecCmd("getenforce"); err == nil {
				if strings.TrimSpace(string(out)) == "Enforcing" {
					p.SELinux = true
				}
			}
		}
	}

	return p
}
