package engine

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateDefaultLevels creates 10 default levels (including the original level)
func CreateDefaultLevels(width, height int) error {
	// Create levels directory if it doesn't exist
	levelsDir := "levels"
	if err := os.MkdirAll(levelsDir, 0755); err != nil {
		return fmt.Errorf("failed to create levels directory: %v", err)
	}

	// Create 10 levels with increasing difficulty
	for i := 1; i <= 10; i++ {
		levelName := fmt.Sprintf("level%d", i)
		levelPath := filepath.Join(levelsDir, levelName+".json")

		// Skip if level file already exists
		if _, err := os.Stat(levelPath); err == nil {
			fmt.Printf("Level %s already exists, skipping\n", levelName)
			continue
		}

		// Create level
		level := createLevelByIndex(i, width, height)

		// Save level to file
		if err := level.SaveToFile(levelPath); err != nil {
			return fmt.Errorf("failed to save level %s: %v", levelName, err)
		}

		fmt.Printf("Created level %s\n", levelName)
	}

	return nil
}

// createLevelByIndex creates a level with the specified index
func createLevelByIndex(index, width, height int) *Level {
	levelName := fmt.Sprintf("Level %d", index)
	level := NewLevel(levelName)

	// Set player start position
	level.SetPlayerStart(100, 100)

	// Create ground platform (common to all levels)
	groundWidth := float64(width * 5) // 5 times the screen width
	level.AddPlatform(0, float64(height-50), groundWidth, 50)

	// Create level-specific platforms and enemies
	switch index {
	case 1:
		// Level 1: Original level (tutorial)
		createLevel1(level, width, height)
	case 2:
		// Level 2: Staircase level
		createLevel2(level, width, height)
	case 3:
		// Level 3: Vertical challenge
		createLevel3(level, width, height)
	case 4:
		// Level 4: Precision jumps
		createLevel4(level, width, height)
	case 5:
		// Level 5: Enemy gauntlet
		createLevel5(level, width, height)
	case 6:
		// Level 6: Platform maze
		createLevel6(level, width, height)
	case 7:
		// Level 7: Vertical maze
		createLevel7(level, width, height)
	case 8:
		// Level 8: Long jumps
		createLevel8(level, width, height)
	case 9:
		// Level 9: Enemy fortress
		createLevel9(level, width, height)
	case 10:
		// Level 10: Final challenge
		createLevel10(level, width, height)
	}

	return level
}

// createLevel1 creates the original level (tutorial)
func createLevel1(level *Level, width, height int) {
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
}

// createLevel2 creates a staircase level
func createLevel2(level *Level, width, height int) {
	// Staircase platforms
	stepCount := 15
	stepWidth := 150.0
	stepHeight := 20.0
	stepSpacing := 200.0

	for i := 0; i < stepCount; i++ {
		x := 100.0 + float64(i) * stepSpacing
		y := float64(height - 150) - float64(i) * 50.0
		level.AddPlatform(x, y, stepWidth, stepHeight)

		// Add enemy on some steps
		if i % 2 == 0 {
			level.AddEnemySpawn(x + stepWidth/2, y - 50, "kidney")
		}
	}

	// Create finish flag at the end of the staircase
	flagX := 100.0 + float64(stepCount-1) * stepSpacing + stepWidth + 100
	flagY := float64(height - 150) - float64(stepCount-1) * 50.0 - 64
	level.SetFlagPosition(flagX, flagY)

	// Add a platform under the flag
	level.AddPlatform(flagX - 50, flagY + 64, 150, 20)
}

// createLevel3 creates a vertical challenge level
func createLevel3(level *Level, width, height int) {
	// Starting platform
	level.AddPlatform(100, 400, 200, 20)

	// Vertical platforms
	level.AddPlatform(400, 350, 50, 200)
	level.AddPlatform(600, 250, 50, 300)
	level.AddPlatform(800, 150, 50, 400)

	// Horizontal platforms between vertical ones
	level.AddPlatform(450, 350, 150, 20)
	level.AddPlatform(650, 250, 150, 20)
	level.AddPlatform(850, 150, 150, 20)

	// Final platform
	level.AddPlatform(1100, 100, 200, 20)

	// Create finish flag at the end
	flagX := 1200.0
	flagY := 36.0
	level.SetFlagPosition(flagX, flagY)

	// Add enemies
	level.AddEnemySpawn(450, 300, "kidney")
	level.AddEnemySpawn(650, 200, "navy")
	level.AddEnemySpawn(850, 100, "kidney")
	level.AddEnemySpawn(1150, 50, "navy")
}

