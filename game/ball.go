package game

import (
	"errors"
	"fmt"
	"math"
)

type Ball struct {
	X float64		// X-axis coordination
	Y float64		// Y-axis coordination
	Rotation int 	// Degrees of rotation 0-359
	Speed int		// Speed of ball in pixels per Tick
	MaxSpeed int	// Maximum speed of ball in pixels per tick
}

func UpdateBall(server *GameServer) error {
	if server == nil {
		return errors.New("unable to update ball - game server cannot be null")
	}

	if server.Ball == nil {
		return errors.New("unable to update ball - ball cannot be null")
	}
	ball := server.Ball
	radians := float64(ball.Rotation) * (math.Pi / 180)
	velocity_x := math.Cos(radians)
	velocity_y := math.Sin(radians)

	ball.X += velocity_x * float64(ball.Speed)
	ball.Y += velocity_y * float64(ball.Speed)

	// Bounce right wall
	if ball.X >= float64(server.WIDTH) {
		ball.Rotation = int(math.Mod(float64(180 - ball.Rotation), 360))
	}

	// Bounce left wall
	if ball.X <= float64(0) {
		ball.Rotation = int(math.Mod(float64(180 - ball.Rotation), 360))
	}

	// Bounce top wall
	if ball.Y <= float64(0) {
		ball.Rotation = int(math.Mod(float64(360 - ball.Rotation), 360))
	}

	// Bounce bottom wall
	if ball.Y >= float64(server.HEIGHT) {
		ball.Rotation = int(math.Mod(float64(360 - ball.Rotation), 360))
	}

	fmt.Printf("new coords: %f, %f, rotation: %d\n", ball.X, ball.Y, ball.Rotation)


	return nil
}