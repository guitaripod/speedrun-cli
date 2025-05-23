package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
)

const (
	APIBase               = "https://www.speedrun.com/api/v1"
	UserAgent             = "speedrun-cli/1.0"
	DefaultTimeout        = 30
	MaxRetries            = 3
	BackoffBase           = 2
	MaxRankWithMedal      = 3
	DefaultColumnWidth    = 20
	CommentMaxWidth       = 25
	EmptyValuePlaceholder = "—"
)

type Colors struct {
	Gold   string
	Silver string
	Bronze string
	Reset  string
	Green  string
	Red    string
	Blue   string
}

var DefaultColors = Colors{
	Gold:   "\033[33m",
	Silver: "\033[37m",
	Bronze: "\033[31m",
	Reset:  "\033[0m",
	Green:  "\033[32m",
	Red:    "\033[31m",
	Blue:   "\033[34m",
}

func debugLog(format string, args ...interface{}) {
	if os.Getenv("SPEEDRUN_DEBUG") != "" {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func truncateString(s string, maxLen int) string {
	if s == "" {
		return EmptyValuePlaceholder
	}
	
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

func cleanComment(comment string) string {
	if comment == "" {
		return EmptyValuePlaceholder
	}
	
	var cleaned strings.Builder
	for _, r := range comment {
		if !unicode.IsControl(r) {
			cleaned.WriteRune(r)
		} else if r == '\n' || r == '\r' || r == '\t' {
			cleaned.WriteRune(' ')
		}
	}
	
	result := strings.TrimSpace(cleaned.String())
	if result == "" {
		return EmptyValuePlaceholder
	}
	return result
}

func formatTime(timeStr string) string {
	if timeStr == "" || timeStr == "null" {
		return EmptyValuePlaceholder
	}
	
	if strings.HasPrefix(timeStr, "PT") {
		return parsePTFormat(timeStr)
	}
	
	if seconds, err := strconv.ParseFloat(timeStr, 64); err == nil {
		return formatSeconds(seconds)
	}
	
	return timeStr
}

func parsePTFormat(ptTime string) string {
	ptTime = strings.TrimPrefix(ptTime, "PT")
	
	var hours, minutes float64
	var seconds float64
	
	if hIdx := strings.Index(ptTime, "H"); hIdx != -1 {
		if h, err := strconv.ParseFloat(ptTime[:hIdx], 64); err == nil {
			hours = h
		}
		ptTime = ptTime[hIdx+1:]
	}
	
	if mIdx := strings.Index(ptTime, "M"); mIdx != -1 {
		if m, err := strconv.ParseFloat(ptTime[:mIdx], 64); err == nil {
			minutes = m
		}
		ptTime = ptTime[mIdx+1:]
	}
	
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
		if seconds == float64(int(seconds)) {
			return fmt.Sprintf("%d:%02d:%02.0f", hours, minutes, seconds)
		}
		return fmt.Sprintf("%d:%02d:%06.3f", hours, minutes, seconds)
	}
	if minutes > 0 || totalSeconds >= 60 {
		if seconds == float64(int(seconds)) {
			return fmt.Sprintf("%d:%02.0f", minutes, seconds)
		}
		return fmt.Sprintf("%d:%06.3f", minutes, seconds)
	}
	if seconds == float64(int(seconds)) {
		return fmt.Sprintf("%.0f", seconds)
	}
	return fmt.Sprintf("%.3f", seconds)
}


func calculateDynamicWidth(content []string, maxWidth int) int {
	width := 0
	for _, item := range content {
		runes := []rune(item)
		if len(runes) > width {
			width = len(runes)
		}
	}
	if width > maxWidth {
		return maxWidth
	}
	if width < 5 {
		return 5
	}
	return width
}