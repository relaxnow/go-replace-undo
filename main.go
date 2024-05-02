package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go /path/to/go.mod")
		return
	}

	modFile := os.Args[1]

	// Get the directory containing the modFile
	dir := filepath.Dir(modFile)

	// Set the vendor folder path in the same directory as the modFile
	vendorDir := filepath.Join(dir, "vendor")

	// Create vendor folder if it doesn't exist
	if err := os.MkdirAll(vendorDir, 0755); err != nil {
		fmt.Printf("Error creating vendor folder: %v\n", err)
		return
	}

	// Run 'go mod vendor' in the directory containing the go.mod file
	cmd := exec.Command("go", "mod", "vendor")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running 'go mod vendor': %v\n", err)
		return
	}

	fmt.Println("Vendoring completed successfully.")

	// Move replaced directories
	copied, err := moveReplaced(modFile, vendorDir)
	if err != nil {
		fmt.Printf("Error moving replaced directories: %v\n", err)
		return
	}
	fmt.Printf("Replaced directories moved successfully: %d\n", copied)
}

// moveReplaced copies directories specified in 'replace' directives to the vendor directory
func moveReplaced(modFile, vendorDir string) (int, error) {
	// Get the directory containing the modFile
	dir := filepath.Dir(modFile)
	// Read the content of go.mod file
	modContent, err := os.ReadFile(modFile)
	if err != nil {
		return 0, fmt.Errorf("error reading %s: %v", modFile, err)
	}

	// Parse the content of go.mod file
	f, err := modfile.Parse(modFile, modContent, nil)
	if err != nil {
		return 0, fmt.Errorf("error parsing %s: %v", modFile, err)
	}

	// Counter for replaced directories moved
	moved := 0

	// Iterate over the replace directives
	for _, r := range f.Replace {
		source := r.New.Path
		destination := r.Old.Path

		// Skip replace of current directory, which is unnecessary if go modules are used
		if source == "./" {
			continue
		}

		// Skip if the destination is not a relative path
		if !strings.HasPrefix(source, "./") {
			continue
		}

		sourcePath := filepath.Join(dir, source)
		destPath := filepath.Join(vendorDir, destination)

		// Create the destination directory in vendor if it doesn't exist
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return moved, fmt.Errorf("error creating %s: %v", destPath, err)
			}
		}

		// Copy the directory
		if err := moveDir(sourcePath, destPath); err != nil {
			return moved, fmt.Errorf("error moving %s to %s: %v", sourcePath, destPath, err)
		}

		relDestPath, err := filepath.Rel(dir, destPath)
		if err != nil {
			return moved, err
		}
		fmt.Printf("Moved %s to %s\n", source, relDestPath)
		moved++
	}
	return moved, nil
}

// moveDir moves a directory and its contents recursively
func moveDir(src, dest string) error {
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Check if the base of the path is "vendor"
		base := filepath.Base(relPath)
		if base == "vendor" {
			fmt.Printf("Skipping %s, vendor files are not moved\n", path)
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		destPath := filepath.Join(dest, relPath)

		// If it's a directory, create it in destination
		if info.IsDir() {
			if err := os.MkdirAll(destPath, info.Mode()); err != nil {
				return err
			}
		} else {
			// If the file already exists at the destination, remove it
			if _, err := os.Stat(destPath); err == nil {
				if err := os.Remove(destPath); err != nil {
					return err
				}
			}

			// Move the file
			if err := os.Rename(path, destPath); err != nil {
				return err
			}
			fmt.Printf("Moved %s to %s\n", path, destPath)
		}
		return nil
	})
	return err
}
