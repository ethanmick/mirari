package mirari

import (
	"encoding/json"
	"errors"
	"log"
	"regexp"
	"strings"
)

// Event\.MatchCreated[\r\n\s\w\W]*matchId":\s"39130126-411d-4071-a4c7-e5a1e9596442"
const (
	logSplitRegex           = `(\[UnityCrossThreadLogger\]|\[Client GRE\])`
	isCollectionRegex       = `<==\sPlayerInventory\.GetPlayerCardsV3\(\d*\)`
	isDeckListRegex         = `<==\sDeck\.GetDeckLists\(\d*\)`
	isPlayerInventoryRegex  = `<==\sPlayerInventory\.GetPlayerInventory\(\d*\)`
	isPlayerConnectionRegex = `ClientToMatchServiceMessageType_AuthenticateRequest`
	isRankInfoRegex         = `<==\sEvent\.GetCombinedRankInfo\(\d*\)`
	isMatchPlayerCourse     = `<==\sEvent\.GetPlayerCourse\(\d*\)`
	isMatchStartRegex       = `Incoming\sEvent\.MatchCreated`
	isMatchEndRegex         = `DuelScene\.GameStop`
)

// ArenaCollection is a parsed collection object from MTGA
type ArenaCollection map[string]int

// ArenaDeck is the log format for a Deck
type ArenaDeck struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Format      string          `json:"format"`
	ResourceID  string          `json:"resourceId"`
	DeckTileID  int             `json:"deckTileId"`
	MainDeck    []ArenaDeckCard `json:"mainDeck"`
	Sideboard   []ArenaDeckCard `json:"sideboard"`
}

// ArenaDeckCard hold the info of the cards in a deck
type ArenaDeckCard struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
}

// ArenaPlayerInventory is your player profile details
type ArenaPlayerInventory struct {
	PlayerID        string  `json:"playerId"`
	WcCommon        int     `json:"wcCommon"`
	WcUncommon      int     `json:"wcUncommon"`
	WcRare          int     `json:"wcRare"`
	WcMythic        int     `json:"wcMythic"`
	Gold            int     `json:"gold"`
	Gems            int     `json:"gems"`
	DraftTokens     int     `json:"draftTokens"`
	SealedTokens    int     `json:"sealedTokens"`
	WcTrackPosition int     `json:"wcTrackPosition"`
	VaultProgress   float64 `json:"vaultProgress"`
}

// ArenaRankInfo contains a players rank info
type ArenaRankInfo struct {
	ConstructedSeasonOrdinal *int    `json:"constructedSeasonOrdinal"`
	ConstructedClass         *string `json:"constructedClass"`
	ConstructedLevel         *int    `json:"constructedLevel"`
	ConstructedStep          *int    `json:"constructedStep"`
	ConstructedMatchesWon    *int    `json:"constructedMatchesWon"`
	ConstructedMatchesLost   *int    `json:"constructedMatchesLost"`
	ConstructedMatchesDrawn  *int    `json:"constructedMatchesDrawn"`
	LimitedSeasonOrdinal     *int    `json:"limitedSeasonOrdinal"`
	LimitedClass             *string `json:"limitedClass"`
	LimitedLevel             *int    `json:"limitedLevel"`
	LimitedStep              *int    `json:"limitedStep"`
	LimitedMatchesWon        *int    `json:"limitedMatchesWon"`
	LimitedMatchesLost       *int    `json:"limitedMatchesLost"`
	LimitedMatchesDrawn      *int    `json:"limitedMatchesDrawn"`
}

// ArenaAuthRequestPayload is the payload the Arena client sends when
// authenticating. We only are interested in the Player's name
type ArenaAuthRequestPayload struct {
	PlayerName string `json:"PlayerName"`
}

// ArenaAuthRequest is the base structure which wraps a payload
type ArenaAuthRequest struct {
	Payload ArenaAuthRequestPayload `json:"Payload"`
}

// ArenaMatch is a match in Arena. May not be completed yet
type ArenaMatch struct {
	MatchID                        string     `json:"matchId"`
	OpponentScreenName             string     `json:"opponentScreenName"`
	OpponentIsWotc                 bool       `json:"opponentIsWotc"`
	OpponentRankingClass           string     `json:"opponentRankingClass"`
	OpponentRankingTier            int        `json:"opponentRankingTier"`
	OpponentMythicPercentile       float64    `json:"opponentMythicPercentile"`
	OpponentMythicLeaderboardPlace int        `json:"opponentMythicLeaderboardPlace"`
	EventID                        string     `json:"eventId"`
	SeatID                         *int       `json:"seatId"`
	TeamID                         *int       `json:"teamId"`
	GameNumber                     *int       `json:"gameNumber"`
	WinningTeamID                  *int       `json:"winningTeamId"`
	WinningReason                  *string    `json:"winningReason"`
	TurnCount                      *int       `json:"turnCount"`
	SecondsCount                   *int       `json:"secondsCount"`
	CourseDeck                     *ArenaDeck `json:"CourseDeck"`
}

// ArenaMatchEndParams are the params which hold the results of the match
type ArenaMatchEndParams struct {
	PayloadObject *ArenaMatch `json:"payloadObject"`
}

// ArenaMatchEnd is the outer structure
type ArenaMatchEnd struct {
	Params *ArenaMatchEndParams `json:"params"`
}

