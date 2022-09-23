package reddit

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rgood/account_verification/pkg/discord"
	"github.com/rgood/account_verification/pkg/invite"
	"github.com/rgood/go-reddit/reddit"
)

var (
	username = os.Getenv("REDDIT_USERNAME")
	password = os.Getenv("REDDIT_PASSWORD")
	id       = os.Getenv("REDDIT_CLIENT_ID")
	secret   = os.Getenv("REDDIT_SECRET_ID")

	messageSubject = os.Getenv("MESSAGE_SUBJECT")
)

func Run(s *discordgo.Session, inviteService *invite.InviteService) {
	credentials := reddit.Credentials{
		ID:       id,
		Secret:   secret,
		Username: username,
		Password: password,
	}
	client, _ := reddit.NewClient(credentials)

	for range time.Tick(5 * time.Second) {
		_, m2, _, _ := client.Message.InboxUnread(context.Background(), nil)
		handled := []string{}
		for _, message := range m2 {

			handled = append(handled, "t4_"+message.ID)

			if message.Subject == messageSubject {
				code := strings.TrimSpace(message.Text)
				fmt.Printf("%s sent code: %s", message.Author, code)
				ok, user, _ := inviteService.CheckCode(code)
				if ok {
					// Check if there's already a valid association for that reddit user
					if !inviteService.IsValidated(strings.ToLower(message.Author)) {

						// Approve the user
						guildIDs := inviteService.ExpireCode(code)
						for _, guildID := range guildIDs {
							err := s.GuildMemberNickname(guildID, user.Id, message.Author)
							if err != nil {
								println("Error setting name in guild: %s; %v\n", guildID, err)
							}
						}
					} else {
						discord.SendUserMessage(s, user.Id, fmt.Sprintf("Could not complete validation, because /u/%s is already associate with a Discord account.", message.Author))
					}
				}
			}
		}

		if len(handled) > 0 {
			_, err := client.Message.Read(context.Background(), handled...)
			if err != nil {
				println(err.Error())
			}
		}
	}
}
