package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	VERSION = "0.0.2-git"
)

var (
	flagToken   string
	flagChannel string
	flagGuild   string
	flagUser    string
	flagMessage string
	flagDiscrim string
	flagSkip    string

	actions = map[string]func(*discordgo.Session) error{
		"sendmessage":   SendMessage,
		"gateway":       Gateway,
		"dumpall":       DumpAll,
		"guildroles":    GuildRoles,
		"guild":         Guild,
		"discrimsearch": DiscrimSearch,
		"dumpuser":      DumpUser,
	}
)

func init() {
	flag.StringVar(&flagToken, "t", "", "Token to use")
	flag.StringVar(&flagChannel, "c", "", "Select a channel")
	flag.StringVar(&flagGuild, "g", "", "Select a guild/server")
	flag.StringVar(&flagUser, "u", "", "Select a user")
	flag.StringVar(&flagMessage, "m", "", "Message to send")
	flag.StringVar(&flagDiscrim, "d", "", "discrim to search")
	flag.StringVar(&flagSkip, "s", "", "skip results that match this")
	flag.Parse()
}

func logln(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
}

func logf(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format, v...)
}

func main() {

	if flagToken == "" {
		flagToken = os.Getenv("DG_TOKEN")

		if flagToken == "" {
			logln("No token specified (either env var DG_TOKEN or arg)")
			os.Exit(1)
		}
	}

	session, err := discordgo.New(flagToken)
	if err != nil {
		logln("Error creating session:", err)
		os.Exit(1)
	}

	action := strings.ToLower(flag.Arg(0))

	actionFunc, ok := actions[action]
	if !ok {
		logln("Unknown action: ", action)
		PrintActions()
		os.Exit(1)
		return
	}

	err = actionFunc(session)
	if err != nil {
		logln("An error occured:", err)
		os.Exit(1)
	} else {
		logln("Success.")
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

// Connects to the gateway, and requests guild members from all joined guilds
func DiscrimSearch(s *discordgo.Session) error {
	var wg sync.WaitGroup

	readyHandler := func(session *discordgo.Session, r *discordgo.Ready) {
		for _, g := range r.Guilds {
			session.RequestGuildMembers(g.ID, "", 0)
			time.Sleep(time.Second)
		}
		logln("Done")
		wg.Done()
	}

	guildMembersChunkHandler := func(session *discordgo.Session, gm *discordgo.GuildMembersChunk) {
		for _, member := range gm.Members {
			if member.User.Discriminator == flagDiscrim && (flagSkip == "" || flagSkip != member.User.ID) {
				log.Printf("%q#%s (%s)", member.User.Username, member.User.Discriminator, member.User.ID)
			}
		}
	}

	s.AddHandler(readyHandler)
	s.AddHandler(guildMembersChunkHandler)

	wg.Add(1)
	err := s.Open()
	if err != nil {
		return err
	}

	wg.Wait()
	return s.Close()
}

func DumpUser(s *discordgo.Session) error {
	if flagUser == "" {
		logln("No user specified, dumping '@me'")
		flagUser = "@me"
	}
	me, err := s.User(flagUser)
	if err != nil {
		return err
	}

	out, err := json.MarshalIndent(me, "", " ")
	if err != nil {
		return err
	}

	fmt.Println(string(out))
	return nil
}
