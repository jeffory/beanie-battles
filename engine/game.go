package engine

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Game represents the main game state and implements ebiten.Game interface
type Game struct {
	width          int
	height         int
	world          *World
	player         *Player
	platforms      []*Platform
	bullets        []*Bullet
	enemies        []*Enemy
	flag           *Flag
	gravity        float64
	debug          bool
	frameCount     int
	enemySpawnTimer int
	levelComplete  bool
}

// NewGame creates a new Game instance
func NewGame(width, height int) (*Game, error) {
	// Create a new game instance
	g := &Game{
		width:          width,
		height:         height,
		gravity:        0.5,
		debug:          true,
		bullets:        make([]*Bullet, 0),
		enemies:        make([]*Enemy, 0),
		frameCount:     0,
		enemySpawnTimer: 300, // Spawn first enemy after 5 seconds
		levelComplete:  false,
	}

 	// Initialize the world with camera
	g.world = NewWorld(g.gravity, width, height)

	// Create player
	g.player = NewPlayer(100, 100)
	g.world.AddEntity(g.player)

	// Create platforms
	g.createPlatforms()

	return g, nil
}

// createPlatforms initializes the game platforms
func (g *Game) createPlatforms() {
	// Create a much wider ground platform to allow for side-scrolling
	groundWidth := float64(g.width * 5) // 5 times the screen width
	ground := NewPlatform(0, float64(g.height-50), groundWidth, 50)
	g.platforms = append(g.platforms, ground)
	g.world.AddEntity(ground)

	// Create platforms across the wider level
	// First screen
	p1 := NewPlatform(100, 400, 200, 20)
	g.platforms = append(g.platforms, p1)
	g.world.AddEntity(p1)

	p2 := NewPlatform(400, 300, 200, 20)
	g.platforms = append(g.platforms, p2)
	g.world.AddEntity(p2)

	p3 := NewPlatform(200, 200, 200, 20)
	g.platforms = append(g.platforms, p3)
	g.world.AddEntity(p3)

	// Second screen
	p4 := NewPlatform(float64(g.width) + 100, 400, 200, 20)
	g.platforms = append(g.platforms, p4)
	g.world.AddEntity(p4)

	p5 := NewPlatform(float64(g.width) + 400, 300, 200, 20)
	g.platforms = append(g.platforms, p5)
	g.world.AddEntity(p5)

	// Third screen
	p6 := NewPlatform(float64(g.width*2) + 100, 350, 200, 20)
	g.platforms = append(g.platforms, p6)
	g.world.AddEntity(p6)

	p7 := NewPlatform(float64(g.width*2) + 400, 250, 200, 20)
	g.platforms = append(g.platforms, p7)
	g.world.AddEntity(p7)

	// Fourth screen
	p8 := NewPlatform(float64(g.width*3) + 100, 300, 200, 20)
	g.platforms = append(g.platforms, p8)
	g.world.AddEntity(p8)

	p9 := NewPlatform(float64(g.width*3) + 400, 200, 200, 20)
	g.platforms = append(g.platforms, p9)
	g.world.AddEntity(p9)

	// Add some vertical platforms for variety
	v1 := NewPlatform(float64(g.width*4) - 100, 350, 50, 200)
	g.platforms = append(g.platforms, v1)
	g.world.AddEntity(v1)

	// Create finish flag at the end of the level
	flagX := float64(g.width*4) + 100 // Place flag at the far right of the level
	flagY := float64(g.height - 114)  // Place flag on the ground
	g.flag = NewFlag(flagX, flagY)
	g.world.AddEntity(g.flag)

	// Add a platform under the flag
	flagPlatform := NewPlatform(flagX - 50, flagY + 64, 150, 20)
	g.platforms = append(g.platforms, flagPlatform)
	g.world.AddEntity(flagPlatform)
}

