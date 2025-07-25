package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type SpeedrunAPI struct {
	client    *http.Client
	userCache map[string]*User
	cacheMux  sync.RWMutex
}

func NewSpeedrunAPI() *SpeedrunAPI {
	return &SpeedrunAPI{
		client: &http.Client{
			Timeout: 0, // No timeout on client, we'll handle it with context
		},
		userCache: make(map[string]*User),
	}
}

func (api *SpeedrunAPI) makeRequest(endpoint string) ([]byte, error) {
	return api.makeRequestWithRetry(endpoint, MaxRetries)
}

func (api *SpeedrunAPI) makeRequestWithRetry(endpoint string, retries int) ([]byte, error) {
	var lastErr error
	
	for attempt := 0; attempt <= retries; attempt++ {
		if attempt > 0 {
			backoffDuration := time.Duration(BackoffBase<<(attempt-1)) * time.Second
			debugLog("Retrying request to %s after %v (attempt %d/%d)", endpoint, backoffDuration, attempt+1, retries+1)
			time.Sleep(backoffDuration)
		}
		
		req, err := http.NewRequest("GET", APIBase+endpoint, nil)
		if err != nil {
			return nil, &APIError{
				Message:    fmt.Sprintf("failed to create request: %v", err),
				StatusCode: 0,
				URL:        APIBase + endpoint,
				Context:    "request creation",
			}
		}
		
		req.Header.Set("User-Agent", UserAgent)
		
		ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
		defer cancel()
		req = req.WithContext(ctx)
		
		resp, err := api.client.Do(req)
		
		if err != nil {
			lastErr = &APIError{
				Message:    fmt.Sprintf("request failed: %v", err),
				StatusCode: 0,
				URL:        APIBase + endpoint,
				Context:    "network error",
			}
			continue
		}
		
		if resp.StatusCode == 429 && attempt < retries {
			resp.Body.Close()
			debugLog("Rate limited (429), retrying...")
			continue
		}
		
		if resp.StatusCode >= 500 && attempt < retries {
			resp.Body.Close()
			debugLog("Server error (%d), retrying...", resp.StatusCode)
			continue
		}
		
		if resp.StatusCode != 200 {
			resp.Body.Close()
			return nil, &APIError{
				Message:    fmt.Sprintf("API request failed with status %d", resp.StatusCode),
				StatusCode: resp.StatusCode,
				URL:        APIBase + endpoint,
				Context:    "API response",
			}
		}
		
		// Read the response body before the context expires
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		if err != nil {
			lastErr = &APIError{
				Message: fmt.Sprintf("failed to read response body: %v", err),
				Context: "response reading",
			}
			continue
		}
		
		return body, nil
	}
	
	return nil, lastErr
}

func (api *SpeedrunAPI) SearchGames(query string) ([]Game, error) {
	debugLog("Searching for games with query: %s", query)
	
	encodedQuery := url.QueryEscape(query)
	fmt.Print("🔍 Searching for games...")
	body, err := api.makeRequest(fmt.Sprintf("/games?name=%s&max=20&embed=categories", encodedQuery))
	fmt.Print("\r                         \r") // Clear the loading message
	
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, &APIError{
			Message: fmt.Sprintf("failed to parse JSON: %v", err),
			Context: "JSON parsing",
		}
	}

	var games []Game
	if err := json.Unmarshal(apiResp.Data, &games); err != nil {
		return nil, &APIError{
			Message: fmt.Sprintf("failed to parse games data: %v", err),
			Context: "games data parsing",
		}
	}

	debugLog("Found %d games", len(games))
	return games, nil
}

func (api *SpeedrunAPI) GetGameCategories(gameID string) ([]Category, error) {
	debugLog("Fetching categories for game: %s", gameID)
	
	body, err := api.makeRequest(fmt.Sprintf("/games/%s/categories", gameID))
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, &APIError{
			Message: fmt.Sprintf("failed to parse JSON: %v", err),
			Context: "JSON parsing",
		}
	}

	var categories []Category
	if err := json.Unmarshal(apiResp.Data, &categories); err != nil {
		return nil, &APIError{
			Message: fmt.Sprintf("failed to parse categories data: %v", err),
			Context: "categories data parsing",
		}
	}

	debugLog("Found %d categories", len(categories))
	return categories, nil
}

func (api *SpeedrunAPI) GetGamePlatforms(gameID string) ([]Platform, error) {
	debugLog("Fetching platforms for game: %s", gameID)
	
	body, err := api.makeRequest(fmt.Sprintf("/games/%s?embed=platforms", gameID))
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
		return nil, &APIError{
			Message: fmt.Sprintf("failed to parse JSON: %v", err),
			Context: "JSON parsing",
		}
	}

	debugLog("Found %d platforms", len(apiResp.Data.Platforms.Data))
	return apiResp.Data.Platforms.Data, nil
}

