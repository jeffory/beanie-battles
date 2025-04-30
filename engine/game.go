package engine

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

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
	rockets        []*Rocket
	grenades       []*Grenade
	enemies        []*Enemy
	flag           *Flag
	bloodParticles []*BloodParticle
	gravity        float64
	debug          bool
	frameCount     int
	enemySpawnTimer int
	levelComplete  bool
	mouseX         int
	mouseY         int
	prevKeys       map[ebiten.Key]bool
	prevMouseButtons map[ebiten.MouseButton]bool
	isZooming      bool
	zoomFactor     float64
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
		rockets:        make([]*Rocket, 0),
		grenades:       make([]*Grenade, 0),
		enemies:        make([]*Enemy, 0),
		bloodParticles: make([]*BloodParticle, 0),
		frameCount:     0,
		enemySpawnTimer: 300, // Spawn first enemy after 5 seconds
		levelComplete:  false,
		mouseX:         0,
		mouseY:         0,
		prevKeys:       make(map[ebiten.Key]bool),
		prevMouseButtons: make(map[ebiten.MouseButton]bool),
		isZooming:      false,
		zoomFactor:     1.0, // Default zoom factor (no zoom)
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

	// Calculate bullet position based on player position and gun angle
	gunLength := g.player.Width * 0.6 // Length of the gun
	gunOffsetY := g.player.Height * 0.4 // Y position of the gun relative to player

	// Calculate gun position
	gunX := g.player.X + g.player.Width/2
	gunY := g.player.Y + gunOffsetY

	// Calculate bullet starting position at the end of the gun
	bulletX := gunX + math.Cos(g.player.GunAngle) * gunLength
	bulletY := gunY + math.Sin(g.player.GunAngle) * gunLength

	// Calculate bullet velocity based on gun angle
	bulletSpeed := 12.0 // Bullet speed
	bulletVelX := math.Cos(g.player.GunAngle) * bulletSpeed
	bulletVelY := math.Sin(g.player.GunAngle) * bulletSpeed

	// Create new bullet with calculated trajectory
	bullet := NewBullet(bulletX, bulletY, bulletVelX, bulletVelY)

	// Add bullet to game and world
	g.bullets = append(g.bullets, bullet)
	g.world.AddEntity(bullet)

	// Update player state
	g.player.CurrentClip--
	g.player.LastFired = g.frameCount
}

// FireRocket creates a new rocket and adds it to the game
func (g *Game) FireRocket() {
	// Check if player is reloading rocket launcher
	if g.player.IsRocketReloading {
		return
	}

	// Check if rocket clip is empty
	if g.player.CurrentRocket <= 0 {
		// Auto-reload if clip is empty and player has rockets
		if g.player.RocketCount > 0 {
			g.player.StartRocketReload()
		}
		return
	}

	// Check if enough time has passed since last rocket shot
	if g.frameCount - g.player.LastRocketFired < g.player.RocketFireRate {
		return
	}

	// Calculate rocket position based on player position and gun angle
	gunLength := g.player.Width * 0.6 // Length of the gun
	gunOffsetY := g.player.Height * 0.4 // Y position of the gun relative to player

	// Calculate gun position
	gunX := g.player.X + g.player.Width/2
	gunY := g.player.Y + gunOffsetY

	// Calculate rocket starting position at the end of the gun
	rocketX := gunX + math.Cos(g.player.GunAngle) * gunLength
	rocketY := gunY + math.Sin(g.player.GunAngle) * gunLength

	// Calculate rocket velocity based on gun angle
	rocketSpeed := 8.0 // Rocket speed (slower than bullets)
	rocketVelX := math.Cos(g.player.GunAngle) * rocketSpeed
	rocketVelY := math.Sin(g.player.GunAngle) * rocketSpeed

	// Create new rocket with calculated trajectory
	rocket := NewRocket(rocketX, rocketY, rocketVelX, rocketVelY)

	// Add rocket to game and world
	g.rockets = append(g.rockets, rocket)
	g.world.AddEntity(rocket)

	// Update player state
	g.player.CurrentRocket--
	g.player.LastRocketFired = g.frameCount
}

