package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Config struct {
	TarDest   string
	Dest      string
	BinDir    string
	LibDir    string
	IndexURL  string
	Version   string
}

type Logger struct {
	colorReset  string
	colorRed    string
	colorGreen  string
	colorYellow string
	colorBlue   string
	colorCyan   string
}

func (l Logger) info(format string, a ...interface{}) {
	fmt.Printf("💡 "+l.colorBlue+"info:"+l.colorReset+" %s\n", fmt.Sprintf(format, a...))
}

func (l Logger) success(format string, a ...interface{}) {
	fmt.Printf("✅ "+l.colorGreen+"success:"+l.colorReset+" %s\n", fmt.Sprintf(format, a...))
}

func (l Logger) warning(format string, a ...interface{}) {
	fmt.Printf("⚠️  "+l.colorYellow+"warning:"+l.colorReset+" %s\n", fmt.Sprintf(format, a...))
}

func (l Logger) error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "❌ "+l.colorRed+"error:"+l.colorReset+" %s\n", fmt.Sprintf(format, a...))
}

func (l Logger) step(format string, a ...interface{}) {
	fmt.Printf("👉 "+l.colorCyan+"step:"+l.colorReset+" %s\n", fmt.Sprintf(format, a...))
}

var logger = Logger{
	colorReset:  "\033[0m",
	colorRed:    "\033[31m",
	colorGreen:  "\033[32m",
	colorYellow: "\033[33m",
	colorBlue:   "\033[34m",
	colorCyan:   "\033[36m",
}

func getConfig() Config {
	var cfg Config

	flag.StringVar(&cfg.TarDest, "tar-dest", getEnv("ZIG_TAR_DEST", "/tmp/zig.tar.xz"), "Path to download the Zig tarball")
	flag.StringVar(&cfg.Dest, "dest", getEnv("ZIG_DEST", "/tmp/zig"), "Temporary directory for extraction")
	flag.StringVar(&cfg.BinDir, "bin-dir", getEnv("ZIG_BIN_DIR", "/usr/local/bin"), "Installation directory for Zig binary")
	flag.StringVar(&cfg.LibDir, "lib-dir", getEnv("ZIG_LIB_DIR", "/usr/local/lib"), "Installation directory for Zig libraries")
	flag.StringVar(&cfg.IndexURL, "index-url", getEnv("ZIG_INDEX_URL", "https://ziglang.org/download/index.json"), "URL for Zig download index")
	flag.StringVar(&cfg.Version, "version", getEnv("ZIG_VERSION", "master"), "Zig version to install (e.g., master, 0.11.0)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nEnvironment variables:\n")
		fmt.Fprintf(os.Stderr, "  ZIG_TAR_DEST   Path to download the Zig tarball\n")
		fmt.Fprintf(os.Stderr, "  ZIG_DEST       Temporary directory for extraction\n")
		fmt.Fprintf(os.Stderr, "  ZIG_BIN_DIR    Installation directory for Zig binary\n")
		fmt.Fprintf(os.Stderr, "  ZIG_LIB_DIR    Installation directory for Zig libraries\n")
		fmt.Fprintf(os.Stderr, "  ZIG_INDEX_URL  URL for Zig download index\n")
		fmt.Fprintf(os.Stderr, "  ZIG_VERSION    Zig version to install (e.g., master, 0.11.0)\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func checkDependencies() error {
	deps := []string{"tar"}
	for _, dep := range deps {
		_, err := exec.LookPath(dep)
		if err != nil {
			return fmt.Errorf("missing dependency: %s", dep)
		}
	}
	return nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func verifyChecksum(file, expectedSum string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	sum := hex.EncodeToString(h.Sum(nil))
	if sum != expectedSum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedSum, sum)
	}
	return nil
}

func extractTarball(src, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	args := []string{"-xf", src, "-C", dest, "--strip-components=1"}
	if strings.HasSuffix(src, ".tar.xz") {
		args = append([]string{"-J"}, args...)
	}
	cmd := exec.Command("tar", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tar extraction failed: %v: %s", err, out)
	}
	return nil
}

func ensureDirectoryExists(path string) error {
	return os.MkdirAll(path, 0755)
}

func getPlatformKey() string {
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	} else if arch == "386" {
		arch = "x86"
	}
	
	os := runtime.GOOS
	if os == "darwin" {
		os = "macos"
	}
	return fmt.Sprintf("%s-%s", arch, os)
}

