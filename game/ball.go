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

	fmt.Printf("new coords: %f, %f\n", ball.X, ball.Y)

	if ball.X >= float64(server.WIDTH) {
		// Bounce by right wall
		if ball.Rotation > 0 && ball.Rotation < 90 {
			// Right bounce down
			ball.Rotation = 90 + ball.Rotation
		}

		if ball.Rotation > 270 && ball.Rotation <= 359 {
			// Right bounce up
			ball.Rotation = 270 - (ball.Rotation - 270)
		}
	}

	if ball.X <= float64(0) {
		// Bounce by left wall
		if ball.Rotation > 90 && ball.Rotation < 180 {
			// Left bounce down
			ball.Rotation = 90 - (ball.Rotation - 90)
		}

		if ball.Rotation > 180 && ball.Rotation < 270 {
			// Left bounce up
			ball.Rotation = 270 + (ball.Rotation - 180)
		}
	}


	return nil
}