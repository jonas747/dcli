package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"strings"
	"sync"
)

const (
	VERSION = "0.0.1"
)

var (
	flagToken   string
	flagChannel string
	flagMessage string

	actions = map[string]func(*discordgo.Session) error{
		"sendmessage": SendMessage,
		"gateway":     Gateway,
	}
)

func init() {
	flag.StringVar(&flagToken, "t", "", "Token to use")
	flag.StringVar(&flagChannel, "c", "", "Select a channel")
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
