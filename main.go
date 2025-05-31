package main

import (
	"fmt"
	"os"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("speedrun-cli version %s\n", Version)
			fmt.Printf("Built: %s\n", BuildTime)
			fmt.Printf("Commit: %s\n", Commit)
			return
		case "--help", "-h":
			showHelp()
			return
		}
	}

	api := NewSpeedrunAPI()
	nav := NewNavigationStack()
	
	fmt.Printf("🏃 Speedrun.com CLI v%s - Game Leaderboard Browser\n", Version)
	fmt.Println("==============================================")
	fmt.Println("Type 'h' or 'help' for instructions")
	
	for {
		query := getUserInput("\nEnter game name to search (or 'u' for user search, 'q' to quit): ")
		
		choice := parseUserInput(query)
		
		if choice.IsQuit {
			fmt.Println("Goodbye! 👋")
			break
		}
		
		if choice.IsHelp {
			showHelp()
			continue
		}
		
		if choice.IsUser {
			handleUserSearch(api)
			continue
		}
		
		if query == "" {
			continue
		}
		
		games, err := api.SearchGames(query)
		if err != nil {
			fmt.Printf("Error searching games: %v\n", err)
			continue
		}
		
		selectedGame := selectGame(games)
		if selectedGame == nil {
			continue
		}
		
		nav.Push("game")
		
		for nav.Current() == "game" {
			fmt.Printf("\n📋 Loading categories for %s...\n", selectedGame.Names.International)
			categories, err := api.GetGameCategories(selectedGame.ID)
			if err != nil {
				fmt.Printf("Error loading categories: %v\n", err)
				nav.Pop()
				break
			}
			
			selectedCategory := selectCategory(categories)
			if selectedCategory == nil {
				nav.Pop()
				break
			}
			
			if selectedCategory.ID == "BACK" {
				nav.Pop()
				break
			}
			
			nav.Push("category")
			
			for nav.Current() == "category" {
				fmt.Printf("\n🏷️  Loading subcategories for %s - %s...\n", selectedGame.Names.International, selectedCategory.Name)
				subCategories, err := api.GetCategoryVariables(selectedCategory.ID)
				if err != nil {
					fmt.Printf("Error loading subcategories: %v\n", err)
					nav.Pop()
					break
				}
				
				selectedSubCategory := selectSubCategory(subCategories)
				if selectedSubCategory == nil {
					nav.Pop()
					break
				}
				
				if selectedSubCategory.ID == "BACK" {
					nav.Pop()
					break
				}
				
				nav.Push("subcategory")
				
				for nav.Current() == "subcategory" {
					fmt.Printf("\n🏆 Loading leaderboard for %s - %s (%s)...\n", 
						selectedGame.Names.International, selectedCategory.Name, selectedSubCategory.Label)
					
					leaderboard, err := api.GetLeaderboard(selectedGame.ID, selectedCategory.ID, "", selectedSubCategory.ID)
					if err != nil {
						fmt.Printf("Error loading leaderboard: %v\n", err)
						nav.Pop()
						break
					}
					
					displayLeaderboard(leaderboard)
					
					choice := handleLeaderboardNavigation()
					
					if choice.IsQuit {
						fmt.Println("Goodbye! 👋")
						os.Exit(0)
					}
					
					if choice.IsBack {
						nav.Pop()
						break
					}
					
					if choice.IsCategory {
						nav.Pop()
						nav.Pop()
						break
					}
					
					if choice.IsRefresh {
						continue
					}
					
					if choice.IsHelp {
						showHelp()
						continue
					}
				}
			}
		}
	}
}

func handleUserSearch(api *SpeedrunAPI) {
	for {
		userQuery := getUserInput("\nEnter username to search (or 'b' to go back): ")
		
		choice := parseUserInput(userQuery)
		
		if choice.IsQuit {
			fmt.Println("Goodbye! 👋")
			return
		}
		
		if choice.IsBack {
			return
		}
		
		if choice.IsHelp {
			showHelp()
			continue
		}
		
		if userQuery == "" {
			continue
		}
		
		users, err := api.SearchUsers(userQuery)
		if err != nil {
			fmt.Printf("Error searching users: %v\n", err)
			continue
		}
		
		selectedUser := selectUser(users)
		if selectedUser == nil {
			continue
		}
		
		runs, err := api.GetUserRuns(selectedUser.ID)
		if err != nil {
			fmt.Printf("Error loading user runs: %v\n", err)
			continue
		}
		
		displayUserRuns(selectedUser, runs)
		
		fmt.Println("\nPress Enter to continue, 'b' to go back, 'q' to quit:")
		input := getUserInput("")
		navChoice := parseUserInput(input)
		
		if navChoice.IsQuit {
			fmt.Println("Goodbye! 👋")
			return
		}
		
		if navChoice.IsBack {
			continue
		}
		
		return
	}
}