func (api *SpeedrunAPI) CheckPlatformForCategory(gameID, categoryID, platformID string) bool {
	endpoint := fmt.Sprintf("/leaderboards/%s/category/%s?platform=%s", gameID, categoryID, platformID)
	_, err := api.makeRequest(endpoint)
	if err != nil {
		debugLog("Platform check failed for %s/%s/%s: %v", gameID, categoryID, platformID, err)
		return false
	}
	
	return true
}

func (api *SpeedrunAPI) GetPlatformsForCategory(gameID, categoryID string) ([]Platform, error) {
	debugLog("Fetching platforms for category: %s/%s", gameID, categoryID)
	
	allPlatforms, err := api.GetGamePlatforms(gameID)
	if err != nil {
		return nil, err
	}

	validPlatforms := make([]Platform, 0, len(allPlatforms))
	resultChan := make(chan struct {
		platform Platform
		valid    bool
	}, len(allPlatforms))

	fmt.Printf("🔍 Checking %d platforms...", len(allPlatforms))
	
	for _, platform := range allPlatforms {
		go func(p Platform) {
			valid := api.CheckPlatformForCategory(gameID, categoryID, p.ID)
			resultChan <- struct {
				platform Platform
				valid    bool
			}{platform: p, valid: valid}
		}(platform)
	}

	for i := 0; i < len(allPlatforms); i++ {
		result := <-resultChan
		if result.valid {
			validPlatforms = append(validPlatforms, result.platform)
		}
	}
	
	fmt.Print("\r                                \r") // Clear the checking message

	debugLog("Found %d valid platforms for category", len(validPlatforms))
	return validPlatforms, nil
}

func (api *SpeedrunAPI) GetCategoryVariables(categoryID string) ([]SubCategory, error) {
	debugLog("Fetching variables for category: %s", categoryID)
	
	body, err := api.makeRequest(fmt.Sprintf("/categories/%s/variables", categoryID))
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, &APIError{
			Message: fmt.Sprintf("failed to parse JSON: %v", err),
			Context: "JSON parsing",
		}
	}

	var variables []Variable
	if err := json.Unmarshal(apiResp.Data, &variables); err != nil {
		return nil, &APIError{
			Message: fmt.Sprintf("failed to parse variables data: %v", err),
			Context: "variables data parsing",
		}
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

	debugLog("Found %d subcategories", len(subCategories))
	return subCategories, nil
}

func (api *SpeedrunAPI) GetVariableIDForCategory(categoryID string) string {
	body, err := api.makeRequest(fmt.Sprintf("/categories/%s/variables", categoryID))
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

func (api *SpeedrunAPI) GetLeaderboard(gameID, categoryID, platformID, variableID string) (*Leaderboard, error) {
	debugLog("Fetching leaderboard for %s/%s (platform: %s, variable: %s)", gameID, categoryID, platformID, variableID)
	
	var endpoint string
	queryParams := "embed=game,category,platforms" // Remove players and regions embed for performance
	
	if platformID != "" {
		queryParams += "&platform=" + platformID
	}
	
	if variableID != "" {
		varID := api.GetVariableIDForCategory(categoryID)
		if varID != "" {
			queryParams += "&var-" + varID + "=" + variableID
		}
	}
	
	endpoint = fmt.Sprintf("/leaderboards/%s/category/%s?%s", gameID, categoryID, queryParams)
	
	// Show progress for potentially slow leaderboard requests
	fmt.Print("⏳ Loading leaderboard data...")
	body, err := api.makeRequest(endpoint)
	fmt.Print("\r                              \r") // Clear the loading message
	
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
		return nil, &APIError{
			Message: fmt.Sprintf("failed to parse JSON: %v", err),
			Context: "JSON parsing",
		}
	}

	platformMap := make(map[string]string)
	for _, platform := range apiResp.Data.Platforms.Data {
		platformMap[platform.ID] = platform.Name
	}

	leaderboard := &Leaderboard{
		Weblink:     apiResp.Data.Weblink,
		Game:        apiResp.Data.Game,
		Category:    apiResp.Data.Category,
		Runs:        apiResp.Data.Runs,
		PlatformMap: platformMap,
	}

	debugLog("Fetched leaderboard with %d runs", len(leaderboard.Runs))
	return leaderboard, nil
}

