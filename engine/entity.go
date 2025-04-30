package engine

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Camera represents the view into the game world
type Camera struct {
	X      float64
	Y      float64
	Width  int
	Height int
	Zoom   float64 // Zoom factor (1.0 = no zoom)
}

// NewCamera creates a new camera with the given dimensions
func NewCamera(width, height int) *Camera {
	return &Camera{
		X:      0,
		Y:      0,
		Width:  width,
		Height: height,
		Zoom:   1.0, // Default zoom factor (no zoom)
	}
}

// Follow makes the camera follow a target entity
func (c *Camera) Follow(target Entity, screenWidth, screenHeight int) {
	// Center the camera on the target
	c.X = target.GetX() - float64(screenWidth)/2 + target.GetWidth()/2

	// Optional: You can also make the camera follow vertically
	// c.Y = target.GetY() - float64(screenHeight)/2 + target.GetHeight()/2

	// Keep the camera within bounds (if needed)
	if c.X < 0 {
		c.X = 0
	}
}

// Entity represents any game object with position and collision
type Entity interface {
    Update()
    Draw(screen *ebiten.Image)
    GetX() float64
    GetY() float64
    GetWidth() float64
    GetHeight() float64
    SetX(x float64)
    SetY(y float64)
    SetVelX(vx float64)
    SetVelY(vy float64)
    GetVelX() float64
    GetVelY() float64
    IsCollidable() bool
}

// World manages all game entities and physics
type World struct {
    entities []Entity
    gravity  float64
    camera   *Camera
}

// NewWorld creates a new game world with physics
func NewWorld(gravity float64, width, height int) *World {
    return &World{
        entities: make([]Entity, 0),
        gravity:  gravity,
        camera:   NewCamera(width, height),
    }
}

// GetCamera returns the world's camera
func (w *World) GetCamera() *Camera {
    return w.camera
}

// AddEntity adds an entity to the world
func (w *World) AddEntity(e Entity) {
    w.entities = append(w.entities, e)
}

// RemoveEntity removes an entity from the world
func (w *World) RemoveEntity(e Entity) {
    for i, entity := range w.entities {
        if entity == e {
            // Remove the entity by replacing it with the last entity
            // and then truncating the slice
            w.entities[i] = w.entities[len(w.entities)-1]
            w.entities = w.entities[:len(w.entities)-1]
            return
        }
    }
}

// Update updates all entities and handles physics
func (w *World) Update() {
    // Update all entities
    for _, e := range w.entities {
        e.Update()
    }

    // Apply gravity and handle collisions
    w.applyPhysics()
}

// applyPhysics applies gravity and resolves collisions
func (w *World) applyPhysics() {
    // Apply gravity to all entities
    for _, e := range w.entities {
        // Skip non-movable entities
        if !e.IsCollidable() {
            continue
        }

        // Apply gravity (with reduced effect for projectiles)
        if bullet, isBullet := e.(*Bullet); isBullet {
            // Apply reduced gravity to bullets (1/10 of normal gravity)
            bullet.VelY += w.gravity * 0.1
        } else if rocket, isRocket := e.(*Rocket); isRocket {
            // Apply reduced gravity to rockets (1/10 of normal gravity)
            rocket.VelY += w.gravity * 0.1
        } else if grenade, isGrenade := e.(*Grenade); isGrenade {
            // Apply reduced gravity to grenades (1/5 of normal gravity)
            grenade.VelY += w.gravity * 0.2
        } else {
            // Apply normal gravity to other entities
            e.SetVelY(e.GetVelY() + w.gravity)
        }

        // Update position based on velocity
        e.SetX(e.GetX() + e.GetVelX())
        e.SetY(e.GetY() + e.GetVelY())

        // Check for collisions with other entities
        w.resolveCollisions(e)
    }
}

// resolveCollisions handles collision detection and response
func (w *World) resolveCollisions(e Entity) {
    for _, other := range w.entities {
        // Skip self or non-collidable entities
        if e == other || !other.IsCollidable() {
            continue
        }

        // Check for collision
        if w.checkCollision(e, other) {
            w.resolveCollision(e, other)
        }
    }
}

// checkCollision detects if two entities are colliding
func (w *World) checkCollision(a, b Entity) bool {
    // Simple AABB collision detection
    return a.GetX() < b.GetX()+b.GetWidth() &&
        a.GetX()+a.GetWidth() > b.GetX() &&
        a.GetY() < b.GetY()+b.GetHeight() &&
        a.GetY()+a.GetHeight() > b.GetY()
}

// resolveCollision handles collision response
func (w *World) resolveCollision(a, b Entity) {
    // Check if entity a is a player and b is a platform
    player, isPlayer := a.(*Player)
    _, isPlatform := b.(*Platform)

    // Check if entity a is a bullet
    bullet, isBullet := a.(*Bullet)

    // Check if entity a is a grenade
    grenade, isGrenade := a.(*Grenade)

    // Check if entity a is an enemy
    enemy, isEnemy := a.(*Enemy)

    // Handle player vs platform collision
    if isPlayer && isPlatform {
        // Calculate overlap
        overlapX := min(a.GetX()+a.GetWidth(), b.GetX()+b.GetWidth()) - max(a.GetX(), b.GetX())
        overlapY := min(a.GetY()+a.GetHeight(), b.GetY()+b.GetHeight()) - max(a.GetY(), b.GetY())

        // Resolve collision based on smallest overlap
        if overlapX < overlapY {
            // Horizontal collision
            if a.GetX() < b.GetX() {
                a.SetX(b.GetX() - a.GetWidth())
            } else {
                a.SetX(b.GetX() + b.GetWidth())
            }
            a.SetVelX(0)
        } else {
            // Vertical collision
            if a.GetY() < b.GetY() {
                // Collision from above
                a.SetY(b.GetY() - a.GetHeight())
                a.SetVelY(0)
                player.OnGround = true
            } else {
                // Collision from below
                a.SetY(b.GetY() + b.GetHeight())
                a.SetVelY(0)
            }
        }
    }

    // Handle enemy vs platform collision
    if isEnemy && isPlatform {
        // Calculate overlap
        overlapX := min(a.GetX()+a.GetWidth(), b.GetX()+b.GetWidth()) - max(a.GetX(), b.GetX())
        overlapY := min(a.GetY()+a.GetHeight(), b.GetY()+b.GetHeight()) - max(a.GetY(), b.GetY())

        // Resolve collision based on smallest overlap
        if overlapX < overlapY {
            // Horizontal collision - enemy hit a wall
            if a.GetX() < b.GetX() {
                a.SetX(b.GetX() - a.GetWidth())
            } else {
                a.SetX(b.GetX() + b.GetWidth())
            }
            // Change direction when hitting a wall
            enemy.MoveDir *= -1
            a.SetVelX(float64(enemy.MoveDir) * enemy.Speed)
        } else {
            // Vertical collision
            if a.GetY() < b.GetY() {
                // Collision from above
                a.SetY(b.GetY() - a.GetHeight())
                a.SetVelY(0)
                enemy.OnGround = true
            } else {
                // Collision from below
                a.SetY(b.GetY() + b.GetHeight())
                a.SetVelY(0)
            }
        }
    }

    // Handle bullet vs platform collision
    if isBullet && isPlatform {
        // Deactivate bullet when it hits a platform
        bullet.Active = false
    }

    // Handle grenade vs platform collision
    if isGrenade && isPlatform {
        // Calculate overlap
        overlapX := min(a.GetX()+a.GetWidth(), b.GetX()+b.GetWidth()) - max(a.GetX(), b.GetX())
        overlapY := min(a.GetY()+a.GetHeight(), b.GetY()+b.GetHeight()) - max(a.GetY(), b.GetY())

        // Resolve collision based on smallest overlap
        if overlapX < overlapY {
            // Horizontal collision - bounce with reduced velocity
            if a.GetX() < b.GetX() {
                a.SetX(b.GetX() - a.GetWidth())
                a.SetVelX(-grenade.VelX * 0.6) // Bounce with 60% of velocity
            } else {
                a.SetX(b.GetX() + b.GetWidth())
                a.SetVelX(-grenade.VelX * 0.6) // Bounce with 60% of velocity
            }
            // Decrease bounce counter
            grenade.HandleBounce()
        } else {
            // Vertical collision
            if a.GetY() < b.GetY() {
                // Collision from above
                a.SetY(b.GetY() - a.GetHeight())
                a.SetVelY(-grenade.VelY * 0.5) // Bounce with 50% of velocity
                // Decrease bounce counter
                grenade.HandleBounce()
            } else {
                // Collision from below
                a.SetY(b.GetY() + b.GetHeight())
                a.SetVelY(0)
            }
        }
    }
}

// Draw draws all entities in the world
func (w *World) Draw(screen *ebiten.Image, editorMode bool) {
    for _, e := range w.entities {
        // Draw the entity with camera offset
        x := e.GetX() - w.camera.X
        y := e.GetY()

        // Skip drawing entities that are off-screen
        if x+e.GetWidth() < 0 || x > float64(w.camera.Width) {
            continue
        }

        // Draw entities with improved visuals
        if player, ok := e.(*Player); ok {
            // Skip drawing player if in editor mode
            if !editorMode {
                // Draw player with more details
                drawPlayer(screen, x, y, player)
            }
        } else if platform, ok := e.(*Platform); ok {
            // Draw platform with texture
            drawPlatform(screen, x, y, platform)
        } else if bullet, ok := e.(*Bullet); ok {
            if bullet.Active {
                // Draw bullet with trail effect
                drawBullet(screen, x, y, bullet)
            }
        } else if rocket, ok := e.(*Rocket); ok {
            if rocket.Active {
                // Draw rocket with effects
                drawRocket(screen, x, y, rocket)
            }
        } else if grenade, ok := e.(*Grenade); ok {
            if grenade.Active {
                // Draw grenade with effects
                drawGrenade(screen, x, y, grenade)
            }
        } else if enemy, ok := e.(*Enemy); ok {
            if enemy.Active && !enemy.Dead {
                // Draw enemy with more details
                drawEnemy(screen, x, y, enemy)
            }
        } else if flag, ok := e.(*Flag); ok {
            if flag.Active {
                // Draw flag with improved animation
                drawFlag(screen, x, y, flag)
            }
        }
    }
}

