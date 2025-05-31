package main

import (
	"fmt"
	"strings"
)

func displayLeaderboard(lb *Leaderboard) {
	colors := DefaultColors
	
	fmt.Printf("\n🏆 %s - %s\n", lb.Game.Data.Names.International, lb.Category.Data.Name)
	fmt.Printf("📊 %s\n\n", lb.Weblink)
	
	if len(lb.Runs) == 0 {
		fmt.Println("No runs found for this category.")
		return
	}
	
	playerNames := make([]string, len(lb.Runs))
	platforms := make([]string, len(lb.Runs))
	comments := make([]string, len(lb.Runs))
	
	for i, entry := range lb.Runs {
		playerNames[i] = getPlayerDisplayName(entry.Run)
		platforms[i] = getPlatformName(entry.Run, lb.PlatformMap)
		comments[i] = cleanComment(entry.Run.Comment)
	}
	
	playerWidth := calculateDynamicWidth(playerNames, 25)
	platformWidth := calculateDynamicWidth(platforms, 20)
	commentWidth := calculateDynamicWidth(comments, 30)
	
	headerFormat := fmt.Sprintf("%%-%ds%%-%ds %%-15s %%-%ds %%-10s %%-5s %%-3s %%s\n", 
		6, playerWidth, platformWidth)
	rowFormat := fmt.Sprintf("%%s%%-%ds %%-15s %%-%ds %%-10s %%-5s %%-3s %%s\n", 
		playerWidth, platformWidth)
	
	fmt.Printf(headerFormat, "Rank", "Player", "Time", "Platform", "Date", "Video", "Emu", "Comment")
	fmt.Println(strings.Repeat("─", 6+playerWidth+15+platformWidth+10+5+3+commentWidth+8))
	
	for _, entry := range lb.Runs {
		playerName := getPlayerDisplayName(entry.Run)
		
		time := getBestTime(entry.Run)
		platform := getPlatformName(entry.Run, lb.PlatformMap)
		
		hasVideo := "❌"
		if len(entry.Run.Videos.Links) > 0 && entry.Run.Videos.Links[0].URI != "" {
			hasVideo = "✅"
		}
		
		emulated := "❌"
		if entry.Run.System.Emulated {
			emulated = "✅"
		}
		
		comment := cleanComment(entry.Run.Comment)
		
		rank := formatRank(entry.Place, colors)
		
		fmt.Printf(rowFormat,
			rank,
			truncateString(playerName, playerWidth),
			time,
			truncateString(platform, platformWidth),
			entry.Run.Date,
			hasVideo,
			emulated,
			truncateString(comment, commentWidth))
	}
	
	fmt.Printf("\n📈 Showing %d runs\n", len(lb.Runs))
}

func getPlayerDisplayName(run Run) string {
	if len(run.Players) == 0 {
		return "Guest"
	}
	
	player := run.Players[0]
	if player.Name != "" {
		return player.Name
	}
	
	if player.ID != "" {
		if userData := fetchUserData(player.ID); userData != nil {
			if userData.Names.International != "" {
				return userData.Names.International
			}
		}
	}
	
	return "Guest"
}

func getBestTime(run Run) string {
	times := []string{
		run.Times.Primary,
		run.Times.Realtime,
		run.Times.RealtimeNoLoads,
		run.Times.Ingame,
	}
	
	for _, timeStr := range times {
		formatted := formatTime(timeStr)
		if formatted != EmptyValuePlaceholder {
			return formatted
		}
	}
	
	return EmptyValuePlaceholder
}

func getPlatformName(run Run, platformMap map[string]string) string {
	if run.System.Platform == "" {
		return "Unknown"
	}
	
	if platformName, exists := platformMap[run.System.Platform]; exists {
		return platformName
	}
	
	return run.System.Platform
}

func formatRank(place int, colors Colors) string {
	switch place {
	case 1:
		return fmt.Sprintf("%s🥇%s    ", colors.Gold, colors.Reset)
	case 2:
		return fmt.Sprintf("%s🥈%s    ", colors.Silver, colors.Reset)
	case 3:
		return fmt.Sprintf("%s🥉%s    ", colors.Bronze, colors.Reset)
	default:
		return fmt.Sprintf("%-6d", place)
	}
}

func displayUserRuns(user *User, runs []UserRun) {
	colors := DefaultColors
	
	fmt.Printf("\n👤 %s - Recent Submitted Runs\n", user.Names.International)
	fmt.Printf("📊 Showing %d verified runs\n\n", len(runs))
	
	if len(runs) == 0 {
		fmt.Println("No verified runs found for this user.")
		return
	}
	
	gameNames := make([]string, len(runs))
	categoryNames := make([]string, len(runs))
	comments := make([]string, len(runs))
	
	for i, run := range runs {
		gameNames[i] = run.Game.Names.International
		categoryNames[i] = run.Category.Name
		comments[i] = cleanComment(run.Comment)
	}
	
	gameWidth := calculateDynamicWidth(gameNames, 25)
	categoryWidth := calculateDynamicWidth(categoryNames, 20)
	commentWidth := calculateDynamicWidth(comments, 25)
	
	headerFormat := fmt.Sprintf("%%-%ds%%-%ds %%-15s %%-%ds %%-10s %%-6s %%-5s %%-3s %%s\n", 
		6, gameWidth, categoryWidth)
	rowFormat := fmt.Sprintf("%%s%%-%ds %%-15s %%-%ds %%-10s %%-6s %%-5s %%-3s %%s\n", 
		gameWidth, categoryWidth)
	
	fmt.Printf(headerFormat, "Place", "Game", "Time", "Category", "Date", "Status", "Video", "Emu", "Comment")
	fmt.Println(strings.Repeat("─", 6+gameWidth+15+categoryWidth+10+6+5+3+commentWidth+9))
	
	for _, run := range runs {
		gameName := run.Game.Names.International
		categoryName := run.Category.Name
		
		time := getUserRunTime(run)
		
		hasVideo := "❌"
		if len(run.Videos.Links) > 0 && run.Videos.Links[0].URI != "" {
			hasVideo = "✅"
		}
		
		emulated := "❌"
		if run.System.Emulated {
			emulated = "✅"
		}
		
		comment := cleanComment(run.Comment)
		
		var rank string
		var status string
		if run.Place > 0 {
			rank = formatRank(run.Place, colors)
			status = fmt.Sprintf("#%d", run.Place)
		} else {
			rank = "    - "
			status = "Unranked"
		}
		
		fmt.Printf(rowFormat,
			rank,
			truncateString(gameName, gameWidth),
			time,
			truncateString(categoryName, categoryWidth),
			run.Date,
			status,
			hasVideo,
			emulated,
			truncateString(comment, commentWidth))
	}
	
	fmt.Printf("\n📈 Showing %d runs\n", len(runs))
}

func getUserRunTime(run UserRun) string {
	times := []string{
		run.Times.Primary,
		run.Times.Realtime,
		run.Times.RealtimeNoLoads,
		run.Times.Ingame,
	}
	
	for _, timeStr := range times {
		formatted := formatTime(timeStr)
		if formatted != EmptyValuePlaceholder {
			return formatted
		}
	}
	
	return EmptyValuePlaceholder
}