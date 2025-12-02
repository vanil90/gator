package main

import (
	"context"
	"database/sql"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"os"

	_ "github.com/lib/pq"
)

func (c *Commands) run(s *state, cmd command) error {
	handler, ok := c.commands[cmd.Name]
	if !ok {
		return fmt.Errorf("invalid command name: '%s'", cmd.Name)
	}
	return handler(s, cmd)
}

func (c *Commands) register(name string, f func(*state, command) error) {
	c.commands[name] = f
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.Config.CurrentUsername)
		if err != nil {
			return err
		}

		return handler(s, cmd, user)
	}
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		panic(err)
	}
	s := state{
		Config: cfg,
	}
	cmds := Commands{
		commands: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handleLogin)
	cmds.register("register", handleRegister)
	cmds.register("reset", handleReset)
	cmds.register("users", handleUsers)
	cmds.register("agg", handleAgg)
	cmds.register("addfeed", middlewareLoggedIn(handleAddFeed))
	cmds.register("feeds", handleListFeeds)
	cmds.register("follow", middlewareLoggedIn(handleFollow))
	cmds.register("following", middlewareLoggedIn(handleFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handleUnfollow))
	cmds.register("browse", middlewareLoggedIn(handleBrowse))

	db, err := sql.Open("postgres", cfg.DbUrl)
	dbQueries := database.New(db)
	s.db = dbQueries

	if len(os.Args) < 2 {
		panic("error: missing command")
	}
	args := os.Args[2:]
	cmdName := os.Args[1]

	fmt.Println(args)
	cmd := command{
		Name: cmdName,
		Args: args,
	}

	err = cmds.run(&s, cmd)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