// FireBullet creates a new bullet and adds it to the game
func (g *Game) FireBullet() {
	// Check if player is reloading
	if g.player.IsReloading {
		return
	}

	// Check if clip is empty
	if g.player.CurrentClip <= 0 {
		// Auto-reload if clip is empty and player has ammo
		if g.player.BulletCount > 0 {
			g.player.StartReload()
		}
		return
	}

	// Check if enough time has passed since last shot
	if g.frameCount - g.player.LastFired < g.player.FireRate {
		return
	}

	// Determine bullet position and velocity based on player direction
	bulletX := g.player.X
	bulletVelX := 10.0 // Default bullet speed

	if g.player.FacingRight {
		bulletX += g.player.Width // Bullet starts at right edge of player
	} else {
		bulletX -= 8 // Bullet starts at left edge of player
		bulletVelX = -10.0 // Bullet moves left
	}

	// Create new bullet
	bullet := NewBullet(bulletX, g.player.Y + g.player.Height/2 - 2, bulletVelX)

	// Add bullet to game and world
	g.bullets = append(g.bullets, bullet)
	g.world.AddEntity(bullet)

	// Update player state
	g.player.CurrentClip--
	g.player.LastFired = g.frameCount
}

// SpawnEnemy creates a new enemy and adds it to the game
func (g *Game) SpawnEnemy() {
	// Determine spawn position based on player position
	playerX := g.player.X

	// Spawn enemies at different locations based on where the player is in the level
	var spawnX, spawnY float64

	// Determine which screen section the player is in
	screenSection := int(playerX / float64(g.width))

	// Choose a spawn location based on screen section
	switch screenSection {
	case 0: // First screen
		// Spawn on first screen platforms
		spawnLocations := []struct{ x, y float64 }{
			{300, 350},  // On platform p1
			{500, 250},  // On platform p2
			{300, 150},  // On platform p3
		}
		spawnPoint := spawnLocations[g.frameCount%len(spawnLocations)]
		spawnX = spawnPoint.x
		spawnY = spawnPoint.y
	case 1: // Second screen
		// Spawn on second screen platforms
		spawnLocations := []struct{ x, y float64 }{
			{float64(g.width) + 200, 350},  // On platform p4
			{float64(g.width) + 500, 250},  // On platform p5
		}
		spawnPoint := spawnLocations[g.frameCount%len(spawnLocations)]
		spawnX = spawnPoint.x
		spawnY = spawnPoint.y
	case 2: // Third screen
		// Spawn on third screen platforms
		spawnLocations := []struct{ x, y float64 }{
			{float64(g.width*2) + 200, 300},  // On platform p6
			{float64(g.width*2) + 500, 200},  // On platform p7
		}
		spawnPoint := spawnLocations[g.frameCount%len(spawnLocations)]
		spawnX = spawnPoint.x
		spawnY = spawnPoint.y
	case 3, 4: // Fourth and fifth screens (near the end)
		// Spawn on fourth screen platforms or ahead of player
		spawnLocations := []struct{ x, y float64 }{
			{float64(g.width*3) + 200, 250},  // On platform p8
			{float64(g.width*3) + 500, 150},  // On platform p9
			{playerX + float64(g.width/2), float64(g.height - 150)}, // Ahead of player
		}
		spawnPoint := spawnLocations[g.frameCount%len(spawnLocations)]
		spawnX = spawnPoint.x
		spawnY = spawnPoint.y
	default: // Beyond the level
		// Spawn near the flag to protect it
		spawnX = g.flag.X - 200 + float64(g.frameCount%400)
		spawnY = float64(g.height - 150)
	}

	// Create new enemy with random variations
	enemy := NewEnemy(spawnX, spawnY)

	// Create bean enemies: kidney beans (dark red) and navy beans (dark blue)
	if g.frameCount%2 == 0 {
		// Kidney bean (faster)
		enemy.Speed = 3
		enemy.Color = color.RGBA{139, 0, 0, 255} // Dark red for kidney beans
		enemy.Health = 50
		enemy.MaxHealth = 50
	} else {
		// Navy bean (tougher)
		enemy.Speed = 2
		enemy.Health = 75
		enemy.MaxHealth = 75
		enemy.Color = color.RGBA{0, 0, 128, 255} // Dark blue for navy beans
	}

	// Add enemy to game and world
	g.enemies = append(g.enemies, enemy)
	g.world.AddEntity(enemy)
}