// FireGrenade creates a new grenade and adds it to the game
func (g *Game) FireGrenade() {
	// Check if player has any grenades left
	if g.player.GrenadeCount <= 0 {
		return
	}

	// Check if enough time has passed since last grenade throw
	if g.frameCount - g.player.LastGrenadeFired < g.player.GrenadeFireRate {
		return
	}

	// Calculate grenade position based on player position and gun angle
	gunLength := g.player.Width * 0.6 // Length of the gun
	gunOffsetY := g.player.Height * 0.4 // Y position of the gun relative to player

	// Calculate gun position
	gunX := g.player.X + g.player.Width/2
	gunY := g.player.Y + gunOffsetY

	// Calculate grenade starting position at the end of the gun
	grenadeX := gunX + math.Cos(g.player.GunAngle) * gunLength
	grenadeY := gunY + math.Sin(g.player.GunAngle) * gunLength

	// Calculate grenade velocity based on gun angle
	grenadeSpeed := 10.0 // Grenade initial speed
	grenadeVelX := math.Cos(g.player.GunAngle) * grenadeSpeed
	grenadeVelY := math.Sin(g.player.GunAngle) * grenadeSpeed

	// Create new grenade with calculated trajectory
	grenade := NewGrenade(grenadeX, grenadeY, grenadeVelX, grenadeVelY)

	// Add grenade to game and world
	g.grenades = append(g.grenades, grenade)
	g.world.AddEntity(grenade)

	// Update player state
	g.player.GrenadeCount--
	g.player.LastGrenadeFired = g.frameCount
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

// CreateBloodEffect creates a blood effect at the specified position
func (g *Game) CreateBloodEffect(x, y float64) {
	// Create multiple blood particles with random velocities
	particleCount := 20
	for i := 0; i < particleCount; i++ {
		// Random velocity
		angle := rand.Float64() * 2 * math.Pi
		speed := 1.0 + rand.Float64()*3.0
		velX := math.Cos(angle) * speed
		velY := math.Sin(angle) * speed - 2.0 // Initial upward velocity

		// Random radius
		radius := 2.0 + rand.Float64()*3.0

		// Random lifetime
		lifetime := 30 + rand.Intn(30)

		// Create blood particle
		particle := &BloodParticle{
			X:        x,
			Y:        y,
			VelX:     velX,
			VelY:     velY,
			Radius:   radius,
			Color:    color.RGBA{200, 0, 0, 255}, // Dark red
			Lifetime: lifetime,
			Active:   true,
		}

		// Add particle to game
		g.bloodParticles = append(g.bloodParticles, particle)
	}
}

// cleanupBloodParticles removes inactive blood particles
func (g *Game) cleanupBloodParticles() {
	activeParticles := make([]*BloodParticle, 0)
	for _, particle := range g.bloodParticles {
		if particle.Active {
			activeParticles = append(activeParticles, particle)
		}
	}
	g.bloodParticles = activeParticles
}

// updateBloodParticles updates all blood particles
func (g *Game) updateBloodParticles() {
	for _, particle := range g.bloodParticles {
		// Update position
		particle.X += particle.VelX
		particle.Y += particle.VelY

		// Apply gravity
		particle.VelY += g.gravity * 0.2

		// Update lifetime
		particle.Lifetime--
		if particle.Lifetime <= 0 {
			particle.Active = false
		}

		// Fade out color as lifetime decreases
		if particle.Lifetime < 15 {
			alpha := uint8(float64(particle.Lifetime) / 15.0 * 255.0)
			particle.Color.A = alpha
		}
	}
}

// drawBloodParticle draws a blood particle
func drawBloodParticle(screen *ebiten.Image, x, y float64, particle *BloodParticle) {
	// Draw blood particle as a circle
	drawCircle(screen, x, y, particle.Radius, particle.Color)
}

// Update updates the game state
func (g *Game) Update() error {
	// Increment frame counter
	g.frameCount++

	// Update mouse position
	g.mouseX, g.mouseY = ebiten.CursorPosition()

	// Update player's gun angle based on mouse position
	g.updatePlayerGunAngle()

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

	// Update blood particles
	g.updateBloodParticles()

	// Clean up inactive bullets, rockets, grenades, enemies, and blood particles
	g.cleanupBullets()
	g.cleanupRockets()
	g.cleanupGrenades()
	g.cleanupEnemies()
	g.cleanupBloodParticles()

	// Check for bullet-enemy collisions
	g.checkBulletEnemyCollisions()

	// Check for rocket-enemy collisions
	g.checkRocketEnemyCollisions()

	// Check for grenade-enemy collisions
	g.checkGrenadeEnemyCollisions()

	// Check for player-enemy collisions
	g.checkPlayerEnemyCollisions()

	// Check for player-flag collision
	g.checkPlayerFlagCollision()

	// Update input state for the next frame
	g.updateInputState()

	return nil
}

// updatePlayerGunAngle calculates the angle between the player and the mouse cursor
func (g *Game) updatePlayerGunAngle() {
	// Get the camera and zoom factor
	camera := g.world.GetCamera()
	zoomFactor := camera.Zoom

	// Calculate the center of the screen
	centerX := float64(camera.Width) / 2
	centerY := float64(camera.Height) / 2

	// Calculate the position of the player in screen coordinates
	playerWorldX := g.player.X
	playerWorldY := g.player.Y + g.player.Height/2 // Use the middle of the player for aiming

	// Convert player world coordinates to screen coordinates with zoom
	playerScreenX := centerX + (playerWorldX - camera.X - centerX) * zoomFactor
	playerScreenY := centerY + (playerWorldY - centerY) * zoomFactor

	// Calculate the angle between the player and the mouse cursor
	dx := float64(g.mouseX) - playerScreenX
	dy := float64(g.mouseY) - playerScreenY
	angle := math.Atan2(dy, dx)

	// Update the player's gun angle
	g.player.GunAngle = angle

	// Update the player's facing direction based on the mouse position
	if dx > 0 {
		g.player.FacingRight = true
	} else {
		g.player.FacingRight = false
	}
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

				// Create blood effect
				g.CreateBloodEffect(enemy.X+enemy.Width/2, enemy.Y+enemy.Height/2)

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

			// Check if player is landing on top of the enemy
			playerBottom := g.player.Y + g.player.Height
			enemyTop := enemy.Y
			playerFalling := g.player.VelY > 0

			// If player's bottom is near the enemy's top and player is falling, kill the enemy
			if playerBottom < enemyTop+10 && playerFalling {
				// Kill the enemy
				enemy.TakeDamage(enemy.MaxHealth) // Ensure enemy dies

				// Create blood effect
				g.CreateBloodEffect(enemy.X+enemy.Width/2, enemy.Y+enemy.Height/2)

				// Make player bounce
				g.player.VelY = -8.0 // Bounce upward
			} else {
				// Enemy hit player
				g.player.TakeDamage(10) // Enemy deals 10 damage
			}
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
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.MoveLeft()
	} else if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.MoveRight()
	} else {
		g.player.StopHorizontal()
	}

	// Jump
	if ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.player.Jump()
	}

	// Weapon switching (1 key for machine gun, 2 key for rocket launcher, 3 key for grenades)
	if ebiten.IsKeyPressed(ebiten.Key1) && !g.isKeyPressedPreviously(ebiten.Key1) {
		g.player.CurrentWeapon = 0 // Switch to machine gun
	} else if ebiten.IsKeyPressed(ebiten.Key2) && !g.isKeyPressedPreviously(ebiten.Key2) {
		g.player.CurrentWeapon = 1 // Switch to rocket launcher
	} else if ebiten.IsKeyPressed(ebiten.Key3) && !g.isKeyPressedPreviously(ebiten.Key3) {
		g.player.CurrentWeapon = 2 // Switch to grenades
	}

	// Shoot based on current weapon using primary mouse button (left click)
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if g.player.CurrentWeapon == 0 {
			// Fire machine gun
			g.FireBullet()
		} else if g.player.CurrentWeapon == 1 {
			// Fire rocket launcher
			g.FireRocket()
		} else if g.player.CurrentWeapon == 2 {
			// Throw grenade
			g.FireGrenade()
		}
	}

	// Toggle zoom with secondary mouse button (right click)
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		if !g.isMouseButtonPressedPreviously(ebiten.MouseButtonRight) {
			// Toggle zoom
			g.isZooming = !g.isZooming

			// Set zoom factor
			if g.isZooming {
				g.zoomFactor = 1.5 // Zoomed in
			} else {
				g.zoomFactor = 1.0 // Normal view
			}

			// Update camera zoom
			g.world.GetCamera().Zoom = g.zoomFactor
		}
	}

	// Toggle debug mode
	if ebiten.IsKeyPressed(ebiten.KeyD) && !g.isKeyPressedPreviously(ebiten.KeyD) {
		g.debug = !g.debug
	}

	// Reload (R key)
	if ebiten.IsKeyPressed(ebiten.KeyR) && !g.isKeyPressedPreviously(ebiten.KeyR) {
		if g.player.CurrentWeapon == 0 {
			// Reload machine gun
			g.player.StartReload()
		} else if g.player.CurrentWeapon == 1 {
			// Reload rocket launcher
			g.player.StartRocketReload()
		}
		// Grenades don't need reloading
	}

	// Testing keys

	// Refill ammo (F key)
	if ebiten.IsKeyPressed(ebiten.KeyF) && !g.isKeyPressedPreviously(ebiten.KeyF) {
		g.player.BulletCount = 200
		g.player.CurrentClip = g.player.ClipSize
		g.player.RocketCount = 10
		g.player.CurrentRocket = g.player.RocketClipSize
		g.player.GrenadeCount = 5
	}

	// Take damage (T key)
	if ebiten.IsKeyPressed(ebiten.KeyT) && !g.isKeyPressedPreviously(ebiten.KeyT) {
		g.player.TakeDamage(20)
	}
}

