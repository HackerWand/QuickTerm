package services

import (
	"QuickTerm/database"
	"QuickTerm/models"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

func AutoSaveCommand(workspaceID int64, commandContent string) error {
	if commandContent == "" {
		return nil
	}

	trimmedContent := strings.TrimSpace(commandContent)
	if trimmedContent == "" {
		return nil
	}

	workspace, err := database.GetWorkspaceByID(workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	if shouldIgnoreCommand(trimmedContent, workspace.IgnoredCommands) {
		return nil
	}

	exists, err := database.CommandExists(workspaceID, trimmedContent)
	if err != nil {
		return fmt.Errorf("failed to check command existence: %w", err)
	}
	if exists {
		return nil
	}

	commandName := trimmedContent
	parts := strings.Fields(commandName)
	if len(parts) > 0 {
		commandName = parts[0]
	}
	if strings.Contains(commandName, "/") {
		slashParts := strings.Split(commandName, "/")
		commandName = slashParts[len(slashParts)-1]
	}

	cmd := &models.Command{
		Name:        commandName,
		Content:     trimmedContent,
		Description: "",
		WorkspaceID: workspaceID,
	}

	if err := database.CreateCommand(cmd); err != nil {
		return fmt.Errorf("failed to create command: %w", err)
	}
	return nil
}

func shouldIgnoreCommand(content string, rules []models.IgnoreRule) bool {
	for _, rule := range rules {
		if rule.IsRegex {
			re, err := regexp.Compile(rule.Pattern)
			if err != nil {
				continue
			}
			if re.MatchString(content) {
				return true
			}
		} else {
			trimmedContent := strings.TrimLeftFunc(content, unicode.IsSpace)
			if strings.HasPrefix(trimmedContent, rule.Pattern) {
				return true
			}
		}
	}
	return false
}
