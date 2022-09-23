package discord

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/rgood/account_verification/pkg/invite"
	"github.com/rgood/account_verification/pkg/utils"
)

var (
	messageSubect = os.Getenv("MESSAGE_SUBJECT")
	botUsername   = os.Getenv("REDDIT_USERNAME")
)

var WelcomeMessage = "This Discord manages your identity through Reddit.\n" +
	"Before we let you in, we first need to know who you are.\n\n" +
	"Click here and hit \"Send\" to verify your account:\n"

func SendUserMessage(s *discordgo.Session, discordId string, message string) {
	channel, err := s.UserChannelCreate(discordId)
	if err != nil {
		return
	}

	_, err = s.ChannelMessageSend(channel.ID, message)
	if err != nil {
		return
	}
}

func Run(s *discordgo.Session, inviteService *invite.InviteService) {
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {

		if m.Type == 7 {
			inviteService.LogJoin(m.Author)

			// Check if user already has an associated account
			accountAssociation, ok := inviteService.CheckAssociation(m.Author.ID)

			if ok {
				// If so, assign the username
				err := s.GuildMemberNickname(m.GuildID, accountAssociation.DiscordId, accountAssociation.RedditUsername)
				if err != nil {
					fmt.Printf("Error setting username: %v\n", err)
				} else {
					fmt.Printf("Re-verified: %s\n", accountAssociation.RedditUsername)
				}
			} else {
				// Otherwise create a channel and send them a code

				channel, err := s.UserChannelCreate(m.Author.ID)
				if err != nil {
					// If an error occurred, we failed to create the channel.
					//
					// Some common causes are:
					// 1. We don't share a server with the user (not possible here).
					// 2. We opened enough DM channels quickly enough for Discord to
					//    label us as abusing the endpoint, blocking us from opening
					//    new ones.
					fmt.Println("error creating channel:", err)
					s.ChannelMessageSend(
						m.ChannelID,
						"Something went wrong while sending the DM!",
					)
					return
				}

				userCode, _ := inviteService.GenerateCode(m.Author, m.GuildID)
				_, err = s.ChannelMessageSend(channel.ID, WelcomeMessage+utils.CreateMessageURL(botUsername, messageSubect, userCode))
				if err != nil {
					// If an error occurred, we failed to send the message.
					//
					// It may occur either when we do not share a server with the
					// user (highly unlikely as we just received a message) or
					// the user disabled DM in their settings (more likely).
					fmt.Println("error sending DM message:", err)
					s.ChannelMessageSend(
						m.ChannelID,
						"Failed to send you a DM. "+
							"Did you disable DM in your privacy settings?",
					)
				}
			}
		}
	})

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	<-stop
}