// cleanupRockets removes inactive rockets from the game
func (g *Game) cleanupRockets() {
	activeRockets := make([]*Rocket, 0)

	for _, rocket := range g.rockets {
		if rocket.Active {
			activeRockets = append(activeRockets, rocket)
		}
	}

	g.rockets = activeRockets
}

// cleanupGrenades removes inactive grenades from the game
func (g *Game) cleanupGrenades() {
	activeGrenades := make([]*Grenade, 0)

	for _, grenade := range g.grenades {
		if grenade.Active {
			activeGrenades = append(activeGrenades, grenade)
		}
	}

	g.grenades = activeGrenades
}

// checkRocketEnemyCollisions checks for collisions between rockets and enemies
func (g *Game) checkRocketEnemyCollisions() {
	for _, rocket := range g.rockets {
		if !rocket.Active || rocket.Exploded {
			continue
		}

		for _, enemy := range g.enemies {
			if !enemy.Active || enemy.Dead {
				continue
			}

			// Simple AABB collision detection
			if rocket.X < enemy.X+enemy.Width &&
				rocket.X+rocket.Width > enemy.X &&
				rocket.Y < enemy.Y+enemy.Height &&
				rocket.Y+rocket.Height > enemy.Y {

				// Rocket hit enemy
				enemy.TakeDamage(rocket.Damage)

				// Create blood effect
				g.CreateBloodEffect(enemy.X+enemy.Width/2, enemy.Y+enemy.Height/2)

				rocket.Explode()
				break
			}
		}
	}
}

