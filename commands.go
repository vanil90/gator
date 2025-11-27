package main

import (
	"gator/internal/config"
	"gator/internal/database"
)

type state struct {
	db     *database.Queries
	Config config.Config
}

type command struct {
	Name string
	Args []string
}

type Commands struct {
	commands map[string]func(*state, command) error
}