// drawPlayer draws the player with more details
func drawPlayer(screen *ebiten.Image, x, y float64, player *Player) {
    // Main body - pinto bean shaped (light brown with spots)
    bodyColor := player.Color

    // Draw bean body with more realistic pinto bean appearance
    beanWidth := player.Width * 1.2
    beanHeight := player.Height * 0.9
    beanX := x - (beanWidth - player.Width) / 2
    beanY := y + player.Height * 0.1

    // Draw main bean body (pinto bean shape - slightly asymmetric)
    for i := 0.0; i < beanWidth; i += 1.0 {
        // Calculate position along the width
        relativePos := i / beanWidth

        // Create pinto bean shape (slightly asymmetric)
        distFromCenter := math.Abs(i - beanWidth/2)

        // Make slightly asymmetric for natural bean shape
        asymmetry := 0.0
        if i > beanWidth/2 {
            asymmetry = (i - beanWidth/2) * 0.02 // Slight curve on right side
        }

        // Use more natural curve for bean shape
        heightReduction := math.Pow(distFromCenter / (beanWidth/2), 1.5) * (beanWidth/7) + asymmetry
        columnHeight := beanHeight - heightReduction

        // Determine if this pixel should have a spot (pinto bean characteristic)
        // Create a few larger spots in specific areas
        hasSpot := false
        spotIntensity := 0.0

        // Create a few defined spots in specific positions
        spots := []struct{
            centerX, centerY, radius float64
        }{
            {beanWidth * 0.3, beanHeight * 0.3, beanWidth * 0.15},
            {beanWidth * 0.7, beanHeight * 0.5, beanWidth * 0.12},
            {beanWidth * 0.5, beanHeight * 0.7, beanWidth * 0.1},
            {beanWidth * 0.2, beanHeight * 0.6, beanWidth * 0.08},
            {beanWidth * 0.8, beanHeight * 0.2, beanWidth * 0.09},
        }

        // Check if current pixel is within any spot
        for _, spot := range spots {
            // Calculate distance from spot center
            pixelY := beanY + columnHeight/2 // Approximate y position
            distX := math.Abs(i - spot.centerX)
            distY := math.Abs(pixelY - (beanY + spot.centerY))
            dist := math.Sqrt(distX*distX + distY*distY)

            // If within spot radius, mark as spot with intensity based on distance
            if dist < spot.radius {
                hasSpot = true
                // Fade spot intensity at edges
                spotIntensity = math.Max(spotIntensity, 1.0 - dist/spot.radius)
            }
        }

        // Add slight texture variation
        colorVariation := uint8(math.Sin(relativePos*math.Pi*5) * 8)
        pixelColor := color.RGBA{
            R: addColorValue(bodyColor.R, colorVariation),
            G: addColorValue(bodyColor.G, colorVariation),
            B: addColorValue(bodyColor.B, colorVariation/2),
            A: bodyColor.A,
        }

        // Apply spot color if this pixel is in a spot
        if hasSpot {
            // Darker brown for spots
            spotColor := color.RGBA{
                R: subtractColorValue(bodyColor.R, uint8(40.0 * spotIntensity)),
                G: subtractColorValue(bodyColor.G, uint8(30.0 * spotIntensity)),
                B: subtractColorValue(bodyColor.B, uint8(20.0 * spotIntensity)),
                A: bodyColor.A,
            }
            pixelColor = spotColor
        }

        ebitenutil.DrawRect(screen, beanX + i, beanY, 1, columnHeight, pixelColor)
    }

    // Draw bean "crease" line with more natural curve
    creaseColor := color.RGBA{
        R: subtractColorValue(bodyColor.R, 40),
        G: subtractColorValue(bodyColor.G, 40),
        B: subtractColorValue(bodyColor.B, 40),
        A: bodyColor.A,
    }

    // Draw curved crease instead of straight line
    creaseWidth := beanWidth * 0.7
    creaseSegments := 10
    segmentWidth := creaseWidth / float64(creaseSegments)
    creaseX := beanX + (beanWidth - creaseWidth) / 2
    creaseBaseY := beanY + beanHeight * 0.4

    for i := 0; i < creaseSegments; i++ {
        segmentX := creaseX + float64(i) * segmentWidth

        // Create slight curve in the crease
        curveOffset := math.Sin(float64(i)/float64(creaseSegments-1)*math.Pi) * 2.5
        segmentY := creaseBaseY - curveOffset

        // Vary thickness slightly
        thickness := 1.5 + math.Sin(float64(i)/float64(creaseSegments-1)*math.Pi) * 0.8

        ebitenutil.DrawRect(screen, segmentX, segmentY, segmentWidth, thickness, creaseColor)
    }

    // Draw head (slightly lighter color)
    headColor := color.RGBA{
        R: addColorValue(bodyColor.R, 30),
        G: addColorValue(bodyColor.G, 30),
        B: addColorValue(bodyColor.B, 30),
        A: bodyColor.A,
    }
    headSize := player.Width * 0.8
    headX := x + (player.Width-headSize)/2
    headY := y + player.Height*0.1
    ebitenutil.DrawRect(screen, headX, headY, headSize, headSize, headColor)

    // Draw cowboy hat
    hatBrimColor := color.RGBA{139, 69, 19, 255} // Brown
    hatTopColor := color.RGBA{160, 82, 45, 255} // Sienna

    // Hat dimensions
    hatWidth := headSize * 1.4
    hatHeight := headSize * 0.4
    hatTopHeight := headSize * 0.3
    hatX := headX + (headSize - hatWidth) / 2
    hatY := headY - hatHeight

    // Draw hat brim
    ebitenutil.DrawRect(screen, hatX, hatY + hatTopHeight, hatWidth, hatHeight - hatTopHeight, hatBrimColor)

    // Draw hat top (slightly narrower than brim)
    hatTopWidth := hatWidth * 0.7
    hatTopX := hatX + (hatWidth - hatTopWidth) / 2
    ebitenutil.DrawRect(screen, hatTopX, hatY, hatTopWidth, hatTopHeight, hatTopColor)

    // Draw hat band
    bandColor := color.RGBA{255, 0, 0, 255} // Red band
    bandHeight := hatHeight * 0.2
    ebitenutil.DrawRect(screen, hatTopX, hatY + hatTopHeight - bandHeight, hatTopWidth, bandHeight, bandColor)

    // Draw eyes
    eyeColor := color.RGBA{255, 255, 255, 255}
    eyeSize := player.Width * 0.15
    leftEyeX := headX + headSize*0.25 - eyeSize/2
    rightEyeX := headX + headSize*0.75 - eyeSize/2
    eyeY := headY + headSize*0.4 - eyeSize/2
    ebitenutil.DrawRect(screen, leftEyeX, eyeY, eyeSize, eyeSize, eyeColor)
    ebitenutil.DrawRect(screen, rightEyeX, eyeY, eyeSize, eyeSize, eyeColor)

    // Draw pupils (looking in direction of movement)
    pupilColor := color.RGBA{0, 0, 0, 255}
    pupilSize := eyeSize * 0.6
    pupilOffset := 0.0
    if player.FacingRight {
        pupilOffset = eyeSize * 0.2
    } else {
        pupilOffset = -eyeSize * 0.2
    }
    ebitenutil.DrawRect(screen, leftEyeX + (eyeSize-pupilSize)/2 + pupilOffset, eyeY + (eyeSize-pupilSize)/2, pupilSize, pupilSize, pupilColor)
    ebitenutil.DrawRect(screen, rightEyeX + (eyeSize-pupilSize)/2 + pupilOffset, eyeY + (eyeSize-pupilSize)/2, pupilSize, pupilSize, pupilColor)

    // Draw gun based on current weapon and angle
    // Calculate gun position at center of player
    gunCenterX := x + player.Width/2
    gunCenterY := y + player.Height * 0.4

    // Draw different gun based on current weapon
    if player.CurrentWeapon == 0 {
        // Machine gun (more detailed)
        gunBarrelColor := color.RGBA{70, 70, 70, 255} // Dark gray
        gunBodyColor := color.RGBA{100, 100, 100, 255} // Medium gray
        gunGripColor := color.RGBA{139, 69, 19, 255} // Brown wooden grip

        // Gun dimensions
        gunBarrelWidth := player.Width * 0.7
        gunBarrelHeight := player.Height * 0.08
        gunBodyWidth := player.Width * 0.3
        gunBodyHeight := player.Height * 0.15
        gunGripWidth := player.Width * 0.15
        gunGripHeight := player.Height * 0.2

        // Draw gun barrel (longer, thinner part)
        drawRotatedRectangle(screen, gunCenterX, gunCenterY, gunBarrelWidth, gunBarrelHeight, player.GunAngle, gunBarrelColor)

        // Draw gun body (wider part behind barrel)
        // Calculate position offset from center based on angle
        bodyOffsetX := -math.Cos(player.GunAngle) * (gunBarrelWidth/2 - gunBodyWidth/2)
        bodyOffsetY := -math.Sin(player.GunAngle) * (gunBarrelWidth/2 - gunBodyWidth/2)
        drawRotatedRectangle(screen, gunCenterX + bodyOffsetX, gunCenterY + bodyOffsetY, gunBodyWidth, gunBodyHeight, player.GunAngle, gunBodyColor)

        // Draw gun grip (handle)
        gripOffsetX := -math.Cos(player.GunAngle) * (gunBarrelWidth/2 - gunBodyWidth/4)
        gripOffsetY := -math.Sin(player.GunAngle) * (gunBarrelWidth/2 - gunBodyWidth/4)
        // Grip is perpendicular to barrel
        gripAngle := player.GunAngle + math.Pi/2
        drawRotatedRectangle(screen, gunCenterX + gripOffsetX, gunCenterY + gripOffsetY, gunGripWidth, gunGripHeight, gripAngle, gunGripColor)

        // Add a muzzle detail at the end of the barrel
        muzzleColor := color.RGBA{50, 50, 50, 255} // Darker gray
        muzzleWidth := gunBarrelHeight * 1.5
        muzzleHeight := gunBarrelHeight * 1.2
        muzzleOffsetX := math.Cos(player.GunAngle) * (gunBarrelWidth/2 - muzzleWidth/4)
        muzzleOffsetY := math.Sin(player.GunAngle) * (gunBarrelWidth/2 - muzzleWidth/4)
        drawRotatedRectangle(screen, gunCenterX + muzzleOffsetX, gunCenterY + muzzleOffsetY, muzzleWidth, muzzleHeight, player.GunAngle, muzzleColor)
    } else if player.CurrentWeapon == 1 {
        // Rocket launcher (more detailed)
        launcherBodyColor := color.RGBA{60, 60, 60, 255} // Dark gray
        launcherTubeColor := color.RGBA{80, 80, 80, 255} // Medium gray
        launcherDetailColor := color.RGBA{40, 40, 40, 255} // Very dark gray

        // Launcher dimensions
        launcherWidth := player.Width * 0.6
        launcherHeight := player.Height * 0.25
        launcherTubeWidth := player.Width * 0.7
        launcherTubeHeight := player.Height * 0.18

        // Draw main launcher tube
        drawRotatedRectangle(screen, gunCenterX, gunCenterY, launcherTubeWidth, launcherTubeHeight, player.GunAngle, launcherTubeColor)

        // Draw launcher body (back part)
        bodyOffsetX := -math.Cos(player.GunAngle) * (launcherTubeWidth/2 - launcherWidth/4)
        bodyOffsetY := -math.Sin(player.GunAngle) * (launcherTubeWidth/2 - launcherWidth/4)
        drawRotatedRectangle(screen, gunCenterX + bodyOffsetX, gunCenterY + bodyOffsetY, launcherWidth * 0.4, launcherHeight, player.GunAngle, launcherBodyColor)

        // Draw sight on top
        sightOffsetX := math.Cos(player.GunAngle + math.Pi/2) * (launcherHeight/4)
        sightOffsetY := math.Sin(player.GunAngle + math.Pi/2) * (launcherHeight/4)
        drawRotatedRectangle(screen, gunCenterX + sightOffsetX, gunCenterY + sightOffsetY, launcherWidth * 0.3, launcherHeight * 0.15, player.GunAngle, launcherDetailColor)

        // Draw handle/grip
        gripOffsetX := -math.Cos(player.GunAngle) * (launcherWidth/4)
        gripOffsetY := -math.Sin(player.GunAngle) * (launcherWidth/4)
        gripAngle := player.GunAngle + math.Pi/2
        drawRotatedRectangle(screen, gunCenterX + gripOffsetX, gunCenterY + gripOffsetY, launcherHeight * 0.2, launcherHeight * 0.5, gripAngle, launcherDetailColor)

        // Draw rocket in tube if loaded
        if player.CurrentRocket > 0 && !player.IsRocketReloading {
            rocketColor := color.RGBA{180, 70, 20, 255} // Orange-red
            rocketWidth := launcherTubeWidth * 0.5
            rocketHeight := launcherTubeHeight * 0.7
            rocketOffsetX := math.Cos(player.GunAngle) * (launcherTubeWidth/4)
            rocketOffsetY := math.Sin(player.GunAngle) * (launcherTubeWidth/4)
            drawRotatedRectangle(screen, gunCenterX + rocketOffsetX, gunCenterY + rocketOffsetY, rocketWidth, rocketHeight, player.GunAngle, rocketColor)
        }
    } else {
        // Grenade launcher/thrower
        launcherColor := color.RGBA{50, 100, 50, 255} // Green
        gripColor := color.RGBA{70, 50, 30, 255} // Brown

        // Launcher dimensions
        launcherWidth := player.Width * 0.4
        launcherHeight := player.Height * 0.2
        gripWidth := player.Width * 0.15
        gripHeight := player.Height * 0.25

        // Draw launcher body (short, wide tube)
        drawRotatedRectangle(screen, gunCenterX, gunCenterY, launcherWidth, launcherHeight, player.GunAngle, launcherColor)

        // Draw grip
        gripOffsetX := -math.Cos(player.GunAngle) * (launcherWidth/4)
        gripOffsetY := -math.Sin(player.GunAngle) * (launcherWidth/4)
        gripAngle := player.GunAngle + math.Pi/2
        drawRotatedRectangle(screen, gunCenterX + gripOffsetX, gunCenterY + gripOffsetY, gripWidth, gripHeight, gripAngle, gripColor)

        // Draw grenade in launcher if available
        if player.GrenadeCount > 0 {
            grenadeColor := color.RGBA{0, 100, 0, 255} // Dark green
            grenadeSize := launcherHeight * 0.8
            grenadeOffsetX := math.Cos(player.GunAngle) * (launcherWidth/2 - grenadeSize/2)
            grenadeOffsetY := math.Sin(player.GunAngle) * (launcherWidth/2 - grenadeSize/2)

            // Draw grenade as a circle
            drawCircle(screen, gunCenterX + grenadeOffsetX, gunCenterY + grenadeOffsetY, grenadeSize/2, grenadeColor)

            // Draw grenade pin
            pinColor := color.RGBA{200, 200, 200, 255} // Silver
            pinOffsetX := math.Cos(player.GunAngle + math.Pi/4) * (grenadeSize/2)
            pinOffsetY := math.Sin(player.GunAngle + math.Pi/4) * (grenadeSize/2)
            drawRotatedRectangle(screen, gunCenterX + grenadeOffsetX + pinOffsetX, gunCenterY + grenadeOffsetY + pinOffsetY, grenadeSize * 0.2, grenadeSize * 0.1, player.GunAngle, pinColor)
        }
    }

    // Draw legs
    legWidth := player.Width * 0.2
    legHeight := player.Height * 0.3
    legY := y + player.Height * 0.7

    // Left leg
    ebitenutil.DrawRect(screen, x + player.Width*0.2, legY, legWidth, legHeight, bodyColor)

    // Right leg
    ebitenutil.DrawRect(screen, x + player.Width*0.6, legY, legWidth, legHeight, bodyColor)

    // If player is invincible, draw a shield effect
    if player.InvincibleTime > 0 && player.InvincibleTime % 4 < 2 {
        shieldColor := color.RGBA{255, 255, 255, 100}
        ebitenutil.DrawRect(screen, x-2, y-2, player.Width+4, player.Height+4, shieldColor)
    }

    // Draw reload spinner above the player if reloading
    if player.IsReloading || player.IsRocketReloading {
        // Position spinner above the player's head
        spinnerX := x + player.Width/2
        spinnerY := y - 20 // Position above the player's head

        // Determine spinner rotation based on reload progress
        var progress float64
        if player.IsReloading {
            progress = float64(player.ReloadTime - player.ReloadTimer) / float64(player.ReloadTime)
        } else {
            progress = float64(player.RocketReloadTime - player.RocketReloadTimer) / float64(player.RocketReloadTime)
        }

        // Spinner properties
        spinnerRadius := 10.0
        numSegments := 8
        segmentRadius := 3.0

        // Semi-transparent color for the spinner
        spinnerColor := color.RGBA{255, 255, 255, 150} // Semi-transparent white
        activeSegmentColor := color.RGBA{255, 215, 0, 200} // Semi-transparent gold for active segment

        // Draw spinner segments
        for i := 0; i < numSegments; i++ {
            angle := float64(i) * (2 * math.Pi / float64(numSegments))

            // Calculate segment position
            segX := spinnerX + math.Cos(angle) * spinnerRadius
            segY := spinnerY + math.Sin(angle) * spinnerRadius

            // Determine if this segment should be highlighted based on progress
            segmentProgress := float64(i) / float64(numSegments)

            // Use different color for segments that represent completed progress
            if segmentProgress <= progress {
                drawCircle(screen, segX, segY, segmentRadius, activeSegmentColor)
            } else {
                drawCircle(screen, segX, segY, segmentRadius, spinnerColor)
            }
        }
    }
}