// checkGrenadeEnemyCollisions checks for collisions between grenades and enemies
func (g *Game) checkGrenadeEnemyCollisions() {
	for _, grenade := range g.grenades {
		if !grenade.Active || grenade.Exploded {
			continue
		}

		for _, enemy := range g.enemies {
			if !enemy.Active || enemy.Dead {
				continue
			}

			// Simple AABB collision detection
			if grenade.X < enemy.X+enemy.Width &&
				grenade.X+grenade.Width > enemy.X &&
				grenade.Y < enemy.Y+enemy.Height &&
				grenade.Y+grenade.Height > enemy.Y {

				// Grenade hit enemy directly - explode
				enemy.TakeDamage(grenade.Damage / 2) // Direct hit does half damage

				// Create blood effect
				g.CreateBloodEffect(enemy.X+enemy.Width/2, enemy.Y+enemy.Height/2)

				grenade.Explode()
				break
			}
		}
	}
}

// isKeyPressedPreviously is a helper to detect key press events
func (g *Game) isKeyPressedPreviously(key ebiten.Key) bool {
	// Check if the key was pressed in the previous frame
	wasPressed, exists := g.prevKeys[key]

	// If the key exists in the map, return its previous state
	// Otherwise, return false (key was not pressed)
	return exists && wasPressed
}

