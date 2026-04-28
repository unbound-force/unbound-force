package sandbox

import (
	"fmt"
	"runtime"
	"strconv"
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

	// UIDMapSupported is true when the Podman machine
	// supports --userns=keep-id UID mapping. Always true
	// on Linux. On macOS, determined by probing the Podman
	// machine's virtiofs support.
	UIDMapSupported bool
}

// DetectPlatform detects the host platform configuration
// (architecture, SELinux status, and UID mapping support)
// for container flag selection. SELinux detection uses the
// injectable ExecCmd and ReadFile on Options for testability.
func DetectPlatform(opts Options) PlatformConfig {
	p := PlatformConfig{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	// SELinux only exists on Linux. Skip detection on macOS
	// and other platforms.
	if p.OS == "linux" {
		// Linux always supports UID mapping natively.
		p.UIDMapSupported = true

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
	} else if p.OS == "darwin" {
		// macOS: probe the Podman machine to check if
		// virtiofs supports keep-id UID mapping.
		p.UIDMapSupported = probeUIDMapping(opts)
	}

	return p
}

// probeImage is the fully-qualified image used for the
// macOS UID mapping probe. Using the full registry path
// avoids Podman short-name resolution ambiguity, which
// can cause the probe to fail on fresh Podman machines
// where registries.conf is not yet configured.
const probeImage = "docker.io/library/busybox:latest"

// probeUIDMapping checks whether the Podman machine supports
// --userns=keep-id UID mapping by running a lightweight
// busybox container that stats a mounted directory. Returns
// true if the file owner inside the container is UID 1000
// (the mapped "dev" user). Returns false on any error or
// unexpected output (fail-safe).
func probeUIDMapping(opts Options) bool {
	fmt.Fprintf(opts.Stderr, "Checking Podman machine UID mapping...\n")

	out, err := opts.ExecCmd("podman",
		"run", "--rm",
		"--entrypoint", "stat",
		"--userns=keep-id:uid=1000,gid=1000",
		"-v", opts.ProjectDir+":/test:ro",
		probeImage,
		"-c", "%u", "/test",
	)
	if err != nil {
		fmt.Fprintf(opts.Stderr,
			"  UID probe failed: %v\n", err)
		return false
	}
	result := strings.TrimSpace(string(out))
	if result != "1000" {
		fmt.Fprintf(opts.Stderr,
			"  UID probe returned %q (expected 1000)\n", result)
		return false
	}
	return true
}

// parsePodmanVersion runs `podman --version` and parses the
// major and minor version numbers from the output format
// "podman version X.Y.Z".
func parsePodmanVersion(opts Options) (int, int, error) {
	out, err := opts.ExecCmd("podman", "--version")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get podman version: %w", err)
	}

	// Expected format: "podman version X.Y.Z"
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) < 3 {
		return 0, 0, fmt.Errorf("unexpected podman version output: %s", string(out))
	}

	version := parts[len(parts)-1]
	segments := strings.SplitN(version, ".", 3)
	if len(segments) < 2 {
		return 0, 0, fmt.Errorf("cannot parse podman version: %s", version)
	}

	major, err := strconv.Atoi(segments[0])
	if err != nil {
		return 0, 0, fmt.Errorf("cannot parse major version: %w", err)
	}
	minor, err := strconv.Atoi(segments[1])
	if err != nil {
		return 0, 0, fmt.Errorf("cannot parse minor version: %w", err)
	}

	return major, minor, nil
}

// isRootlessPodman checks whether Podman is running in
// rootless mode by querying `podman info`. Returns true if
// the output is "true". Returns false on any error or
// unexpected output (fail-safe).
func isRootlessPodman(opts Options) bool {
	out, err := opts.ExecCmd("podman", "info",
		"--format", "{{.Host.Security.Rootless}}")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}