// createLevel4 creates a precision jumps level
func createLevel4(level *Level, width, height int) {
	// Starting platform
	level.AddPlatform(100, 400, 200, 20)

	// Small platforms with increasing gaps
	platformCount := 10
	platformWidth := 80.0
	platformHeight := 20.0
	startX := 400.0
	y := 400.0

	for i := 0; i < platformCount; i++ {
		gap := 100.0 + float64(i) * 20.0
		x := startX + float64(i) * (platformWidth + gap)
		level.AddPlatform(x, y, platformWidth, platformHeight)

		// Add enemy on some platforms
		if i % 3 == 0 {
			level.AddEnemySpawn(x + platformWidth/2, y - 50, "navy")
		}
	}

	// Final platform
	finalX := startX + float64(platformCount) * (platformWidth + 200.0)
	level.AddPlatform(finalX, y, 200, 20)

	// Create finish flag at the end
	flagX := finalX + 100.0
	flagY := y - 64.0
	level.SetFlagPosition(flagX, flagY)
}

// createLevel5 creates an enemy gauntlet level
func createLevel5(level *Level, width, height int) {
	// Long platform with many enemies
	platformWidth := float64(width * 3)
	level.AddPlatform(100, 400, platformWidth, 20)

	// Add many enemies along the platform
	enemyCount := 15
	enemySpacing := platformWidth / float64(enemyCount)

	for i := 0; i < enemyCount; i++ {
		x := 100.0 + float64(i) * enemySpacing
		enemyType := "kidney"
		if i % 2 == 0 {
			enemyType = "navy"
		}
		level.AddEnemySpawn(x, 350, enemyType)
	}

	// Create finish flag at the end
	flagX := 100.0 + platformWidth + 100.0
	flagY := 400.0 - 64.0
	level.SetFlagPosition(flagX, flagY)

	// Add a platform under the flag
	level.AddPlatform(flagX - 50, flagY + 64, 150, 20)
}

// createLevel6 creates a platform maze level
func createLevel6(level *Level, width, height int) {
	// Starting platform
	level.AddPlatform(100, 400, 200, 20)

	// Maze of platforms
	// Row 1
	level.AddPlatform(400, 350, 100, 20)
	level.AddPlatform(600, 300, 100, 20)
	level.AddPlatform(800, 350, 100, 20)
	level.AddPlatform(1000, 300, 100, 20)

	// Row 2
	level.AddPlatform(300, 250, 100, 20)
	level.AddPlatform(500, 200, 100, 20)
	level.AddPlatform(700, 250, 100, 20)
	level.AddPlatform(900, 200, 100, 20)
	level.AddPlatform(1100, 250, 100, 20)

	// Row 3
	level.AddPlatform(400, 150, 100, 20)
	level.AddPlatform(600, 100, 100, 20)
	level.AddPlatform(800, 150, 100, 20)
	level.AddPlatform(1000, 100, 100, 20)

	// Final platform
	level.AddPlatform(1200, 150, 200, 20)

	// Add enemies
	level.AddEnemySpawn(400, 300, "kidney")
	level.AddEnemySpawn(600, 250, "navy")
	level.AddEnemySpawn(800, 300, "kidney")
	level.AddEnemySpawn(1000, 250, "navy")
	level.AddEnemySpawn(500, 150, "kidney")
	level.AddEnemySpawn(700, 200, "navy")
	level.AddEnemySpawn(900, 150, "kidney")
	level.AddEnemySpawn(1100, 200, "navy")
	level.AddEnemySpawn(1200, 100, "kidney")

	// Create finish flag at the end
	flagX := 1300.0
	flagY := 86.0
	level.SetFlagPosition(flagX, flagY)
}

// createLevel7 creates a vertical maze level
func createLevel7(level *Level, width, height int) {
	// Starting platform
	level.AddPlatform(100, 400, 200, 20)

	// Vertical maze
	// Column 1
	level.AddPlatform(400, 450, 20, 100)
	level.AddPlatform(400, 250, 20, 100)

	// Column 2
	level.AddPlatform(600, 350, 20, 100)
	level.AddPlatform(600, 150, 20, 100)

	// Column 3
	level.AddPlatform(800, 450, 20, 100)
	level.AddPlatform(800, 250, 20, 100)

	// Column 4
	level.AddPlatform(1000, 350, 20, 100)
	level.AddPlatform(1000, 150, 20, 100)

	// Horizontal platforms
	level.AddPlatform(400, 350, 200, 20)
	level.AddPlatform(600, 250, 200, 20)
	level.AddPlatform(800, 150, 200, 20)
	level.AddPlatform(1000, 50, 200, 20)

	// Add enemies
	level.AddEnemySpawn(500, 300, "kidney")
	level.AddEnemySpawn(700, 200, "navy")
	level.AddEnemySpawn(900, 100, "kidney")

	// Create finish flag at the end
	flagX := 1100.0
	flagY := -14.0
	level.SetFlagPosition(flagX, flagY)
}