// isMouseButtonPressedPreviously is a helper to detect mouse button press events
func (g *Game) isMouseButtonPressedPreviously(button ebiten.MouseButton) bool {
	// Check if the button was pressed in the previous frame
	wasPressed, exists := g.prevMouseButtons[button]

	// If the button exists in the map, return its previous state
	// Otherwise, return false (button was not pressed)
	return exists && wasPressed
}

// updateInputState updates the state of all keys and mouse buttons for the next frame
func (g *Game) updateInputState() {
	// Update key states for weapon switching keys
	g.prevKeys[ebiten.Key1] = ebiten.IsKeyPressed(ebiten.Key1)
	g.prevKeys[ebiten.Key2] = ebiten.IsKeyPressed(ebiten.Key2)
	g.prevKeys[ebiten.Key3] = ebiten.IsKeyPressed(ebiten.Key3)

	// Update key states for other keys
	g.prevKeys[ebiten.KeyEnter] = ebiten.IsKeyPressed(ebiten.KeyEnter)
	g.prevKeys[ebiten.KeyD] = ebiten.IsKeyPressed(ebiten.KeyD)
	g.prevKeys[ebiten.KeyR] = ebiten.IsKeyPressed(ebiten.KeyR)
	g.prevKeys[ebiten.KeyF] = ebiten.IsKeyPressed(ebiten.KeyF)
	g.prevKeys[ebiten.KeyT] = ebiten.IsKeyPressed(ebiten.KeyT)

	// Update mouse button states
	g.prevMouseButtons[ebiten.MouseButtonLeft] = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	g.prevMouseButtons[ebiten.MouseButtonRight] = ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
}

// Draw draws the game
func (g *Game) Draw(screen *ebiten.Image) {
	// Hide the actual mouse cursor
	ebiten.SetCursorMode(ebiten.CursorModeHidden)

	// Draw a gradient background
	g.drawBackground(screen)

	// Draw all entities
	g.world.Draw(screen)

	// Draw blood particles
	g.drawBloodParticles(screen)

	// Draw UI elements
	g.drawUI(screen)

	// Draw crosshair at mouse position
	g.drawCrosshair(screen)

	// Draw debug info
	if g.debug {
		g.drawDebugInfo(screen)
	}
}

// drawBloodParticles draws all blood particles
func (g *Game) drawBloodParticles(screen *ebiten.Image) {
	camera := g.world.GetCamera()
	for _, particle := range g.bloodParticles {
		if !particle.Active {
			continue
		}

		// Calculate screen position with camera offset
		x := particle.X - camera.X
		y := particle.Y

		// Skip particles that are off-screen
		if x+particle.Radius*2 < 0 || x-particle.Radius*2 > float64(camera.Width) {
			continue
		}

		// Draw the particle
		drawBloodParticle(screen, x, y, particle)
	}
}

