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
	Size int		// Ball diameter (circle radius) (eg. 1/2 of diameter)
}

func isRotatedLeft(rotation int) bool {
	return rotation > 90 && rotation < 270
}

func isRotateRight(rotation int) bool {
	return rotation > 270 && rotation <= 359 || rotation > 0 && rotation < 90
}

// Returns if given value is in range given by base, left transform and right transform
func InRangeAround(value float64 ,base float64, left float64, right float64) bool {
	left_limit := base - left
	right_limit := base + right

	if value >= left_limit && value <= right_limit {
		return true
	}
	return false
}

func UpdateBall(server *GameServer) error {
	if server == nil {
		return errors.New("unable to update ball - game server cannot be null")
	}

	if server.Ball == nil {
		return errors.New("unable to update ball - ball cannot be null")
	}
	radians := float64(server.Ball.Rotation) * (math.Pi / 180)
	velocityX := math.Cos(radians)
	velocityY := math.Sin(radians)


	if velocityX > 0 {
		// Move right
		velocityRight := math.Abs(velocityX * float64(server.Ball.Speed))
		if server.Ball.X + velocityRight + float64(server.Ball.Size) > float64(server.WIDTH) {
			// Stop at boundary
			server.Ball.X = float64(server.WIDTH) - float64(server.Ball.Size)
			// Bounce
			server.Ball.Rotation = int(math.Mod(float64(180 - server.Ball.Rotation), 360))
		} else {
			// Move normally
			server.Ball.X += velocityRight
		}
	} else {
		velocityLeft := math.Abs(velocityX * float64(server.Ball.Speed))
		if server.Ball.X - velocityLeft - float64(server.Ball.Size) < 0 {
			// Stop at boundary
			server.Ball.X = 0 + float64(server.Ball.Size)
			// Bounce
			server.Ball.Rotation = int(math.Mod(float64(180 - server.Ball.Rotation), 360))
		} else {
			// Move normally
			server.Ball.X -= velocityLeft
		}
	}

	if velocityY > 0 {
		// Move down
		velocityDown := math.Abs(velocityY * float64(server.Ball.Speed))
		if server.Player2 != nil {
			// Player2 joined - check his boundary
			if InRangeAround(server.Ball.X, server.Player2.x, float64(server.Player2.width) / 2 + float64(server.Ball.Size), float64(server.Player2.width) / 2 + float64(server.Ball.Size)) {
				// Ball is in range of player - cautious move
				if server.Ball.Y + velocityDown + float64(server.Ball.Size) > server.Player2.y {
					// Stop at players location
					server.Ball.Y = server.Player2.y - float64(server.Ball.Size)
					// Bounce
					server.Ball.Rotation = int(math.Mod(float64(360 - server.Ball.Rotation), 360))
					// Add speed
					server.Ball.Speed++
					// Check max speed
					if server.Ball.Speed > server.Ball.MaxSpeed {
						server.Ball.Speed = server.Ball.MaxSpeed
					}
				} else {
					// Move normally
					server.Ball.Y += velocityDown
				}
			} else {
				// Ball is not in range of player - normal move
				server.Ball.Y += velocityDown
			}
		} else {
			// Player2 did not join - bounce by bottom wall
			server.Ball.Y += velocityDown
			if server.Ball.Y >= float64(server.HEIGHT) {
				server.Ball.Rotation = int(math.Mod(float64(360 - server.Ball.Rotation), 360))
			}
		}
	} else {
		// Move up
		velocityUp := math.Abs(velocityY * float64(server.Ball.Speed))
		if server.Player1 != nil {
			// Player1 joined - check his boundary
			if InRangeAround(server.Ball.X, server.Player1.x, float64(server.Player1.width) / 2 + float64(server.Ball.Size), float64(server.Player1.width) / 2 + float64(server.Ball.Size)) {
				// Ball is in range of player - cautious move
				if server.Ball.Y - velocityUp - float64(server.Ball.Size) < server.Player1.y {
					// Stop at players location
					server.Ball.Y = server.Player1.y + float64(server.Ball.Size)
					// Bounce
					server.Ball.Rotation = int(math.Mod(float64(360 - server.Ball.Rotation), 360))
					// Add speed
					server.Ball.Speed++
					// Check max speed
					if server.Ball.Speed > server.Ball.MaxSpeed {
						server.Ball.Speed = server.Ball.MaxSpeed
					}
				} else {
					// Move normally
					server.Ball.Y -= velocityUp
				}
			} else {
				// Ball is not in range of player - normal move
				server.Ball.Y -= velocityUp
			}
		} else {
			// Player 1 did not join - bounce by top wall
			server.Ball.Y -= velocityUp
			if server.Ball.Y <= float64(0) {
				server.Ball.Rotation = int(math.Mod(float64(360 - server.Ball.Rotation), 360))
			}
		}
	}

	if server.Ball.Y <= -1 * float64(server.Ball.Size){
		// Top player missed ball - add point to player2
		server.Score2++

		// Reset ball
		server.Ball.X = float64(server.WIDTH / 2)
		server.Ball.Y = float64(server.HEIGHT /2)
		server.Ball.Speed = 3

		randMax := 359
		randMin := 0

		// Generate new random rotation while we get cardinal directions
		for server.Ball.Rotation == 0 || server.Ball.Rotation == 90 || server.Ball.Rotation == 180 || server.Ball.Rotation == 270 {
			server.Ball.Rotation = rand.Intn((randMax - randMin) + randMin)
		}
	}

	if server.Ball.Y >= float64(server.HEIGHT) + float64(server.Ball.Size) {
		// Bottom player missed ball - add point to player1
		server.Score1++

		// Reset ball
		server.Ball.X = float64(server.WIDTH / 2)
		server.Ball.Y = float64(server.HEIGHT /2)
		server.Ball.Speed = 3

		randMax := 359
		randMin := 0

		// Generate new random rotation while we get cardinal directions
		for server.Ball.Rotation == 0 || server.Ball.Rotation == 90 || server.Ball.Rotation == 180 || server.Ball.Rotation == 270 {
			server.Ball.Rotation = rand.Intn((randMax - randMin) + randMin)
		}
	}

	return nil
}