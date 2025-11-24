// Audio Tool - Batch audio file processing tool
// Copyright (c) 2025 baiyutang
// Licensed under the MIT License

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const version = "1.0.0"

// findCommonPrefix finds the longest common prefix among all strings
func findCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return ""
	}

	// Find the shortest string length
	minLen := len(strs[0])
	for _, s := range strs {
		if len(s) < minLen {
			minLen = len(s)
		}
	}

	// Byte-by-byte comparison to find common prefix
	prefixLen := 0
	for i := 0; i < minLen; i++ {
		char := strs[0][i]
		allMatch := true
		for _, s := range strs {
			if s[i] != char {
				allMatch = false
				break
			}
		}
		if allMatch {
			prefixLen = i + 1
		} else {
			break
		}
	}

	if prefixLen == 0 {
		return ""
	}

	// Extract raw byte prefix
	prefix := strs[0][:prefixLen]

	// Smart trimming: ensure cutting at separator positions
	// Search backwards for the last separator
	lastSep := -1
	for i := len(prefix) - 1; i >= 0; i-- {
		ch := prefix[i]
		if ch == '-' || ch == '_' || ch == ' ' || ch == ')' || ch == ']' {
			lastSep = i + 1
			break
		}
		// Check for Chinese 】 symbol (UTF-8: E3 80 91)
		if i >= 2 && prefix[i-2] == 0xE3 && prefix[i-1] == 0x80 && prefix[i] == 0x91 {
			lastSep = i + 1
			break
		}
	}

	if lastSep > 0 && lastSep < len(prefix) {
		return prefix[:lastSep]
	}

	return prefix
}