// Update updates the game state
func (g *Game) Update() error {
	// Increment frame counter
	g.frameCount++

	// Handle player input
	g.handleInput()

	// Update physics
	g.world.Update()

	// Check if player has fallen off the level
	g.checkPlayerDeath()

	// Make camera follow the player
	g.world.GetCamera().Follow(g.player, g.width, g.height)

	// Handle enemy spawning (only if level not complete)
	if !g.levelComplete {
		g.enemySpawnTimer--
		if g.enemySpawnTimer <= 0 {
			g.SpawnEnemy()
			g.enemySpawnTimer = 180 // Spawn enemies more frequently (every 3 seconds)
		}
	}

	// Clean up inactive bullets and enemies
	g.cleanupBullets()
	g.cleanupEnemies()

	// Check for bullet-enemy collisions
	g.checkBulletEnemyCollisions()

	// Check for player-enemy collisions
	g.checkPlayerEnemyCollisions()

	// Check for player-flag collision
	g.checkPlayerFlagCollision()

	return nil
}

// checkPlayerFlagCollision checks if the player has reached the flag
func (g *Game) checkPlayerFlagCollision() {
	// Skip if flag is already collected or not active
	if g.flag == nil || !g.flag.Active || g.flag.Collected || g.levelComplete {
		return
	}

	// Simple AABB collision detection
	if g.player.X < g.flag.X+g.flag.Width &&
		g.player.X+g.player.Width > g.flag.X &&
		g.player.Y < g.flag.Y+g.flag.Height &&
		g.player.Y+g.player.Height > g.flag.Y {

		// Player reached the flag
		g.flag.Collected = true
		g.levelComplete = true

		// Refill player health and ammo as a reward
		g.player.Health = g.player.MaxHealth
		g.player.BulletCount = 200
		g.player.CurrentClip = g.player.ClipSize
	}
}

// cleanupEnemies removes inactive enemies from the game
func (g *Game) cleanupEnemies() {
	activeEnemies := make([]*Enemy, 0)

	for _, enemy := range g.enemies {
		if enemy.Active {
			activeEnemies = append(activeEnemies, enemy)
		}
	}

	g.enemies = activeEnemies
}

// checkBulletEnemyCollisions checks for collisions between bullets and enemies
func (g *Game) checkBulletEnemyCollisions() {
	for _, bullet := range g.bullets {
		if !bullet.Active {
			continue
		}

		for _, enemy := range g.enemies {
			if !enemy.Active || enemy.Dead {
				continue
			}

			// Simple AABB collision detection
			if bullet.X < enemy.X+enemy.Width &&
				bullet.X+bullet.Width > enemy.X &&
				bullet.Y < enemy.Y+enemy.Height &&
				bullet.Y+bullet.Height > enemy.Y {

				// Bullet hit enemy
				enemy.TakeDamage(bullet.Damage)
				bullet.Active = false
				break
			}
		}
	}
}

// checkPlayerEnemyCollisions checks for collisions between player and enemies
func (g *Game) checkPlayerEnemyCollisions() {
	for _, enemy := range g.enemies {
		if !enemy.Active || enemy.Dead {
			continue
		}

		// Simple AABB collision detection
		if g.player.X < enemy.X+enemy.Width &&
			g.player.X+g.player.Width > enemy.X &&
			g.player.Y < enemy.Y+enemy.Height &&
			g.player.Y+g.player.Height > enemy.Y {

			// Enemy hit player
			g.player.TakeDamage(10) // Enemy deals 10 damage
			break
		}
	}
}

// cleanupBullets removes inactive bullets from the game
func (g *Game) cleanupBullets() {
	activeBullets := make([]*Bullet, 0)

	for _, bullet := range g.bullets {
		if bullet.Active {
			activeBullets = append(activeBullets, bullet)
		}
	}

	g.bullets = activeBullets
}

