package main

import (
	"gator/internal/config"
	"gator/internal/database"
)

type State struct {
	db     *database.Queries
	Config config.Config
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	commands map[string]func(*State, Command) error
}
