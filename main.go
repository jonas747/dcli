package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"strings"
	"sync"
)

const (
	VERSION = "0.0.2-git"
)

var (
	flagToken   string
	flagChannel string
	flagGuild   string
	flagMessage string

	actions = map[string]func(*discordgo.Session) error{
		"sendmessage": SendMessage,
		"gateway":     Gateway,
		"dumpall":     DumpAll,
		"guildroles":  GuildRoles,
		"guild":       Guild,
	}
)

func init() {
	flag.StringVar(&flagToken, "t", "", "Token to use")
	flag.StringVar(&flagChannel, "c", "", "Select a channel")
	flag.StringVar(&flagGuild, "g", "", "Select a guild/server")
	flag.StringVar(&flagMessage, "m", "", "Message to send")
	flag.Parse()
}

func main() {

	if flagToken == "" {
		fmt.Println("No token specified")
		os.Exit(1)
	}

	session, err := discordgo.New(flagToken)
	if err != nil {
		fmt.Println("Error creating session:", err)
		os.Exit(1)
	}

	action := strings.ToLower(flag.Arg(0))

	actionFunc, ok := actions[action]
	if !ok {
		fmt.Println("Unknown actions")
		PrintActions()
		return
	}

	err = actionFunc(session)
	if err != nil {
		fmt.Println("An error occured:", err)
		os.Exit(1)
	}
}

// Prints all the available actions
func PrintActions() {
	fmt.Println("Available actions:")
	for k, _ := range actions {
		fmt.Println(k)
	}
}

// Sends a message in a channel
func SendMessage(s *discordgo.Session) error {
	if flagChannel == "" {
		return errors.New("No channel specified (-c channel)")
	}

	if flagMessage == "" {
		return errors.New("No message specified (-m message)")
	}

	_, err := s.ChannelMessageSend(flagChannel, flagMessage)
	return err
}

// Connects to the gateway and waits for ready then exits
func Gateway(s *discordgo.Session) error {
	var wg sync.WaitGroup

	readyHandler := func(session *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("Ready received! Sucessfully connected to gateway, exiting...")
		wg.Done()
	}

	s.AddHandler(readyHandler)

	wg.Add(1)
	err := s.Open()
	if err != nil {
		return err
	}

	wg.Wait()
	return s.Close()
}

// Dumps all discord events to stdout
func DumpAll(s *discordgo.Session) error {
	s.Debug = true
	s.LogLevel = discordgo.LogDebug

	err := s.Open()
	if err != nil {
		return err
	}

	fmt.Println("Runnning. ctrl-c to exit.")
	select {}
	return nil
}

// Dumps the guild roles to stdout
func GuildRoles(s *discordgo.Session) error {
	roles, err := s.GuildRoles(flagGuild)
	if err != nil {
		return err
	}

	for _, v := range roles {
		fmt.Printf("Role %s, ID: %s, Position: %d, Perms: %d, Hoist: %b, Color: %d, Managed: %b\n", v.Name, v.ID, v.Position, v.Permissions, v.Hoist, v.Color, v.Managed)
	}

	fmt.Println(len(roles), "Guild roles")
	return nil
}

// Dumps the guild roles to stdout
func Guild(s *discordgo.Session) error {
	guild, err := s.Guild(flagGuild)
	if err != nil {
		return err
	}

	out, err := json.MarshalIndent(guild, "", " ")
	if err != nil {
		return err
	}

	fmt.Println(string(out))
	return nil
}
