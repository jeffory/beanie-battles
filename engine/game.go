package engine

import (
    "fmt"
    "image/color"
    "math"
    "math/rand"
    "path/filepath"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Game represents the main game state and implements ebiten.Game interface
type Game struct {
    width            int
    height           int
    world            *World
    player           *Player
    platforms        []*Platform
    bullets          []*Bullet
    rockets          []*Rocket
    grenades         []*Grenade
    enemies          []*Enemy
    flag             *Flag
    bloodParticles   []*BloodParticle
    gravity          float64
    debug            bool
    frameCount       int
    enemySpawnTimer  int
    levelComplete    bool
    mouseX           int
    mouseY           int
    prevKeys         map[ebiten.Key]bool
    prevMouseButtons map[ebiten.MouseButton]bool
    isZooming        bool
    zoomFactor       float64

    // Level system
    currentLevel      *Level
    currentLevelIndex int
    levels            []string
    levelsDir         string

    // Editor mode
    editorMode                bool
    editorTool                string // "platform", "enemy", "flag", "player"
    editorPlatformWidth       float64
    editorPlatformHeight      float64
    editorPlatformType        string // "normal", "small", "moving", "spike"
    editorMovingPlatformMoveX float64
    editorMovingPlatformMoveY float64
    editorMovingPlatformSpeed float64
    editorEnemyType           string
    selectedEntity            interface{}
    dragging                  bool
    dragOffsetX               float64
    dragOffsetY               float64
}

// NewGame creates a new Game instance
func NewGame(width, height int) (*Game, error) {
    // Create a new game instance
    g := &Game{
        width:            width,
        height:           height,
        gravity:          0.5,
        debug:            false,
        bullets:          make([]*Bullet, 0),
        rockets:          make([]*Rocket, 0),
        grenades:         make([]*Grenade, 0),
        enemies:          make([]*Enemy, 0),
        bloodParticles:   make([]*BloodParticle, 0),
        frameCount:       0,
        enemySpawnTimer:  300, // Spawn first enemy after 5 seconds
        levelComplete:    false,
        mouseX:           0,
        mouseY:           0,
        prevKeys:         make(map[ebiten.Key]bool),
        prevMouseButtons: make(map[ebiten.MouseButton]bool),
        isZooming:        false,
        zoomFactor:       1.0, // Default zoom factor (no zoom)

        // Initialize level system
        currentLevelIndex: 0,
        levels:            make([]string, 0),
        levelsDir:         "levels",

        // Initialize editor mode
        editorMode:                false,
        editorTool:                "platform",
        editorPlatformWidth:       200,
        editorPlatformHeight:      20,
        editorPlatformType:        "normal",
        editorMovingPlatformMoveX: 100,
        editorMovingPlatformMoveY: 0,
        editorMovingPlatformSpeed: 1.0,
        editorEnemyType:           "kidney",
        dragging:                  false,
    }

    // Initialize the world with camera
    g.world = NewWorld(g.gravity, width, height)

    // Create player
    g.player = NewPlayer(100, 100)
    g.world.AddEntity(g.player)

    // Load levels or create default levels if none exist
    levels, err := GetLevelList(g.levelsDir)
    if err != nil {
        fmt.Println("Error loading levels:", err)
    }

    if len(levels) > 0 {
        g.levels = levels
        // Load the first level
        if err := g.LoadLevel(0); err != nil {
            fmt.Println("Error loading level:", err)
            // If level loading fails, create default level
            g.createDefaultLevel()
        }
    } else {
        // No levels found, create and save default levels
        fmt.Println("No levels found, creating default levels...")
        if err := CreateDefaultLevels(width, height); err != nil {
            fmt.Println("Error creating default levels:", err)
            // If creating default levels fails, create a single default level
            g.createDefaultLevel()

            // Save the default level
            defaultLevel := CreateDefaultLevel(width, height)
            err := defaultLevel.SaveToFile(filepath.Join(g.levelsDir, "level1.json"))
            if err != nil {
                fmt.Println("Error saving default level:", err)
            }

            // Add default level to levels list
            g.levels = append(g.levels, "level1")
        } else {
            // Load the newly created levels
            levels, _ := GetLevelList(g.levelsDir)
            g.levels = levels
            // Load the first level
            g.LoadLevel(0)
        }
    }

    return g, nil
}

// LoadLevel loads a level from a file
func (g *Game) LoadLevel(index int) error {
    if index < 0 || index >= len(g.levels) {
        return fmt.Errorf("invalid level index: %d", index)
    }

    // Load level from file
    levelName := g.levels[index]
    levelPath := filepath.Join(g.levelsDir, levelName+".json")
    level, err := LoadLevelFromFile(levelPath)
    if err != nil {
        return fmt.Errorf("failed to load level: %v", err)
    }

    // Set current level
    g.currentLevel = level
    g.currentLevelIndex = index

    // Reset level complete flag
    g.levelComplete = false

    // Clear existing platforms, enemies, and flag
    g.clearLevel()

    // Create platforms from level data
    g.createPlatformsFromLevel()

    // Set player position
    g.player.X = g.currentLevel.PlayerStart.X
    g.player.Y = g.currentLevel.PlayerStart.Y

    // Create flag
    flagX := g.currentLevel.FlagPosition.X
    flagY := g.currentLevel.FlagPosition.Y
    g.flag = NewFlag(flagX, flagY)
    g.world.AddEntity(g.flag)

    return nil
}

// LoadNextLevel loads the next level
func (g *Game) LoadNextLevel() error {
    nextIndex := g.currentLevelIndex + 1
    if nextIndex >= len(g.levels) {
        // Wrap around to the first level
        nextIndex = 0
    }
    return g.LoadLevel(nextIndex)
}

// SaveCurrentLevel saves the current level to a file
func (g *Game) SaveCurrentLevel() error {
    if g.currentLevel == nil {
        return fmt.Errorf("no current level to save")
    }

    // Update level data from game state
    g.updateLevelFromGameState()

    // Save level to file
    levelName := g.levels[g.currentLevelIndex]
    levelPath := filepath.Join(g.levelsDir, levelName+".json")
    return g.currentLevel.SaveToFile(levelPath)
}

// updateLevelFromGameState updates the current level data from the game state
func (g *Game) updateLevelFromGameState() {
    if g.currentLevel == nil {
        return
    }

    // Clear existing platforms and enemy spawns
    g.currentLevel.Platforms = make([]PlatformData, 0)
    g.currentLevel.EnemySpawns = make([]EnemySpawnData, 0)

    // Add platforms
    for _, platform := range g.platforms {
        g.currentLevel.AddPlatform(platform.X, platform.Y, platform.Width, platform.Height)
    }

    // Set flag position
    if g.flag != nil {
        g.currentLevel.SetFlagPosition(g.flag.X, g.flag.Y)
    }

    // Set player start position
    g.currentLevel.SetPlayerStart(g.player.X, g.player.Y)
}

// createDefaultLevel creates the default level
func (g *Game) createDefaultLevel() {
    // Clear existing platforms, enemies, and flag
    g.clearLevel()

    // Create default level
    g.currentLevel = CreateDefaultLevel(g.width, g.height)

    // Create platforms from level data
    g.createPlatformsFromLevel()

    // Set player position
    g.player.X = g.currentLevel.PlayerStart.X
    g.player.Y = g.currentLevel.PlayerStart.Y

    // Create flag
    flagX := g.currentLevel.FlagPosition.X
    flagY := g.currentLevel.FlagPosition.Y
    g.flag = NewFlag(flagX, flagY)
    g.world.AddEntity(g.flag)
}

// clearLevel removes all platforms, enemies, and the flag from the game
func (g *Game) clearLevel() {
    // Remove platforms from world
    for _, platform := range g.platforms {
        g.world.RemoveEntity(platform)
    }
    g.platforms = make([]*Platform, 0)

    // Remove enemies from world
    for _, enemy := range g.enemies {
        g.world.RemoveEntity(enemy)
    }
    g.enemies = make([]*Enemy, 0)

    // Remove flag from world
    if g.flag != nil {
        g.world.RemoveEntity(g.flag)
        g.flag = nil
    }
}

// createPlatformsFromLevel creates platforms from the current level data
func (g *Game) createPlatformsFromLevel() {
    if g.currentLevel == nil {
        return
    }

    // Create platforms from level data
    for _, platformData := range g.currentLevel.Platforms {
        // Create different platform types based on the Type field
        switch platformData.Type {
        case "small":
            // Create small platform
            platform := NewSmallPlatform(
                platformData.X,
                platformData.Y,
            )
            g.platforms = append(g.platforms, platform)
            g.world.AddEntity(platform)

        case "moving":
            // Create moving platform
            movingPlatform := NewMovingPlatform(
                platformData.X,
                platformData.Y,
                platformData.Width,
                platformData.Height,
                platformData.MoveX,
                platformData.MoveY,
                platformData.Speed,
            )
            g.world.AddEntity(movingPlatform)

        case "spike":
            // Create spike pit
            spikePit := NewSpikePit(
                platformData.X,
                platformData.Y,
                platformData.Width,
                platformData.Height,
            )
            g.world.AddEntity(spikePit)

        default: // "normal" or any other type
            // Create normal platform
            platform := NewPlatform(
                platformData.X,
                platformData.Y,
                platformData.Width,
                platformData.Height,
            )
            g.platforms = append(g.platforms, platform)
            g.world.AddEntity(platform)
        }
    }
}

// createPlatforms initializes the game platforms (legacy function, kept for compatibility)
func (g *Game) createPlatforms() {
    // Create default level
    g.createDefaultLevel()
}

// FireEnemyBullet creates a new bullet fired by an enemy and adds it to the game
func (g *Game) FireEnemyBullet(enemy *Enemy) {
    // Skip if enemy is dead or inactive
    if enemy.Dead || !enemy.Active || !enemy.CanShoot {
        return
    }

    // Check if enough time has passed since last shot
    if enemy.ShootTimer < enemy.ShootRate {
        return
    }

    // Calculate angle to player
    playerCenterX := g.player.X + g.player.Width/2
    playerCenterY := g.player.Y + g.player.Height/2
    enemyCenterX := enemy.X + enemy.Width/2
    enemyCenterY := enemy.Y + enemy.Height/2

    dx := playerCenterX - enemyCenterX
    dy := playerCenterY - enemyCenterY
    angle := math.Atan2(dy, dx)

    // Calculate bullet starting position
    bulletX := enemyCenterX + math.Cos(angle)*enemy.Width/2
    bulletY := enemyCenterY + math.Sin(angle)*enemy.Height/2

    // Calculate bullet velocity
    bulletVelX := math.Cos(angle) * enemy.BulletSpeed
    bulletVelY := math.Sin(angle) * enemy.BulletSpeed

    // Create new bullet with calculated trajectory
    bullet := NewBullet(bulletX, bulletY, bulletVelX, bulletVelY)
    bullet.Damage = enemy.BulletDamage

    // Add bullet to game and world
    g.bullets = append(g.bullets, bullet)
    g.world.AddEntity(bullet)

    // Reset enemy's shoot timer
    enemy.ShootTimer = 0
    enemy.LastShot = g.frameCount
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
    if g.frameCount-g.player.LastFired < g.player.FireRate {
        return
    }

    // Calculate bullet position based on player position and gun angle
    gunLength := g.player.Width * 0.6   // Length of the gun
    gunOffsetY := g.player.Height * 0.4 // Y position of the gun relative to player

    // Calculate gun position
    gunX := g.player.X + g.player.Width/2
    gunY := g.player.Y + gunOffsetY

    // Calculate bullet starting position at the end of the gun
    bulletX := gunX + math.Cos(g.player.GunAngle)*gunLength
    bulletY := gunY + math.Sin(g.player.GunAngle)*gunLength

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
    if g.frameCount-g.player.LastRocketFired < g.player.RocketFireRate {
        return
    }

    // Calculate rocket position based on player position and gun angle
    gunLength := g.player.Width * 0.6   // Length of the gun
    gunOffsetY := g.player.Height * 0.4 // Y position of the gun relative to player

    // Calculate gun position
    gunX := g.player.X + g.player.Width/2
    gunY := g.player.Y + gunOffsetY

    // Calculate rocket starting position at the end of the gun
    rocketX := gunX + math.Cos(g.player.GunAngle)*gunLength
    rocketY := gunY + math.Sin(g.player.GunAngle)*gunLength

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
    if g.frameCount-g.player.LastGrenadeFired < g.player.GrenadeFireRate {
        return
    }

    // Calculate grenade position based on player position and gun angle
    gunLength := g.player.Width * 0.6   // Length of the gun
    gunOffsetY := g.player.Height * 0.4 // Y position of the gun relative to player

    // Calculate gun position
    gunX := g.player.X + g.player.Width/2
    gunY := g.player.Y + gunOffsetY

    // Calculate grenade starting position at the end of the gun
    grenadeX := gunX + math.Cos(g.player.GunAngle)*gunLength
    grenadeY := gunY + math.Sin(g.player.GunAngle)*gunLength

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
            {300, 350}, // On platform p1
            {500, 250}, // On platform p2
            {300, 150}, // On platform p3
        }
        spawnPoint := spawnLocations[g.frameCount%len(spawnLocations)]
        spawnX = spawnPoint.x
        spawnY = spawnPoint.y
    case 1: // Second screen
        // Spawn on second screen platforms
        spawnLocations := []struct{ x, y float64 }{
            {float64(g.width) + 200, 350}, // On platform p4
            {float64(g.width) + 500, 250}, // On platform p5
        }
        spawnPoint := spawnLocations[g.frameCount%len(spawnLocations)]
        spawnX = spawnPoint.x
        spawnY = spawnPoint.y
    case 2: // Third screen
        // Spawn on third screen platforms
        spawnLocations := []struct{ x, y float64 }{
            {float64(g.width*2) + 200, 300}, // On platform p6
            {float64(g.width*2) + 500, 200}, // On platform p7
        }
        spawnPoint := spawnLocations[g.frameCount%len(spawnLocations)]
        spawnX = spawnPoint.x
        spawnY = spawnPoint.y
    case 3, 4: // Fourth and fifth screens (near the end)
        // Spawn on fourth screen platforms or ahead of player
        spawnLocations := []struct{ x, y float64 }{
            {float64(g.width*3) + 200, 250},                         // On platform p8
            {float64(g.width*3) + 500, 150},                         // On platform p9
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
    var enemy *Enemy

    // Randomly choose enemy type (20% chance for shooter enemy)
    if rand.Intn(5) == 0 {
        // Create shooter enemy
        enemy = NewShooterEnemy(spawnX, spawnY)
    } else {
        // Create regular enemy
        enemy = NewEnemy(spawnX, spawnY)

        // Create bean enemies: kidney beans (dark red) and navy beans (dark blue)
        if g.frameCount%2 == 0 {
            // Kidney bean (faster)
            enemy.Speed = 3
            enemy.Color = color.RGBA{139, 0, 0, 255} // Dark red for kidney beans
            enemy.Health = 50
            enemy.MaxHealth = 50
            enemy.EnemyType = "kidney"
        } else {
            // Navy bean (tougher)
            enemy.Speed = 2
            enemy.Health = 75
            enemy.MaxHealth = 75
            enemy.Color = color.RGBA{0, 0, 128, 255} // Dark blue for navy beans
            enemy.EnemyType = "navy"
        }
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
        velY := math.Sin(angle)*speed - 2.0 // Initial upward velocity

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

    // Check if any enemies should shoot at the player
    g.checkEnemyShooting()

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
    playerScreenX := centerX + (playerWorldX-camera.X-centerX)*zoomFactor
    playerScreenY := centerY + (playerWorldY-centerY)*zoomFactor

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

// checkEnemyShooting checks if any enemies should shoot at the player
func (g *Game) checkEnemyShooting() {
    // Skip if player is dead or level is complete
    if g.player.Dead || g.levelComplete {
        return
    }

    // Check each enemy
    for _, enemy := range g.enemies {
        // Skip if enemy is not a shooter or is dead
        if !enemy.CanShoot || enemy.Dead || !enemy.Active {
            continue
        }

        // Calculate distance to player
        playerCenterX := g.player.X + g.player.Width/2
        playerCenterY := g.player.Y + g.player.Height/2
        enemyCenterX := enemy.X + enemy.Width/2
        enemyCenterY := enemy.Y + enemy.Height/2

        dx := playerCenterX - enemyCenterX
        dy := playerCenterY - enemyCenterY
        distance := math.Sqrt(dx*dx + dy*dy)

        // Only shoot if player is within range (500 pixels)
        if distance < 500 {
            // Try to fire a bullet
            g.FireEnemyBullet(enemy)
        }
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
        // Press Enter to continue to the next level
        if ebiten.IsKeyPressed(ebiten.KeyEnter) && !g.isKeyPressedPreviously(ebiten.KeyEnter) {
            // Load the next level
            err := g.LoadNextLevel()
            if err != nil {
                fmt.Println("Error loading next level:", err)
                // If loading the next level fails, reset the current level
                g.levelComplete = false
                g.flag.Collected = false

                // Reset player position
                g.player.X = g.currentLevel.PlayerStart.X
                g.player.Y = g.currentLevel.PlayerStart.Y
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
        }
    }

    // Press E to enter/exit editor mode (can be accessed anytime)
    if ebiten.IsKeyPressed(ebiten.KeyE) && !g.isKeyPressedPreviously(ebiten.KeyE) {
        g.editorMode = !g.editorMode
        if g.editorMode {
            fmt.Println("Entered editor mode")
            // Reset level complete flag when entering editor mode
            g.levelComplete = false
        } else {
            fmt.Println("Exited editor mode")
            // Save the current level when exiting editor mode
            g.SaveCurrentLevel()
        }
    }

    // Check if in editor mode
    if g.editorMode {
        // Handle editor input
        g.handleEditorInput()
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

    // Set debug mode to false when D is pressed
    if ebiten.IsKeyPressed(ebiten.KeyD) && !g.isKeyPressedPreviously(ebiten.KeyD) {
        g.debug = false
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

// handleEditorInput handles input for the level editor
func (g *Game) handleEditorInput() {
    // Get mouse position in world coordinates
    mouseWorldX, mouseWorldY := g.getMouseWorldPosition()

    // Tool selection
    if ebiten.IsKeyPressed(ebiten.Key1) && !g.isKeyPressedPreviously(ebiten.Key1) {
        g.editorTool = "platform"
        fmt.Println("Selected platform tool")
    } else if ebiten.IsKeyPressed(ebiten.Key2) && !g.isKeyPressedPreviously(ebiten.Key2) {
        g.editorTool = "enemy"
        fmt.Println("Selected enemy tool")
    } else if ebiten.IsKeyPressed(ebiten.Key3) && !g.isKeyPressedPreviously(ebiten.Key3) {
        g.editorTool = "flag"
        fmt.Println("Selected flag tool")
    } else if ebiten.IsKeyPressed(ebiten.Key4) && !g.isKeyPressedPreviously(ebiten.Key4) {
        g.editorTool = "player"
        fmt.Println("Selected player start tool")
    }

    // Enemy type selection (when enemy tool is selected)
    if g.editorTool == "enemy" {
        if ebiten.IsKeyPressed(ebiten.KeyK) && !g.isKeyPressedPreviously(ebiten.KeyK) {
            g.editorEnemyType = "kidney"
            fmt.Println("Selected kidney bean enemy")
        } else if ebiten.IsKeyPressed(ebiten.KeyN) && !g.isKeyPressedPreviously(ebiten.KeyN) {
            g.editorEnemyType = "navy"
            fmt.Println("Selected navy bean enemy")
        } else if ebiten.IsKeyPressed(ebiten.KeyS) && !g.isKeyPressedPreviously(ebiten.KeyS) {
            g.editorEnemyType = "shooter"
            fmt.Println("Selected shooter enemy")
        }
    }

    // Platform type and size adjustment (when platform tool is selected)
    if g.editorTool == "platform" {
        // Platform type selection
        if ebiten.IsKeyPressed(ebiten.KeyN) && !g.isKeyPressedPreviously(ebiten.KeyN) {
            g.editorPlatformType = "normal"
            fmt.Println("Selected normal platform")
        } else if ebiten.IsKeyPressed(ebiten.KeyS) && !g.isKeyPressedPreviously(ebiten.KeyS) {
            g.editorPlatformType = "small"
            fmt.Println("Selected small platform")
        } else if ebiten.IsKeyPressed(ebiten.KeyM) && !g.isKeyPressedPreviously(ebiten.KeyM) {
            g.editorPlatformType = "moving"
            fmt.Println("Selected moving platform")
        } else if ebiten.IsKeyPressed(ebiten.KeyP) && !g.isKeyPressedPreviously(ebiten.KeyP) {
            g.editorPlatformType = "spike"
            fmt.Println("Selected spike pit")
        }

        // Adjust width (except for small platforms which have fixed size)
        if g.editorPlatformType != "small" {
            if ebiten.IsKeyPressed(ebiten.KeyLeft) {
                g.editorPlatformWidth = math.Max(20, g.editorPlatformWidth-10)
            } else if ebiten.IsKeyPressed(ebiten.KeyRight) {
                g.editorPlatformWidth += 10
            }
        }

        // Adjust height (except for small platforms which have fixed size)
        if g.editorPlatformType != "small" {
            if ebiten.IsKeyPressed(ebiten.KeyDown) {
                g.editorPlatformHeight = math.Max(10, g.editorPlatformHeight-5)
            } else if ebiten.IsKeyPressed(ebiten.KeyUp) {
                g.editorPlatformHeight += 5
            }
        }

        // Moving platform configuration (when platform type is "moving")
        if g.editorPlatformType == "moving" {
            // Adjust horizontal movement distance with H + Left/Right
            if ebiten.IsKeyPressed(ebiten.KeyH) {
                if ebiten.IsKeyPressed(ebiten.KeyLeft) {
                    g.editorMovingPlatformMoveX = math.Max(0, g.editorMovingPlatformMoveX-10)
                } else if ebiten.IsKeyPressed(ebiten.KeyRight) {
                    g.editorMovingPlatformMoveX += 10
                }
            }

            // Adjust vertical movement distance with V + Up/Down
            if ebiten.IsKeyPressed(ebiten.KeyV) {
                if ebiten.IsKeyPressed(ebiten.KeyDown) {
                    g.editorMovingPlatformMoveY = math.Max(0, g.editorMovingPlatformMoveY-10)
                } else if ebiten.IsKeyPressed(ebiten.KeyUp) {
                    g.editorMovingPlatformMoveY += 10
                }
            }

            // Adjust speed with + and -
            if ebiten.IsKeyPressed(ebiten.KeyMinus) {
                g.editorMovingPlatformSpeed = math.Max(0.1, g.editorMovingPlatformSpeed-0.1)
            } else if ebiten.IsKeyPressed(ebiten.KeyEqual) { // + key
                g.editorMovingPlatformSpeed += 0.1
            }
        }
    }

    // Check for left mouse button press
    if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && !g.isMouseButtonPressedPreviously(ebiten.MouseButtonLeft) {
        // Place or select entity based on current tool
        switch g.editorTool {
        case "platform":
            // Create platform based on selected type
            switch g.editorPlatformType {
            case "small":
                // Create small platform
                platform := NewSmallPlatform(
                    mouseWorldX-25, // Small platforms have fixed width of 50
                    mouseWorldY-5,  // Small platforms have fixed height of 10
                )
                g.platforms = append(g.platforms, platform)
                g.world.AddEntity(platform)
                // Add to level data
                g.currentLevel.AddPlatformWithType(
                    mouseWorldX-25,
                    mouseWorldY-5,
                    50, // Fixed width
                    10, // Fixed height
                    "small",
                    0, 0, 0, // No movement parameters
                )
                fmt.Printf("Added small platform at (%.1f, %.1f)\n", mouseWorldX, mouseWorldY)

            case "moving":
                // Create moving platform
                movingPlatform := NewMovingPlatform(
                    mouseWorldX-g.editorPlatformWidth/2,
                    mouseWorldY-g.editorPlatformHeight/2,
                    g.editorPlatformWidth,
                    g.editorPlatformHeight,
                    g.editorMovingPlatformMoveX,
                    g.editorMovingPlatformMoveY,
                    g.editorMovingPlatformSpeed,
                )
                g.world.AddEntity(movingPlatform)
                // Add to level data
                g.currentLevel.AddPlatformWithType(
                    mouseWorldX-g.editorPlatformWidth/2,
                    mouseWorldY-g.editorPlatformHeight/2,
                    g.editorPlatformWidth,
                    g.editorPlatformHeight,
                    "moving",
                    g.editorMovingPlatformMoveX,
                    g.editorMovingPlatformMoveY,
                    g.editorMovingPlatformSpeed,
                )
                fmt.Printf("Added moving platform at (%.1f, %.1f) with movement (%.1f, %.1f) and speed %.1f\n",
                    mouseWorldX, mouseWorldY, g.editorMovingPlatformMoveX, g.editorMovingPlatformMoveY, g.editorMovingPlatformSpeed)

            case "spike":
                // Create spike pit
                spikePit := NewSpikePit(
                    mouseWorldX-g.editorPlatformWidth/2,
                    mouseWorldY-g.editorPlatformHeight/2,
                    g.editorPlatformWidth,
                    g.editorPlatformHeight,
                )
                g.world.AddEntity(spikePit)
                // Add to level data
                g.currentLevel.AddPlatformWithType(
                    mouseWorldX-g.editorPlatformWidth/2,
                    mouseWorldY-g.editorPlatformHeight/2,
                    g.editorPlatformWidth,
                    g.editorPlatformHeight,
                    "spike",
                    0, 0, 0, // No movement parameters
                )
                fmt.Printf("Added spike pit at (%.1f, %.1f)\n", mouseWorldX, mouseWorldY)

            default: // "normal" or any other type
                // Create normal platform
                platform := NewPlatform(
                    mouseWorldX-g.editorPlatformWidth/2,
                    mouseWorldY-g.editorPlatformHeight/2,
                    g.editorPlatformWidth,
                    g.editorPlatformHeight,
                )
                g.platforms = append(g.platforms, platform)
                g.world.AddEntity(platform)
                // Add to level data
                g.currentLevel.AddPlatformWithType(
                    mouseWorldX-g.editorPlatformWidth/2,
                    mouseWorldY-g.editorPlatformHeight/2,
                    g.editorPlatformWidth,
                    g.editorPlatformHeight,
                    "normal",
                    0, 0, 0, // No movement parameters
                )
                fmt.Printf("Added normal platform at (%.1f, %.1f)\n", mouseWorldX, mouseWorldY)
            }

        case "enemy":
            // Create new enemy spawn point
            g.currentLevel.AddEnemySpawn(mouseWorldX, mouseWorldY, g.editorEnemyType)
            fmt.Printf("Added %s enemy spawn at (%.1f, %.1f)\n", g.editorEnemyType, mouseWorldX, mouseWorldY)

        case "flag":
            // Move flag
            if g.flag != nil {
                g.world.RemoveEntity(g.flag)
            }
            g.flag = NewFlag(mouseWorldX, mouseWorldY)
            g.world.AddEntity(g.flag)
            fmt.Printf("Moved flag to (%.1f, %.1f)\n", mouseWorldX, mouseWorldY)

        case "player":
            // Set player start position
            g.player.X = mouseWorldX
            g.player.Y = mouseWorldY
            fmt.Printf("Set player start to (%.1f, %.1f)\n", mouseWorldX, mouseWorldY)
        }
    }

    // Delete selected entity with Delete key
    if ebiten.IsKeyPressed(ebiten.KeyDelete) && !g.isKeyPressedPreviously(ebiten.KeyDelete) {
        // Find entity under mouse cursor
        mouseWorldX, mouseWorldY := g.getMouseWorldPosition()

        // Check platforms
        for i, platform := range g.platforms {
            if mouseWorldX >= platform.X && mouseWorldX <= platform.X+platform.Width &&
                mouseWorldY >= platform.Y && mouseWorldY <= platform.Y+platform.Height {
                // Remove platform
                g.world.RemoveEntity(platform)
                g.platforms = append(g.platforms[:i], g.platforms[i+1:]...)
                fmt.Printf("Deleted platform at (%.1f, %.1f)\n", platform.X, platform.Y)
                break
            }
        }
    }

    // Save level with S key
    if ebiten.IsKeyPressed(ebiten.KeyS) && !g.isKeyPressedPreviously(ebiten.KeyS) {
        if err := g.SaveCurrentLevel(); err != nil {
            fmt.Println("Error saving level:", err)
        } else {
            fmt.Println("Level saved successfully")
        }
    }

    // Create new level with N key
    if ebiten.IsKeyPressed(ebiten.KeyN) && !g.isKeyPressedPreviously(ebiten.KeyN) {
        g.createNewLevel()
    }

    // Camera movement
    cameraSpeed := 10.0
    if ebiten.IsKeyPressed(ebiten.KeyA) {
        g.world.GetCamera().X -= cameraSpeed
    } else if ebiten.IsKeyPressed(ebiten.KeyD) {
        g.world.GetCamera().X += cameraSpeed
    }
}

// getMouseWorldPosition converts screen mouse coordinates to world coordinates
func (g *Game) getMouseWorldPosition() (float64, float64) {
    // Get mouse position in screen coordinates
    mouseX, mouseY := ebiten.CursorPosition()

    // Convert to world coordinates
    camera := g.world.GetCamera()
    mouseWorldX := float64(mouseX) + camera.X
    mouseWorldY := float64(mouseY)

    return mouseWorldX, mouseWorldY
}

// createNewLevel creates a new empty level
func (g *Game) createNewLevel() {
    // Generate a new level name
    levelNum := len(g.levels) + 1
    levelName := fmt.Sprintf("level%d", levelNum)

    // Create new empty level
    newLevel := NewLevel(fmt.Sprintf("Level %d", levelNum))

    // Set default player start position
    newLevel.SetPlayerStart(100, 100)

    // Add ground platform
    groundWidth := float64(g.width * 5)
    newLevel.AddPlatform(0, float64(g.height-50), groundWidth, 50)

    // Set flag position
    flagX := float64(g.width * 2)
    flagY := float64(g.height - 114)
    newLevel.SetFlagPosition(flagX, flagY)

    // Save new level
    levelPath := filepath.Join(g.levelsDir, levelName+".json")
    if err := newLevel.SaveToFile(levelPath); err != nil {
        fmt.Println("Error saving new level:", err)
        return
    }

    // Add to levels list
    g.levels = append(g.levels, levelName)

    // Load the new level
    g.currentLevelIndex = len(g.levels) - 1
    g.LoadLevel(g.currentLevelIndex)

    fmt.Printf("Created new level: %s\n", levelName)
}

// updateInputState updates the state of all keys and mouse buttons for the next frame
func (g *Game) updateInputState() {
    // Update key states for weapon switching keys
    g.prevKeys[ebiten.Key1] = ebiten.IsKeyPressed(ebiten.Key1)
    g.prevKeys[ebiten.Key2] = ebiten.IsKeyPressed(ebiten.Key2)
    g.prevKeys[ebiten.Key3] = ebiten.IsKeyPressed(ebiten.Key3)
    g.prevKeys[ebiten.Key4] = ebiten.IsKeyPressed(ebiten.Key4)

    // Update key states for other keys
    g.prevKeys[ebiten.KeyEnter] = ebiten.IsKeyPressed(ebiten.KeyEnter)
    g.prevKeys[ebiten.KeyD] = ebiten.IsKeyPressed(ebiten.KeyD)
    g.prevKeys[ebiten.KeyE] = ebiten.IsKeyPressed(ebiten.KeyE)
    g.prevKeys[ebiten.KeyR] = ebiten.IsKeyPressed(ebiten.KeyR)
    g.prevKeys[ebiten.KeyF] = ebiten.IsKeyPressed(ebiten.KeyF)
    g.prevKeys[ebiten.KeyT] = ebiten.IsKeyPressed(ebiten.KeyT)
    g.prevKeys[ebiten.KeyS] = ebiten.IsKeyPressed(ebiten.KeyS)
    g.prevKeys[ebiten.KeyN] = ebiten.IsKeyPressed(ebiten.KeyN)
    g.prevKeys[ebiten.KeyK] = ebiten.IsKeyPressed(ebiten.KeyK)
    g.prevKeys[ebiten.KeyDelete] = ebiten.IsKeyPressed(ebiten.KeyDelete)

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
    g.world.Draw(screen, g.editorMode)

    // Draw blood particles
    g.drawBloodParticles(screen)

    // Draw UI elements
    if g.editorMode {
        // Draw editor UI
        g.drawEditorUI(screen)
    } else {
        // Draw game UI
        g.drawUI(screen)
    }

    // Draw crosshair at mouse position
    g.drawCrosshair(screen)

    // Draw debug info
    if g.debug {
        g.drawDebugInfo(screen)
    }
}

// drawEditorUI draws the editor UI
func (g *Game) drawEditorUI(screen *ebiten.Image) {
    // Draw editor info box
    boxWidth := 300
    boxHeight := 180
    boxX := 20
    boxY := 20

    // Draw background with border
    ebitenutil.DrawRect(screen, float64(boxX), float64(boxY), float64(boxWidth), float64(boxHeight), color.RGBA{40, 40, 40, 220})

    // Draw border
    borderSize := 2.0
    borderColor := color.RGBA{255, 215, 0, 255} // Gold
    ebitenutil.DrawRect(screen, float64(boxX), float64(boxY), float64(boxWidth), borderSize, borderColor)
    ebitenutil.DrawRect(screen, float64(boxX), float64(boxY+boxHeight-int(borderSize)), float64(boxWidth), borderSize, borderColor)
    ebitenutil.DrawRect(screen, float64(boxX), float64(boxY), borderSize, float64(boxHeight), borderColor)
    ebitenutil.DrawRect(screen, float64(boxX+boxWidth-int(borderSize)), float64(boxY), borderSize, float64(boxHeight), borderColor)

    // Draw title
    titleY := boxY + 10
    ebitenutil.DebugPrintAt(screen, "LEVEL EDITOR", boxX+10, titleY)

    // Draw current level info
    levelY := titleY + 20
    levelInfo := fmt.Sprintf("Level: %s (%d/%d)", g.levels[g.currentLevelIndex], g.currentLevelIndex+1, len(g.levels))
    ebitenutil.DebugPrintAt(screen, levelInfo, boxX+10, levelY)

    // Draw current tool info
    toolY := levelY + 20
    toolInfo := fmt.Sprintf("Tool: %s", g.editorTool)
    ebitenutil.DebugPrintAt(screen, toolInfo, boxX+10, toolY)

    // Draw tool-specific info
    infoY := toolY + 20
    switch g.editorTool {
    case "platform":
        // Show platform type
        platformTypeInfo := fmt.Sprintf("Platform Type: %s", g.editorPlatformType)
        ebitenutil.DebugPrintAt(screen, platformTypeInfo, boxX+10, infoY)
        infoY += 20

        // Show platform type selection controls
        ebitenutil.DebugPrintAt(screen, "Press N: Normal, S: Small, M: Moving, P: Spike", boxX+10, infoY)
        infoY += 20

        // Show platform size (except for small platforms which have fixed size)
        if g.editorPlatformType != "small" {
            platformInfo := fmt.Sprintf("Platform Size: %.0f x %.0f", g.editorPlatformWidth, g.editorPlatformHeight)
            ebitenutil.DebugPrintAt(screen, platformInfo, boxX+10, infoY)
            infoY += 20
            ebitenutil.DebugPrintAt(screen, "Use arrow keys to adjust size", boxX+10, infoY)
            infoY += 20
        }

        // Show moving platform configuration
        if g.editorPlatformType == "moving" {
            moveInfo := fmt.Sprintf("Movement: X=%.0f Y=%.0f Speed=%.1f",
                g.editorMovingPlatformMoveX, g.editorMovingPlatformMoveY, g.editorMovingPlatformSpeed)
            ebitenutil.DebugPrintAt(screen, moveInfo, boxX+10, infoY)
            infoY += 20
            ebitenutil.DebugPrintAt(screen, "Hold H + Left/Right: Adjust horizontal movement", boxX+10, infoY)
            infoY += 20
            ebitenutil.DebugPrintAt(screen, "Hold V + Up/Down: Adjust vertical movement", boxX+10, infoY)
            infoY += 20
            ebitenutil.DebugPrintAt(screen, "Press +/-: Adjust speed", boxX+10, infoY)
        }
    case "enemy":
        enemyInfo := fmt.Sprintf("Enemy Type: %s", g.editorEnemyType)
        ebitenutil.DebugPrintAt(screen, enemyInfo, boxX+10, infoY)
        infoY += 20
        ebitenutil.DebugPrintAt(screen, "Press K for kidney bean, N for navy bean, S for shooter", boxX+10, infoY)
    }

    // Draw controls
    controlsY := infoY + 30
    ebitenutil.DebugPrintAt(screen, "Controls:", boxX+10, controlsY)
    controlsY += 15
    ebitenutil.DebugPrintAt(screen, "1-4: Select tool (Platform, Enemy, Flag, Player)", boxX+10, controlsY)
    controlsY += 15
    ebitenutil.DebugPrintAt(screen, "Left Click: Place/Select entity", boxX+10, controlsY)
    controlsY += 15
    ebitenutil.DebugPrintAt(screen, "Delete: Delete entity under cursor", boxX+10, controlsY)
    controlsY += 15
    ebitenutil.DebugPrintAt(screen, "S: Save level", boxX+10, controlsY)
    controlsY += 15
    ebitenutil.DebugPrintAt(screen, "N: Create new level", boxX+10, controlsY)
    controlsY += 15
    ebitenutil.DebugPrintAt(screen, "E: Exit editor mode", boxX+10, controlsY)

    // Draw entity preview at mouse position
    mouseX, mouseY := ebiten.CursorPosition()
    mouseWorldX, mouseWorldY := g.getMouseWorldPosition()

    switch g.editorTool {
    case "platform":
        // Draw platform preview
        previewX := mouseWorldX - g.editorPlatformWidth/2 - g.world.GetCamera().X
        previewY := mouseWorldY - g.editorPlatformHeight/2
        previewColor := color.RGBA{0, 255, 0, 128} // Semi-transparent green
        ebitenutil.DrawRect(screen, previewX, previewY, g.editorPlatformWidth, g.editorPlatformHeight, previewColor)

    case "enemy":
        // Draw enemy preview
        previewRadius := 16.0
        previewX := float64(mouseX)
        previewY := float64(mouseY)
        previewColor := color.RGBA{139, 0, 0, 128} // Semi-transparent dark red for kidney bean
        if g.editorEnemyType == "navy" {
            previewColor = color.RGBA{0, 0, 128, 128} // Semi-transparent dark blue for navy bean
        }
        drawCircle(screen, previewX, previewY, previewRadius, previewColor)

    case "flag":
        // Draw flag preview
        previewWidth := 32.0
        previewHeight := 64.0
        previewX := float64(mouseX) - previewWidth/2
        previewY := float64(mouseY) - previewHeight/2
        previewColor := color.RGBA{255, 215, 0, 128} // Semi-transparent gold
        ebitenutil.DrawRect(screen, previewX, previewY, previewWidth, previewHeight, previewColor)

    case "player":
        // Draw player preview
        previewWidth := 32.0
        previewHeight := 64.0
        previewX := float64(mouseX) - previewWidth/2
        previewY := float64(mouseY) - previewHeight/2
        previewColor := color.RGBA{160, 120, 80, 128} // Semi-transparent light brown
        ebitenutil.DrawRect(screen, previewX, previewY, previewWidth, previewHeight, previewColor)
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
    crosshairColor := color.RGBA{255, 255, 255, 255}  // White
    crosshairOutlineColor := color.RGBA{0, 0, 0, 255} // Black

    // Get mouse position
    mouseX := float64(g.mouseX)
    mouseY := float64(g.mouseY)

    // Draw crosshair outline (black)
    // Horizontal line
    ebitenutil.DrawRect(screen, mouseX-crosshairSize-1, mouseY-crosshairThickness/2-1, crosshairSize*2+2, crosshairThickness+2, crosshairOutlineColor)
    // Vertical line
    ebitenutil.DrawRect(screen, mouseX-crosshairThickness/2-1, mouseY-crosshairSize-1, crosshairThickness+2, crosshairSize*2+2, crosshairOutlineColor)

    // Draw crosshair (white)
    // Horizontal line
    ebitenutil.DrawRect(screen, mouseX-crosshairSize, mouseY-crosshairThickness/2, crosshairSize*2, crosshairThickness, crosshairColor)
    // Vertical line
    ebitenutil.DrawRect(screen, mouseX-crosshairThickness/2, mouseY-crosshairSize, crosshairThickness, crosshairSize*2, crosshairColor)

    // Draw a small gap in the center
    gapSize := 2.0
    ebitenutil.DrawRect(screen, mouseX-gapSize, mouseY-gapSize, gapSize*2, gapSize*2, crosshairOutlineColor)
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
            reloadProgress := float64(g.player.ReloadTime-g.player.ReloadTimer) / float64(g.player.ReloadTime) * 100
            ammoText = fmt.Sprintf("RELOADING... %.0f%%", reloadProgress)
        }
    } else if g.player.CurrentWeapon == 1 {
        // Rocket launcher
        weaponName = "ROCKET LAUNCHER"
        ammoText = fmt.Sprintf("%d / %d", g.player.CurrentRocket, g.player.RocketCount)
        iconColor = color.RGBA{255, 100, 0, 255} // Orange

        // Draw reload indicator if reloading
        if g.player.IsRocketReloading {
            reloadProgress := float64(g.player.RocketReloadTime-g.player.RocketReloadTimer) / float64(g.player.RocketReloadTime) * 100
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
    ebitenutil.DrawRect(screen, ammoBoxX, ammoBoxY, ammoBoxWidth, borderSize, color.RGBA{255, 215, 0, 255})                          // Gold top
    ebitenutil.DrawRect(screen, ammoBoxX, ammoBoxY+ammoBoxHeight-borderSize, ammoBoxWidth, borderSize, color.RGBA{255, 215, 0, 255}) // Gold bottom
    ebitenutil.DrawRect(screen, ammoBoxX, ammoBoxY, borderSize, ammoBoxHeight, color.RGBA{255, 215, 0, 255})                         // Gold left
    ebitenutil.DrawRect(screen, ammoBoxX+ammoBoxWidth-borderSize, ammoBoxY, borderSize, ammoBoxHeight, color.RGBA{255, 215, 0, 255}) // Gold right

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
        ebitenutil.DrawRect(screen, iconX, iconY-2, iconWidth/3, iconHeight/2, iconColor)
        ebitenutil.DrawRect(screen, iconX, iconY+iconHeight, iconWidth/3, iconHeight/2, iconColor)
    } else {
        // Draw grenade icon (circle with pin)
        // Draw grenade body (circle)
        drawCircle(screen, iconX+iconWidth/2, iconY+iconHeight/2, iconWidth/2, iconColor)
        // Draw pin
        pinColor := color.RGBA{200, 200, 200, 255} // Silver
        ebitenutil.DrawRect(screen, iconX+iconWidth/2-1, iconY-3, 2, 3, pinColor)
    }

    // Draw text with offset to center vertically
    textY := int(ammoBoxY) + int(ammoBoxHeight)/2 - 3
    ebitenutil.DebugPrintAt(screen, fullText, int(ammoBoxX)+20, textY)

    // Draw lives as hearts
    heartSize := 25.0
    heartSpacing := 10.0
    heartsX := 20.0
    heartsY := 20.0

    // Calculate number of hearts based on max health (1 heart = 20 health)
    maxHearts := (g.player.MaxHealth + 19) / 20  // Ceiling division
    currentHearts := (g.player.Health + 19) / 20 // Ceiling division

    // Draw heart background with border
    backgroundWidth := float64(maxHearts)*(heartSize+heartSpacing) + heartSpacing
    backgroundHeight := heartSize + heartSpacing*2

    // Draw background
    ebitenutil.DrawRect(screen, heartsX-heartSpacing, heartsY-heartSpacing,
        backgroundWidth, backgroundHeight,
        color.RGBA{40, 40, 40, 220})

    // Draw border
    heartBorderSize := 2.0
    borderColor := color.RGBA{255, 215, 0, 255} // Gold

    // Top border
    ebitenutil.DrawRect(screen, heartsX-heartSpacing, heartsY-heartSpacing,
        backgroundWidth, heartBorderSize, borderColor)

    // Bottom border
    ebitenutil.DrawRect(screen, heartsX-heartSpacing, heartsY-heartSpacing+backgroundHeight-heartBorderSize,
        backgroundWidth, heartBorderSize, borderColor)

    // Left border
    ebitenutil.DrawRect(screen, heartsX-heartSpacing, heartsY-heartSpacing,
        heartBorderSize, backgroundHeight, borderColor)

    // Right border
    ebitenutil.DrawRect(screen, heartsX-heartSpacing+backgroundWidth-heartBorderSize, heartsY-heartSpacing,
        heartBorderSize, backgroundHeight, borderColor)

    // Draw hearts
    for i := 0; i < maxHearts; i++ {
        heartX := heartsX + float64(i)*(heartSize+heartSpacing)
        heartY := heartsY

        // Determine heart color based on whether it's a filled or empty heart
        var heartColor color.RGBA
        if i < currentHearts {
            // Filled heart (red)
            heartColor = color.RGBA{255, 0, 0, 255}
        } else {
            // Empty heart (semi-transparent gray)
            heartColor = color.RGBA{150, 150, 150, 150}
        }

        // Draw the heart
        drawHeart(screen, heartX+heartSize/2, heartY+heartSize/2, heartSize, heartColor)
    }

    // Draw "LIVES" text
    livesText := "LIVES"
    textX := heartsX + backgroundWidth + 10
    livesTextY := int(heartsY + heartSize/2)

    // Draw text shadow
    shadowOffset := 1
    ebitenutil.DebugPrintAt(screen, livesText, int(textX)+shadowOffset, livesTextY+shadowOffset)

    // Draw text
    ebitenutil.DebugPrintAt(screen, livesText, int(textX), livesTextY)

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
            baseAngle := float64(g.frameCount%360) * 0.01
            angleRad := baseAngle + float64(i)*(2*3.14159/float64(starCount))

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
            msgX+(msgWidth-len(completeMsg)*7)/2+shadowOffset,
            msgY+50+shadowOffset)

        // Draw text in gold
        ebitenutil.DebugPrintAt(screen, completeMsg,
            msgX+(msgWidth-len(completeMsg)*7)/2,
            msgY+50)

        // Draw congratulatory message
        congratsMsg := "Congratulations!"
        ebitenutil.DebugPrintAt(screen, congratsMsg,
            msgX+(msgWidth-len(congratsMsg)*7)/2,
            msgY+80)

        // Draw instruction to continue with animation effect
        continueMsg := "Press ENTER to continue"

        // Make the text pulse by using the frame count
        pulseOffset := 0
        if g.frameCount%60 < 30 {
            pulseOffset = 1
        }

        ebitenutil.DebugPrintAt(screen, continueMsg,
            msgX+(msgWidth-len(continueMsg)*7)/2,
            msgY+110+pulseOffset)
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
