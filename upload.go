package mirari

// UploadData encapsulates the data to send to the server
type UploadData struct {
	Collection *ArenaCollection      `json:"collection"`
	Decks      *[]ArenaDeck          `json:"deck"`
	Inventory  *ArenaPlayerInventory `json:"inventory"`
	Rank       *ArenaRankInfo        `json:"rank"`
	Auth       *ArenaAuthRequest     `json:"auth"`
	Matches    *[]ArenaMatch         `json:"matches"`
}
