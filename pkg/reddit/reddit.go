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

	messageSubject       = os.Getenv("MESSAGE_SUBJECT")
	mobileMessageSubject = strings.ReplaceAll(messageSubject, " ", "+")
)

const (
	tmpRoleId      = "1311701000247181392"
	tmpAdminRoleId = "1311701252358275102"
)

func Run(s *discordgo.Session, inviteService *invite.InviteService) {
	println("Running Reddit Inbox Checker!")
	defer println("Inbox checker exited!")

	credentials := reddit.Credentials{
		ID:       id,
		Secret:   secret,
		Username: username,
		Password: password,
	}
	client, _ := reddit.NewClient(credentials)

	for range time.Tick(5 * time.Second) {
		_, m2, _, err := client.Message.InboxUnread(context.Background(), nil)
		if err != nil {
			fmt.Printf("Error fetching inbox: %s\n", err.Error())
			continue
		}

		handled := []string{}
		for _, message := range m2 {

			handled = append(handled, "t4_"+message.ID)

			if message.Subject == messageSubject || message.Subject == mobileMessageSubject {
				code := strings.TrimSpace(message.Text)
				fmt.Printf("%s sent code: %s\n", message.Author, code)
				ok, user, err := inviteService.CheckCode(code)
				if err != nil {
					client.Comment.Submit(context.Background(), message.FullID, "Uh oh! That code doesn't lead anywhere. Please try leaving and re-joining the server!")
					continue
				}
				if ok {

					// Check if there's already a valid association for that reddit user
					if !inviteService.IsValidated(strings.ToLower(message.Author)) {

						// Approve the user
						guildIDs, err := inviteService.ExpireCode(code)
						if err != nil {
							fmt.Printf("Expire threw error: %s\n", err.Error())
							continue
						}
						_, err = fmt.Printf("Updating %s's name in the guilds: %v\n", message.Author, guildIDs)
						if err != nil {
							println("fmt.Printf Err:", err.Error())
						}

						redditUser, _, err := client.User.Get(context.Background(), message.Author)
						toAssign := []string{tmpRoleId}
						if err == nil && redditUser.IsEmployee {
							toAssign = append(toAssign, tmpAdminRoleId)
						}

						for _, guildID := range guildIDs {
							_, err := s.GuildMemberEdit(guildID, user.Id, &discordgo.GuildMemberParams{
								Nick:  message.Author,
								Roles: &toAssign,
							})
							if err != nil {
								fmt.Printf("Error setting name for user %s in guild: %s; %v\n", user.Id, guildID, err)
							}
						}
					} else {
						println("Couldn't approve.")
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