// checkPlayerDeath checks if the player has died and handles respawning
func (g *Game) checkPlayerDeath() {
	// Define death boundary (some distance below the bottom of the screen)
	deathBoundary := float64(g.height + 200)

	// Check if player has fallen below the death boundary
	if g.player.Y > deathBoundary && !g.player.Dead {
		g.player.Dead = true
	}

	// Check if player is dead (from health or falling)
	if g.player.Dead {
		g.player.Reset()
	}
}

// handleInput processes player input
func (g *Game) handleInput() {
	// If level is complete, only handle continue input
	if g.levelComplete {
		// Press Enter to continue
		if ebiten.IsKeyPressed(ebiten.KeyEnter) && !g.isKeyPressedPreviously(ebiten.KeyEnter) {
			// Reset the level (in a real game, you'd load the next level)
			g.levelComplete = false
			g.flag.Collected = false

			// Reset player position
			g.player.X = 100
			g.player.Y = 100
			g.player.VelX = 0
			g.player.VelY = 0

			// Reset camera
			g.world.GetCamera().X = 0

			// Clear enemies
			for _, enemy := range g.enemies {
				enemy.Active = false
			}
			g.cleanupEnemies()
		}
		return
	}

	// Left/Right movement
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.player.MoveLeft()
		g.player.FacingRight = false
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.player.MoveRight()
		g.player.FacingRight = true
	} else {
		g.player.StopHorizontal()
	}

	// Jump
	if ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.player.Jump()
	}

	// Shoot
	if ebiten.IsKeyPressed(ebiten.KeyX) {
		g.FireBullet()
	}

	// Toggle debug mode
	if ebiten.IsKeyPressed(ebiten.KeyD) && !g.isKeyPressedPreviously(ebiten.KeyD) {
		g.debug = !g.debug
	}

	// Reload (R key)
	if ebiten.IsKeyPressed(ebiten.KeyR) && !g.isKeyPressedPreviously(ebiten.KeyR) {
		g.player.StartReload()
	}

	// Testing keys

	// Refill ammo (F key)
	if ebiten.IsKeyPressed(ebiten.KeyF) && !g.isKeyPressedPreviously(ebiten.KeyF) {
		g.player.BulletCount = 200
		g.player.CurrentClip = g.player.ClipSize
	}

	// Take damage (T key)
	if ebiten.IsKeyPressed(ebiten.KeyT) && !g.isKeyPressedPreviously(ebiten.KeyT) {
		g.player.TakeDamage(20)
	}
}

// isKeyPressedPreviously is a helper to detect key press events
func (g *Game) isKeyPressedPreviously(key ebiten.Key) bool {
	// This is a simplified version - in a real game you'd track previous state
	return false
}

// Draw draws the game
func (g *Game) Draw(screen *ebiten.Image) {
	// Draw a gradient background
	g.drawBackground(screen)

	// Draw all entities
	g.world.Draw(screen)

	// Draw UI elements
	g.drawUI(screen)

	// Draw debug info
	if g.debug {
		g.drawDebugInfo(screen)
	}
}