// ParseCollection looks for a MTGA Collection JSON object in a given input
func ParseCollection(raw string) (ArenaCollection, error) {
	isCollection := regexp.MustCompile(isCollectionRegex)
	texts := strings.Split(raw, "[UnityCrossThreadLogger]")
	match := getLastRegex(texts, isCollection, 2)
	if match == "" {
		return nil, errors.New("collection not found")
	}
	var collection ArenaCollection
	err := json.Unmarshal([]byte(match), &collection)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

// ParseDecks looks through an Arena log file and returns the decks
func ParseDecks(raw string) ([]ArenaDeck, error) {
	isDeck := regexp.MustCompile(isDeckListRegex)
	texts := strings.Split(raw, "[UnityCrossThreadLogger]")
	match := getLastRegex(texts, isDeck, 2)
	if match == "" {
		return nil, errors.New("decks not found")
	}
	var decks []ArenaDeck
	err := json.Unmarshal([]byte(match), &decks)
	if err != nil {
		return nil, err
	}
	return decks, nil
}

// ParsePlayerInventory gets a players details
func ParsePlayerInventory(raw string) (*ArenaPlayerInventory, error) {
	isPlayerInventory := regexp.MustCompile(isPlayerInventoryRegex)
	texts := strings.Split(raw, "[UnityCrossThreadLogger]")
	match := getLastRegex(texts, isPlayerInventory, 2)
	if match == "" {
		return nil, errors.New("inventory not found")
	}
	var inventory ArenaPlayerInventory
	err := json.Unmarshal([]byte(match), &inventory)
	if err != nil {
		return nil, err
	}
	return &inventory, nil
}

// ParseRankInfo finds a users rank info
func ParseRankInfo(raw string) (*ArenaRankInfo, error) {
	isRankInfo := regexp.MustCompile(isRankInfoRegex)
	texts := splitLogText(raw, logSplitRegex)
	rankJSON := getLastRegex(texts, isRankInfo, 2)
	if rankJSON == "" {
		return nil, errors.New("rank info not found")
	}
	var rank ArenaRankInfo
	err := json.Unmarshal([]byte(rankJSON), &rank)
	if err != nil {
		return nil, err
	}
	return &rank, nil
}

// ParseAuthRequest parses the auth request (for username)
func ParseAuthRequest(raw string) (*ArenaAuthRequest, error) {
	isPlayerConnection := regexp.MustCompile(isPlayerConnectionRegex)
	texts := splitLogText(raw, logSplitRegex)
	match := getLastRegex(texts, isPlayerConnection, 1)
	if match == "" {
		return nil, errors.New("auth request not found")
	}
	var auth ArenaAuthRequest
	err := json.Unmarshal([]byte(match), &auth)
	if err != nil {
		return nil, err
	}
	return &auth, nil
}

// ParseMatches finds the matches in a log
func ParseMatches(raw string) []ArenaMatch {
	texts := splitLogText(raw, logSplitRegex)
	var match *ArenaMatch
	var matches []ArenaMatch
	for _, t := range texts {
		isMatchDeck := regexp.MustCompile(isMatchPlayerCourse)
		isMatchStart := regexp.MustCompile(isMatchStartRegex)
		isMatchEnd := regexp.MustCompile(isMatchEndRegex)
		// The Player Course shows what they started searching for, and with
		// which deck, which we need to know what they played with
		if isMatchDeck.MatchString(t) {
			if match != nil {
				match = nil
			}
			playerCourse := strings.SplitN(t, "\n", 3)[2]
			if err := parseJSONBackoff(playerCourse, &match); err != nil {
				log.Printf("Error Parsing Player Course: %v", err.Error())
				continue
			}
		}
		if isMatchStart.MatchString(t) {
			incomingMatchJSON := strings.SplitN(t, "\n", 2)[1]
			// Need to chomp off the first part until we get to the JSON
			incomingMatchJSON = strings.TrimPrefix(incomingMatchJSON, "(-1) Incoming Event.MatchCreated ")
			if err := json.Unmarshal([]byte(incomingMatchJSON), &match); err != nil {
				log.Printf("Error Parsing Match Start: %v", err.Error())
				continue
			}
		}
		// Okay, we have a match, now what was the result?
		if isMatchEnd.MatchString(t) && match != nil {
			matchEndJSON := strings.SplitN(t, "\n", 3)[2]
			var result ArenaMatchEnd
			err := json.Unmarshal([]byte(matchEndJSON), &result)
			if err != nil {
				log.Printf("Error Parsing Match: %v", err.Error())
				continue
			}
			match.SeatID = result.Params.PayloadObject.SeatID
			match.TeamID = result.Params.PayloadObject.TeamID
			match.GameNumber = result.Params.PayloadObject.GameNumber
			match.WinningTeamID = result.Params.PayloadObject.WinningTeamID
			match.WinningReason = result.Params.PayloadObject.WinningReason
			match.TurnCount = result.Params.PayloadObject.TurnCount
			match.SecondsCount = result.Params.PayloadObject.SecondsCount
			matches = append(matches, *match)
			match = nil
		}
	}
	return matches
}

func splitLogText(raw, split string) []string {
	r := regexp.MustCompile(split)
	return r.Split(raw, -1)
}

func getLastRegex(texts []string, regex *regexp.Regexp, i int) string {
	var match string
	for _, rawLogText := range texts {
		if regex.MatchString(rawLogText) {
			split := strings.SplitN(rawLogText, "\n", i+1)
			if len(split) != i+1 {
				continue
			}
			match = split[i]
		}
	}
	return match
}

func parseJSONBackoff(s string, res interface{}) error {
	if s == "" {
		log.Println("json backoff failed with empty string")
		return errors.New("unable to parse")
	}
	log.Printf("parsing: %v\n", s)
	if err := json.Unmarshal([]byte(s), &res); err != nil {
		split := strings.Split(s, "\n")
		split = split[:len(split)-1]
		return parseJSONBackoff(strings.Join(split, "\n"), res)
	}
	return nil
}
