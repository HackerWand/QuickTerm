package main

import (
	"QuickTerm/database"
	"QuickTerm/models"
	"QuickTerm/pty"
	"QuickTerm/services"
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/application"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

type App struct {
	ctx                context.Context
	ptys               map[string]*pty.PTY
	ptyMux             sync.Mutex
	terminalWorkspaces map[string]int64
	currentWorkspaceID int64
}

func NewApp() *App {
	return &App{
		ptys:               make(map[string]*pty.PTY),
		terminalWorkspaces: make(map[string]int64),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := ensureDefaultWorkspace(); err != nil {
		log.Printf("Warning: Failed to ensure default workspace: %v", err)
	}
}

func (a *App) shutdown(ctx context.Context) {
	// 清理所有终端
	a.closeAllTerminals()

	if err := database.Close(); err != nil {
		log.Printf("Warning: Failed to close database: %v", err)
	}
}

func (a *App) closeAllTerminals() {
	// 先复制需要关闭的终端列表
	var toClose []*pty.PTY
	a.ptyMux.Lock()
	for _, term := range a.ptys {
		toClose = append(toClose, term)
	}
	// 清空 map
	a.ptys = make(map[string]*pty.PTY)
	a.ptyMux.Unlock()

	// 在锁外关闭终端
	for _, term := range toClose {
		if err := term.Close(); err != nil {
			log.Printf("Warning: Failed to close terminal: %v", err)
		}
	}
}

func (a *App) ClearAllTerminals() error {
	a.closeAllTerminals()
	return nil
}

func (a *App) GetWorkspaces() ([]models.Workspace, error) {
	return database.GetWorkspaces()
}

func (a *App) CreateWorkspace(name string, path string) (*models.Workspace, error) {
	if path == "" {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		path = usr.HomeDir
	}

	workspace := &models.Workspace{
		Name:            name,
		Path:            path,
		IgnoredCommands: []models.IgnoreRule{},
	}

	if err := database.CreateWorkspace(workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

func (a *App) UpdateWorkspace(workspace *models.Workspace) error {
	return database.UpdateWorkspace(workspace)
}

func (a *App) DeleteWorkspace(id int64) error {
	count, err := database.GetWorkspaceCount()
	if err != nil {
		return err
	}
	if count <= 1 {
		return fmt.Errorf("cannot delete the last workspace")
	}
	return database.DeleteWorkspace(id)
}

func (a *App) SetCurrentWorkspace(workspaceID int64) error {
	a.currentWorkspaceID = workspaceID
	return nil
}

func (a *App) GetCommandGroups(workspaceID int64) ([]models.CommandGroup, error) {
	return database.GetCommandGroupsByWorkspace(workspaceID)
}

func (a *App) CreateCommandGroup(group *models.CommandGroup) error {
	return database.CreateCommandGroup(group)
}

func (a *App) UpdateCommandGroup(group *models.CommandGroup) error {
	return database.UpdateCommandGroup(group)
}

func (a *App) DeleteCommandGroup(id int64) error {
	return database.DeleteCommandGroup(id)
}

func (a *App) GetCommands(workspaceID int64) ([]models.Command, error) {
	return database.GetCommandsByWorkspace(workspaceID)
}

func (a *App) CreateCommand(cmd *models.Command) error {
	exists, err := database.CommandExists(cmd.WorkspaceID, cmd.Content)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("command already exists")
	}
	return database.CreateCommand(cmd)
}

func (a *App) UpdateCommand(cmd *models.Command) error {
	return database.UpdateCommand(cmd)
}

func (a *App) DeleteCommand(id int64) error {
	return database.DeleteCommand(id)
}

func (a *App) GetRecentPaths(workspaceID int64) ([]models.RecentPath, error) {
	return database.GetRecentPaths(workspaceID)
}

func (a *App) AddRecentPath(workspaceID int64, path string) error {
	return database.AddRecentPath(workspaceID, path)
}

func (a *App) DeleteRecentPath(id int64) error {
	return database.DeleteRecentPath(id)
}

func (a *App) ClearRecentPaths(workspaceID int64) error {
	return database.ClearRecentPaths(workspaceID)
}

func (a *App) AutoSaveCommand(workspaceID int64, commandContent string) error {
	return services.AutoSaveCommand(workspaceID, commandContent)
}

func (a *App) OpenWorkspaceWindow(workspaceID int64) error {
	workspace, err := database.GetWorkspaceByID(workspaceID)
	if err != nil {
		return err
	}
	go services.OpenWorkspaceWindow(a.ctx, workspace)
	return nil
}

func (a *App) CreateTerminal(id string, shell string, cwd string, env []string) error {
	// 检查是否已存在
	a.ptyMux.Lock()
	if _, exists := a.ptys[id]; exists {
		a.ptyMux.Unlock()
		return fmt.Errorf("terminal already exists")
	}
	a.ptyMux.Unlock()

	// 创建新的PTY
	term, err := pty.New(shell, cwd, env)
	if err != nil {
		return err
	}

	// 存储终端
	a.ptyMux.Lock()
	a.ptys[id] = term
	a.ptyMux.Unlock()

	// 开始读取输出并发送到前端
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := term.Read(buf)
			if err != nil {
				break
			}
			runtime.EventsEmit(a.ctx, "terminal-output-"+id, string(buf[:n]))
		}
		// 终端关闭后清理
		a.ptyMux.Lock()
		delete(a.ptys, id)
		a.ptyMux.Unlock()
		runtime.EventsEmit(a.ctx, "terminal-closed-"+id)
	}()

	return nil
}

func (a *App) WriteTerminal(id string, data string) error {
	a.ptyMux.Lock()
	term, exists := a.ptys[id]
	a.ptyMux.Unlock()

	if !exists {
		return fmt.Errorf("terminal not found")
	}

	_, err := term.Write([]byte(data))
	return err
}

func (a *App) ResizeTerminal(id string, rows, cols int) error {
	a.ptyMux.Lock()
	term, exists := a.ptys[id]
	a.ptyMux.Unlock()

	if !exists {
		return fmt.Errorf("terminal not found")
	}

	err := term.Resize(rows, cols)
	return err
}

func (a *App) CloseTerminal(id string) error {
	a.ptyMux.Lock()
	term, exists := a.ptys[id]
	if exists {
		delete(a.ptys, id)
	}
	a.ptyMux.Unlock()

	if !exists {
		return fmt.Errorf("terminal not found")
	}

	err := term.Close()
	return err
}

func (a *App) SelectPathDialog() (string, error) {
	result, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择文件或文件夹",
	})
	return result, err
}

