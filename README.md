# 🏃 Speedrun CLI

![image](https://github.com/user-attachments/assets/910e96d7-daca-4c25-b986-a5dbebd62356)


A production-ready command-line interface for browsing speedrun.com leaderboards. Search for games, navigate categories, and view detailed run information directly from your terminal.

## ✨ Features

- **🔍 Smart Game Search**: Fuzzy search across speedrun.com's game database
- **📊 Detailed Leaderboards**: View comprehensive run data including times, platforms, videos, and more
- **🎮 Category Navigation**: Browse all categories for any game
- **⌨️  Vim-style Controls**: Familiar navigation with vim-inspired commands
- **🌍 Cross-platform**: Runs on Linux, macOS, and Windows
- **🚀 Zero Dependencies**: Uses only Go standard library
- **📱 Responsive Display**: Clean, compact formatting that works in any terminal

## 🚀 Installation

### Pre-built Binaries

1. Run the build script to generate binaries for all platforms:
```bash
./build.sh
```

2. Copy the appropriate binary for your platform:
```bash
# Linux
sudo cp build/speedrun-cli-linux-amd64 /usr/local/bin/speedrun-cli

# macOS (Intel)
sudo cp build/speedrun-cli-macos-amd64 /usr/local/bin/speedrun-cli

# macOS (Apple Silicon)
sudo cp build/speedrun-cli-macos-arm64 /usr/local/bin/speedrun-cli

# Windows
# Copy build/speedrun-cli-windows-amd64.exe to a directory in your PATH
```

### Build from Source

Requirements: Go 1.18 or later

```bash
git clone git@github.com:marcusziade/speedrun-cli.git
cd speedrun-cli
go build -o speedrun-cli .
```

## 🎮 Usage

### Basic Usage

```bash
speedrun-cli
```

### Navigation Controls

| Command | Action |
|---------|--------|
| `[game name]` | Search for a game |
| `[number]` | Select from numbered lists |
| `q` or `:q` | Quit application |
| `b` or `:b` | Go back to previous menu |
| `r` | Refresh current view |
| `h` or `help` | Show help information |

### Example Workflow

1. **Search for a game**:
   ```
   Enter game name to search: Super Mario Bros
   ```

2. **Select from results**:
   ```
   Found 5 games:
   1. Super Mario Bros. (smb1) - 1985
   2. Super Mario Bros.: The Lost Levels (smb2j) - 1986
   3. Super Mario Bros. 2 (smb2) - 1988
   4. Super Mario Bros. 3 (smb3) - 1988
   5. Super Mario Bros. 35 (smb35) - 2020
   
   Enter number (1-5): 1
   ```

3. **Choose a category**:
   ```
   Categories:
   1. Any% (fullgame)
   2. Any% Warpless (fullgame)
   3. 8-4 IL (level)
   4. All Levels (level)
   
   Enter number (1-4): 1
   ```

4. **View the leaderboard**:
   ```
   🏆 Super Mario Bros. - Any%
   📊 https://www.speedrun.com/smb1#Any
   
   #    Player             Time         Platform     Date       Video  Emu    Comment
   ────────────────────────────────────────────────────────────────────────────────────
   🥇   niftski            4:54.798     NES          2021-04-15 ✅     ❌     World Record!
   🥈   Kosmic             4:55.230     NES          2018-12-04 ✅     ❌     -
   🥉   somewes            4:56.245     NES          2016-01-05 ✅     ❌     -
   ```

## 🛠️ Development

### Project Structure

```
speedrun-cli/
├── main.go          # Main application code
├── build.sh         # Cross-platform build script
├── go.mod           # Go module definition
└── README.md        # This file
```

### API Integration

The application uses the official speedrun.com REST API:
- **Base URL**: `https://www.speedrun.com/api/v1`
- **Authentication**: Not required for read-only operations
- **Rate Limiting**: Respects API rate limits
- **Documentation**: [speedrun.com API docs](https://github.com/speedruncomorg/api)

### Key Features Implementation

- **Game Search**: Uses `/games?name=query` with fuzzy matching
- **Categories**: Fetches via `/games/{id}/categories`
- **Leaderboards**: Retrieved from `/leaderboards/{game}/category/{category}`
- **Time Parsing**: Handles multiple time formats (PT format, seconds)
- **Cross-platform**: Pure Go standard library, no external dependencies

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is open source and available under the [MIT License](LICENSE).

## 🙏 Acknowledgments

- [speedrun.com](https://speedrun.com) for providing the excellent API
- The speedrunning community for creating amazing content

---

**Note**: This is an unofficial tool and is not affiliated with speedrun.com.
