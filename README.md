# Audio Tool

A command-line tool written in Go for batch processing audio files.

## Features

- 🔍 **Recursive Traversal**: Automatically traverses specified directories and all subdirectories
- 🚫 **Directory Exclusion**: Skip specific directories like `@eaDir`, `.git`, or `Thumbnails`
- 🎵 **File Type Filtering**: Process only specific file types (music, video, etc.)
- 🎯 **Smart Detection**: Automatically detects common prefixes in filenames within the same directory
- 🧠 **Intelligent Matching**: Handles outliers - if 70%+ of files share a prefix, processes only those files
- ✅ **Safe Confirmation**: Provides preview mode and interactive confirmation to avoid mistakes
- 📊 **Detailed Display**: Shows before/after filename comparisons
- ⚡ **Batch Processing**: Supports processing multiple directories at once

## Use Case Example

Suppose you have files like this:

```
01-【Music Collection】Taylor Swift Greatest Hits p01 Shake It Off.m4a
02-【Music Collection】Taylor Swift Greatest Hits p02 Blank Space.m4a
03-【Music Collection】Taylor Swift Greatest Hits p03 Love Story.m4a
```

After running the tool, the common prefix `【Music Collection】Taylor Swift Greatest Hits ` will be removed:

```
01-p01 Shake It Off.m4a
02-p02 Blank Space.m4a
03-p03 Love Story.m4a
```

## Installation

### Option 1: Using go install (Recommended)

```bash
go install github.com/baiyutang/audio-tool@latest
```

### Option 2: Download Pre-built Binaries

Download the appropriate pre-built binary for your system from [GitHub Releases](https://github.com/baiyutang/audio-tool/releases).

### Option 3: Build from Source

```bash
git clone git@github.com:baiyutang/audio-tool.git
cd audio-tool
go build -o audiotool
# Or using Makefile
make build
```

## Usage

### View Help

```bash
# View all available commands
./audiotool help

# View help for a specific command
./audiotool removeprefix -h
```

### Command: removeprefix

Remove common prefixes from filenames. Particularly useful for batch-downloaded music files like playlists.

#### Basic Usage

```bash
# Preview mode (recommended for first use)
./audiotool removeprefix -dir /path/to/music -dry-run

# Execute with interactive confirmation
./audiotool removeprefix -dir /path/to/music

# Auto-confirm all operations
./audiotool removeprefix -dir /path/to/music -y

# Process only music files (mp3, m4a, flac, wav)
./audiotool removeprefix -dir /path/to/music -exts mp3,m4a,flac,wav

# Process music and video files
./audiotool removeprefix -dir /path/to/media -exts mp3,m4a,flac,wav,mp4,mkv,avi

# Exclude specific directories (e.g., @eaDir, .git, Thumbnails)
./audiotool removeprefix -dir /path/to/music -exclude-dirs @eaDir,.git,Thumbnails

# Process current directory
./audiotool removeprefix
```

#### Command-line Options

| Option | Description | Default |
|--------|-------------|---------||
| `-dir` | Directory path to process | `.` (current directory) |
| `-dry-run` | Preview mode, don't actually rename files | `false` |
| `-y` | Auto-confirm all operations without asking | `false` |
| `-exclude-dirs` | Comma-separated list of directory names to exclude (e.g., `@eaDir,.git`) | `@eaDir` |
| `-exts` | Comma-separated list of file extensions to process (e.g., `mp3,m4a,flac,wav,mp4,mkv`) | `` (all files) |

#### Common Use Cases

**Process only audio files:**
```bash
./audiotool removeprefix -dir /path/to/music -exts mp3,m4a,flac,wav,aac,ogg
```

**Process audio and video files:**
```bash
./audiotool removeprefix -dir /path/to/media -exts mp3,m4a,mp4,mkv,avi,mov
```

**Skip system directories (NAS/Synology):**
```bash
./audiotool removeprefix -dir /volume1/music -exclude-dirs @eaDir,#recycle
```

**Combined: music files only, exclude system dirs:**
```bash
./audiotool removeprefix -dir /path/to/music -exts mp3,m4a,flac -exclude-dirs @eaDir,.git,Thumbnails -dry-run
```

#### Recommended Workflow

1. **Preview Stage**: Use `-dry-run` to see the effect first
   ```bash
   ./audiotool removeprefix -dir /path/to/music -dry-run
   ```

2. **Execute After Confirmation**: If preview looks good, remove `-dry-run`
   ```bash
   ./audiotool removeprefix -dir /path/to/music
   ```

3. **Batch Auto Mode**: If processing many directories and confident, use `-y`
   ```bash
   ./audiotool removeprefix -dir /path/to/music -y
   ```

## How It Works

1. Recursively scans specified directory and all subdirectories (excluding specified directories)
2. Filters files by extension if specified
3. Groups files by directory
4. Finds common prefix for filenames in each directory
5. **Smart majority matching**: If not all files share a prefix, finds prefix that matches at least 70% of files
6. Smart prefix trimming (cuts at separators to avoid breaking filename structure)
7. Displays rename preview with match statistics
8. Asks for user confirmation (unless `-y` is used)
9. Executes rename operations only on matching files

## Important Notes

⚠️ **Important**:

- **Always test first**: Use `-dry-run` mode or test in a safe directory before processing important files
- **Smart outlier handling**: Tool automatically handles cases where some files don't match the common pattern (e.g., collaboration songs with different artist names)
- **Majority rule**: If 70%+ of files share a prefix, only those files will be processed while outliers are skipped
- **Directory exclusion**: By default, `@eaDir` directories are excluded. Customize with `-exclude-dirs` for other system folders
- **File filtering**: Use `-exts` to process only specific file types (recommended for mixed-content directories)
- **Minimum prefix length**: Tool skips cases where common prefix is less than 3 characters
- **Empty filename protection**: Files will be skipped if removing the prefix results in an empty filename
- **Backup recommendation**: Backup important files before processing
- **Minimum file count**: At least 2 files are needed in the same directory for processing

## Example Output

```
$ ./audiotool removeprefix -dir /path/to/music -exts mp3,m4a -dry-run

Processing directory: /path/to/music
Mode: Preview mode (files will not be modified)
Excluding directories: [@eaDir]
Processing only files with extensions: [mp3 m4a]

Found 58 files in total
Involving 1 directory

Directory: /path/to/music
Common prefix found: 【Music Collection】Taylor Swift Greatest Hits  (length: 52 bytes)
File count: 58
Example filename: 【Music Collection】Taylor Swift Greatest Hits p01 Shake It Off.m4a

Rename preview (showing first 5):
  01-【Music Collection】Taylor Swift Greatest Hits p01 Shake It Off.m4a
  -> 01-p01 Shake It Off.m4a

  02-【Music Collection】Taylor Swift Greatest Hits p02 Blank Space.m4a
  -> 02-p02 Blank Space.m4a

  ... and 53 more files

[Preview Mode] No actual renaming performed

Processing complete!
```


## Extending with New Features

The project uses a subcommand architecture, making it easy to add new features:

1. Add a new command handler function in `main.go`
2. Register the new command in the `main()` switch statement
3. Update the command list in `printUsage()`

## License

MIT

## Repository

GitHub: [baiyutang/audio-tool](https://github.com/baiyutang/audio-tool)