func (a *App) GetDirectoryPath(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return path, nil
	}
	dir := filepath.Dir(path)
	return dir, nil
}

func (a *App) ExportWorkspace(workspaceID int64) (*models.WorkspaceExport, error) {
	return database.ExportWorkspace(workspaceID)
}

func (a *App) ImportWorkspace(exportData *models.WorkspaceExport, path string) (*models.Workspace, error) {
	return database.ImportWorkspace(exportData, path)
}

func (a *App) SaveFileDialog(defaultName string) (string, error) {
	result, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "保存工作空间导出",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON文件 (*.json)",
				Pattern:     "*.json",
			},
		},
	})
	return result, err
}

func (a *App) OpenFileDialog() (string, error) {
	result, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择要导入的工作空间文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON文件 (*.json)",
				Pattern:     "*.json",
			},
		},
	})
	return result, err
}

func (a *App) OpenFileSelectorDialog() (string, error) {
	result, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择文件",
	})
	return result, err
}

func (a *App) OpenDirectorySelectorDialog() (string, error) {
	result, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择目录",
	})
	return result, err
}

func (a *App) WriteFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func (a *App) ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (a *App) ExportDatabase() (*models.DatabaseBackup, error) {
	return database.ExportDatabase()
}

func (a *App) ImportDatabase(backup *models.DatabaseBackup) error {
	return database.ImportDatabase(backup)
}

func (a *App) DatabaseHasData() (bool, error) {
	return database.DatabaseHasData()
}

func (a *App) SaveDatabaseBackupDialog(defaultName string) (string, error) {
	result, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "保存数据库备份",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON文件 (*.json)",
				Pattern:     "*.json",
			},
		},
	})
	return result, err
}

func (a *App) OpenDatabaseBackupDialog() (string, error) {
	result, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择要恢复的数据库备份文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON文件 (*.json)",
				Pattern:     "*.json",
			},
		},
	})
	return result, err
}

func ensureDefaultWorkspace() error {
	workspaces, err := database.GetWorkspaces()
	if err != nil {
		return err
	}
	if len(workspaces) == 0 {
		usr, err := user.Current()
		if err != nil {
			return err
		}
		workspace := &models.Workspace{
			Name:            "默认工作空间",
			Path:            usr.HomeDir,
			IgnoredCommands: []models.IgnoreRule{},
		}
		return database.CreateWorkspace(workspace)
	}
	return nil
}

func main() {
	app := NewApp()

	options := &options.App{
		Title:             "QuickTerm",
		Width:             800,
		Height:            600,
		MinWidth:          600,
		MinHeight:         400,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 30, G: 30, B: 30, A: 255},
		// Menu:              appMenu,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Bind: []interface{}{
			app,
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: false,
			CSSDropProperty:    "--wails-drop-target",
			CSSDropValue:       "drop",
		},
		OnStartup: app.startup,
		OnDomReady: func(ctx context.Context) {
			// 前端 DOM 准备好时清理之前可能存在的终端
			app.closeAllTerminals()
		},
		OnBeforeClose:    func(ctx context.Context) (prevent bool) { return false },
		OnShutdown:       app.shutdown,
		WindowStartState: options.Normal,
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarDefault(),
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "QuickTerm",
				Message: "现代化跨平台终端管理工具",
			},
		},
		Linux: &linux.Options{
			WindowIsTranslucent: false,
		},
	}

	if err := application.NewWithOptions(options).Run(); err != nil {
		log.Fatalf("Error: %v", err)
		os.Exit(1)
	}
}
