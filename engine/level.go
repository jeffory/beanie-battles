package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Level represents a game level with platforms, enemy spawn points, and a flag
type Level struct {
	Name           string             `json:"name"`
	Platforms      []PlatformData     `json:"platforms"`
	EnemySpawns    []EnemySpawnData   `json:"enemySpawns"`
	FlagPosition   FlagData           `json:"flagPosition"`
	PlayerStart    PlayerStartData    `json:"playerStart"`
	BackgroundType string             `json:"backgroundType"`
}

// PlatformData represents the data for a platform
type PlatformData struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// EnemySpawnData represents an enemy spawn point
type EnemySpawnData struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Type string  `json:"type"` // "kidney" or "navy"
}

// FlagData represents the position of the level's flag
type FlagData struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// PlayerStartData represents the player's starting position
type PlayerStartData struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NewLevel creates a new empty level
func NewLevel(name string) *Level {
	return &Level{
		Name:           name,
		Platforms:      make([]PlatformData, 0),
		EnemySpawns:    make([]EnemySpawnData, 0),
		FlagPosition:   FlagData{X: 0, Y: 0},
		PlayerStart:    PlayerStartData{X: 100, Y: 100},
		BackgroundType: "default",
	}
}

// AddPlatform adds a platform to the level
func (l *Level) AddPlatform(x, y, width, height float64) {
	platform := PlatformData{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
	l.Platforms = append(l.Platforms, platform)
}

// AddEnemySpawn adds an enemy spawn point to the level
func (l *Level) AddEnemySpawn(x, y float64, enemyType string) {
	spawn := EnemySpawnData{
		X:    x,
		Y:    y,
		Type: enemyType,
	}
	l.EnemySpawns = append(l.EnemySpawns, spawn)
}

// SetFlagPosition sets the position of the level's flag
func (l *Level) SetFlagPosition(x, y float64) {
	l.FlagPosition = FlagData{X: x, Y: y}
}

// SetPlayerStart sets the player's starting position
func (l *Level) SetPlayerStart(x, y float64) {
	l.PlayerStart = PlayerStartData{X: x, Y: y}
}

// SaveToFile saves the level to a JSON file
func (l *Level) SaveToFile(filename string) error {
	// Create levels directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Marshal level to JSON
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal level: %v", err)
	}

	// Write JSON to file
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write level file: %v", err)
	}

	return nil
}

// LoadFromFile loads a level from a JSON file
func LoadLevelFromFile(filename string) (*Level, error) {
	// Read file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read level file: %v", err)
	}

	// Unmarshal JSON to level
	var level Level
	if err := json.Unmarshal(data, &level); err != nil {
		return nil, fmt.Errorf("failed to unmarshal level: %v", err)
	}

	return &level, nil
}

// GetLevelList returns a list of available level files
func GetLevelList(directory string) ([]string, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	// Get list of JSON files in directory
	files, err := filepath.Glob(filepath.Join(directory, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list level files: %v", err)
	}

	// Extract level names from filenames
	var levels []string
	for _, file := range files {
		// Get filename without extension
		name := filepath.Base(file)
		name = name[:len(name)-len(filepath.Ext(name))]
		levels = append(levels, name)
	}

	return levels, nil
}

// CreateDefaultLevel creates the default level (same as the original hardcoded level)
func CreateDefaultLevel(width, height int) *Level {
	level := NewLevel("Default Level")
	
	// Set player start position
	level.SetPlayerStart(100, 100)
	
	// Create a much wider ground platform to allow for side-scrolling
	groundWidth := float64(width * 5) // 5 times the screen width
	level.AddPlatform(0, float64(height-50), groundWidth, 50)
	
	// Create platforms across the wider level
	// First screen
	level.AddPlatform(100, 400, 200, 20)
	level.AddPlatform(400, 300, 200, 20)
	level.AddPlatform(200, 200, 200, 20)
	
	// Second screen
	level.AddPlatform(float64(width) + 100, 400, 200, 20)
	level.AddPlatform(float64(width) + 400, 300, 200, 20)
	
	// Third screen
	level.AddPlatform(float64(width*2) + 100, 350, 200, 20)
	level.AddPlatform(float64(width*2) + 400, 250, 200, 20)
	
	// Fourth screen
	level.AddPlatform(float64(width*3) + 100, 300, 200, 20)
	level.AddPlatform(float64(width*3) + 400, 200, 200, 20)
	
	// Add some vertical platforms for variety
	level.AddPlatform(float64(width*4) - 100, 350, 50, 200)
	
	// Create finish flag at the end of the level
	flagX := float64(width*4) + 100 // Place flag at the far right of the level
	flagY := float64(height - 114)  // Place flag on the ground
	level.SetFlagPosition(flagX, flagY)
	
	// Add a platform under the flag
	level.AddPlatform(flagX - 50, flagY + 64, 150, 20)
	
	// Add enemy spawn points
	// First screen
	level.AddEnemySpawn(300, 350, "kidney")
	level.AddEnemySpawn(500, 250, "navy")
	level.AddEnemySpawn(300, 150, "kidney")
	
	// Second screen
	level.AddEnemySpawn(float64(width) + 200, 350, "navy")
	level.AddEnemySpawn(float64(width) + 500, 250, "kidney")
	
	// Third screen
	level.AddEnemySpawn(float64(width*2) + 200, 300, "navy")
	level.AddEnemySpawn(float64(width*2) + 500, 200, "kidney")
	
	// Fourth screen
	level.AddEnemySpawn(float64(width*3) + 200, 250, "navy")
	level.AddEnemySpawn(float64(width*3) + 500, 150, "kidney")
	
	return level
}