func (api *SpeedrunAPI) GetUserData(userID string) *User {
	api.cacheMux.RLock()
	if user, exists := api.userCache[userID]; exists {
		api.cacheMux.RUnlock()
		return user
	}
	api.cacheMux.RUnlock()

	body, err := api.makeRequest(fmt.Sprintf("/users/%s", userID))
	if err != nil {
		debugLog("Failed to fetch user data for %s: %v", userID, err)
		return nil
	}

	var response struct {
		Data User `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		debugLog("Failed to decode user data for %s: %v", userID, err)
		return nil
	}

	api.cacheMux.Lock()
	api.userCache[userID] = &response.Data
	api.cacheMux.Unlock()

	return &response.Data
}

func (api *SpeedrunAPI) SearchUsers(query string) ([]User, error) {
	debugLog("Searching for users with query: %s", query)
	
	encodedQuery := url.QueryEscape(query)
	fmt.Print("🔍 Searching for users...")
	body, err := api.makeRequest(fmt.Sprintf("/users?lookup=%s&max=20", encodedQuery))
	fmt.Print("\r                         \r")
	
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, &APIError{
			Message: fmt.Sprintf("failed to parse JSON: %v", err),
			Context: "JSON parsing",
		}
	}

	var users []User
	if err := json.Unmarshal(apiResp.Data, &users); err != nil {
		return nil, &APIError{
			Message: fmt.Sprintf("failed to parse users data: %v", err),
			Context: "users data parsing",
		}
	}

	debugLog("Found %d users", len(users))
	return users, nil
}

func (api *SpeedrunAPI) GetUserRuns(userID string) ([]UserRun, error) {
	debugLog("Fetching runs for user: %s", userID)
	
	fmt.Print("⏳ Loading user runs...")
	body, err := api.makeRequest(fmt.Sprintf("/runs?user=%s&embed=game,category&orderby=date&direction=desc&max=25", userID))
	fmt.Print("\r                      \r")
	
	if err != nil {
		return nil, err
	}

	var apiResp struct {
		Data []struct {
			ID       string          `json:"id"`
			Weblink  string          `json:"weblink"`
			Game     json.RawMessage `json:"game"`
			Category json.RawMessage `json:"category"`
			Date     string          `json:"date"`
			Submitted time.Time      `json:"submitted"`
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
				Platform string `json:"platform"`
				Emulated bool   `json:"emulated"`
				Region   string `json:"region"`
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
			Place   int    `json:"place"`
		} `json:"data"`
		Embedded struct {
			Games []struct {
				Data Game `json:"data"`
			} `json:"games"`
			Categories []struct {
				Data Category `json:"data"`
			} `json:"categories"`
		} `json:"embedded"`
	}
	
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, &APIError{
			Message: fmt.Sprintf("failed to parse JSON: %v", err),
			Context: "JSON parsing",
		}
	}

	gameMap := make(map[string]Game)
	for _, gameWrapper := range apiResp.Embedded.Games {
		gameMap[gameWrapper.Data.ID] = gameWrapper.Data
	}

	categoryMap := make(map[string]Category)
	for _, categoryWrapper := range apiResp.Embedded.Categories {
		categoryMap[categoryWrapper.Data.ID] = categoryWrapper.Data
	}

	var userRuns []UserRun
	for _, runData := range apiResp.Data {
		if runData.Status.Status != "verified" {
			continue
		}

		userRun := UserRun{
			ID:       runData.ID,
			Weblink:  runData.Weblink,
			Date:     runData.Date,
			Submitted: runData.Submitted,
			Times:    runData.Times,
			Players:  runData.Players,
			System:   runData.System,
			Videos:   runData.Videos,
			Comment:  runData.Comment,
			Place:    runData.Place,
		}

		// Try to parse game as string ID first, then as object
		var gameID string
		if err := json.Unmarshal(runData.Game, &gameID); err != nil {
			// If it's not a string, try parsing as an object
			var gameData struct {
				Data Game `json:"data"`
			}
			if err := json.Unmarshal(runData.Game, &gameData); err == nil {
				userRun.Game = gameData.Data
			}
		} else {
			// It's a string ID, look it up in the map
			if game, exists := gameMap[gameID]; exists {
				userRun.Game = game
			}
		}

		// Try to parse category as string ID first, then as object
		var categoryID string
		if err := json.Unmarshal(runData.Category, &categoryID); err != nil {
			// If it's not a string, try parsing as an object
			var categoryData struct {
				Data Category `json:"data"`
			}
			if err := json.Unmarshal(runData.Category, &categoryData); err == nil {
				userRun.Category = categoryData.Data
			}
		} else {
			// It's a string ID, look it up in the map
			if category, exists := categoryMap[categoryID]; exists {
				userRun.Category = category
			}
		}

		userRuns = append(userRuns, userRun)
	}

	debugLog("Found %d verified runs for user", len(userRuns))
	return userRuns, nil
}

func fetchUserData(userID string) *struct {
	Names struct {
		International string `json:"international"`
	} `json:"names"`
} {
	api := NewSpeedrunAPI()
	user := api.GetUserData(userID)
	if user == nil {
		return nil
	}
	
	return &struct {
		Names struct {
			International string `json:"international"`
		} `json:"names"`
	}{
		Names: struct {
			International string `json:"international"`
		}{
			International: user.Names.International,
		},
	}
}