# Beanie Battles

A 2D platformer game where you play as a bean character battling against enemy beans.

## Game Controls

- **WASD Keys**: Move left/right (A/D) and jump (W)
- **Mouse Movement**: Aim weapons
- **Left Mouse Button**: Fire current weapon
- **Right Mouse Button**: Toggle zoom for precision aiming
- **1**: Switch to machine gun
- **2**: Switch to rocket launcher
- **3**: Switch to grenades
- **R**: Reload current weapon (machine gun or rocket launcher)
- **F**: Refill ammo (debug feature)
- **D**: Toggle debug mode
- **E**: Toggle level editor mode
- **Enter**: Continue to next level (when level is complete)

## Weapons

### Machine Gun
- Fast firing rate
- Limited clip size, requires reloading
- Moderate damage

### Rocket Launcher
- Slow firing rate
- Powerful explosions
- Limited ammo, requires reloading
- Rockets are aimed with the mouse cursor

### Grenades
- Bounces off surfaces
- Explodes after 3 bounces or on direct enemy hit
- Limited to 5 grenades total
- Grenades are aimed with the mouse cursor
- Does not require reloading

## Level Editor

The game includes a built-in level editor that allows you to create and modify levels. To access the level editor, simply press **E** at any time during gameplay.

### Editor Controls

- **1**: Select platform tool
- **2**: Select enemy spawn tool
- **3**: Select flag tool
- **4**: Select player start tool
- **Arrow Keys**: Adjust platform size (when platform tool is selected)
- **K/N**: Select enemy type (kidney/navy) when enemy tool is selected
- **Left Mouse Button**: Place selected entity
- **Delete**: Delete entity under cursor
- **S**: Save current level
- **N**: Create new level
- **E**: Exit editor mode (saves the level)
- **A/D**: Move camera left/right

### Creating a Level

1. Enter the editor mode
2. Use the platfoarm tool (1) to create platforms
3. Use the enemy spawn tool (2) to place enemy spawn points
4. Use the flag tool (3) to place the level's finish flag
5. Use the player start tool (4) to set the player's starting position
6. Press **S** to save the level
7. Press **E** to exit the editor and test your level

## Features

- Side-scrolling platformer gameplay
- Multiple enemy types
- Physics-based projectiles with reduced gravity
- Weapon switching system
- Health and ammo management
- Jump on enemies to defeat them with blood effects
- Bounce off enemies when defeating them by jumping
- Level editor for creating custom levels
- Multiple levels with progression