// createLevel8 creates a long jumps level
func createLevel8(level *Level, width, height int) {
	// Starting platform
	level.AddPlatform(100, 400, 200, 20)

	// Platforms with increasing gaps
	level.AddPlatform(500, 400, 100, 20)
	level.AddPlatform(800, 350, 100, 20)
	level.AddPlatform(1200, 350, 100, 20)
	level.AddPlatform(1600, 300, 100, 20)
	level.AddPlatform(2100, 300, 100, 20)
	level.AddPlatform(2600, 250, 100, 20)
	level.AddPlatform(3200, 250, 200, 20)

	// Add enemies
	level.AddEnemySpawn(500, 350, "kidney")
	level.AddEnemySpawn(800, 300, "navy")
	level.AddEnemySpawn(1200, 300, "kidney")
	level.AddEnemySpawn(1600, 250, "navy")
	level.AddEnemySpawn(2100, 250, "kidney")
	level.AddEnemySpawn(2600, 200, "navy")
	level.AddEnemySpawn(3200, 200, "kidney")
	level.AddEnemySpawn(3300, 200, "navy")

	// Create finish flag at the end
	flagX := 3300.0
	flagY := 186.0
	level.SetFlagPosition(flagX, flagY)
}

// createLevel9 creates an enemy fortress level
func createLevel9(level *Level, width, height int) {
	// Starting platform
	level.AddPlatform(100, 400, 200, 20)

	// Fortress walls
	level.AddPlatform(800, 100, 20, 350) // Left wall
	level.AddPlatform(1200, 100, 20, 350) // Right wall
	level.AddPlatform(800, 100, 420, 20) // Top wall

	// Platforms leading to fortress
	level.AddPlatform(400, 350, 100, 20)
	level.AddPlatform(600, 300, 100, 20)

	// Platforms inside fortress
	level.AddPlatform(850, 300, 100, 20)
	level.AddPlatform(1000, 250, 100, 20)
	level.AddPlatform(850, 200, 100, 20)

	// Add many enemies guarding the fortress
	// Outside
	level.AddEnemySpawn(400, 300, "kidney")
	level.AddEnemySpawn(600, 250, "navy")
	level.AddEnemySpawn(750, 350, "kidney")

	// Inside
	level.AddEnemySpawn(850, 250, "navy")
	level.AddEnemySpawn(1000, 200, "kidney")
	level.AddEnemySpawn(850, 150, "navy")
	level.AddEnemySpawn(950, 150, "kidney")
	level.AddEnemySpawn(1050, 150, "navy")
	level.AddEnemySpawn(1150, 150, "kidney")

	// Create finish flag inside the fortress
	flagX := 1100.0
	flagY := 36.0
	level.SetFlagPosition(flagX, flagY)
}

// createLevel10 creates the final challenge level
func createLevel10(level *Level, width, height int) {
	// Starting platform
	level.AddPlatform(100, 400, 200, 20)

	// First section: Precision jumps
	level.AddPlatform(400, 400, 50, 20)
	level.AddPlatform(550, 350, 50, 20)
	level.AddPlatform(700, 400, 50, 20)
	level.AddPlatform(850, 350, 50, 20)

	// Second section: Vertical challenge
	level.AddPlatform(1000, 400, 200, 20)
	level.AddPlatform(1300, 350, 20, 200)
	level.AddPlatform(1320, 250, 100, 20)
	level.AddPlatform(1500, 200, 20, 200)
	level.AddPlatform(1400, 150, 100, 20)

	// Third section: Enemy gauntlet
	level.AddPlatform(1600, 150, 500, 20)

	// Fourth section: Final challenge
	level.AddPlatform(2200, 200, 50, 20)
	level.AddPlatform(2350, 250, 50, 20)
	level.AddPlatform(2500, 200, 50, 20)
	level.AddPlatform(2650, 150, 50, 20)
	level.AddPlatform(2800, 100, 200, 20)

	// Add enemies throughout the level
	// First section
	level.AddEnemySpawn(400, 350, "kidney")
	level.AddEnemySpawn(550, 300, "navy")
	level.AddEnemySpawn(700, 350, "kidney")
	level.AddEnemySpawn(850, 300, "navy")

	// Second section
	level.AddEnemySpawn(1100, 350, "kidney")
	level.AddEnemySpawn(1320, 200, "navy")
	level.AddEnemySpawn(1400, 100, "kidney")

	// Third section (enemy gauntlet)
	for i := 0; i < 8; i++ {
		x := 1650.0 + float64(i) * 50.0
		enemyType := "kidney"
		if i % 2 == 0 {
			enemyType = "navy"
		}
		level.AddEnemySpawn(x, 100, enemyType)
	}

	// Fourth section
	level.AddEnemySpawn(2200, 150, "kidney")
	level.AddEnemySpawn(2350, 200, "navy")
	level.AddEnemySpawn(2500, 150, "kidney")
	level.AddEnemySpawn(2650, 100, "navy")
	level.AddEnemySpawn(2800, 50, "kidney")
	level.AddEnemySpawn(2900, 50, "navy")

	// Create finish flag at the end
	flagX := 2900.0
	flagY := 36.0
	level.SetFlagPosition(flagX, flagY)
}