// drawRotatedRectangle draws a rectangle rotated around its center
func drawRotatedRectangle(screen *ebiten.Image, centerX, centerY, width, height, angle float64, clr color.RGBA) {
    // Calculate the four corners of the rectangle
    halfWidth := width / 2
    halfHeight := height / 2

    // Calculate sin and cos of the angle once
    sinAngle := math.Sin(angle)
    cosAngle := math.Cos(angle)

    // Calculate rotated corners
    x1 := centerX + cosAngle*halfWidth - sinAngle*halfHeight
    y1 := centerY + sinAngle*halfWidth + cosAngle*halfHeight

    x2 := centerX + cosAngle*halfWidth + sinAngle*halfHeight
    y2 := centerY + sinAngle*halfWidth - cosAngle*halfHeight

    x3 := centerX - cosAngle*halfWidth + sinAngle*halfHeight
    y3 := centerY - sinAngle*halfWidth - cosAngle*halfHeight

    x4 := centerX - cosAngle*halfWidth - sinAngle*halfHeight
    y4 := centerY - sinAngle*halfWidth + cosAngle*halfHeight

    // Draw the rotated rectangle as two triangles
    ebitenutil.DrawLine(screen, x1, y1, x2, y2, clr)
    ebitenutil.DrawLine(screen, x2, y2, x3, y3, clr)
    ebitenutil.DrawLine(screen, x3, y3, x4, y4, clr)
    ebitenutil.DrawLine(screen, x4, y4, x1, y1, clr)
}

