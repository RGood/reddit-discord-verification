package invite

import (
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rgood/account_verification/pkg/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InviteService struct {
	db *gorm.DB
}

type PendingInvite struct {
	DiscordId string
	GuildIDs  []string
	Code      string
}

type UserJoin struct {
	DiscordId string
	JoinDate  time.Time
}

type DiscordUser struct {
	Id            string
	Username      string
	Discriminator string
}

type AccountAssociation struct {
	DiscordId      string
	RedditUsername string
}

func NewInviteService(host string, port int, username string, password string, database string) *InviteService {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", host, username, password, database, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return &InviteService{
		db: db,
	}
}

func (is *InviteService) LogJoin(user *discordgo.User) {
	is.db.Clauses(
		clause.OnConflict{DoNothing: true},
	).Create(&DiscordUser{
		Id:            user.ID,
		Username:      user.Username,
		Discriminator: user.Discriminator,
	})

	is.db.Create(&UserJoin{
		DiscordId: user.ID,
		JoinDate:  time.Now(),
	})
}

func (is *InviteService) CheckAssociation(discordId string) (*AccountAssociation, bool) {
	var accountAssociation AccountAssociation
	result := is.db.Where(AccountAssociation{
		DiscordId: discordId,
	}).First(&accountAssociation)

	if result.Error != nil {
		return nil, false
	}

	return &accountAssociation, true
}

func (is *InviteService) IsValidated(redditUsername string) bool {
	var accountAssociation AccountAssociation
	result := is.db.Where(AccountAssociation{
		RedditUsername: redditUsername,
	}).First(&accountAssociation)

	return !errors.Is(result.Error, gorm.ErrRecordNotFound)
}

func (is *InviteService) GenerateCode(user *discordgo.User, guildID string) (string, error) {
	pendingInvite := PendingInvite{
		DiscordId: user.ID,
		Code:      utils.RandStringRunes(32),
		GuildIDs:  []string{guildID},
	}

	result := is.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "discord_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"guild_ids": gorm.Expr("guild_ids || ?", []string{guildID}),
		}),
	}).Create(&pendingInvite)

	if result.Error != nil {
		return "", result.Error
	}

	return pendingInvite.Code, nil
}

func (is *InviteService) CheckCode(code string) (bool, *DiscordUser, error) {
	var pendingInvite PendingInvite
	result := is.db.Where(&PendingInvite{Code: code}).First(&pendingInvite)
	if result.Error != nil {
		return false, nil, result.Error
	}

	return pendingInvite.Code == code, &DiscordUser{
		Id: pendingInvite.DiscordId,
	}, nil
}

func (is *InviteService) ExpireCode(code string) []string {
	var pendingInvites []PendingInvite
	is.db.Clauses(clause.Returning{}).Where(&PendingInvite{Code: code}).Delete(&pendingInvites)

	if len(pendingInvites) > 0 {
		return pendingInvites[0].GuildIDs
	}

	return []string{}
}

func (is *InviteService) AssociateAccounts(discordId string, redditUsername string) {
	is.db.Create(&AccountAssociation{
		DiscordId:      discordId,
		RedditUsername: redditUsername,
	})
}