func main() {
	cfg := getConfig()

	if err := checkDependencies(); err != nil {
		logger.error("%v", err)
		os.Exit(1)
	}

	// Ensure parent directories exist
	if err := ensureDirectoryExists(filepath.Dir(cfg.TarDest)); err != nil {
		logger.error("failed to create tarball directory: %v", err)
		os.Exit(1)
	}

	// Clean up previous files
	os.Remove(cfg.TarDest)
	os.RemoveAll(cfg.Dest)

	// Fetch release information
	resp, err := http.Get(cfg.IndexURL)
	if err != nil {
		logger.error("failed to fetch index: %v", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.error("failed to fetch index: HTTP %d", resp.StatusCode)
		os.Exit(1)
	}

	var index map[string]map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		logger.error("failed to parse index: %v", err)
		os.Exit(1)
	}

	// Get the version info
	versionInfo, ok := index[cfg.Version]
	if !ok {
		logger.error("version %s not found in index", cfg.Version)
		os.Exit(1)
	}

	// Get platform-specific release
	platformKey := getPlatformKey()
	platformRelease, ok := versionInfo[platformKey].(map[string]interface{})
	if !ok {
		logger.error("no release found for platform %s and version %s", platformKey, cfg.Version)
		os.Exit(1)
	}

	tarballURL, ok := platformRelease["tarball"].(string)
	if !ok {
		logger.error("invalid tarball URL in index")
		os.Exit(1)
	}

	shasum, ok := platformRelease["shasum"].(string)
	if !ok {
		logger.error("invalid shasum in index")
		os.Exit(1)
	}

	// Check if we already have a valid tarball
	needsDownload := true
	if _, err := os.Stat(cfg.TarDest); err == nil {
		logger.info("found existing file, checking checksum...")
		if err := verifyChecksum(cfg.TarDest, shasum); err == nil {
			logger.success("existing file matches checksum, skipping download")
			needsDownload = false
		} else {
			logger.warning("existing file has incorrect checksum, will download fresh copy")
			os.Remove(cfg.TarDest)
		}
	}

	// Download tarball if needed
	if needsDownload {
		logger.step("downloading Zig %s for %s...", cfg.Version, platformKey)
		if err := downloadFile(tarballURL, cfg.TarDest); err != nil {
			logger.error("failed to download tarball: %v", err)
			os.Exit(1)
		}

		// Verify checksum of downloaded file
		logger.step("verifying checksum...")
		if err := verifyChecksum(cfg.TarDest, shasum); err != nil {
			os.Remove(cfg.TarDest)
			logger.error("checksum verification failed: %v", err)
			os.Exit(1)
		}
	}

	logger.step("extracting...")
	if err := extractTarball(cfg.TarDest, cfg.Dest); err != nil {
		logger.error("failed to extract tarball: %v", err)
		os.Exit(1)
	}

	// Ensure installation directories exist
	if err := ensureDirectoryExists(cfg.BinDir); err != nil {
		logger.error("failed to create bin directory: %v", err)
		os.Exit(1)
	}
	if err := ensureDirectoryExists(cfg.LibDir); err != nil {
		logger.error("failed to create lib directory: %v", err)
		os.Exit(1)
	}

	// Install zig
	logger.step("installing...")
	os.Remove(filepath.Join(cfg.BinDir, "zig"))
	os.RemoveAll(filepath.Join(cfg.LibDir, "zig"))

	if err := os.Rename(filepath.Join(cfg.Dest, "zig"), filepath.Join(cfg.BinDir, "zig")); err != nil {
		logger.error("failed to install zig binary: %v", err)
		os.Exit(1)
	}

	// First ensure lib directory exists
	libSrcPath := filepath.Join(cfg.Dest, "lib")
	if _, err := os.ReadDir(libSrcPath); err != nil {
		logger.error("failed to read lib directory: %v", err)
		os.Exit(1)
	}

	// Move the entire lib directory
	if err := os.Rename(libSrcPath, filepath.Join(cfg.LibDir, "zig")); err != nil {
		logger.error("failed to install zig libraries: %v", err)
		os.Exit(1)
	}

	// Cleanup
	logger.step("cleaning up...")
	os.Remove(cfg.TarDest)
	os.RemoveAll(cfg.Dest)

	logger.success("Zig %s installed successfully! 🎉", cfg.Version)
}