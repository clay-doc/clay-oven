package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// RunOven is the main orchestration function. It loads configuration,
// scans the docs directory, generates a structure file, and downloads
// the clay release.
func RunOven(args map[string]string, sink OutputSink) {
	_, noConfirm := args["-nc"]

	// Resolve arguments with defaults
	config := argOrDefault(args, "-c", "clay.yaml")
	docsDir := argOrDefault(args, "-d", "./docs")
	output := argOrDefault(args, "-o", "./output")
	dirMeta := argOrDefault(args, "-fm", "dir-meta.yaml")

	_, ciMode := args["-ci"]

	// Print parameters
	sink.Header("Configuration")
	sink.KeyVal("Config file", config)
	sink.KeyVal("Docs directory", docsDir)
	sink.KeyVal("Output directory", output)
	sink.KeyVal("Dir meta file", dirMeta)
	sink.KeyVal("No confirm", fmt.Sprintf("%v", noConfirm))
	sink.KeyVal("Verbose", fmt.Sprintf("%v", Verbose))
	sink.KeyVal("CI mode", fmt.Sprintf("%v", ciMode))

	// --- Load config ---
	sink.Header("Loading")

	cfg, err := LoadConfigYaml(config)
	if err != nil {
		sink.Error(fmt.Sprintf("Could not load config: %v", err))
		sink.Done()
		return
	}
	sink.Success(fmt.Sprintf("Config loaded: %s", cfg.Title))

	// --- Load directory tree ---
	rootNode := DirNode{PathName: "docs"}
	directoryTree, err := LoadDirectoryTree(rootNode, docsDir)
	if err != nil {
		sink.Error(fmt.Sprintf("Could not load docs directory: %v", err))
		sink.Done()
		return
	}
	sink.Success("Docs directory loaded")

	// --- Load meta tree ---
	metaTree, err := LoadMetaTree(dirMeta)
	if err != nil {
		sink.Warn(fmt.Sprintf("Could not load dir meta: %v", err))
		sink.Info("Continuing without directory metadata")
		metaTree = MetaNode{}
	} else {
		sink.Success("Dir meta loaded")
	}

	// --- Download clay release ---
	sink.Header("Download")

	cwd, err := os.Getwd()
	if err != nil {
		sink.Error(fmt.Sprintf("Could not get working directory: %v", err))
		sink.Done()
		return
	}

	zipPath := filepath.Join(cwd, "clay-dist.zip")
	sink.Info(fmt.Sprintf("Clay release will be downloaded to: %s", zipPath))

	if !noConfirm {
		if !sink.Confirm("Download clay release?") {
			sink.Warn("Download cancelled")
			sink.Done()
			return
		}
	}

	downloadedPath, err := downloadClayRelease(cwd, sink)
	if err != nil {
		sink.Error(fmt.Sprintf("Download failed: %v", err))
		sink.Done()
		return
	}
	sink.Success(fmt.Sprintf("Clay release downloaded to: %s", downloadedPath))

	// --- Unzip clay release ---
	sink.Header("Extract")

	publishDir := filepath.Join(cwd, "publish")
	sink.Info(fmt.Sprintf("Extracting to: %s", publishDir))

	if err := unzipFile(downloadedPath, publishDir, sink); err != nil {
		sink.Error(fmt.Sprintf("Extraction failed: %v", err))
		sink.Done()
		return
	}
	sink.Success(fmt.Sprintf("Clay release extracted to: %s", publishDir))

	// --- Clean up downloaded zip ---
	if err := os.Remove(downloadedPath); err != nil {
		sink.Warn(fmt.Sprintf("Could not delete zip: %v", err))
	} else {
		sink.Success("Downloaded zip deleted")
	}

	// --- Replace base URL placeholder ---
	sink.Header("Base URL")
	publicDir := filepath.Join(publishDir, "public")

	baseURL := strings.Trim(cfg.BaseURL, "/")
	if baseURL == "" {
		sink.Info("baseURL is empty or \"/\", placeholder will be removed")
	} else {
		sink.Info(fmt.Sprintf("baseURL resolved to: %q", baseURL))
	}

	sink.Info(fmt.Sprintf("Replacing ##BASE_URL_BUILD## in %s", publicDir))

	if !noConfirm {
		if !sink.Confirm("Replace base URL placeholders?") {
			sink.Warn("Base URL replacement cancelled")
			sink.Done()
			return
		}
	}

	if err := replaceInDir(publicDir, "##BASE_URL_BUILD##", baseURL, sink); err != nil {
		sink.Error(fmt.Sprintf("Base URL replacement failed: %v", err))
		sink.Done()
		return
	}
	sink.Success("Base URL placeholders replaced")

	// --- Copy clay.yaml to publish/public ---
	sink.Header("Copy Config")
	configDest := filepath.Join(publicDir, "clay.yaml")
	sink.Info(fmt.Sprintf("Copying %s to %s", config, configDest))

	if !noConfirm {
		if !sink.Confirm("Copy config file?") {
			sink.Warn("Config copy cancelled")
			sink.Done()
			return
		}
	}

	if err := copyFile(config, configDest); err != nil {
		sink.Error(fmt.Sprintf("Could not copy config: %v", err))
		sink.Done()
		return
	}
	sink.Success(fmt.Sprintf("Config copied to: %s", configDest))

	// --- Copy assets to publish/public ---
	sink.Header("Copy Assets")

	assets := collectAssets(cfg)
	for _, asset := range assets {
		sink.Info(fmt.Sprintf("Asset: %s", asset))
	}

	if !noConfirm {
		if !sink.Confirm("Copy asset files?") {
			sink.Warn("Asset copy cancelled")
			sink.Done()
			return
		}
	}

	for _, asset := range assets {
		assetDest := filepath.Join(publicDir, asset)
		if err := os.MkdirAll(filepath.Dir(assetDest), 0755); err != nil {
			sink.Error(fmt.Sprintf("Could not create directory for asset %s: %v", asset, err))
			sink.Done()
			return
		}
		if err := copyFile(asset, assetDest); err != nil {
			sink.Warn(fmt.Sprintf("Could not copy asset %s: %v", asset, err))
		} else {
			sink.Success(fmt.Sprintf("Copied asset: %s", asset))
		}
	}

	// --- Copy markdown files to publish/public ---
	sink.Header("Copy Docs")
	docsDestDir := filepath.Join(publicDir, "docs")
	sink.Info(fmt.Sprintf("Copying markdown files from %s to %s", docsDir, docsDestDir))

	if !noConfirm {
		if !sink.Confirm("Copy markdown files?") {
			sink.Warn("Markdown copy cancelled")
			sink.Done()
			return
		}
	}

	if err := copyMarkdownFiles(docsDir, docsDestDir, sink); err != nil {
		sink.Error(fmt.Sprintf("Could not copy markdown files: %v", err))
		sink.Done()
		return
	}
	sink.Success("Markdown files copied")

	// --- Generate structure file ---
	if err := genStructureFile(directoryTree, metaTree, noConfirm, sink); err != nil {
		sink.Error(fmt.Sprintf("Structure generation failed: %v", err))
		sink.Done()
		return
	}

	// --- Summary ---
	sink.Header("Summary")
	sink.KeyVal("Output directory", publishDir)
	sink.KeyVal("Public directory", publicDir)
	sink.KeyVal("Base URL", baseURL)
	sink.KeyVal("Config file", configDest)
	sink.KeyVal("Docs copied to", docsDestDir)
	sink.KeyVal("Structure file", filepath.Join(publicDir, "clay-structure.yaml"))
	sink.Success("Build complete")

	_ = output
	sink.Done()
}

