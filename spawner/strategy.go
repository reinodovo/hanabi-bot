package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

var lock = &sync.Mutex{}

type SpawnerStrategy interface {
	Spawn(tableId int, tablePassword string) error
}

type LocalSpawnerStrategy struct {
	usedBots map[int]bool
}

func NewLocalSpawnerStrategy() *LocalSpawnerStrategy {
	return &LocalSpawnerStrategy{
		usedBots: make(map[int]bool),
	}
}

func (s *LocalSpawnerStrategy) nextAvailableBotIndex() int {
	lock.Lock()
	botIndex := -1
	for i, _ := range s.usedBots {
		if botIndex+1 != i {
			botIndex++
			s.usedBots[botIndex] = true
			return botIndex
		}
		botIndex = i
	}
	botIndex++
	s.usedBots[botIndex] = true
	lock.Unlock()
	return botIndex
}

func (s *LocalSpawnerStrategy) Spawn(tableId int, tablePassword string) error {
	botIndex := s.nextAvailableBotIndex()
	botName := fmt.Sprintf("ovo-test-%v", botIndex)

	envVars := append(os.Environ(), fmt.Sprintf("BOT_PASSWORD=%v", os.Getenv("BOT_PASSWORD")))
	envVars = append(envVars, fmt.Sprintf("BOT_NAME=%v", botName))
	envVars = append(envVars, fmt.Sprintf("TABLE_ID=%v", fmt.Sprintf("%v", tableId)))
	envVars = append(envVars, fmt.Sprintf("TABLE_PASSWORD=%v", tablePassword))

	cmdBuild := exec.Command("go", "run", "./player")
	cmdBuild.Stdout = os.Stdout
	cmdBuild.Stderr = os.Stderr
	cmdBuild.Env = envVars

	err := cmdBuild.Run()

	lock.Lock()
	delete(s.usedBots, botIndex)
	lock.Unlock()

	return err
}