// drawRocket draws a rocket with effects
func drawRocket(screen *ebiten.Image, x, y float64, rocket *Rocket) {
    if rocket.Exploded {
        // Draw explosion
        explosionProgress := float64(rocket.ExplosionTimer) / float64(rocket.ExplosionTime)
        explosionRadius := rocket.BlastRadius * explosionProgress
        explosionAlpha := uint8(255 * (1 - explosionProgress))

        // Draw explosion as concentric circles with fading colors
        outerColor := color.RGBA{255, 100, 0, explosionAlpha} // Orange
        middleColor := color.RGBA{255, 200, 0, explosionAlpha} // Yellow
        innerColor := color.RGBA{255, 255, 255, explosionAlpha} // White

        // Draw outer circle
        drawCircle(screen, x + rocket.Width/2, y + rocket.Height/2, explosionRadius, outerColor)

        // Draw middle circle
        drawCircle(screen, x + rocket.Width/2, y + rocket.Height/2, explosionRadius * 0.7, middleColor)

        // Draw inner circle
        drawCircle(screen, x + rocket.Width/2, y + rocket.Height/2, explosionRadius * 0.3, innerColor)
    } else {
        // Draw rocket body
        rocketColor := rocket.Color
        ebitenutil.DrawRect(screen, x, y, rocket.Width, rocket.Height, rocketColor)

        // Calculate angle from velocity
        angle := math.Atan2(rocket.VelY, rocket.VelX)

        // Draw rocket nose cone
        noseLength := rocket.Width * 0.5
        noseX := x + rocket.Width
        noseY := y + rocket.Height/2

        // Calculate nose tip position based on angle
        noseTipX := noseX + math.Cos(angle) * noseLength
        noseTipY := noseY + math.Sin(angle) * noseLength

        // Draw nose cone as a triangle
        ebitenutil.DrawLine(screen, noseX, noseY - rocket.Height/2, noseTipX, noseTipY, rocketColor)
        ebitenutil.DrawLine(screen, noseTipX, noseTipY, noseX, noseY + rocket.Height/2, rocketColor)

        // Draw rocket fins
        finColor := color.RGBA{150, 150, 150, 255}
        finWidth := rocket.Width * 0.3
        finHeight := rocket.Height * 0.8

        // Draw top fin
        ebitenutil.DrawRect(screen, x, y - finHeight, finWidth, finHeight, finColor)

        // Draw bottom fin
        ebitenutil.DrawRect(screen, x, y + rocket.Height, finWidth, finHeight, finColor)

        // Draw rocket exhaust/trail
        trailLength := rocket.Width * 2

        // Calculate trail start position (back of rocket)
        trailStartX := x
        trailStartY := y + rocket.Height/2

        // Calculate trail end position based on opposite of rocket angle
        trailEndX := trailStartX - math.Cos(angle) * trailLength
        trailEndY := trailStartY - math.Sin(angle) * trailLength

        // Draw trail as a triangle
        trailColor1 := color.RGBA{255, 100, 0, 200} // Orange
        trailColor2 := color.RGBA{255, 200, 0, 150} // Yellow
        trailColor3 := color.RGBA{255, 255, 255, 100} // White

        // Draw three overlapping triangles for the trail effect
        ebitenutil.DrawLine(screen, trailStartX, trailStartY - rocket.Height/4, trailEndX, trailEndY, trailColor1)
        ebitenutil.DrawLine(screen, trailStartX, trailStartY + rocket.Height/4, trailEndX, trailEndY, trailColor1)

        // Draw inner trail
        innerTrailEndX := trailStartX - math.Cos(angle) * trailLength * 0.7
        innerTrailEndY := trailStartY - math.Sin(angle) * trailLength * 0.7
        ebitenutil.DrawLine(screen, trailStartX, trailStartY - rocket.Height/6, innerTrailEndX, innerTrailEndY, trailColor2)
        ebitenutil.DrawLine(screen, trailStartX, trailStartY + rocket.Height/6, innerTrailEndX, innerTrailEndY, trailColor2)

        // Draw core trail
        coreTrailEndX := trailStartX - math.Cos(angle) * trailLength * 0.4
        coreTrailEndY := trailStartY - math.Sin(angle) * trailLength * 0.4
        ebitenutil.DrawLine(screen, trailStartX, trailStartY - rocket.Height/10, coreTrailEndX, coreTrailEndY, trailColor3)
        ebitenutil.DrawLine(screen, trailStartX, trailStartY + rocket.Height/10, coreTrailEndX, coreTrailEndY, trailColor3)
    }
}

// drawGrenade draws a grenade with effects
func drawGrenade(screen *ebiten.Image, x, y float64, grenade *Grenade) {
    if grenade.Exploded {
        // Draw explosion
        explosionProgress := float64(grenade.ExplosionTimer) / float64(grenade.ExplosionTime)
        explosionRadius := grenade.BlastRadius * explosionProgress
        explosionAlpha := uint8(255 * (1 - explosionProgress))

        // Draw explosion as concentric circles with fading colors
        outerColor := color.RGBA{0, 150, 0, explosionAlpha} // Green
        middleColor := color.RGBA{100, 200, 0, explosionAlpha} // Light green
        innerColor := color.RGBA{255, 255, 255, explosionAlpha} // White

        // Draw outer circle
        drawCircle(screen, x + grenade.Width/2, y + grenade.Height/2, explosionRadius, outerColor)

        // Draw middle circle
        drawCircle(screen, x + grenade.Width/2, y + grenade.Height/2, explosionRadius * 0.7, middleColor)

        // Draw inner circle
        drawCircle(screen, x + grenade.Width/2, y + grenade.Height/2, explosionRadius * 0.3, innerColor)

        // Draw debris particles
        debrisCount := 20
        for i := 0; i < debrisCount; i++ {
            angle := float64(i) * (2 * math.Pi / float64(debrisCount))
            distance := explosionRadius * 0.8 * explosionProgress

            debrisX := x + grenade.Width/2 + math.Cos(angle) * distance
            debrisY := y + grenade.Height/2 + math.Sin(angle) * distance

            debrisSize := 2.0 + float64(i%3)
            debrisColor := color.RGBA{50, 100, 50, explosionAlpha}
            ebitenutil.DrawRect(screen, debrisX, debrisY, debrisSize, debrisSize, debrisColor)
        }
    } else {
        // Draw grenade body (circle)
        grenadeColor := grenade.Color
        drawCircle(screen, x + grenade.Width/2, y + grenade.Height/2, grenade.Width/2, grenadeColor)

        // Draw grenade details
        // Draw pin at the top
        pinColor := color.RGBA{200, 200, 200, 255} // Silver
        pinWidth := grenade.Width * 0.2
        pinHeight := grenade.Height * 0.3
        pinX := x + grenade.Width/2 - pinWidth/2
        pinY := y - pinHeight * 0.7

        // Draw pin
        ebitenutil.DrawRect(screen, pinX, pinY, pinWidth, pinHeight, pinColor)

        // Draw handle
        handleColor := pinColor
        handleWidth := grenade.Width * 0.4
        handleHeight := grenade.Height * 0.1
        handleX := x + grenade.Width/2
        handleY := y - handleHeight

        // Draw handle as a curved line
        ebitenutil.DrawLine(screen, handleX, handleY, handleX + handleWidth, handleY, handleColor)

        // Draw highlight on grenade for 3D effect
        highlightColor := color.RGBA{
            R: addColorValue(grenadeColor.R, 50),
            G: addColorValue(grenadeColor.G, 50),
            B: addColorValue(grenadeColor.B, 50),
            A: 200,
        }

        // Draw highlight as a smaller circle
        drawCircle(screen, x + grenade.Width*0.4, y + grenade.Height*0.4, grenade.Width*0.2, highlightColor)
    }
}

// drawCircle draws a filled circle
func drawCircle(screen *ebiten.Image, centerX, centerY, radius float64, clr color.RGBA) {
    // Draw a circle by drawing lines from the center to the edge
    segments := 32 // Number of segments to approximate the circle
    for i := 0; i < segments; i++ {
        angle1 := float64(i) * 2 * math.Pi / float64(segments)
        angle2 := float64(i+1) * 2 * math.Pi / float64(segments)

        x1 := centerX + math.Cos(angle1) * radius
        y1 := centerY + math.Sin(angle1) * radius

        x2 := centerX + math.Cos(angle2) * radius
        y2 := centerY + math.Sin(angle2) * radius

        // Draw a line from the center to the edge
        ebitenutil.DrawLine(screen, centerX, centerY, x1, y1, clr)
        ebitenutil.DrawLine(screen, centerX, centerY, x2, y2, clr)
        ebitenutil.DrawLine(screen, x1, y1, x2, y2, clr)
    }
}