// argOrDefault returns the argument value if present, otherwise the default.
func argOrDefault(args map[string]string, key, fallback string) string {
	if val, ok := args[key]; ok {
		return val
	}
	return fallback
}

// progressWriter wraps an io.Writer to report download progress.
type progressWriter struct {
	sink     OutputSink
	total    int64
	received int64
	inner    io.Writer
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.inner.Write(p)
	pw.received += int64(n)
	pw.sink.DownloadProgress(pw.received, pw.total)
	return n, err
}

// downloadClayRelease downloads the latest clay dist.zip to the given directory.
func downloadClayRelease(dir string, sink OutputSink) (string, error) {
	url := "https://github.com/clay-doc/clay/releases/latest/download/dist.zip"

	destPath := filepath.Join(dir, "clay-dist.zip")

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("creating output file: %w", err)
	}
	defer out.Close()

	pw := &progressWriter{
		sink:  sink,
		total: resp.ContentLength,
		inner: out,
	}

	if _, err = io.Copy(pw, resp.Body); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return destPath, nil
}

// genStructureFile generates and optionally writes the structure.yaml file.
func genStructureFile(dirTree DirNode, metaTree MetaNode, noConfirm bool, sink OutputSink) error {
	sink.Header("Structure Generation")

	structure := StructureFile{Lines: []string{"docs:\n"}}
	GenerateStructureFile(&structure, dirTree, &metaTree, 1, sink)

	// Preview the generated content
	sink.Info("Generated structure file content:")
	for _, line := range structure.Lines {
		sink.StructLine(line)
	}

	// Write to disk
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}
	publicDir := filepath.Join(cwd, "publish", "public")
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		return fmt.Errorf("creating publish/public directory: %w", err)
	}
	destPath := filepath.Join(publicDir, "clay-structure.yaml")

	// Check if file already exists
	if _, err := os.Stat(destPath); err == nil {
		sink.Warn(fmt.Sprintf("File already exists: %s", destPath))
		if !noConfirm {
			if !sink.Confirm("Overwrite structure file?") {
				sink.Warn("Structure file write cancelled")
				return nil
			}
		}
	} else {
		sink.Info(fmt.Sprintf("Structure file will be written to: %s", destPath))
		if !noConfirm {
			if !sink.Confirm("Write structure file?") {
				sink.Warn("Structure file write cancelled")
				return nil
			}
		}
	}

	content := strings.Join(structure.Lines, "")
	if err := os.WriteFile(destPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing structure file: %w", err)
	}

	sink.Success(fmt.Sprintf("Structure file written to: %s", destPath))
	return nil
}

