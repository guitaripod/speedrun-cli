package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type NavigationStack struct {
	stack []string
}

func NewNavigationStack() *NavigationStack {
	return &NavigationStack{
		stack: make([]string, 0),
	}
}

func (ns *NavigationStack) Push(level string) {
	ns.stack = append(ns.stack, level)
}

func (ns *NavigationStack) Pop() string {
	if len(ns.stack) == 0 {
		return ""
	}
	
	last := ns.stack[len(ns.stack)-1]
	ns.stack = ns.stack[:len(ns.stack)-1]
	return last
}

func (ns *NavigationStack) Current() string {
	if len(ns.stack) == 0 {
		return "main"
	}
	return ns.stack[len(ns.stack)-1]
}

func (ns *NavigationStack) Size() int {
	return len(ns.stack)
}

type UserChoice struct {
	Index    int
	Command  string
	IsQuit   bool
	IsBack   bool
	IsRefresh bool
	IsCategory bool
	IsHelp   bool
	IsUser   bool
}

func getUserInput(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func parseUserInput(input string) UserChoice {
	input = strings.TrimSpace(strings.ToLower(input))
	
	choice := UserChoice{
		Index: -1,
	}
	
	switch input {
	case "q", "quit", ":q":
		choice.IsQuit = true
	case "b", "back", ":b":
		choice.IsBack = true
	case "r", "refresh":
		choice.IsRefresh = true
	case "c", "category", ":c":
		choice.IsCategory = true
	case "h", "help":
		choice.IsHelp = true
	case "u", "user":
		choice.IsUser = true
	default:
		if index, err := strconv.Atoi(input); err == nil && index > 0 {
			choice.Index = index - 1
		} else {
			choice.Command = input
		}
	}
	
	return choice
}

func getUserChoice(prompt string, maxOptions int, allowBack bool) UserChoice {
	for {
		input := getUserInput(prompt)
		choice := parseUserInput(input)
		
		if choice.IsQuit || choice.IsHelp || choice.IsUser {
			return choice
		}
		
		if choice.IsBack && allowBack {
			return choice
		}
		
		if choice.IsRefresh || choice.IsCategory {
			return choice
		}
		
		if choice.Index >= 0 && choice.Index < maxOptions {
			return choice
		}
		
		showInputError(maxOptions, allowBack)
	}
}

func showInputError(maxOptions int, allowBack bool) {
	fmt.Printf("Invalid selection. Please try again.\n")
	
	var options []string
	if maxOptions > 0 {
		options = append(options, fmt.Sprintf("1-%d to select", maxOptions))
	}
	options = append(options, "'q' to quit")
	if allowBack {
		options = append(options, "'b' to go back")
	}
	options = append(options, "'h' for help")
	
	fmt.Printf("Available options: %s\n", strings.Join(options, ", "))
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
	
	choice := getUserChoice("\nEnter number to select, 'q' to quit: ", len(games), false)
	
	if choice.IsQuit {
		return nil
	}
	
	if choice.Index >= 0 {
		return &games[choice.Index]
	}
	
	return nil
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
	
	choice := getUserChoice("\nEnter number to select, 'b' to go back, 'q' to quit: ", len(platforms), true)
	
	if choice.IsQuit {
		return nil
	}
	
	if choice.IsBack {
		return &Platform{ID: "BACK"}
	}
	
	if choice.Index >= 0 {
		return &platforms[choice.Index]
	}
	
	return nil
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
	
	choice := getUserChoice("\nEnter number to select, 'b' to go back, 'q' to quit: ", len(categories), true)
	
	if choice.IsQuit {
		return nil
	}
	
	if choice.IsBack {
		return &Category{ID: "BACK"}
	}
	
	if choice.Index >= 0 {
		return &categories[choice.Index]
	}
	
	return nil
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
	
	choice := getUserChoice("\nEnter number to select, 'b' to go back, 'q' to quit: ", len(subCategories), true)
	
	if choice.IsQuit {
		return nil
	}
	
	if choice.IsBack {
		return &SubCategory{ID: "BACK"}
	}
	
	if choice.Index >= 0 {
		return &subCategories[choice.Index]
	}
	
	return nil
}

func handleLeaderboardNavigation() UserChoice {
	fmt.Println("\nControls: [Enter] continue, 'b' back to subcategories, 'c' back to categories, 'q' quit, 'r' refresh")
	input := getUserInput("Action: ")
	return parseUserInput(input)
}

func selectUser(users []User) *User {
	if len(users) == 0 {
		fmt.Println("No users found.")
		return nil
	}
	
	if len(users) == 1 {
		fmt.Printf("Found exact match: %s\n", users[0].Names.International)
		return &users[0]
	}
	
	fmt.Printf("\nFound %d users:\n", len(users))
	for i, user := range users {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, user.Names.International, user.ID)
	}
	
	choice := getUserChoice("\nEnter number to select, 'q' to quit: ", len(users), false)
	
	if choice.IsQuit {
		return nil
	}
	
	if choice.Index >= 0 {
		return &users[choice.Index]
	}
	
	return nil
}

func showHelp() {
	fmt.Println("\n📚 Help - Speedrun.com CLI")
	fmt.Println("============================")
	fmt.Println("Navigation Flow:")
	fmt.Println("  1. Search for a game OR search for a user")
	fmt.Println("     • Game: Search for a game → Select categories → View leaderboard")
	fmt.Println("     • User: Search for a user → View their recent runs with placements")
	fmt.Println("  2. For games: Select platform category and subcategory")
	fmt.Println("  3. View leaderboard or user runs")
	fmt.Println("\nControls:")
	fmt.Println("  • Use numbers to select from lists")
	fmt.Println("  • 'q' or ':q' - quit")
	fmt.Println("  • 'b' or ':b' - go back")
	fmt.Println("  • 'c' or ':c' - back to categories (from leaderboard)")
	fmt.Println("  • 'r' - refresh current view")
	fmt.Println("  • 'u' or 'user' - search for users instead of games")
	fmt.Println("  • 'h' or 'help' - show this help")
	fmt.Println("\nFeatures:")
	fmt.Println("  • Fuzzy game and user search")
	fmt.Println("  • Platform categories with subcategories")
	fmt.Println("  • Detailed leaderboards with filtering")
	fmt.Println("  • User run history with placements and medals")
	fmt.Println("  • Run times, players, platforms, videos")
	fmt.Println()
}