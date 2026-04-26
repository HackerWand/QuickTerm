package services

import (
	"QuickTerm/models"
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var activeWindows = make(map[int64]bool)

func OpenWorkspaceWindow(ctx context.Context, workspace *models.Workspace) {
	if activeWindows[workspace.ID] {
		runtime.LogWarning(ctx, fmt.Sprintf("Workspace window %d already active", workspace.ID))
		return
	}

	activeWindows[workspace.ID] = true
	defer delete(activeWindows, workspace.ID)

	runtime.LogInfo(ctx, fmt.Sprintf("Opening workspace window: %s", workspace.Name))
}