// unzipFile extracts a zip archive to the given destination directory.
func unzipFile(src, dest string, sink OutputSink) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("opening zip: %w", err)
	}
	defer r.Close()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	for i, f := range r.File {
		fPath := filepath.Join(dest, f.Name)

		// Prevent zip slip
		if !strings.HasPrefix(filepath.Clean(fPath), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fPath, f.Mode()); err != nil {
				return fmt.Errorf("creating directory: %w", err)
			}
			sink.TaskProgress("Extracting", i+1, len(r.File))
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(fPath), 0755); err != nil {
			return fmt.Errorf("creating parent directory: %w", err)
		}

		outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("creating file: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("opening file in zip: %w", err)
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			return fmt.Errorf("extracting file: %w", err)
		}

		rc.Close()
		outFile.Close()

		sink.TaskProgress("Extracting", i+1, len(r.File))
		sink.Verbose(fmt.Sprintf("Extracted: %s", f.Name))
	}

	return nil
}

// replaceInDir walks the given directory and replaces all occurrences of
// oldStr with newStr in every file.
func replaceInDir(dir, oldStr, newStr string, sink OutputSink) error {
	// First pass: count files
	var total int
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		total++
		return nil
	})

	// Second pass: replace with progress
	current := 0
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		content := string(data)
		if strings.Contains(content, oldStr) {
			replaced := strings.ReplaceAll(content, oldStr, newStr)

			// Clean up any double slashes caused by the replacement (but preserve "://")
			for strings.Contains(replaced, "//") {
				replaced = strings.ReplaceAll(replaced, "://", "::PROTO::")
				replaced = strings.ReplaceAll(replaced, "//", "/")
				replaced = strings.ReplaceAll(replaced, "::PROTO::", "://")
			}

			if err := os.WriteFile(path, []byte(replaced), info.Mode()); err != nil {
				return fmt.Errorf("writing %s: %w", path, err)
			}

			sink.Verbose(fmt.Sprintf("Replaced in: %s", path))
		}

		current++
		sink.TaskProgress("Replacing placeholders", current, total)
		return nil
	})
}

// copyFile copies a file from src to dst, preserving permissions.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source: %w", err)
	}
	defer srcFile.Close()

	info, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copying: %w", err)
	}

	return nil
}

// collectAssets returns a deduplicated list of asset file paths referenced
// in the config (favicon and navbar logo).
func collectAssets(cfg Config) []string {
	seen := map[string]bool{}
	var assets []string

	for _, path := range []string{cfg.Favicon, cfg.Navbar.Logo} {
		if path != "" && !seen[path] {
			seen[path] = true
			assets = append(assets, path)
		}
	}

	return assets
}

// copyMarkdownFiles walks srcDir and copies all .md files into destDir,
// preserving the directory structure relative to srcDir.
func copyMarkdownFiles(srcDir, destDir string, sink OutputSink) error {
	// First pass: count markdown files
	var total int
	_ = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) == ".md" {
			total++
		}
		return nil
	})

	// Second pass: copy with progress
	current := 0
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) != ".md" {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("computing relative path: %w", err)
		}

		dest := filepath.Join(destDir, relPath)
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", dest, err)
		}

		if err := copyFile(path, dest); err != nil {
			return fmt.Errorf("copying %s: %w", relPath, err)
		}

		current++
		sink.TaskProgress("Copying docs", current, total)
		sink.Verbose(fmt.Sprintf("Copied: %s", relPath))
		return nil
	})
}
