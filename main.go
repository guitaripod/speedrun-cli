package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	apiBase = "https://www.speedrun.com/api/v1"
	userAgent = "speedrun-cli/1.0"
)

type Game struct {
	ID           string `json:"id"`
	Names        struct {
		International string `json:"international"`
		Japanese      string `json:"japanese"`
	} `json:"names"`
	Abbreviation string `json:"abbreviation"`
	Released     int    `json:"released"`
	Weblink      string `json:"weblink"`
}

type Category struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Rules  string `json:"rules"`
	Weblink string `json:"weblink"`
}

type Run struct {
	ID       string `json:"id"`
	Weblink  string `json:"weblink"`
	Game     string `json:"game"`
	Category string `json:"category"`
	Date     string `json:"date"`
	Submitted time.Time `json:"submitted"`
	Times    struct {
		Primary        string `json:"primary"`
		Realtime       string `json:"realtime"`
		RealtimeNoLoads string `json:"realtime_noloads"`
		Ingame         string `json:"ingame"`
	} `json:"times"`
	Players []struct {
		Rel  string `json:"rel"`
		ID   string `json:"id"`
		Name string `json:"name"`
		URI  string `json:"uri"`
	} `json:"players"`
	System struct {
		Platform  string `json:"platform"`
		Emulated  bool   `json:"emulated"`
		Region    string `json:"region"`
	} `json:"system"`
	Status struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	} `json:"status"`
	Videos struct {
		Text  string `json:"text"`
		Links []struct {
			URI string `json:"uri"`
		} `json:"links"`
	} `json:"videos"`
	Comment string `json:"comment"`
}

type Leaderboard struct {
	Weblink string `json:"weblink"`
	Game    struct {
		Data Game `json:"data"`
	} `json:"game"`
	Category struct {
		Data Category `json:"data"`
	} `json:"category"`
	Runs []struct {
		Place int `json:"place"`
		Run   Run `json:"run"`
	} `json:"runs"`
	PlatformMap map[string]string `json:"-"`
}

type APIResponse struct {
	Data       json.RawMessage `json:"data"`
	Pagination struct {
		Offset int `json:"offset"`
		Max    int `json:"max"`
		Size   int `json:"size"`
		Links  []struct {
			Rel string `json:"rel"`
			URI string `json:"uri"`
		} `json:"links"`
	} `json:"pagination"`
}

type User struct {
	ID    string `json:"id"`
	Names struct {
		International string `json:"international"`
		Japanese      string `json:"japanese"`
	} `json:"names"`
}

type Platform struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Released int    `json:"released"`
}

type Region struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Variable struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Values   struct {
		Values map[string]struct {
			Label string `json:"label"`
			Rules string `json:"rules"`
		} `json:"values"`
	} `json:"values"`
	IsSubcategory bool `json:"is-subcategory"`
}

type SubCategory struct {
	ID    string
	Label string
	Rules string
}

func makeRequest(endpoint string) (*http.Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", apiBase+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return client.Do(req)
}

func searchGames(query string) ([]Game, error) {
	encodedQuery := url.QueryEscape(query)
	resp, err := makeRequest(fmt.Sprintf("/games?name=%s&max=20&embed=categories", encodedQuery))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var games []Game
	if err := json.Unmarshal(apiResp.Data, &games); err != nil {
		return nil, err
	}

	return games, nil
}

