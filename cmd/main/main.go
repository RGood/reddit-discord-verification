package main

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/rgood/account_verification/pkg/discord"
	"github.com/rgood/account_verification/pkg/invite"
	"github.com/rgood/account_verification/pkg/reddit"
)

var (
	botToken = os.Getenv("DISCORD_BOT_TOKEN")
)

func main() {
	inviteService := invite.NewInviteService("postgres", 5432, "postgres", "postgres", "postgres")

	s, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Invalid Discord bot parameters: %v", err)
	}

	go discord.Run(s, inviteService)
	reddit.Run(s, inviteService)
}
