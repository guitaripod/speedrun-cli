package main

import (
	"fmt"
	"os"
)

func main() {
	api := NewSpeedrunAPI()
	nav := NewNavigationStack()
	
	fmt.Println("🏃 Speedrun.com CLI - Game Leaderboard Browser")
	fmt.Println("==============================================")
	fmt.Println("Type 'h' or 'help' for instructions")
	
	for {
		query := getUserInput("\nEnter game name to search (or 'q' to quit): ")
		
		choice := parseUserInput(query)
		
		if choice.IsQuit {
			fmt.Println("Goodbye! 👋")
			break
		}
		
		if choice.IsHelp {
			showHelp()
			continue
		}
		
		if query == "" {
			continue
		}
		
		fmt.Println("Searching for games...")
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
			fmt.Printf("\nLoading categories for %s...\n", selectedGame.Names.International)
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
				fmt.Printf("\nLoading subcategories for %s - %s...\n", selectedGame.Names.International, selectedCategory.Name)
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
					fmt.Printf("\nLoading leaderboard for %s - %s (%s)...\n", 
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