// drawGrass draws animated grass blades on top of a platform
func drawGrass(screen *ebiten.Image, x, y float64, width float64, timer int) {
    // Base grass color
    grassColor := color.RGBA{34, 139, 34, 255} // Forest green

    // Lighter grass color for highlights
    lightGrassColor := color.RGBA{50, 205, 50, 255} // Lime green

    // Draw base grass layer
    ebitenutil.DrawRect(screen, x, y, width, 5, grassColor)

    // Draw animated grass blades
    bladeCount := int(width / 8) // One blade every 8 pixels
    bladeWidth := 2.0
    maxBladeHeight := 8.0

    for i := 0; i < bladeCount; i++ {
        // Calculate blade position
        bladeX := x + float64(i) * 8.0 + 2.0 // Distribute evenly with a small offset

        // Calculate wave offset based on timer and position
        // This creates a wave-like motion where different blades move at slightly different times
        waveOffset := float64(timer) / 20.0
        positionOffset := float64(i) / float64(bladeCount) * 6.28 // 2π
        waveAmount := math.Sin(waveOffset + positionOffset) * 3.0

        // Randomize blade height slightly based on position
        bladeHeight := maxBladeHeight - float64(i%3)

        // Choose color (alternate between regular and light green)
        bladeColor := grassColor
        if i%2 == 0 {
            bladeColor = lightGrassColor
        }

        // Draw the blade with a slight lean based on the wave
        ebitenutil.DrawRect(
            screen,
            bladeX + waveAmount, // Apply horizontal wave offset
            y - bladeHeight,
            bladeWidth,
            bladeHeight,
            bladeColor,
        )
    }
}

// drawPlatform draws a platform with texture
func drawPlatform(screen *ebiten.Image, x, y float64, platform *Platform) {
    // Main platform body
    ebitenutil.DrawRect(screen, x, y, platform.Width, platform.Height, platform.Color)

    // Add texture details to make it look more like grass/dirt
    if platform.Height >= 20 {
        // Draw animated grass on top
        drawGrass(screen, x, y, platform.Width, platform.GrassWaveTimer)

        // Draw dirt texture lines
        dirtColor := color.RGBA{101, 67, 33, 255} // Brown
        lineSpacing := 10.0

        for i := 0.0; i < platform.Width; i += lineSpacing {
            lineHeight := platform.Height - 5 // Exclude grass layer
            if lineHeight > 0 {
                ebitenutil.DrawRect(screen, x+i, y+5, 2, lineHeight, dirtColor)
            }
        }
    } else {
        // For thin platforms, just add some highlights
        highlightColor := color.RGBA{
            R: addColorValue(platform.Color.R, 50),
            G: addColorValue(platform.Color.G, 50),
            B: addColorValue(platform.Color.B, 50),
            A: platform.Color.A,
        }
        ebitenutil.DrawRect(screen, x, y, platform.Width, 2, highlightColor)
    }
}

// drawBullet draws a bullet with trail effect
func drawBullet(screen *ebiten.Image, x, y float64, bullet *Bullet) {
    // Calculate angle from velocity
    angle := math.Atan2(bullet.VelY, bullet.VelX)

    // Calculate center of bullet
    centerX := x + bullet.Width/2
    centerY := y + bullet.Height/2

    // Draw main bullet using rotated rectangle
    drawRotatedRectangle(screen, centerX, centerY, bullet.Width, bullet.Height, angle, bullet.Color)

    // Draw bullet trail
    trailLength := 5
    trailAlpha := uint8(200)
    trailStep := trailAlpha / uint8(trailLength)

    for i := 1; i <= trailLength; i++ {
        // Calculate trail position based on angle
        trailDistance := float64(i*2)
        trailX := centerX - math.Cos(angle) * trailDistance
        trailY := centerY - math.Sin(angle) * trailDistance

        trailColor := color.RGBA{
            bullet.Color.R,
            bullet.Color.G,
            bullet.Color.B,
            trailAlpha - (trailStep * uint8(i)),
        }

        trailWidth := bullet.Width / 2
        trailHeight := bullet.Height / 2

        // Draw trail segment as rotated rectangle
        drawRotatedRectangle(
            screen,
            trailX,
            trailY,
            trailWidth,
            trailHeight,
            angle,
            trailColor,
        )
    }
}

// drawEnemy draws an enemy with more details
func drawEnemy(screen *ebiten.Image, x, y float64, enemy *Enemy) {
    // Determine bean type based on color
    isKidneyBean := enemy.Color.R > enemy.Color.B // Kidney beans are red, navy beans are blue

    // Draw bean-shaped body
    beanWidth := enemy.Width * 1.2
    beanHeight := enemy.Height * 0.9
    beanX := x - (beanWidth - enemy.Width) / 2
    beanY := y + enemy.Height * 0.1

    // Draw main bean body with shape based on bean type
    if isKidneyBean {
        // Kidney bean - more curved, asymmetric shape
        for i := 0.0; i < beanWidth; i += 1.0 {
            // Calculate position along the width
            relativePos := i / beanWidth

            // Create kidney bean shape (more curved on one side)
            distFromCenter := math.Abs(i - beanWidth/2)

            // Make one side more curved for kidney bean shape
            asymmetry := 0.0
            if i < beanWidth/2 {
                asymmetry = (beanWidth/2 - i) * 0.04 // More curve on left side
            }

            heightReduction := (distFromCenter / (beanWidth/2)) * (beanWidth/8) + asymmetry
            columnHeight := beanHeight - heightReduction

            // Add slight indentation in middle (kidney bean characteristic)
            if relativePos > 0.4 && relativePos < 0.6 {
                columnHeight *= 0.95 // Slight indentation
            }

            // Draw with slight color variation for texture
            colorVariation := uint8(math.Sin(relativePos*math.Pi*4) * 10)
            pixelColor := color.RGBA{
                R: addColorValue(enemy.Color.R, colorVariation),
                G: addColorValue(enemy.Color.G, colorVariation/2),
                B: addColorValue(enemy.Color.B, colorVariation/2),
                A: enemy.Color.A,
            }

            ebitenutil.DrawRect(screen, beanX + i, beanY, 1, columnHeight, pixelColor)
        }
    } else {
        // Navy bean - more rounded, symmetric shape
        for i := 0.0; i < beanWidth; i += 1.0 {
            // Calculate position along the width
            relativePos := i / beanWidth

            // Create rounder shape for navy beans
            distFromCenter := math.Abs(i - beanWidth/2)

            // Use quadratic function for rounder shape
            heightReduction := math.Pow(distFromCenter / (beanWidth/2), 2) * (beanWidth/6)
            columnHeight := beanHeight - heightReduction

            // Add slight texture variation
            colorVariation := uint8(math.Sin(relativePos*math.Pi*6) * 8)
            pixelColor := color.RGBA{
                R: addColorValue(enemy.Color.R, colorVariation/2),
                G: addColorValue(enemy.Color.G, colorVariation/2),
                B: addColorValue(enemy.Color.B, colorVariation),
                A: enemy.Color.A,
            }

            ebitenutil.DrawRect(screen, beanX + i, beanY, 1, columnHeight, pixelColor)
        }
    }

    // Draw bean "crease" line with more natural curve
    creaseColor := color.RGBA{
        R: subtractColorValue(enemy.Color.R, 40),
        G: subtractColorValue(enemy.Color.G, 40),
        B: subtractColorValue(enemy.Color.B, 40),
        A: enemy.Color.A,
    }

    // Draw curved crease instead of straight line
    creaseWidth := beanWidth * 0.7
    creaseSegments := 10
    segmentWidth := creaseWidth / float64(creaseSegments)
    creaseX := beanX + (beanWidth - creaseWidth) / 2
    creaseBaseY := beanY + beanHeight * 0.4

    for i := 0; i < creaseSegments; i++ {
        segmentX := creaseX + float64(i) * segmentWidth

        // Create slight curve in the crease
        curveOffset := math.Sin(float64(i)/float64(creaseSegments-1)*math.Pi) * 3.0
        segmentY := creaseBaseY - curveOffset

        // Vary thickness slightly
        thickness := 1.5 + math.Sin(float64(i)/float64(creaseSegments-1)*math.Pi) * 1.0

        ebitenutil.DrawRect(screen, segmentX, segmentY, segmentWidth, thickness, creaseColor)
    }

    // Draw eyes (angry looking)
    eyeColor := color.RGBA{255, 255, 255, 255}
    eyeWidth := enemy.Width * 0.2
    eyeHeight := enemy.Height * 0.15
    eyeY := beanY + beanHeight * 0.25

    // Left eye
    leftEyeX := beanX + beanWidth * 0.25 - eyeWidth/2
    ebitenutil.DrawRect(screen, leftEyeX, eyeY, eyeWidth, eyeHeight, eyeColor)

    // Right eye
    rightEyeX := beanX + beanWidth * 0.75 - eyeWidth/2
    ebitenutil.DrawRect(screen, rightEyeX, eyeY, eyeWidth, eyeHeight, eyeColor)

    // Draw pupils (angry)
    pupilColor := color.RGBA{0, 0, 0, 255} // Black pupils
    pupilWidth := eyeWidth * 0.6
    pupilHeight := eyeHeight * 0.8

    // Direction of pupils based on movement
    pupilOffset := 0.0
    if enemy.MoveDir > 0 {
        pupilOffset = eyeWidth * 0.15
    } else {
        pupilOffset = -eyeWidth * 0.15
    }

    ebitenutil.DrawRect(screen, leftEyeX + (eyeWidth-pupilWidth)/2 + pupilOffset, eyeY + (eyeHeight-pupilHeight)/2, pupilWidth, pupilHeight, pupilColor)
    ebitenutil.DrawRect(screen, rightEyeX + (eyeWidth-pupilWidth)/2 + pupilOffset, eyeY + (eyeHeight-pupilHeight)/2, pupilWidth, pupilHeight, pupilColor)

    // Draw mouth (angry)
    mouthWidth := beanWidth * 0.4
    mouthHeight := beanHeight * 0.08
    mouthX := beanX + (beanWidth - mouthWidth) / 2
    mouthY := beanY + beanHeight * 0.6
    ebitenutil.DrawRect(screen, mouthX, mouthY, mouthWidth, mouthHeight, color.RGBA{0, 0, 0, 255})

    // Draw enemy health bar
    healthBarWidth := beanWidth
    healthBarHeight := 5.0
    healthBarY := beanY - healthBarHeight - 5

    // Draw health bar background (red)
    ebitenutil.DrawRect(screen, beanX, healthBarY, healthBarWidth, healthBarHeight, color.RGBA{255, 0, 0, 255})

    // Draw health bar foreground (green)
    healthWidth := healthBarWidth * float64(enemy.Health) / float64(enemy.MaxHealth)
    ebitenutil.DrawRect(screen, beanX, healthBarY, healthWidth, healthBarHeight, color.RGBA{0, 255, 0, 255})
}

