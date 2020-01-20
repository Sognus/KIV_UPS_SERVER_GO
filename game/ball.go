package game

import (
	"errors"
	"math"
	"math/rand"
)

type Ball struct {
	X float64		// X-axis coordination
	Y float64		// Y-axis coordination
	Rotation int 	// Degrees of rotation 0-359
	Speed int		// Speed of ball in pixels per Tick
	MaxSpeed int	// Maximum speed of ball in pixels per tick
	Size int		// Ball diameter (circle diameter)
}

func UpdateBall(server *GameServer) error {
	if server == nil {
		return errors.New("unable to update ball - game server cannot be null")
	}

	if server.Ball == nil {
		return errors.New("unable to update ball - ball cannot be null")
	}
	radians := float64(server.Ball.Rotation) * (math.Pi / 180)
	velocity_x := math.Cos(radians)
	velocity_y := math.Sin(radians)

	server.Ball.X += velocity_x * float64(server.Ball.Speed)
	server.Ball.Y += velocity_y * float64(server.Ball.Speed)

	// Bounce right wall
	if server.Ball.X >= float64(server.WIDTH) {
		server.Ball.Rotation = int(math.Mod(float64(180 - server.Ball.Rotation), 360))
	}

	// Bounce left wall
	if server.Ball.X <= float64(0) {
		server.Ball.Rotation = int(math.Mod(float64(180 - server.Ball.Rotation), 360))
	}

	// Bounce top wall
	if server.Player1 != nil {
		if server.Ball.Rotation > 180 && server.Ball.Rotation <= 359 && server.Ball.Y <= server.Player1.y + float64(server.Ball.Size) && server.Ball.Y >= server.Player1.y - float64(server.Ball.Size / 2) && server.Ball.X >= server.Player1.x - server.Player1.width / 2 && server.Ball.X <= server.Player1.x + server.Player1.width / 2 {
			server.Ball.Rotation = int(math.Mod(float64(360 - server.Ball.Rotation), 360))
			server.Ball.Speed += 1
			if server.Ball.Speed >= server.Ball.MaxSpeed {
				server.Ball.Speed = server.Ball.MaxSpeed
			}
		}
	} else {
		// There is no player - bounce by wall
		if server.Ball.Y <= float64(0) {
			server.Ball.Rotation = int(math.Mod(float64(360 - server.Ball.Rotation), 360))
		}
	}

	// Bounce bottom wall
	if server.Ball.Y >= float64(server.HEIGHT) {
		server.Ball.Rotation = int(math.Mod(float64(360 - server.Ball.Rotation), 360))
	}

	// Check if ball is out of bounds
	if server.Ball.Y <= -10 {
		// Reset ball
		server.Ball.X = float64(server.WIDTH / 2)
		server.Ball.Y = float64(server.HEIGHT /2)

		randMax := 359
		randMin := 0

		// Generate new random rotation while we get cardinal directions
		for server.Ball.Rotation == 0 || server.Ball.Rotation == 90 || server.Ball.Rotation == 180 || server.Ball.Rotation == 270 {
			server.Ball.Rotation = rand.Intn((randMax - randMin) + randMin)
		}

		// Add score to player2
		server.Score2++

		// Reset ball speed
		server.Ball.Speed = 3
	}

	// Check if ball is out of bounds
	if server.Ball.Y >= float64(server.HEIGHT) + 10 {
		// Reset ball
		server.Ball.X = float64(server.WIDTH / 2)
		server.Ball.Y = float64(server.HEIGHT /2)

		randMax := 359
		randMin := 0

		// Generate new random rotation while we get cardinal directions
		for server.Ball.Rotation == 0 || server.Ball.Rotation == 90 || server.Ball.Rotation == 180 || server.Ball.Rotation == 270 {
			server.Ball.Rotation = rand.Intn((randMax - randMin) + randMin)
		}

		// Add score to player2
		server.Score1++

		// Reset ball speed
		server.Ball.Speed = 3
	}

	return nil
}