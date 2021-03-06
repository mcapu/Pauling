package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // postgres driver

	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Pauling/internal/helpers"
)

var db *sql.DB

func ConnectDB() {
	var err error
	connectInfo := fmt.Sprintf("host=%s port =%s dbname=%s user=%s password=%s sslmode=disable", helpers.DBHost, helpers.DBPort, helpers.DBName, helpers.DBUser, helpers.DBPassword)

	helpers.Logger.Info("Connecting to DB")
	db, err = sql.Open("postgres", connectInfo)
	if err != nil {
		helpers.Logger.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		helpers.Logger.Fatal(err)
	}
}

// GetPlayerID returns a player ID (primary key), given their Steam Community id
func GetPlayerID(steamid string) uint {
	var id uint

	rows, _ := db.Query("SELECT id FROM players WHERE steam_id = $1", steamid)
	for rows.Next() {
		rows.Scan(&id)
	}

	return id
}

// GetTeam returns the player's team, given the player's steamid and the lobby id
func GetTeam(lobbyID uint, lobbyType models.LobbyType, steamID string) string {
	var slot int

	db.QueryRow("SELECT slot FROM lobby_slots WHERE player_id = $1", GetPlayerID(steamID)).Scan(&slot)

	team, _, _ := models.LobbyGetSlotInfoString(lobbyType, slot)

	return team
}

// GetSlotSteamID returns the steam ID for the player occupying the given slot
func GetSlotSteamID(team, class string, lobbyID uint, lobbyType models.LobbyType) string {
	var (
		steamID  string
		playerID uint
	)

	slot, err := models.LobbyGetPlayerSlot(lobbyType, team, class)
	if err != nil {
		return ""
	}

	db.QueryRow("SELECT player_id FROM lobby_slots WHERE slot = $1 AND lobby_id = $2", slot, lobbyID).Scan(&playerID)
	db.QueryRow("SELECT steam_id FROM players WHERE id = $1", playerID).Scan(&steamID)

	return steamID
}

// GetName returns the name for a plyer given their steam ID
func GetName(steamID string) string {
	var name string

	db.QueryRow("SELECT name FROM players WHERE steam_id = $1", steamID).Scan(&name)

	return name
}

func IsAllowed(lobbyID uint, steamID string) bool {
	playerID := GetPlayerID(steamID)
	if playerID == 0 {
		return false
	}

	rows, err := db.Query("SELECT slot FROM lobby_slots WHERE player_id = $1 AND lobby_id = $2", playerID, lobbyID)
	if err != nil {
		return false
	}

	return rows.Next()
}