// drawFlag draws the flag with improved animation
func drawFlag(screen *ebiten.Image, x, y float64, flag *Flag) {
    // Draw the flag pole
    poleWidth := 4.0
    poleColor := color.RGBA{139, 69, 19, 255} // Brown pole
    ebitenutil.DrawRect(screen, x, y, poleWidth, flag.Height, poleColor)

    // Draw pole details
    poleDetailColor := color.RGBA{101, 67, 33, 255} // Darker brown
    detailSpacing := 10.0
    for i := 0.0; i < flag.Height; i += detailSpacing {
        ebitenutil.DrawRect(screen, x, y+i, poleWidth, 2, poleDetailColor)
    }

    // Draw the flag with a wave animation
    flagWidth := flag.Width - poleWidth
    flagHeight := flag.Height / 2

    // Calculate wave offset based on timer for a more natural wave
    waveOffset := float64(flag.WaveTimer) / 10.0
    if waveOffset > 3.0 {
        waveOffset = 6.0 - waveOffset
    }

    // Draw the flag with wave effect and multiple segments for better wave appearance
    segments := 5
    segmentWidth := flagWidth / float64(segments)

    for i := 0; i < segments; i++ {
        // Each segment has a slightly different wave offset
        segmentWaveOffset := waveOffset * float64(i+1) / float64(segments)
        segmentX := x + poleWidth + float64(i)*segmentWidth
        segmentY := y + segmentWaveOffset

        ebitenutil.DrawRect(screen, segmentX, segmentY, segmentWidth, flagHeight, flag.Color)
    }

    // Draw flag details (stripes or pattern)
    stripeColor := color.RGBA{
        R: subtractColorValue(flag.Color.R, 50),
        G: subtractColorValue(flag.Color.G, 50),
        B: subtractColorValue(flag.Color.B, 50),
        A: flag.Color.A,
    }

    for i := 0; i < segments; i++ {
        if i % 2 == 0 {
            segmentWaveOffset := waveOffset * float64(i+1) / float64(segments)
            segmentX := x + poleWidth + float64(i)*segmentWidth
            segmentY := y + segmentWaveOffset + flagHeight/2

            ebitenutil.DrawRect(screen, segmentX, segmentY, segmentWidth, flagHeight/2, stripeColor)
        }
    }

    // If collected, add a celebration effect
    if flag.Collected {
        // Draw more elaborate celebration effects
        sparkleColors := []color.RGBA{
            {255, 255, 255, 255}, // White
            {255, 215, 0, 255},   // Gold
            {255, 255, 0, 255},   // Yellow
        }

        // Use frameCount to animate the sparkles
        sparkleSize := 5.0
        sparkleCount := 10

        for i := 0; i < sparkleCount; i++ {
            // Calculate position using a simple pattern
            radius := 20.0 + float64(i%3)*5.0

            // Use the angle to create circular movement
            offsetX := float64(i%2) * 2.0 * radius
            offsetY := float64(i/2) * radius

            // Add some animation based on the wave timer
            if i % 3 == 0 {
                offsetX += float64(flag.WaveTimer % 10) - 5.0
            }

            sparkleX := x + flag.Width/2 + offsetX - sparkleSize/2
            sparkleY := y + flag.Height/3 + offsetY - sparkleSize/2

            // Alternate colors
            sparkleColor := sparkleColors[i%len(sparkleColors)]

            // Draw sparkle
            ebitenutil.DrawRect(screen, sparkleX, sparkleY, sparkleSize, sparkleSize, sparkleColor)
        }
    }
}

// Helper functions
func min(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}

func max(a, b float64) float64 {
    if a > b {
        return a
    }
    return b
}

func minInt(a, b int) int {
    if a < b {
        return a
    }
    return b
}

// Color helper functions
func minUint8(a, b uint8) uint8 {
    if a < b {
        return a
    }
    return b
}

func maxUint8(a, b uint8) uint8 {
    if a > b {
        return a
    }
    return b
}

func addColorValue(base, add uint8) uint8 {
    // Prevent overflow
    if 255 - base < add {
        return 255
    }
    return base + add
}

func subtractColorValue(base, subtract uint8) uint8 {
    // Prevent underflow
    if base < subtract {
        return 0
    }
    return base - subtract
}

// Player represents the player character
type Player struct {
    X         float64
    Y         float64
    Width     float64
    Height    float64
    VelX      float64
    VelY      float64
    Speed     float64
    JumpForce float64
    OnGround  bool
    Dead      bool
    Color     color.RGBA
    SpawnX    float64
    SpawnY    float64
    // Gun properties
    GunAngle   float64 // Angle of the gun in radians
    // Machine gun properties
    BulletCount int    // Total ammo
    ClipSize    int    // Size of each clip
    CurrentClip int    // Bullets in current clip
    FireRate    int    // Frames between shots
    LastFired   int    // Frame count of last shot
    // Rocket launcher properties
    RocketCount   int  // Total rockets
    RocketClipSize int // Size of rocket clip (1)
    CurrentRocket int  // Current rocket in clip
    RocketFireRate int // Frames between rocket shots
    LastRocketFired int // Frame count of last rocket fired
    RocketReloadTime int // Total frames needed to reload rocket
    RocketReloadTimer int // Current rocket reload timer
    IsRocketReloading bool // Whether rocket launcher is reloading
    // Grenade properties
    GrenadeCount int  // Total grenades (limited to 5)
    LastGrenadeFired int // Frame count of last grenade thrown
    GrenadeFireRate int // Frames between grenade throws
    // Weapon state
    FacingRight bool   // Direction player is facing
    IsReloading bool   // Whether player is currently reloading
    ReloadTime  int    // Total frames needed to reload
    ReloadTimer int    // Current reload timer
    CurrentWeapon int  // 0 = machine gun, 1 = rocket launcher, 2 = grenades
    // Health and lives
    Health     int
    MaxHealth  int
    Lives      int
    InvincibleTime int // Frames of invincibility after taking damage
}

// NewPlayer creates a new player entity
func NewPlayer(x, y float64) *Player {
    return &Player{
        X:          x,
        Y:          y,
        Width:      32,
        Height:     64,
        VelX:       0,
        VelY:       0,
        Speed:      5,
        JumpForce:  -15,
        OnGround:   false,
        Dead:       false,
        Color:      color.RGBA{160, 120, 80, 255}, // Light brown (pinto bean)
        SpawnX:     x,
        SpawnY:     y,
        // Initialize gun properties
        GunAngle:    0,
        // Initialize machine gun properties
        BulletCount:  200,    // Total ammo
        ClipSize:     50,     // 50 bullets per clip
        CurrentClip:  50,     // Start with a full clip
        FireRate:     10,     // Frames between shots
        LastFired:    0,
        // Initialize rocket launcher properties
        RocketCount:   10,    // Total rockets
        RocketClipSize: 1,    // 1 rocket per clip
        CurrentRocket: 1,     // Start with a full clip
        RocketFireRate: 60,   // 1 second between rocket shots
        LastRocketFired: 0,
        RocketReloadTime: 180, // 3 seconds at 60 FPS
        RocketReloadTimer: 0,
        IsRocketReloading: false,
        // Initialize grenade properties
        GrenadeCount:   5,     // Total grenades (limited to 5)
        LastGrenadeFired: 0,
        GrenadeFireRate: 45,   // 0.75 seconds between grenade throws
        // Initialize weapon state
        FacingRight:  true,
        IsReloading:  false,
        ReloadTime:   120,    // 2 seconds at 60 FPS
        ReloadTimer:  0,
        CurrentWeapon: 0,     // Start with machine gun
        // Initialize health and lives
        Health:        100,
        MaxHealth:     100,
        Lives:         3,
        InvincibleTime: 0,
    }
}