// drawBackground creates a gradient sky background
func (g *Game) drawBackground(screen *ebiten.Image) {
	// Sky gradient from top to bottom
	skyTopColor := color.RGBA{100, 181, 246, 255}    // Light blue at top
	skyBottomColor := color.RGBA{187, 222, 251, 255} // Very light blue at bottom

	// Draw gradient by creating horizontal strips
	strips := 20
	stripHeight := float64(g.height) / float64(strips)

	for i := 0; i < strips; i++ {
		// Calculate interpolated color
		t := float64(i) / float64(strips-1)
		r := int(float64(skyTopColor.R)*(1-t) + float64(skyBottomColor.R)*t)
		green := int(float64(skyTopColor.G)*(1-t) + float64(skyBottomColor.G)*t)
		b := int(float64(skyTopColor.B)*(1-t) + float64(skyBottomColor.B)*t)

		stripColor := color.RGBA{uint8(r), uint8(green), uint8(b), 255}
		ebitenutil.DrawRect(screen, 0, float64(i)*stripHeight, float64(g.width), stripHeight, stripColor)
	}

	// Draw some clouds
	cloudColor := color.RGBA{255, 255, 255, 180}

	// Cloud positions based on camera position for parallax effect
	cameraOffset := g.world.GetCamera().X * 0.3 // Clouds move slower than the camera

	// Draw several clouds at different positions
	cloudPositions := []struct{ x, y, width, height float64 }{
		{100 - float64(int(cameraOffset)%1000), 50, 120, 40},
		{400 - float64(int(cameraOffset)%1200), 100, 150, 50},
		{700 - float64(int(cameraOffset)%900), 70, 100, 30},
		{1000 - float64(int(cameraOffset)%1500), 120, 180, 60},
	}

	for _, cloud := range cloudPositions {
		// Draw main cloud body
		ebitenutil.DrawRect(screen, cloud.x, cloud.y, cloud.width, cloud.height, cloudColor)

		// Draw cloud bumps for a more cloud-like shape
		ebitenutil.DrawRect(screen, cloud.x-20, cloud.y+10, 40, cloud.height-20, cloudColor)
		ebitenutil.DrawRect(screen, cloud.x+cloud.width-20, cloud.y+15, 40, cloud.height-25, cloudColor)
		ebitenutil.DrawRect(screen, cloud.x+cloud.width/2-15, cloud.y-10, 50, 20, cloudColor)
	}
}

