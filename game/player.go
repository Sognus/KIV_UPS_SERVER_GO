package game

import (
	"../communication"
)

type Player struct {
	client *communication.Client
	ID int
	userName string
	lastCommunication int64
}