// Update updates the player state
func (p *Player) Update() {
    // Reset ground state each frame
    p.OnGround = false

    // Update invincibility timer
    if p.InvincibleTime > 0 {
        p.InvincibleTime--

        // Flash the player color when invincible
        if p.InvincibleTime % 4 < 2 {
            p.Color = color.RGBA{160, 120, 80, 255} // Light brown (pinto bean)
        } else {
            p.Color = color.RGBA{255, 255, 255, 255} // White
        }
    } else {
        // Reset to normal color
        p.Color = color.RGBA{160, 120, 80, 255} // Light brown (pinto bean)
    }

    // Update machine gun reload timer
    if p.IsReloading {
        p.ReloadTimer--
        if p.ReloadTimer <= 0 {
            // Reload complete
            p.FinishReload()
        }
    }

    // Update rocket launcher reload timer
    if p.IsRocketReloading {
        p.RocketReloadTimer--
        if p.RocketReloadTimer <= 0 {
            // Rocket reload complete
            p.FinishRocketReload()
        }
    }
}

// StartReload begins the reload process for the machine gun
func (p *Player) StartReload() {
    // Don't reload if already reloading or if clip is full or if no ammo left
    if p.IsReloading || p.CurrentClip >= p.ClipSize || p.BulletCount <= 0 {
        return
    }

    p.IsReloading = true
    p.ReloadTimer = p.ReloadTime
}

// FinishReload completes the reload process for the machine gun
func (p *Player) FinishReload() {
    if !p.IsReloading {
        return
    }

    // Calculate how many bullets to add to the clip
    bulletsNeeded := p.ClipSize - p.CurrentClip
    bulletsToAdd := minInt(bulletsNeeded, p.BulletCount)

    // Add bullets to clip and remove from total
    p.CurrentClip += bulletsToAdd
    p.BulletCount -= bulletsToAdd

    // Reset reload state
    p.IsReloading = false
    p.ReloadTimer = 0
}

// StartRocketReload begins the reload process for the rocket launcher
func (p *Player) StartRocketReload() {
    // Don't reload if already reloading or if clip is full or if no ammo left
    if p.IsRocketReloading || p.CurrentRocket >= p.RocketClipSize || p.RocketCount <= 0 {
        return
    }

    p.IsRocketReloading = true
    p.RocketReloadTimer = p.RocketReloadTime
}

// FinishRocketReload completes the reload process for the rocket launcher
func (p *Player) FinishRocketReload() {
    if !p.IsRocketReloading {
        return
    }

    // Calculate how many rockets to add to the clip
    rocketsNeeded := p.RocketClipSize - p.CurrentRocket
    rocketsToAdd := minInt(rocketsNeeded, p.RocketCount)

    // Add rockets to clip and remove from total
    p.CurrentRocket += rocketsToAdd
    p.RocketCount -= rocketsToAdd

    // Reset reload state
    p.IsRocketReloading = false
    p.RocketReloadTimer = 0
}

// TakeDamage applies damage to the player
func (p *Player) TakeDamage(amount int) {
    // If player is invincible, don't take damage
    if p.InvincibleTime > 0 {
        return
    }

    // Apply damage
    p.Health -= amount

    // Check if player is dead
    if p.Health <= 0 {
        p.Health = 0
        p.Dead = true
    } else {
        // Make player invincible for a short time
        p.InvincibleTime = 60 // 1 second at 60 FPS
    }
}

// Draw is a placeholder to satisfy the Entity interface
// The actual drawing is handled by World.Draw
func (p *Player) Draw(screen *ebiten.Image) {
    // Drawing is handled by World.Draw
}

// Movement methods
func (p *Player) MoveLeft() {
    p.VelX = -p.Speed
}

func (p *Player) MoveRight() {
    p.VelX = p.Speed
}

func (p *Player) StopHorizontal() {
    p.VelX = 0
}

func (p *Player) Jump() {
    if p.OnGround {
        p.VelY = p.JumpForce
        p.OnGround = false
    }
}

// Reset resets the player to their spawn position after death
func (p *Player) Reset() {
    p.X = p.SpawnX
    p.Y = p.SpawnY
    p.VelX = 0
    p.VelY = 0
    p.OnGround = false
    p.Dead = false

    // Decrement lives if player died
    if p.Lives > 0 {
        p.Lives--
    }

    // Reset health
    p.Health = p.MaxHealth

    // If player has no lives left, game over
    if p.Lives <= 0 {
        // Game over logic would go here
        // For now, just give the player 3 more lives and reset bullet count
        p.Lives = 3
        p.BulletCount = 200
    }

    // Reset clip system for machine gun
    p.CurrentClip = minInt(p.ClipSize, p.BulletCount)
    p.BulletCount -= p.CurrentClip
    p.IsReloading = false
    p.ReloadTimer = 0

    // Reset clip system for rocket launcher
    p.CurrentRocket = minInt(p.RocketClipSize, p.RocketCount)
    p.RocketCount -= p.CurrentRocket
    p.IsRocketReloading = false
    p.RocketReloadTimer = 0
}

// Entity interface implementation
func (p *Player) GetX() float64      { return p.X }
func (p *Player) GetY() float64      { return p.Y }
func (p *Player) GetWidth() float64  { return p.Width }
func (p *Player) GetHeight() float64 { return p.Height }
func (p *Player) SetX(x float64)     { p.X = x }
func (p *Player) SetY(y float64)     { p.Y = y }
func (p *Player) SetVelX(vx float64) { p.VelX = vx }
func (p *Player) SetVelY(vy float64) { p.VelY = vy }
func (p *Player) GetVelX() float64   { return p.VelX }
func (p *Player) GetVelY() float64   { return p.VelY }
func (p *Player) IsCollidable() bool { return true }

// Bullet represents a projectile fired by the player
type Bullet struct {
    X         float64
    Y         float64
    Width     float64
    Height    float64
    VelX      float64
    VelY      float64
    Active    bool
    Color     color.RGBA
    Lifetime  int
    Damage    int    // Damage dealt to enemies
}

// Rocket represents a rocket projectile fired by the player
type Rocket struct {
    X         float64
    Y         float64
    Width     float64
    Height    float64
    VelX      float64
    VelY      float64
    Active    bool
    Color     color.RGBA
    Lifetime  int
    Damage    int    // Damage dealt to enemies
    Exploded  bool   // Whether the rocket has exploded
    BlastRadius float64 // Explosion radius
    ExplosionTime int // Time the explosion lasts
    ExplosionTimer int // Current explosion timer
}

// Grenade represents a grenade projectile fired by the player
type Grenade struct {
    X         float64
    Y         float64
    Width     float64
    Height    float64
    VelX      float64
    VelY      float64
    Active    bool
    Color     color.RGBA
    Lifetime  int
    Damage    int    // Damage dealt to enemies
    Exploded  bool   // Whether the grenade has exploded
    BlastRadius float64 // Explosion radius
    ExplosionTime int // Time the explosion lasts
    ExplosionTimer int // Current explosion timer
    Bounces   int    // Number of bounces before exploding
}

// NewRocket creates a new rocket entity
func NewRocket(x, y, velX, velY float64) *Rocket {
    return &Rocket{
        X:        x,
        Y:        y,
        Width:    12,
        Height:   6,
        VelX:     velX,
        VelY:     velY,
        Active:   true,
        Color:    color.RGBA{255, 100, 0, 255}, // Orange
        Lifetime: 180, // Frames before rocket disappears
        Damage:   75,  // Damage dealt to enemies
        Exploded: false,
        BlastRadius: 100.0, // Explosion radius
        ExplosionTime: 30, // Explosion lasts for 0.5 seconds
        ExplosionTimer: 0,
    }
}

// Update updates the rocket state
func (r *Rocket) Update() {
    // If exploded, update explosion timer
    if r.Exploded {
        r.ExplosionTimer++
        if r.ExplosionTimer >= r.ExplosionTime {
            r.Active = false
        }
        return
    }

    // Update lifetime
    r.Lifetime--
    if r.Lifetime <= 0 {
        r.Explode()
    }
}

// Explode triggers the rocket explosion
func (r *Rocket) Explode() {
    r.Exploded = true
    r.ExplosionTimer = 0
    // Stop the rocket's movement
    r.VelX = 0
    r.VelY = 0
}

// Draw is a placeholder to satisfy the Entity interface
// The actual drawing is handled by World.Draw
func (r *Rocket) Draw(screen *ebiten.Image) {
    // Drawing is handled by World.Draw
}

// Entity interface implementation
func (r *Rocket) GetX() float64      { return r.X }
func (r *Rocket) GetY() float64      { return r.Y }
func (r *Rocket) GetWidth() float64  { return r.Width }
func (r *Rocket) GetHeight() float64 { return r.Height }
func (r *Rocket) SetX(x float64)     { r.X = x }
func (r *Rocket) SetY(y float64)     { r.Y = y }
func (r *Rocket) SetVelX(vx float64) { r.VelX = vx }
func (r *Rocket) SetVelY(vy float64) { r.VelY = vy }
func (r *Rocket) GetVelX() float64   { return r.VelX }
func (r *Rocket) GetVelY() float64   { return r.VelY }
func (r *Rocket) IsCollidable() bool { return r.Active && !r.Exploded }

// NewGrenade creates a new grenade entity
func NewGrenade(x, y, velX, velY float64) *Grenade {
    return &Grenade{
        X:        x,
        Y:        y,
        Width:    10,
        Height:   10,
        VelX:     velX,
        VelY:     velY,
        Active:   true,
        Color:    color.RGBA{0, 100, 0, 255}, // Dark green
        Lifetime: 180, // Frames before grenade explodes if not bounced
        Damage:   100, // Damage dealt to enemies
        Exploded: false,
        BlastRadius: 120.0, // Explosion radius (larger than rocket)
        ExplosionTime: 30, // Explosion lasts for 0.5 seconds
        ExplosionTimer: 0,
        Bounces:  3, // Explode after 3 bounces
    }
}

