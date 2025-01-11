package invite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rgood/account_verification/pkg/utils"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type InviteService struct {
	db *bun.DB
}

type PendingInvite struct {
	DiscordId string
	GuildIds  []string `bun:",array"`
	Code      string
}

type UserJoin struct {
	DiscordId string
	JoinDate  time.Time
}

type DiscordUser struct {
	Id       string
	Username string
}

type AccountAssociation struct {
	DiscordId      string
	RedditUsername string
}

func NewInviteService(host string, port int, username string, password string, database string) *InviteService {
	//dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", host, username, password, database, port)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", username, password, host, port, database)
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	return &InviteService{
		db: db,
	}
}

func (is *InviteService) LogJoin(user *discordgo.User) {
	joinedUser := &DiscordUser{
		Id:       user.ID,
		Username: user.Username,
	}

	is.db.NewInsert().
		Model(joinedUser).
		On("CONFLICT DO NOTHING").
		Exec(context.Background())

	is.db.NewInsert().
		Model(&UserJoin{
			DiscordId: user.ID,
			JoinDate:  time.Now(),
		}).
		Exec(context.Background())
}

func (is *InviteService) CheckAssociation(discordId string) (*AccountAssociation, bool) {
	var accountAssociation AccountAssociation

	err := is.db.NewSelect().
		Model(&accountAssociation).
		Limit(1).
		Where("discord_id = ?", discordId).
		Scan(context.Background())

	if err != nil {
		println("Error checking association: %s", err.Error())
		return nil, false
	}

	return &accountAssociation, true
}

func (is *InviteService) IsValidated(redditUsername string) bool {
	var accountAssociation AccountAssociation

	err := is.db.NewSelect().
		Model(&accountAssociation).
		Limit(1).
		Where("reddit_username = ?", redditUsername).
		Scan(context.Background())

	return err == nil
}

func (is *InviteService) GenerateCode(user *discordgo.User, guildId string) (string, error) {
	pendingInvite := &PendingInvite{
		DiscordId: user.ID,
		Code:      utils.RandStringRunes(32),
		GuildIds:  []string{guildId},
	}

	_, err := is.db.NewInsert().
		Model(pendingInvite).
		On("CONFLICT (discord_id) DO UPDATE").
		Set("guild_ids = array_append(pending_invite.guild_ids, ?)", guildId).
		Set("code = ?", pendingInvite.Code).
		Exec(context.Background())

	if err != nil {
		return "", err
	}

	return pendingInvite.Code, nil
}

func (is *InviteService) CheckCode(code string) (bool, *DiscordUser, error) {
	pendingInvite := &PendingInvite{}

	err := is.db.NewSelect().Model(pendingInvite).Where("code = ?", code).Limit(1).Scan(context.Background())
	if err != nil {
		return false, nil, err
	}

	println("Checking", code, "==", pendingInvite.Code, code == pendingInvite.Code)
	fmt.Printf("User Id: %s", pendingInvite.DiscordId)

	return pendingInvite.Code == code, &DiscordUser{
		Id: pendingInvite.DiscordId,
	}, nil
}

func (is *InviteService) ExpireCode(code string) ([]string, error) {
	var pendingInvites PendingInvite

	res, err := is.db.NewDelete().Model(&pendingInvites).Where("code = ?", code).Returning("guild_ids").Exec(context.Background())
	if err != nil {
		return nil, err
	}

	if c, _ := res.RowsAffected(); c > 0 {
		fmt.Printf("Should return IDs: %v\n", pendingInvites.GuildIds)
		return pendingInvites.GuildIds, nil
	}

	return []string{}, nil
}

func (is *InviteService) AssociateAccounts(discordId string, redditUsername string) {
	is.db.NewInsert().Model(&AccountAssociation{
		DiscordId:      discordId,
		RedditUsername: redditUsername,
	}).Exec(context.Background())
}
