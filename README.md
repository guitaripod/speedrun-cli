# 🏃 Speedrun CLI

[![Go Version](https://img.shields.io/github/go-mod/go-version/marcusziade/speedrun-cli)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/marcusziade/speedrun-cli)](https://goreportcard.com/report/github.com/marcusziade/speedrun-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/marcusziade/speedrun-cli)](https://github.com/marcusziade/speedrun-cli/releases)
[![Downloads](https://img.shields.io/github/downloads/marcusziade/speedrun-cli/total)](https://github.com/marcusziade/speedrun-cli/releases)
[![GitHub Stars](https://img.shields.io/github/stars/marcusziade/speedrun-cli?style=social)](https://github.com/marcusziade/speedrun-cli/stargazers)
[![GitHub Issues](https://img.shields.io/github/issues/marcusziade/speedrun-cli)](https://github.com/marcusziade/speedrun-cli/issues)
[![GitHub Pull Requests](https://img.shields.io/github/issues-pr/marcusziade/speedrun-cli)](https://github.com/marcusziade/speedrun-cli/pulls)
[![Contributors](https://img.shields.io/github/contributors/marcusziade/speedrun-cli)](https://github.com/marcusziade/speedrun-cli/graphs/contributors)
[![Last Commit](https://img.shields.io/github/last-commit/marcusziade/speedrun-cli)](https://github.com/marcusziade/speedrun-cli/commits/master)
[![Code Size](https://img.shields.io/github/languages/code-size/marcusziade/speedrun-cli)](https://github.com/marcusziade/speedrun-cli)

![image](https://github.com/user-attachments/assets/d5d7322a-8b86-45e4-8ac2-a5b1949f3510)


A production-ready command-line interface for browsing speedrun.com leaderboards. Search for games, navigate categories, and view detailed run information directly from your terminal.

## ✨ Features

- **🔍 Smart Game Search**: Fuzzy search across speedrun.com's game database
- **👤 User Search**: Search for users and their runs
- **📊 Detailed Leaderboards**: View comprehensive run data including times, platforms, videos, and more
- **🎮 Category Navigation**: Browse all categories for any game
- **⌨️  Vim-style Controls**: Familiar navigation with vim-inspired commands
- **🌍 Cross-platform**: Runs on Linux, macOS, and Windows
- **🚀 Zero Dependencies**: Uses only Go standard library
- **📱 Responsive Display**: Clean, compact formatting that works in any terminal
- **📄 Leaderboard Pagination**: Navigate large leaderboards with 25 entries per page

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
| `u` | Search for users |
| `[number]` | Select from numbered lists |
| `q` or `:q` | Quit application |
| `b` or `:b` | Go back to previous menu |
| `c` or `:c` | Go back to categories (from leaderboard) |
| `r` | Refresh current view |
| `n` or `next` | Next page (in leaderboards) |
| `p` or `prev` | Previous page (in leaderboards) |
| `p[number]` | Jump to specific page (e.g., `p3` for page 3) |
| `h` or `help` | Show help information |

### Example Workflow

1. **Search for a game**:
   ```
   Enter game name to search (or 'u' for user search, 'q' to quit): ffx
   ```

2. **Select from results**:
   ```
   Found 7 games:
   1. FFX Runner (ffx_runner) - 2006
   2. Final Fantasy X (ffx) - 2001
   3. Final Fantasy XV (ffxv) - 2016
   4. Final Fantasy X-2 (ffx_2) - 2003
   5. Final Fantasy XIV: Dawntrail (ffxiv) - 2013
   6. Final Fantasy XV: Pocket Edition (ffxv_pocket) - 2018
   7. Final Fantasy XII: Revenant Wings (ffxiirw) - 2007
   
   Enter number (1-7), 'q' to quit: 2
   ```

3. **Choose a platform category**:
   ```
   Loading platform categories for Final Fantasy X...
   
   Categories:
   1. PS2 (per-game)
   2. HD Console (per-game)
   3. PC (per-game)
   4. Cutscene Remover (per-game)
   
   Enter number (1-4), 'q' to quit, 'b' to go back: 1
   ```

4. **Select a subcategory**:
   ```
   Loading subcategories for Final Fantasy X - PS2...
   
   Subcategories:
   1. JP Any%
   2. No Sphere Grid
   3. Any%
   4. Nemesis
   
   Enter number (1-4), 'q' to quit, 'b' to go back: 3
   ```

5. **View the leaderboard**:
   ```
   🏆 Final Fantasy X - PS2
   📊 https://www.speedrun.com/ffx#PS2
   
   Rank Player               Time            Platform        Date       Video Emu Comment
   ────────────────────────────────────────────────────────────────────────────────────────────────────
   🥇1  CaracarnVi           9:32:31.000     PlayStation 2   2025-01-19 ✅     ❌   Got all the skips this...
   🥈2  oddmog               9:37:58.000     PlayStation 2   2024-10-01 ✅     ❌   yeah, we are just not ...
   
   📈 Page 1/3 (Showing 1-25 of 67 runs)
   
   Controls: 'n' next page, 'p1-p3' jump to page, 'b' back, 'c' categories, 'q' quit, 'r' refresh
   ```

### User Search Workflow

1. **Search for users**:
   ```
   Enter game name to search (or 'u' for user search, 'q' to quit): u
   ```

2. **Enter username**:
   ```
   Enter username to search (or 'b' to go back): speedrunner123
   ```

3. **Select from user results**:
   ```
   Found 3 users:
   1. speedrunner123 (John Doe)
   2. speedrunner123_alt (John D.)
   3. speedrunner1234 (Jane Smith)
   
   Enter number (1-3), 'q' to quit, 'b' to go back: 1
   ```

4. **View user's runs**:
   ```
   👤 User: speedrunner123 (John Doe)
   🏃 Recent Runs:
   
   Game                     Category        Time         Rank  Date       Platform
   ────────────────────────────────────────────────────────────────────────────────
   Super Mario 64          Any%            16:12.450    #42   2025-01-15 Nintendo 64
   The Legend of Zelda     Any%            31:45.230    #18   2025-01-10 NES
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
- **User Search**: Uses `/users?name=query` to find users by username
- User Runs: Fetches recents via `/users/{id}/personal-bests`
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
