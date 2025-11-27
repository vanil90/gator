package main

import (
	"database/sql"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"os"

	_ "github.com/lib/pq"
)

func (c *Commands) run(s *State, cmd Command) error {
	handler, ok := c.commands[cmd.Name]
	if !ok {
		return fmt.Errorf("invalid command name: '%s'", cmd.Name)
	}
	return handler(s, cmd)
}

func (c *Commands) register(name string, f func(*State, Command) error) {
	c.commands[name] = f
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		panic(err)
	}
	s := State{
		Config: cfg,
	}
	handlers := Commands{
		commands: make(map[string]func(*State, Command) error),
	}
	handlers.register("login", handleLogin)
	handlers.register("register", handleRegister)
	handlers.register("reset", handleReset)
	handlers.register("users", handleUsers)
	handlers.register("agg", handleAgg)
	handlers.register("addfeed", handleAddFeed)
	handlers.register("feeds", handleListFeeds)
	handlers.register("follow", handleFollow)
	handlers.register("following", handleFollowing)

	db, err := sql.Open("postgres", cfg.DbUrl)
	dbQueries := database.New(db)
	s.db = dbQueries

	if len(os.Args) < 2 {
		panic("error: missing command")
	}
	args := os.Args[2:]
	cmdName := os.Args[1]

	fmt.Println(args)
	cmd := Command{
		Name: cmdName,
		Args: args,
	}

	err = handlers.run(&s, cmd)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