// drawUI draws the UI elements (bullet count, health bar)
func (g *Game) drawUI(screen *ebiten.Image) {
	// Draw ammo info in top right corner with larger text and better styling
	clipText := fmt.Sprintf("AMMO: %d / %d", g.player.CurrentClip, g.player.BulletCount)

	// Draw reload indicator if reloading
	if g.player.IsReloading {
		reloadProgress := float64(g.player.ReloadTime - g.player.ReloadTimer) / float64(g.player.ReloadTime) * 100
		clipText = fmt.Sprintf("RELOADING... %.0f%%", reloadProgress)
	}

	// Draw a stylized background for the ammo text
	textWidth := len(clipText) * 7 // Approximate width based on character count
	ammoBoxWidth := float64(textWidth + 20)
	ammoBoxHeight := 30.0
	ammoBoxX := float64(g.width) - ammoBoxWidth - 10
	ammoBoxY := 10.0

	// Draw main background with border
	ebitenutil.DrawRect(screen, ammoBoxX, ammoBoxY, ammoBoxWidth, ammoBoxHeight, color.RGBA{40, 40, 40, 220})

	// Draw border
	borderSize := 2.0
	ebitenutil.DrawRect(screen, ammoBoxX, ammoBoxY, ammoBoxWidth, borderSize, color.RGBA{255, 215, 0, 255}) // Gold top
	ebitenutil.DrawRect(screen, ammoBoxX, ammoBoxY + ammoBoxHeight - borderSize, ammoBoxWidth, borderSize, color.RGBA{255, 215, 0, 255}) // Gold bottom
	ebitenutil.DrawRect(screen, ammoBoxX, ammoBoxY, borderSize, ammoBoxHeight, color.RGBA{255, 215, 0, 255}) // Gold left
	ebitenutil.DrawRect(screen, ammoBoxX + ammoBoxWidth - borderSize, ammoBoxY, borderSize, ammoBoxHeight, color.RGBA{255, 215, 0, 255}) // Gold right

	// Draw bullet icon
	bulletIconX := ammoBoxX + 5
	bulletIconY := ammoBoxY + ammoBoxHeight/2 - 3
	bulletIconWidth := 10.0
	bulletIconHeight := 6.0
	ebitenutil.DrawRect(screen, bulletIconX, bulletIconY, bulletIconWidth, bulletIconHeight, color.RGBA{255, 255, 0, 255})

	// Draw text with offset to center vertically
	textY := int(ammoBoxY) + int(ammoBoxHeight)/2 - 3
	ebitenutil.DebugPrintAt(screen, clipText, int(ammoBoxX) + 20, textY)

	// Draw health bar with improved styling
	healthBarWidth := 200
	healthBarHeight := 25
	healthBarX := 20
	healthBarY := 20
	healthBarBorderSize := 2

	// Draw health bar container with border
	ebitenutil.DrawRect(screen, float64(healthBarX-healthBarBorderSize), float64(healthBarY-healthBarBorderSize), 
		float64(healthBarWidth+healthBarBorderSize*2), float64(healthBarHeight+healthBarBorderSize*2), 
		color.RGBA{40, 40, 40, 220})

	// Draw health bar background (dark red)
	ebitenutil.DrawRect(screen, float64(healthBarX), float64(healthBarY), 
		float64(healthBarWidth), float64(healthBarHeight), 
		color.RGBA{120, 0, 0, 255})

	// Draw health bar foreground with gradient effect
	healthWidth := float64(healthBarWidth) * float64(g.player.Health) / float64(g.player.MaxHealth)

	// Health color changes based on amount (red->yellow->green)
	var healthColor color.RGBA
	healthPercent := float64(g.player.Health) / float64(g.player.MaxHealth)

	if healthPercent < 0.3 {
		// Red for low health
		healthColor = color.RGBA{255, 0, 0, 255}
	} else if healthPercent < 0.6 {
		// Yellow for medium health
		healthColor = color.RGBA{255, 255, 0, 255}
	} else {
		// Green for high health
		healthColor = color.RGBA{0, 255, 0, 255}
	}

	ebitenutil.DrawRect(screen, float64(healthBarX), float64(healthBarY), 
		healthWidth, float64(healthBarHeight), healthColor)

	// Add highlight at the top of the health bar for 3D effect
	highlightHeight := 3.0
	highlightColor := color.RGBA{255, 255, 255, 100}
	ebitenutil.DrawRect(screen, float64(healthBarX), float64(healthBarY), 
		healthWidth, highlightHeight, highlightColor)

	// Draw health text with shadow for better visibility
	healthText := fmt.Sprintf("HEALTH: %d", g.player.Health)

	// Draw text shadow
	shadowOffset := 1
	ebitenutil.DebugPrintAt(screen, healthText, healthBarX+5+shadowOffset, healthBarY+8+shadowOffset)

	// Draw text
	ebitenutil.DebugPrintAt(screen, healthText, healthBarX+5, healthBarY+8)

	// Draw level completion message if level is complete
	if g.levelComplete {
		// Draw a fancy background for the message
		msgWidth := 400
		msgHeight := 150
		msgX := (g.width - msgWidth) / 2
		msgY := (g.height - msgHeight) / 2

		// Draw outer glow
		glowSize := 10
		glowColor := color.RGBA{255, 215, 0, 100} // Gold with transparency
		ebitenutil.DrawRect(screen, 
			float64(msgX-glowSize), float64(msgY-glowSize), 
			float64(msgWidth+glowSize*2), float64(msgHeight+glowSize*2), 
			glowColor)

		// Draw main background
		ebitenutil.DrawRect(screen, float64(msgX), float64(msgY), 
			float64(msgWidth), float64(msgHeight), 
			color.RGBA{0, 0, 60, 230})

		// Draw decorative border
		borderSize := 3.0
		borderColor := color.RGBA{255, 215, 0, 255} // Gold

		// Top border
		ebitenutil.DrawRect(screen, float64(msgX), float64(msgY), 
			float64(msgWidth), borderSize, borderColor)

		// Bottom border
		ebitenutil.DrawRect(screen, float64(msgX), float64(msgY+msgHeight-int(borderSize)), 
			float64(msgWidth), borderSize, borderColor)

		// Left border
		ebitenutil.DrawRect(screen, float64(msgX), float64(msgY), 
			borderSize, float64(msgHeight), borderColor)

		// Right border
		ebitenutil.DrawRect(screen, float64(msgX+msgWidth-int(borderSize)), float64(msgY), 
			borderSize, float64(msgHeight), borderColor)

		// Draw stars/sparkles around the message
		starCount := 8
		starSize := 8.0
		for i := 0; i < starCount; i++ {
			// Calculate positions in a pattern around the message box
			radius := float64(msgWidth/2 + 20)

			// Use sine and cosine for circular positioning based on frame count for animation
			baseAngle := float64(g.frameCount % 360) * 0.01
			angleRad := baseAngle + float64(i) * (2 * 3.14159 / float64(starCount))

			// Calculate offsets using sine and cosine for circular movement
			offsetX := radius * 0.8 * math.Cos(angleRad) * 0.5
			offsetY := radius * 0.3 * math.Sin(angleRad) * 0.5

			starX := float64(msgX) + float64(msgWidth)/2 + offsetX - starSize/2
			starY := float64(msgY) + float64(msgHeight)/2 + offsetY - starSize/2

			// Alternate star colors
			starColor := color.RGBA{255, 255, 0, 255} // Yellow
			if i%2 == 0 {
				starColor = color.RGBA{255, 215, 0, 255} // Gold
			}

			// Draw star (simple square for now)
			ebitenutil.DrawRect(screen, starX, starY, starSize, starSize, starColor)

			// Add smaller star nearby for effect
			ebitenutil.DrawRect(screen, starX+15, starY-10, starSize/2, starSize/2, starColor)
		}

		// Draw completion message with shadow for better visibility
		completeMsg := "LEVEL COMPLETE!"
		shadowOffset := 2

		// Draw text shadow
		ebitenutil.DebugPrintAt(screen, completeMsg, 
			msgX + (msgWidth-len(completeMsg)*7)/2 + shadowOffset, 
			msgY + 50 + shadowOffset)

		// Draw text in gold
		ebitenutil.DebugPrintAt(screen, completeMsg, 
			msgX + (msgWidth-len(completeMsg)*7)/2, 
			msgY + 50)

		// Draw congratulatory message
		congratsMsg := "Congratulations!"
		ebitenutil.DebugPrintAt(screen, congratsMsg, 
			msgX + (msgWidth-len(congratsMsg)*7)/2, 
			msgY + 80)

		// Draw instruction to continue with animation effect
		continueMsg := "Press ENTER to continue"

		// Make the text pulse by using the frame count
		pulseOffset := 0
		if g.frameCount % 60 < 30 {
			pulseOffset = 1
		}

		ebitenutil.DebugPrintAt(screen, continueMsg, 
			msgX + (msgWidth-len(continueMsg)*7)/2, 
			msgY + 110 + pulseOffset)
	}
}

