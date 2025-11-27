package main

import (
	"context"
	"errors"
	"fmt"
	"gator/internal/database"
	"time"

	"github.com/google/uuid"
)

func handleLogin(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return errors.New("login: invalid arguments for 'login' command")
	}
	username := cmd.Args[0]

	user, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		return fmt.Errorf("login: invalid username '%s'", username)
	}

	err = s.Config.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("User set to '%s'\n", username)

	return nil
}

func handleRegister(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return errors.New("register: invalid arguments for 'register' command")
	}

	ctx := context.Background()

	name := cmd.Args[0]
	_, err := s.db.GetUser(ctx, name)

	if err == nil {
		return fmt.Errorf("register: username '%s' already in use", name)
	}

	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
	}
	fmt.Println(userParams)

	userResult, err := s.db.CreateUser(ctx, userParams)
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}
	fmt.Println(userResult)

	err = s.Config.SetUser(name)
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}

	return nil
}

func handleReset(s *State, cmd Command) error {
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("reset: %w", err)
	}
	fmt.Println("database reset")
	return nil
}

func handleUsers(s *State, cmd Command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("users: %w", err)
	}

	for _, user := range users {
		if user.Name == s.Config.CurrentUsername {
			fmt.Printf("%s: %s (current)\n", user.ID, user.Name)
		} else {
			fmt.Printf("%s: %s\n", user.ID, user.Name)
		}
	}
	return nil
}
