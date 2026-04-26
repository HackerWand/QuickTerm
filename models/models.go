package models

import "encoding/json"

type Workspace struct {
	ID              int64        `json:"id"`
	Name            string       `json:"name"`
	Path            string       `json:"path"`
	IgnoredCommands []IgnoreRule `json:"ignoredCommands"`
}

type IgnoreRule struct {
	Pattern string `json:"pattern"`
	IsRegex bool   `json:"isRegex"`
}

type CommandGroup struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	WorkspaceID int64  `json:"workspaceId"`
}

type TemplateOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type TemplateParam struct {
	Name        string           `json:"name"`
	Type        string           `json:"type"`
	Description string           `json:"description"`
	Options     []TemplateOption `json:"options"`
}

type Command struct {
	ID             int64           `json:"id"`
	Name           string          `json:"name"`
	Content        string          `json:"content"`
	Description    string          `json:"description"`
	GroupID        *int64          `json:"groupId"`
	WorkspaceID    int64           `json:"workspaceId"`
	TemplateParams []TemplateParam `json:"templateParams"`
}

type RecentPath struct {
	ID          int64  `json:"id"`
	WorkspaceID int64  `json:"workspaceId"`
	Path        string `json:"path"`
	Position    int    `json:"position"`
}

func (ws *Workspace) GetIgnoredCommandsJSON() (string, error) {
	if len(ws.IgnoredCommands) == 0 {
		return "", nil
	}
	data, err := json.Marshal(ws.IgnoredCommands)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (ws *Workspace) SetIgnoredCommandsFromJSON(data string) error {
	if data == "" {
		ws.IgnoredCommands = []IgnoreRule{}
		return nil
	}
	var rules []IgnoreRule
	err := json.Unmarshal([]byte(data), &rules)
	if err != nil {
		return err
	}
	ws.IgnoredCommands = rules
	return nil
}

func (c *Command) GetTemplateParamsJSON() (string, error) {
	if len(c.TemplateParams) == 0 {
		return "", nil
	}
	data, err := json.Marshal(c.TemplateParams)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Command) SetTemplateParamsFromJSON(data string) error {
	if data == "" {
		c.TemplateParams = []TemplateParam{}
		return nil
	}
	var params []TemplateParam
	err := json.Unmarshal([]byte(data), &params)
	if err != nil {
		return err
	}
	c.TemplateParams = params
	return nil
}

type WorkspaceExport struct {
	Version  string         `json:"version"`
	Name     string         `json:"name"`
	Groups   []CommandGroup `json:"groups"`
	Commands []Command      `json:"commands"`
	Ignored  []IgnoreRule   `json:"ignored"`
}

func NewWorkspaceExport() *WorkspaceExport {
	return &WorkspaceExport{
		Version:  "1.0",
		Groups:   []CommandGroup{},
		Commands: []Command{},
		Ignored:  []IgnoreRule{},
	}
}

type DatabaseBackup struct {
	Version      string         `json:"version"`
	Workspaces   []Workspace    `json:"workspaces"`
	Groups       []CommandGroup `json:"groups"`
	Commands     []Command      `json:"commands"`
	RecentPaths  []RecentPath   `json:"recentPaths"`
}

func NewDatabaseBackup() *DatabaseBackup {
	return &DatabaseBackup{
		Version:     "1.0",
		Workspaces:  []Workspace{},
		Groups:      []CommandGroup{},
		Commands:    []Command{},
		RecentPaths: []RecentPath{},
	}
}