// drawDebugInfo draws debug information
func (g *Game) drawDebugInfo(screen *ebiten.Image) {
	// Draw FPS and other debug info
	fps := fmt.Sprintf("%.2f", ebiten.ActualFPS())
	playerX := fmt.Sprintf("%.2f", g.player.X)
	playerY := fmt.Sprintf("%.2f", g.player.Y)
	velX := fmt.Sprintf("%.2f", g.player.VelX)
	velY := fmt.Sprintf("%.2f", g.player.VelY)
	onGround := fmt.Sprintf("%t", g.player.OnGround)
	isDead := fmt.Sprintf("%t", g.player.Dead)
	cameraX := fmt.Sprintf("%.2f", g.world.GetCamera().X)
	cameraY := fmt.Sprintf("%.2f", g.world.GetCamera().Y)
	lives := fmt.Sprintf("%d", g.player.Lives)
	enemyCount := fmt.Sprintf("%d", len(g.enemies))
	levelComplete := fmt.Sprintf("%t", g.levelComplete)

	debugText := fmt.Sprintf("FPS: %s\nPlayer Pos: (%s, %s)\nVelocity: (%s, %s)\nOn Ground: %s\nDead: %s\nLives: %s\nEnemies: %s\nLevel Complete: %s\nCamera Pos: (%s, %s)",
		fps, playerX, playerY, velX, velY, onGround, isDead, lives, enemyCount, levelComplete, cameraX, cameraY)

	ebitenutil.DebugPrint(screen, debugText)
}

// Layout implements ebiten.Game's Layout
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.width, g.height
}