// Update updates the grenade state
func (g *Grenade) Update() {
    // If exploded, update explosion timer
    if g.Exploded {
        g.ExplosionTimer++
        if g.ExplosionTimer >= g.ExplosionTime {
            g.Active = false
        }
        return
    }

    // Update lifetime
    g.Lifetime--
    if g.Lifetime <= 0 {
        g.Explode()
    }
}

// Explode triggers the grenade explosion
func (g *Grenade) Explode() {
    g.Exploded = true
    g.ExplosionTimer = 0
    // Stop the grenade's movement
    g.VelX = 0
    g.VelY = 0
}

// HandleBounce decreases bounce counter and explodes if no bounces left
func (g *Grenade) HandleBounce() {
    g.Bounces--
    if g.Bounces <= 0 {
        g.Explode()
    }
}

// Draw is a placeholder to satisfy the Entity interface
// The actual drawing is handled by World.Draw
func (g *Grenade) Draw(screen *ebiten.Image) {
    // Drawing is handled by World.Draw
}

// Entity interface implementation
func (g *Grenade) GetX() float64      { return g.X }
func (g *Grenade) GetY() float64      { return g.Y }
func (g *Grenade) GetWidth() float64  { return g.Width }
func (g *Grenade) GetHeight() float64 { return g.Height }
func (g *Grenade) SetX(x float64)     { g.X = x }
func (g *Grenade) SetY(y float64)     { g.Y = y }
func (g *Grenade) SetVelX(vx float64) { g.VelX = vx }
func (g *Grenade) SetVelY(vy float64) { g.VelY = vy }
func (g *Grenade) GetVelX() float64   { return g.VelX }
func (g *Grenade) GetVelY() float64   { return g.VelY }
func (g *Grenade) IsCollidable() bool { return g.Active && !g.Exploded }

// Enemy represents an enemy character
type Enemy struct {
    X         float64
    Y         float64
    Width     float64
    Height    float64
    VelX      float64
    VelY      float64
    Speed     float64
    Health    int
    MaxHealth int
    Dead      bool
    Color     color.RGBA
    MoveDir   int     // Direction of movement: -1 left, 1 right
    MoveTimer int     // Timer for changing direction
    Active    bool    // Whether the enemy is active
    OnGround  bool    // Whether the enemy is on the ground
}

// NewBullet creates a new bullet entity
func NewBullet(x, y, velX, velY float64) *Bullet {
    return &Bullet{
        X:        x,
        Y:        y,
        Width:    8,
        Height:   4,
        VelX:     velX,
        VelY:     velY,
        Active:   true,
        Color:    color.RGBA{255, 255, 0, 255}, // Yellow
        Lifetime: 120, // Frames before bullet disappears
        Damage:   25,  // Damage dealt to enemies
    }
}

// NewEnemy creates a new enemy entity
func NewEnemy(x, y float64) *Enemy {
    return &Enemy{
        X:         x,
        Y:         y,
        Width:     32,
        Height:    48,
        VelX:      0,
        VelY:      0,
        Speed:     2,
        Health:    50,
        MaxHealth: 50,
        Dead:      false,
        Color:     color.RGBA{139, 69, 19, 255}, // Brown (default bean color)
        MoveDir:   1,  // Start moving right
        MoveTimer: 120, // Change direction every 2 seconds
        Active:    true,
        OnGround:  false,
    }
}

// Update updates the bullet state
func (b *Bullet) Update() {
    // Update lifetime
    b.Lifetime--
    if b.Lifetime <= 0 {
        b.Active = false
    }
}

// Draw is a placeholder to satisfy the Entity interface
// The actual drawing is handled by World.Draw
func (b *Bullet) Draw(screen *ebiten.Image) {
    // Drawing is handled by World.Draw
}

// Entity interface implementation
func (b *Bullet) GetX() float64      { return b.X }
func (b *Bullet) GetY() float64      { return b.Y }
func (b *Bullet) GetWidth() float64  { return b.Width }
func (b *Bullet) GetHeight() float64 { return b.Height }
func (b *Bullet) SetX(x float64)     { b.X = x }
func (b *Bullet) SetY(y float64)     { b.Y = y }
func (b *Bullet) SetVelX(vx float64) { b.VelX = vx }
func (b *Bullet) SetVelY(vy float64) { b.VelY = vy }
func (b *Bullet) GetVelX() float64   { return b.VelX }
func (b *Bullet) GetVelY() float64   { return b.VelY }
func (b *Bullet) IsCollidable() bool { return true }

// Update updates the enemy state
func (e *Enemy) Update() {
    // Skip if enemy is dead
    if e.Dead || !e.Active {
        return
    }

    // Reset ground state each frame
    e.OnGround = false

    // Move in current direction
    e.VelX = float64(e.MoveDir) * e.Speed
}

// TakeDamage applies damage to the enemy
func (e *Enemy) TakeDamage(amount int) {
    // Skip if enemy is already dead
    if e.Dead {
        return
    }

    // Apply damage
    e.Health -= amount

    // Check if enemy is dead
    if e.Health <= 0 {
        e.Health = 0
        e.Dead = true
        e.Active = false
    }
}

// Draw is a placeholder to satisfy the Entity interface
// The actual drawing is handled by World.Draw
func (e *Enemy) Draw(screen *ebiten.Image) {
    // Drawing is handled by World.Draw
}

// Entity interface implementation
func (e *Enemy) GetX() float64      { return e.X }
func (e *Enemy) GetY() float64      { return e.Y }
func (e *Enemy) GetWidth() float64  { return e.Width }
func (e *Enemy) GetHeight() float64 { return e.Height }
func (e *Enemy) SetX(x float64)     { e.X = x }
func (e *Enemy) SetY(y float64)     { e.Y = y }
func (e *Enemy) SetVelX(vx float64) { e.VelX = vx }
func (e *Enemy) SetVelY(vy float64) { e.VelY = vy }
func (e *Enemy) GetVelX() float64   { return e.VelX }
func (e *Enemy) GetVelY() float64   { return e.VelY }
func (e *Enemy) IsCollidable() bool { return e.Active }

// Flag represents the level finish flag
type Flag struct {
    X         float64
    Y         float64
    Width     float64
    Height    float64
    Color     color.RGBA
    Active    bool
    Collected bool
    WaveTimer int // Timer for waving animation
}

// NewFlag creates a new flag entity
func NewFlag(x, y float64) *Flag {
    return &Flag{
        X:         x,
        Y:         y,
        Width:     32,
        Height:    64,
        Color:     color.RGBA{255, 215, 0, 255}, // Gold color
        Active:    true,
        Collected: false,
        WaveTimer: 0,
    }
}

// Update updates the flag state (animation)
func (f *Flag) Update() {
    // Simple animation - increment timer
    f.WaveTimer++
    if f.WaveTimer > 60 {
        f.WaveTimer = 0
    }
}

// Draw is a placeholder to satisfy the Entity interface
// The actual drawing is handled by World.Draw
func (f *Flag) Draw(screen *ebiten.Image) {
    // Drawing is handled by World.Draw
}

// Entity interface implementation
func (f *Flag) GetX() float64      { return f.X }
func (f *Flag) GetY() float64      { return f.Y }
func (f *Flag) GetWidth() float64  { return f.Width }
func (f *Flag) GetHeight() float64 { return f.Height }
func (f *Flag) SetX(x float64)     { f.X = x }
func (f *Flag) SetY(y float64)     { f.Y = y }
func (f *Flag) SetVelX(vx float64) { /* Flags don't move */ }
func (f *Flag) SetVelY(vy float64) { /* Flags don't move */ }
func (f *Flag) GetVelX() float64   { return 0 }
func (f *Flag) GetVelY() float64   { return 0 }
func (f *Flag) IsCollidable() bool { return f.Active && !f.Collected }

// BloodParticle represents a blood particle effect
type BloodParticle struct {
    X         float64
    Y         float64
    VelX      float64
    VelY      float64
    Radius    float64
    Color     color.RGBA
    Lifetime  int
    Active    bool
}

// Platform represents a solid platform
type Platform struct {
    X              float64
    Y              float64
    Width          float64
    Height         float64
    Color          color.RGBA
    GrassWaveTimer int     // Timer for grass waving animation
}

// NewPlatform creates a new platform entity
func NewPlatform(x, y, width, height float64) *Platform {
    return &Platform{
        X:              x,
        Y:              y,
        Width:          width,
        Height:         height,
        Color:          color.RGBA{0, 255, 0, 255}, // Green
        GrassWaveTimer: 0,
    }
}

// Update updates the platform state
func (p *Platform) Update() {
    // Update grass wave timer
    p.GrassWaveTimer++
    if p.GrassWaveTimer > 120 { // Reset after 2 seconds (assuming 60 FPS)
        p.GrassWaveTimer = 0
    }
}

// Draw is a placeholder to satisfy the Entity interface
// The actual drawing is handled by World.Draw
func (p *Platform) Draw(screen *ebiten.Image) {
    // Drawing is handled by World.Draw
}

// Entity interface implementation
func (p *Platform) GetX() float64      { return p.X }
func (p *Platform) GetY() float64      { return p.Y }
func (p *Platform) GetWidth() float64  { return p.Width }
func (p *Platform) GetHeight() float64 { return p.Height }
func (p *Platform) SetX(x float64)     { p.X = x }
func (p *Platform) SetY(y float64)     { p.Y = y }
func (p *Platform) SetVelX(vx float64) { /* Platforms don't move */ }
func (p *Platform) SetVelY(vy float64) { /* Platforms don't move */ }
func (p *Platform) GetVelX() float64   { return 0 }
func (p *Platform) GetVelY() float64   { return 0 }
func (p *Platform) IsCollidable() bool { return true }