// drawCrosshair draws a crosshair at the mouse position
func (g *Game) drawCrosshair(screen *ebiten.Image) {
	// Crosshair properties
	crosshairSize := 10.0
	crosshairThickness := 2.0
	crosshairColor := color.RGBA{255, 255, 255, 255} // White
	crosshairOutlineColor := color.RGBA{0, 0, 0, 255} // Black

	// Get mouse position
	mouseX := float64(g.mouseX)
	mouseY := float64(g.mouseY)

	// Draw crosshair outline (black)
	// Horizontal line
	ebitenutil.DrawRect(screen, mouseX - crosshairSize - 1, mouseY - crosshairThickness/2 - 1, crosshairSize*2 + 2, crosshairThickness + 2, crosshairOutlineColor)
	// Vertical line
	ebitenutil.DrawRect(screen, mouseX - crosshairThickness/2 - 1, mouseY - crosshairSize - 1, crosshairThickness + 2, crosshairSize*2 + 2, crosshairOutlineColor)

	// Draw crosshair (white)
	// Horizontal line
	ebitenutil.DrawRect(screen, mouseX - crosshairSize, mouseY - crosshairThickness/2, crosshairSize*2, crosshairThickness, crosshairColor)
	// Vertical line
	ebitenutil.DrawRect(screen, mouseX - crosshairThickness/2, mouseY - crosshairSize, crosshairThickness, crosshairSize*2, crosshairColor)

	// Draw a small gap in the center
	gapSize := 2.0
	ebitenutil.DrawRect(screen, mouseX - gapSize, mouseY - gapSize, gapSize*2, gapSize*2, crosshairOutlineColor)
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
	// Draw weapon info in top right corner with larger text and better styling
	var ammoText string
	var weaponName string
	var iconColor color.RGBA

	// Set text and icon based on current weapon
	if g.player.CurrentWeapon == 0 {
		// Machine gun
		weaponName = "MACHINE GUN"
		ammoText = fmt.Sprintf("%d / %d", g.player.CurrentClip, g.player.BulletCount)
		iconColor = color.RGBA{255, 255, 0, 255} // Yellow

		// Draw reload indicator if reloading
		if g.player.IsReloading {
			reloadProgress := float64(g.player.ReloadTime - g.player.ReloadTimer) / float64(g.player.ReloadTime) * 100
			ammoText = fmt.Sprintf("RELOADING... %.0f%%", reloadProgress)
		}
	} else if g.player.CurrentWeapon == 1 {
		// Rocket launcher
		weaponName = "ROCKET LAUNCHER"
		ammoText = fmt.Sprintf("%d / %d", g.player.CurrentRocket, g.player.RocketCount)
		iconColor = color.RGBA{255, 100, 0, 255} // Orange

		// Draw reload indicator if reloading
		if g.player.IsRocketReloading {
			reloadProgress := float64(g.player.RocketReloadTime - g.player.RocketReloadTimer) / float64(g.player.RocketReloadTime) * 100
			ammoText = fmt.Sprintf("RELOADING... %.0f%%", reloadProgress)
		}
	} else {
		// Grenades
		weaponName = "GRENADES"
		ammoText = fmt.Sprintf("%d", g.player.GrenadeCount)
		iconColor = color.RGBA{0, 150, 0, 255} // Green
	}

	// Combine weapon name and ammo text
	fullText := fmt.Sprintf("%s: %s", weaponName, ammoText)

	// Draw a stylized background for the ammo text
	textWidth := len(fullText) * 7 // Approximate width based on character count
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

	// Draw weapon icon
	iconX := ammoBoxX + 5
	iconY := ammoBoxY + ammoBoxHeight/2 - 3
	iconWidth := 10.0
	iconHeight := 6.0

	if g.player.CurrentWeapon == 0 {
		// Draw bullet icon for machine gun
		ebitenutil.DrawRect(screen, iconX, iconY, iconWidth, iconHeight, iconColor)
	} else if g.player.CurrentWeapon == 1 {
		// Draw rocket icon for rocket launcher
		ebitenutil.DrawRect(screen, iconX, iconY, iconWidth, iconHeight, iconColor)
		// Add fins to the rocket icon
		ebitenutil.DrawRect(screen, iconX, iconY - 2, iconWidth/3, iconHeight/2, iconColor)
		ebitenutil.DrawRect(screen, iconX, iconY + iconHeight, iconWidth/3, iconHeight/2, iconColor)
	} else {
		// Draw grenade icon (circle with pin)
		// Draw grenade body (circle)
		drawCircle(screen, iconX + iconWidth/2, iconY + iconHeight/2, iconWidth/2, iconColor)
		// Draw pin
		pinColor := color.RGBA{200, 200, 200, 255} // Silver
		ebitenutil.DrawRect(screen, iconX + iconWidth/2 - 1, iconY - 3, 2, 3, pinColor)
	}

	// Draw text with offset to center vertically
	textY := int(ammoBoxY) + int(ammoBoxHeight)/2 - 3
	ebitenutil.DebugPrintAt(screen, fullText, int(ammoBoxX) + 20, textY)

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
	// Update the game's dimensions to match the window size
	g.width = outsideWidth
	g.height = outsideHeight

	// Update the camera's dimensions to match the new window size
	camera := g.world.GetCamera()
	camera.Width = outsideWidth
	camera.Height = outsideHeight

	return g.width, g.height
}
