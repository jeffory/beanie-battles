package main

import (
    "log"

    "github.com/hajimehoshi/ebiten/v2"
    "beanie-battles/engine"
)

const (
    screenWidth  = 800
    screenHeight = 600
    title        = "2D Platformer Engine"
)

func main() {
    log.Println("Starting 2D Platformer Engine...")

    // Create a new game instance
    log.Println("Creating game instance...")
    game, err := engine.NewGame(screenWidth, screenHeight)
    if err != nil {
        log.Fatal("Failed to create game:", err)
    }
    log.Println("Game instance created successfully")

    // Set up the window
    log.Println("Setting up window...")
    ebiten.SetWindowSize(screenWidth, screenHeight)
    ebiten.SetWindowTitle(title)
    ebiten.SetWindowResizable(true)
    log.Println("Window setup complete")

    // Run the game
    log.Println("Running game...")
    if err := ebiten.RunGame(game); err != nil {
        log.Fatal("Game crashed:", err)
    }
    log.Println("Game exited")
}