func getGameCategories(gameID string) ([]Category, error) {
	resp, err := makeRequest(fmt.Sprintf("/games/%s/categories", gameID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var categories []Category
	if err := json.Unmarshal(apiResp.Data, &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

func getGameCategoriesForPlatform(gameID, platformID string) ([]Category, error) {
	// First get all categories
	allCategories, err := getGameCategories(gameID)
	if err != nil {
		return nil, err
	}

	// Filter categories that have runs for this platform
	var validCategories []Category
	for _, category := range allCategories {
		// Try to get leaderboard for this category+platform combination
		// If it succeeds and has any data structure (even with 0 runs), include it
		if hasRunsForPlatform(gameID, category.ID, platformID) {
			validCategories = append(validCategories, category)
		}
	}

	return validCategories, nil
}

func hasRunsForPlatform(gameID, categoryID, platformID string) bool {
	endpoint := fmt.Sprintf("/leaderboards/%s/category/%s?platform=%s", gameID, categoryID, platformID)
	resp, err := makeRequest(endpoint)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// If we get a 200 response, the combination is valid
	// (even if there are 0 runs, the API will return a valid structure)
	return resp.StatusCode == 200
}

func getPlatformsForCategory(gameID, categoryID string) ([]Platform, error) {
	// First get all platforms for the game
	allPlatforms, err := getGamePlatforms(gameID)
	if err != nil {
		return nil, err
	}

	// Filter platforms that have runs for this category
	var validPlatforms []Platform
	for _, platform := range allPlatforms {
		// Try to get leaderboard for this category+platform combination
		if hasRunsForPlatform(gameID, categoryID, platform.ID) {
			validPlatforms = append(validPlatforms, platform)
		}
	}

	return validPlatforms, nil
}

func getCategoryVariables(categoryID string) ([]SubCategory, error) {
	resp, err := makeRequest(fmt.Sprintf("/categories/%s/variables", categoryID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	var variables []Variable
	if err := json.Unmarshal(apiResp.Data, &variables); err != nil {
		return nil, err
	}

	var subCategories []SubCategory
	for _, variable := range variables {
		if variable.IsSubcategory {
			for valueID, value := range variable.Values.Values {
				subCategories = append(subCategories, SubCategory{
					ID:    valueID,
					Label: value.Label,
					Rules: value.Rules,
				})
			}
		}
	}

	return subCategories, nil
}

func getVariableIDForCategory(categoryID string) string {
	// This is a simplified approach - get the first subcategory variable ID
	resp, err := makeRequest(fmt.Sprintf("/categories/%s/variables", categoryID))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return ""
	}

	var variables []Variable
	if err := json.Unmarshal(apiResp.Data, &variables); err != nil {
		return ""
	}

	for _, variable := range variables {
		if variable.IsSubcategory {
			return variable.ID
		}
	}
	return ""
}

func getGamePlatforms(gameID string) ([]Platform, error) {
	resp, err := makeRequest(fmt.Sprintf("/games/%s?embed=platforms", gameID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp struct {
		Data struct {
			Platforms struct {
				Data []Platform `json:"data"`
			} `json:"platforms"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	return apiResp.Data.Platforms.Data, nil
}

func getLeaderboard(gameID, categoryID string, platformID string, variableID string) (*Leaderboard, error) {
	var endpoint string
	queryParams := "embed=game,category,players,platforms,regions"
	
	if platformID != "" {
		queryParams += "&platform=" + platformID
	}
	
	if variableID != "" {
		// Need to find the variable name for this category to build the query
		variables, err := getCategoryVariables(categoryID)
		if err == nil && len(variables) > 0 {
			// For now, assume the first variable is the subcategory variable
			// In a more robust implementation, we'd match the variable properly
			queryParams += "&var-" + getVariableIDForCategory(categoryID) + "=" + variableID
		}
	}
	
	endpoint = fmt.Sprintf("/leaderboards/%s/category/%s?%s", gameID, categoryID, queryParams)
	resp, err := makeRequest(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp struct {
		Data struct {
			Weblink string `json:"weblink"`
			Game    struct {
				Data Game `json:"data"`
			} `json:"game"`
			Category struct {
				Data Category `json:"data"`
			} `json:"category"`
			Runs []struct {
				Place int `json:"place"`
				Run   Run `json:"run"`
			} `json:"runs"`
			Platforms struct {
				Data []Platform `json:"data"`
			} `json:"platforms"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	// Create platform lookup map
	platformMap := make(map[string]string)
	for _, platform := range apiResp.Data.Platforms.Data {
		platformMap[platform.ID] = platform.Name
	}

	leaderboard := &Leaderboard{
		Weblink:  apiResp.Data.Weblink,
		Game:     apiResp.Data.Game,
		Category: apiResp.Data.Category,
		Runs:     apiResp.Data.Runs,
	}

	// Store platform map for later use
	leaderboard.PlatformMap = platformMap

	return leaderboard, nil
}

func formatTime(timeStr string) string {
	if timeStr == "" || timeStr == "null" {
		return "N/A"
	}
	
	// Parse PT format (e.g., "PT1H23M45.678S")
	if strings.HasPrefix(timeStr, "PT") {
		return parsePTFormat(timeStr)
	}
	
	// Try parsing as float seconds
	if seconds, err := strconv.ParseFloat(timeStr, 64); err == nil {
		return formatSeconds(seconds)
	}
	
	return timeStr
}

func parsePTFormat(ptTime string) string {
	ptTime = strings.TrimPrefix(ptTime, "PT")
	
	var hours, minutes float64
	var seconds float64
	
	// Parse hours
	if hIdx := strings.Index(ptTime, "H"); hIdx != -1 {
		if h, err := strconv.ParseFloat(ptTime[:hIdx], 64); err == nil {
			hours = h
		}
		ptTime = ptTime[hIdx+1:]
	}
	
	// Parse minutes
	if mIdx := strings.Index(ptTime, "M"); mIdx != -1 {
		if m, err := strconv.ParseFloat(ptTime[:mIdx], 64); err == nil {
			minutes = m
		}
		ptTime = ptTime[mIdx+1:]
	}
	
	// Parse seconds
	if sIdx := strings.Index(ptTime, "S"); sIdx != -1 {
		if s, err := strconv.ParseFloat(ptTime[:sIdx], 64); err == nil {
			seconds = s
		}
	}
	
	totalSeconds := hours*3600 + minutes*60 + seconds
	return formatSeconds(totalSeconds)
}

func formatSeconds(totalSeconds float64) string {
	hours := int(totalSeconds) / 3600
	minutes := (int(totalSeconds) % 3600) / 60
	seconds := totalSeconds - float64(hours*3600) - float64(minutes*60)
	
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%06.3f", hours, minutes, seconds)
	}
	if minutes > 0 || totalSeconds >= 60 {
		return fmt.Sprintf("%d:%06.3f", minutes, seconds)
	}
	return fmt.Sprintf("%.3f", seconds)
}

func getPlayerName(player struct {
	Rel  string `json:"rel"`
	ID   string `json:"id"`
	Name string `json:"name"`
	URI  string `json:"uri"`
}) string {
	if player.Name != "" {
		return player.Name
	}
	
	// If we have an ID but no name, try to fetch the user data
	if player.ID != "" {
		if userData := fetchUserData(player.ID); userData != nil {
			if userData.Names.International != "" {
				return userData.Names.International
			}
		}
	}
	
	return "Guest"
}

func fetchUserData(userID string) *struct {
	Names struct {
		International string `json:"international"`
	} `json:"names"`
} {
	resp, err := makeRequest(fmt.Sprintf("/users/%s", userID))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var response struct {
		Data struct {
			Names struct {
				International string `json:"international"`
			} `json:"names"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil
	}

	return &response.Data
}

func displayLeaderboard(lb *Leaderboard) {
	fmt.Printf("\n🏆 %s - %s\n", lb.Game.Data.Names.International, lb.Category.Data.Name)
	fmt.Printf("📊 %s\n\n", lb.Weblink)
	
	if len(lb.Runs) == 0 {
		fmt.Println("No runs found for this category.")
		return
	}
	
	// Use fixed-width formatting for proper alignment
	fmt.Printf("%-5s%-20s %-15s %-15s %-10s %-5s %-3s %s\n", 
		"Rank", "Player", "Time", "Platform", "Date", "Video", "Emu", "Comment")
	fmt.Println(strings.Repeat("─", 100))
	
	for _, entry := range lb.Runs {
		playerName := "Guest"
		if len(entry.Run.Players) > 0 {
			playerName = getPlayerName(entry.Run.Players[0])
		}
		
		// Try to get the best available time
		time := formatTime(entry.Run.Times.Primary)
		if time == "N/A" || time == "" {
			time = formatTime(entry.Run.Times.Realtime)
		}
		if time == "N/A" || time == "" {
			time = formatTime(entry.Run.Times.RealtimeNoLoads)
		}
		if time == "N/A" || time == "" {
			time = formatTime(entry.Run.Times.Ingame)
		}
		
		// Get platform name from the map, fallback to ID if not found
		platform := "Unknown"
		if entry.Run.System.Platform != "" {
			if platformName, exists := lb.PlatformMap[entry.Run.System.Platform]; exists {
				platform = platformName
			} else {
				platform = entry.Run.System.Platform
			}
		}
		
		hasVideo := "❌"
		if len(entry.Run.Videos.Links) > 0 && entry.Run.Videos.Links[0].URI != "" {
			hasVideo = "✅"
		}
		
		emulated := "❌"
		if entry.Run.System.Emulated {
			emulated = "✅"
		}
		
		// Clean comment first, then truncate
		comment := entry.Run.Comment
		if comment == "" {
			comment = "-"
		} else {
			// Clean all control characters first
			comment = strings.ReplaceAll(comment, "\n", " ")
			comment = strings.ReplaceAll(comment, "\r", " ")
			comment = strings.ReplaceAll(comment, "\t", " ")
			comment = strings.TrimSpace(comment)
		}
		
		// Handle medals vs numbers with proper alignment
		switch entry.Place {
		case 1:
			// Medal entries need more space compensation since emoji takes visual width
			fmt.Printf("🥇1  %-20s %-15s %-15s %-10s %-5s %-3s %s\n",
				truncateString(playerName, 20),
				time,
				truncateString(platform, 15),
				entry.Run.Date,
				hasVideo,
				emulated,
				truncateString(comment, 25))
		case 2:
			fmt.Printf("🥈2  %-20s %-15s %-15s %-10s %-5s %-3s %s\n",
				truncateString(playerName, 20),
				time,
				truncateString(platform, 15),
				entry.Run.Date,
				hasVideo,
				emulated,
				truncateString(comment, 25))
		case 3:
			fmt.Printf("🥉3  %-20s %-15s %-15s %-10s %-5s %-3s %s\n",
				truncateString(playerName, 20),
				time,
				truncateString(platform, 15),
				entry.Run.Date,
				hasVideo,
				emulated,
				truncateString(comment, 25))
		default:
			// Regular numeric entries
			fmt.Printf("%-5s%-20s %-15s %-15s %-10s %-5s %-3s %s\n",
				fmt.Sprintf("%d", entry.Place),
				truncateString(playerName, 20),
				time,
				truncateString(platform, 15),
				entry.Run.Date,
				hasVideo,
				emulated,
				truncateString(comment, 25))
		}
	}
	
	fmt.Printf("\n📈 Showing %d runs\n", len(lb.Runs))
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func getUserInput(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func selectGame(games []Game) *Game {
	if len(games) == 0 {
		fmt.Println("No games found.")
		return nil
	}
	
	if len(games) == 1 {
		fmt.Printf("Found exact match: %s\n", games[0].Names.International)
		return &games[0]
	}
	
	fmt.Printf("\nFound %d games:\n", len(games))
	for i, game := range games {
		fmt.Printf("%d. %s (%s) - %d\n", i+1, game.Names.International, game.Abbreviation, game.Released)
	}
	
	for {
		input := getUserInput("\nEnter number (1-" + strconv.Itoa(len(games)) + "), 'q' to quit: ")
		
		if input == "q" || input == "quit" {
			return nil
		}
		
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(games) {
			fmt.Println("Invalid selection. Please try again.")
			continue
		}
		
		return &games[choice-1]
	}
}

func selectPlatform(platforms []Platform) *Platform {
	if len(platforms) == 0 {
		fmt.Println("No platforms found.")
		return nil
	}
	
	fmt.Printf("\nPlatforms:\n")
	for i, platform := range platforms {
		fmt.Printf("%d. %s (%d)\n", i+1, platform.Name, platform.Released)
	}
	
	for {
		input := getUserInput("\nEnter number (1-" + strconv.Itoa(len(platforms)) + "), 'q' to quit, 'b' to go back: ")
		
		if input == "q" || input == "quit" || input == ":q" {
			return nil
		}
		
		if input == "b" || input == "back" || input == ":b" {
			return &Platform{ID: "BACK"}
		}
		
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(platforms) {
			fmt.Println("Invalid selection. Please try again.")
			fmt.Println("Controls: [number] select, 'q' quit, 'b' back")
			continue
		}
		
		return &platforms[choice-1]
	}
}

func selectCategory(categories []Category) *Category {
	if len(categories) == 0 {
		fmt.Println("No categories found.")
		return nil
	}
	
	fmt.Printf("\nCategories:\n")
	for i, cat := range categories {
		fmt.Printf("%d. %s (%s)\n", i+1, cat.Name, cat.Type)
	}
	
	for {
		input := getUserInput("\nEnter number (1-" + strconv.Itoa(len(categories)) + "), 'q' to quit, 'b' to go back: ")
		
		if input == "q" || input == "quit" || input == ":q" {
			return nil
		}
		
		if input == "b" || input == "back" || input == ":b" {
			return &Category{ID: "BACK"}
		}
		
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(categories) {
			fmt.Println("Invalid selection. Please try again.")
			fmt.Println("Controls: [number] select, 'q' quit, 'b' back")
			continue
		}
		
		return &categories[choice-1]
	}
}

func selectSubCategory(subCategories []SubCategory) *SubCategory {
	if len(subCategories) == 0 {
		fmt.Println("No subcategories found.")
		return nil
	}
	
	fmt.Printf("\nSubcategories:\n")
	for i, subCat := range subCategories {
		fmt.Printf("%d. %s\n", i+1, subCat.Label)
	}
	
	for {
		input := getUserInput("\nEnter number (1-" + strconv.Itoa(len(subCategories)) + "), 'q' to quit, 'b' to go back: ")
		
		if input == "q" || input == "quit" || input == ":q" {
			return nil
		}
		
		if input == "b" || input == "back" || input == ":b" {
			return &SubCategory{ID: "BACK"}
		}
		
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(subCategories) {
			fmt.Println("Invalid selection. Please try again.")
			fmt.Println("Controls: [number] select, 'q' quit, 'b' back")
			continue
		}
		
		return &subCategories[choice-1]
	}
}



func showHelp() {
	fmt.Println("\n📚 Help - Speedrun.com CLI")
	fmt.Println("============================")
	fmt.Println("Navigation Flow:")
	fmt.Println("  1. Search for a game")
	fmt.Println("  2. Select a platform category (PS2, HD Console, PC, etc.)")
	fmt.Println("  3. Select a subcategory (Any%, 100%, etc.)")
	fmt.Println("  4. View leaderboard")
	fmt.Println("\nControls:")
	fmt.Println("  • Use numbers to select from lists")
	fmt.Println("  • 'q' or ':q' - quit")
	fmt.Println("  • 'b' or ':b' - go back")
	fmt.Println("  • 'c' or ':c' - back to categories (from leaderboard)")
	fmt.Println("  • 'r' - refresh current view")
	fmt.Println("  • 'h' or 'help' - show this help")
	fmt.Println("\nFeatures:")
	fmt.Println("  • Fuzzy game search")
	fmt.Println("  • Platform categories with subcategories")
	fmt.Println("  • Detailed leaderboards with filtering")
	fmt.Println("  • Run times, players, platforms, videos")
	fmt.Println()
}

func main() {
	fmt.Println("🏃 Speedrun.com CLI - Game Leaderboard Browser")
	fmt.Println("==============================================")
	fmt.Println("Type 'h' or 'help' for instructions")
	
	for {
		query := getUserInput("\nEnter game name to search (or 'q' to quit): ")
		
		if query == "q" || query == "quit" || query == ":q" {
			fmt.Println("Goodbye! 👋")
			break
		}
		
		if query == "h" || query == "help" {
			showHelp()
			continue
		}
		
		if query == "" {
			continue
		}
		
		fmt.Println("Searching for games...")
		games, err := searchGames(query)
		if err != nil {
			fmt.Printf("Error searching games: %v\n", err)
			continue
		}
		
		selectedGame := selectGame(games)
		if selectedGame == nil {
			continue
		}
		
		for {
			fmt.Printf("\nLoading platform categories for %s...\n", selectedGame.Names.International)
			categories, err := getGameCategories(selectedGame.ID)
			if err != nil {
				fmt.Printf("Error loading categories: %v\n", err)
				break
			}
			
			selectedCategory := selectCategory(categories)
			if selectedCategory == nil {
				break
			}
			
			if selectedCategory.ID == "BACK" {
				break
			}
			
			for {
				fmt.Printf("\nLoading subcategories for %s - %s...\n", selectedGame.Names.International, selectedCategory.Name)
				subCategories, err := getCategoryVariables(selectedCategory.ID)
				if err != nil {
					fmt.Printf("Error loading subcategories: %v\n", err)
					break
				}
				
				selectedSubCategory := selectSubCategory(subCategories)
				if selectedSubCategory == nil {
					break
				}
				
				if selectedSubCategory.ID == "BACK" {
					break
				}
				
				fmt.Printf("\nLoading leaderboard for %s - %s (%s)...\n", selectedGame.Names.International, selectedCategory.Name, selectedSubCategory.Label)
				leaderboard, err := getLeaderboard(selectedGame.ID, selectedCategory.ID, "", selectedSubCategory.ID)
				if err != nil {
					fmt.Printf("Error loading leaderboard: %v\n", err)
					continue
				}
				
				displayLeaderboard(leaderboard)
				
				fmt.Println("\nControls: [Enter] continue, 'b' back to subcategories, 'c' back to categories, 'q' quit, 'r' refresh")
				input := getUserInput("Action: ")
				
				if input == "q" || input == "quit" || input == ":q" {
					fmt.Println("Goodbye!")
					os.Exit(0)
				}
				if input == "b" || input == "back" || input == ":b" {
					// Go back to subcategory selection
					break
				}
				if input == "c" || input == "category" || input == ":c" {
					// Go back to category selection
					goto categorySelection
				}
				if input == "r" || input == "refresh" {
					// Refresh the current leaderboard
					fmt.Printf("Refreshing leaderboard for %s - %s (%s)...\n", selectedGame.Names.International, selectedCategory.Name, selectedSubCategory.Label)
					leaderboard, err = getLeaderboard(selectedGame.ID, selectedCategory.ID, "", selectedSubCategory.ID)
					if err != nil {
						fmt.Printf("Error refreshing leaderboard: %v\n", err)
						continue
					}
					displayLeaderboard(leaderboard)
					continue
				}
			}
			categorySelection:
		}
	}
}