// findMajorityPrefix finds common prefix for the majority of files (at least 70%)
// This helps handle cases where a few outlier files don't share the common prefix
func findMajorityPrefix(strs []string) string {
	if len(strs) < 2 {
		return ""
	}

	// Try to find a prefix that works for at least 70% of files
	threshold := int(float64(len(strs)) * 0.7)
	if threshold < 2 {
		threshold = 2
	}

	// Try each file as a potential prefix source
	bestPrefix := ""
	bestMatchCount := 0

	for _, sourceFile := range strs {
		// Try different prefix lengths from this file
		for length := len(sourceFile); length >= 3; length-- {
			potentialPrefix := sourceFile[:length]

			// Count how many files have this prefix
			matchCount := 0
			for _, s := range strs {
				if strings.HasPrefix(s, potentialPrefix) {
					matchCount++
				}
			}

			// If this prefix matches more files than our current best
			if matchCount >= threshold && matchCount > bestMatchCount {
				// Find the last separator position
				lastSep := -1
				for i := len(potentialPrefix) - 1; i >= 0; i-- {
					ch := potentialPrefix[i]
					if ch == '-' || ch == '_' || ch == ' ' || ch == ')' || ch == ']' {
						lastSep = i + 1
						break
					}
					// Check for Chinese 】 symbol (UTF-8: E3 80 91)
					if i >= 2 && potentialPrefix[i-2] == 0xE3 && potentialPrefix[i-1] == 0x80 && potentialPrefix[i] == 0x91 {
						lastSep = i + 1
						break
					}
				}

				if lastSep > 0 && lastSep < len(potentialPrefix) {
					finalPrefix := potentialPrefix[:lastSep]
					if len(strings.TrimSpace(finalPrefix)) >= 3 {
						bestPrefix = finalPrefix
						bestMatchCount = matchCount
					}
				}
			}
		}
	}

	return bestPrefix
} // collectFiles recursively collects all files in a directory
// excludeDirs: directories to skip (e.g., @eaDir, .git)
// extensions: allowed file extensions (e.g., .mp3, .m4a); empty means all files
func collectFiles(root string, excludeDirs []string, extensions []string) ([]string, error) {
	var files []string

	// Convert extensions to lowercase for case-insensitive matching
	extMap := make(map[string]bool)
	for _, ext := range extensions {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		extMap[strings.ToLower(ext)] = true
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		if info.IsDir() {
			dirName := filepath.Base(path)
			for _, excludeDir := range excludeDirs {
				if dirName == excludeDir {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Filter by extension if specified
		if len(extMap) > 0 {
			ext := strings.ToLower(filepath.Ext(path))
			if !extMap[ext] {
				return nil
			}
		}

		files = append(files, path)
		return nil
	})
	return files, err
}

// groupFilesByDirectory groups files by their directory
func groupFilesByDirectory(files []string) map[string][]string {
	groups := make(map[string][]string)
	for _, file := range files {
		dir := filepath.Dir(file)
		groups[dir] = append(groups[dir], file)
	}
	return groups
}

// processDirectory processes files in a single directory
func processDirectory(dir string, files []string, dryRun bool, autoYes bool) error {
	if len(files) < 2 {
		return nil // Less than 2 files, no processing needed
	}

	// Extract filenames (without path)
	filenames := make([]string, len(files))
	for i, file := range files {
		filenames[i] = filepath.Base(file)
	}

	// Find common prefix for all files
	prefix := findCommonPrefix(filenames)

	// If prefix is too short, try smart filtering: find prefix for majority of files
	if prefix == "" || len(strings.TrimSpace(prefix)) < 3 {
		prefix = findMajorityPrefix(filenames)
		if prefix == "" || len(strings.TrimSpace(prefix)) < 3 {
			return nil // Still no good prefix, skip processing
		}
	}

	fmt.Printf("\nDirectory: %s\n", dir)
	fmt.Printf("Common prefix found: %s (length: %d bytes)\n", prefix, len(prefix))

	// Count how many files will actually be renamed
	matchCount := 0
	for _, name := range filenames {
		if strings.HasPrefix(name, prefix) {
			matchCount++
		}
	}
	fmt.Printf("File count: %d (matching prefix: %d)\n", len(files), matchCount)

	// Display first filename as example
	if len(filenames) > 0 {
		fmt.Printf("Example filename: %s\n\n", filenames[0])
	} else {
		fmt.Println()
	}

	// Build rename plan
	type RenamePlan struct {
		OldPath string
		NewPath string
		OldName string
		NewName string
	}
	var plans []RenamePlan

	for _, file := range files {
		oldName := filepath.Base(file)
		newName := strings.TrimPrefix(oldName, prefix)
		newName = strings.TrimSpace(newName)

		if newName == "" {
			fmt.Printf("Warning: filename empty after removing prefix, skipping: %s\n", oldName)
			continue
		}

		if oldName != newName {
			newPath := filepath.Join(dir, newName)
			plans = append(plans, RenamePlan{
				OldPath: file,
				NewPath: newPath,
				OldName: oldName,
				NewName: newName,
			})
		}
	}

	if len(plans) == 0 {
		return nil
	}

	// Display first few examples
	fmt.Println("Rename preview (showing first 5):")
	displayCount := 5
	if len(plans) < displayCount {
		displayCount = len(plans)
	}
	for i := 0; i < displayCount; i++ {
		fmt.Printf("  %s\n  -> %s\n\n", plans[i].OldName, plans[i].NewName)
	}
	if len(plans) > displayCount {
		fmt.Printf("  ... and %d more files\n\n", len(plans)-displayCount)
	}

	if dryRun {
		fmt.Println("[Preview Mode] No actual renaming performed")
		return nil
	}

	// Ask for confirmation
	proceed := autoYes
	if !autoYes {
		fmt.Printf("Proceed to rename these %d files? (y/n): ", len(plans))
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		response = strings.ToLower(strings.TrimSpace(response))
		proceed = response == "y" || response == "yes"
	}

	if !proceed {
		fmt.Println("Skipped this directory")
		return nil
	}

	// Execute rename
	successCount := 0
	for _, plan := range plans {
		err := os.Rename(plan.OldPath, plan.NewPath)
		if err != nil {
			fmt.Printf("Error: failed to rename %s: %v\n", plan.OldName, err)
		} else {
			successCount++
		}
	}

	fmt.Printf("Successfully renamed %d/%d files\n", successCount, len(plans))
	return nil
}

// removePrefixCommand is the subcommand to remove common prefix from filenames
func removePrefixCommand(args []string) {
	fs := flag.NewFlagSet("removeprefix", flag.ExitOnError)
	dir := fs.String("dir", ".", "Directory path to process")
	dryRun := fs.Bool("dry-run", false, "Preview mode, don't actually rename files")
	autoYes := fs.Bool("y", false, "Auto-confirm all operations without asking")
	excludeDirs := fs.String("exclude-dirs", "@eaDir", "Comma-separated list of directory names to exclude")
	exts := fs.String("exts", "", "Comma-separated list of file extensions to process (e.g., mp3,m4a,flac,wav,mp4,mkv)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: audiotool removeprefix [options]\n\n")
		fmt.Fprintf(os.Stderr, "Recursively traverse directories and remove common prefixes from filenames\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  audiotool removeprefix -dir /path/to/music -dry-run\n")
		fmt.Fprintf(os.Stderr, "  audiotool removeprefix -dir /path/to/music -y\n")
		fmt.Fprintf(os.Stderr, "  audiotool removeprefix -dir /path/to/music -exts mp3,m4a,flac\n")
		fmt.Fprintf(os.Stderr, "  audiotool removeprefix -dir /path/to/music -exclude-dirs @eaDir,.git\n")
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse flags: %v\n", err)
		os.Exit(1)
	}

	// Get absolute path
	absDir, err := filepath.Abs(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to get absolute path: %v\n", err)
		os.Exit(1)
	}

	// Check if directory exists
	info, err := os.Stat(absDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to access directory %s: %v\n", absDir, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", absDir)
		os.Exit(1)
	}

	fmt.Printf("Processing directory: %s\n", absDir)
	if *dryRun {
		fmt.Println("Mode: Preview mode (files will not be modified)")
	}
	fmt.Println()

	// Parse exclude directories
	var excludeDirList []string
	if *excludeDirs != "" {
		excludeDirList = strings.Split(*excludeDirs, ",")
		for i, dir := range excludeDirList {
			excludeDirList[i] = strings.TrimSpace(dir)
		}
		fmt.Printf("Excluding directories: %v\n", excludeDirList)
	}

	// Parse file extensions
	var extList []string
	if *exts != "" {
		extList = strings.Split(*exts, ",")
		for i, ext := range extList {
			extList[i] = strings.TrimSpace(ext)
		}
		fmt.Printf("Processing only files with extensions: %v\n", extList)
	}

	// Collect all files
	files, err := collectFiles(absDir, excludeDirList, extList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to collect files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No files found")
		return
	}

	fmt.Printf("Found %d files in total\n", len(files))

	// Group by directory
	groups := groupFilesByDirectory(files)
	fmt.Printf("Involving %d directories\n", len(groups))

	// Process each directory
	for dir, dirFiles := range groups {
		err := processDirectory(dir, dirFiles, *dryRun, *autoYes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to process directory %s: %v\n", dir, err)
		}
	}

	fmt.Println("\nProcessing complete!")
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Audio Tool - Batch audio file processing tool v%s\n\n", version)
	fmt.Fprintf(os.Stderr, "Usage: audiotool <command> [options]\n\n")
	fmt.Fprintf(os.Stderr, "Available commands:\n")
	fmt.Fprintf(os.Stderr, "  removeprefix     Remove common prefix from filenames\n")
	fmt.Fprintf(os.Stderr, "  version          Show version information\n")
	fmt.Fprintf(os.Stderr, "  help             Show help information\n")
	fmt.Fprintf(os.Stderr, "\nUse 'audiotool <command> -h' for detailed help on a command\n")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "removeprefix":
		removePrefixCommand(os.Args[2:])
	case "version":
		fmt.Printf("Audio Tool v%s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}
}
