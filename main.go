package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func testDatabaseConnection(db *sql.DB) {
	// Versuche eine einfache Abfrage, um die Verbindung zu testen
	err := db.Ping() // Ping sendet ein Signal, um die Verbindung zu testen
	if err != nil {
		log.Fatalf("Fehler bei der Verbindung zur Datenbank: %v", err)
	}
	fmt.Println("Verbindung zur Datenbank erfolgreich!")
}

func writeTable(db *sql.DB, bank int, streak int, wins int, loses int) {
	insertStmt := `INSERT INTO results (bank, streak, wins, loses) VALUES (?, ?, ?, ?)`

	_, err := db.Exec(insertStmt, bank, streak, wins, loses)
	if err != nil {
		log.Printf("Fehler beim Einfügen der Daten: %v", err) // Fehler hier korrekt behandeln
		return
	}
	fmt.Println("Spielergebnis erfolgreich gespeichert!")
}

func connectDB() *sql.DB {

	dbUser := "root"         //os.Getenv("DB_USER")
	dbPassword := "password" //os.Getenv("DB_PASSWORD")
	dbHost := "localhost"    // localhost <-> mysql                     //os.Getenv("DB_HOST")
	dbPort := "3306"         //os.Getenv("DB_PORT")
	dbName := "blackjack"    //os.Getenv("DB_NAME")

	fmt.Println("DB_USER:", dbUser)
	fmt.Println("DB_PASSWORD:", dbPassword)
	fmt.Println("DB_HOST:", dbHost)
	fmt.Println("DB_PORT:", dbPort)
	fmt.Println("DB_NAME:", dbName)

	if dbUser == "" || dbPassword == "" || dbHost == "" || dbPort == "" || dbName == "" {
		log.Fatal("Fehlende Umgebungsvariablen für die Datenbankverbindung.")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)

	var db *sql.DB
	var err error
	for {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			break
		}
		fmt.Println("Verbindungsversuch fehlgeschlagen. Warte 2 Sekunden und versuche es erneut...")
		time.Sleep(2 * time.Second)
	}

	err = db.Ping() // Ping sendet ein Signal, um die Verbindung zu testen
	if err != nil {
		log.Fatalf("Fehler bei der Verbindung zur Datenbank: %v", err)
	}
	fmt.Println("Verbindung zur Datenbank erfolgreich!")

	createTableSQL := `CREATE TABLE IF NOT EXISTS results (
		bank INT,
		streak INT, 
		wins INT, 
		loses INT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

// Player strukturiert das Spielerkonto und den Gewinn und die Gewinnserie
type Player struct {
	Bank   int
	Bet    int
	Win    int
	Loses  int
	Streak int
}

type Card struct {
	Suit Suit
	Rank Rank
}

type Suit string
type Rank string

const (
	Hearts, Diamonds, Clubs, Spades = "Herz", "Karo", "Kreuz", "Pik"
	Two, Three, Four, Five, Six     = "2", "3", "4", "5", "6"
	Seven, Eight, Nine, Ten         = "7", "8", "9", "10"
	Jack, Queen, King, Ace          = "Bube", "Dame", "König", "Ass"
)

type Deck []Card

// JSON-Datei schreiben
func writeJSON(filename string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

// JSON-Datei lesen
func readJSON(filename string, data interface{}) error {
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, data)
}

// Bank aktualisieren
func updateBank(filename string, player *Player) {
	playerData := map[string]int{
		"bank":   player.Bank,
		"win":    player.Win,
		"streak": player.Streak,
	}
	if err := writeJSON(filename, playerData); err != nil {
		fmt.Println("Fehler beim Schreiben der Datei:", err)
	}
}

// Neues Deck erstellen und mischen
func newDeck() Deck {
	suits := []Suit{Hearts, Diamonds, Clubs, Spades}
	ranks := []Rank{Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King, Ace}
	var deck Deck
	for _, suit := range suits {
		for _, rank := range ranks {
			deck = append(deck, Card{suit, rank})
		}
	}
	deck.shuffle()
	return deck
}

// Deck mischen
func (d Deck) shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(d), func(i, j int) { d[i], d[j] = d[j], d[i] })
}

// Karte ziehen
func (d *Deck) draw() Card {
	card := (*d)[0]
	*d = (*d)[1:]
	return card
}

// Kartenwert berechnen
func (c Card) value() int {
	switch c.Rank {
	case Two, Three, Four, Five, Six, Seven, Eight, Nine:
		return int(c.Rank[0] - '0')
	case Ten, Jack, Queen, King:
		return 10
	case Ace:
		return 11
	}
	return 0
}

// Handwert berechnen, inkl. Ace-Reduzierung
func handValue(hand []Card) int {
	total, aces := 0, 0
	for _, card := range hand {
		total += card.value()
		if card.Rank == Ace {
			aces++
		}
	}
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}
	return total
}

// Eingabe des Spielers zum Wetten
func getBet(bank int) int {
	var bet int
	fmt.Println("Was ist Ihr Einsatz?")
	fmt.Scan(&bet)
	for bet < 0 || bet > bank {
		fmt.Println("Ungültiger Einsatz. Geben Sie einen neuen Betrag ein:")
		fmt.Scan(&bet)
	}
	return bet
}

// Hauptspielablauf
func playBlackjack(player *Player) {
	deck := newDeck()
	playerHand := []Card{deck.draw(), deck.draw()}
	dealerHand := []Card{deck.draw(), deck.draw()}

	fmt.Println("Ihre Hand:")
	displayHand(playerHand)
	fmt.Println("\nDealer zeigt:", dealerHand[0].Rank, "von", dealerHand[0].Suit)
	playerTurn(deck, &playerHand)

	// Wenn Spieler unter 21, dann Dealerzug
	if handValue(playerHand) <= 21 {
		fmt.Println("\nDealer-Hand:")
		displayHand(dealerHand)
		dealerTurn(deck, &dealerHand)
	}

	result := determineWinner(handValue(playerHand), handValue(dealerHand))
	fmt.Println(result)
	updateBankResult(player, result)
}

// Hand anzeigen
func displayHand(hand []Card) {
	for _, card := range hand {
		fmt.Printf(" - %s von %s\n", card.Rank, card.Suit)
	}
	fmt.Printf("Gesamtwert: %d\n", handValue(hand))
}

// Spielerzug
func playerTurn(deck Deck, hand *[]Card) {
	for handValue(*hand) < 21 {
		fmt.Print("Möchten Sie eine Karte ziehen (h) oder halten (s)? ")
		input := bufio.NewScanner(os.Stdin)
		for input.Text() == "" {
			input.Scan()
		}
		switch input.Text() {
		case "h":
			*hand = append(*hand, deck.draw())
			fmt.Println("Neue Karte:", (*hand)[len(*hand)-1].Rank, "von", (*hand)[len(*hand)-1].Suit)
			fmt.Printf("Neuer Gesamtwert: %d\n", handValue(*hand))
		case "s":

			return
		default:
			fmt.Println("Ungültige Eingabe.")
		}
	}
}

// Dealerzug
func dealerTurn(deck Deck, hand *[]Card) {
	for handValue(*hand) < 17 {
		*hand = append(*hand, deck.draw())
		fmt.Println("Dealer zieht:", (*hand)[len(*hand)-1].Rank, "von", (*hand)[len(*hand)-1].Suit)
	}
}

// Gewinnentscheidung
func determineWinner(playerScore, dealerScore int) string {
	switch {
	case playerScore > 21:
		return "Bust! Dealer gewinnt."
	case dealerScore > 21 || playerScore > dealerScore:
		return "Sie gewinnen!"
	case playerScore == dealerScore:
		return "Unentschieden!"
	default:
		return "Dealer gewinnt!"
	}
}

// Bankstand aktualisieren
func updateBankResult(player *Player, result string) {
	switch result {
	case "Sie gewinnen!":
		player.Bank += player.Bet * 2
		player.Streak++
		player.Win++
	case "Unentschieden!":
		player.Bank += player.Bet
	case "Bust! Dealer gewinnt.":
		player.Streak = 0
		player.Loses++
	case "Dealer gewinnt!":
		player.Loses++
		player.Streak = 0
	}

}

// Spiel starten
func main() {
	dbconnection := connectDB()

	filename := "money.json"
	player := &Player{}

	if err := readJSON(filename, player); err != nil {
		player.Bank = 1000 // Default-Wert, falls Datei fehlt
	}

	player.Bet = getBet(player.Bank)
	player.Bank -= player.Bet

	writeTable(dbconnection, player.Bank, player.Streak, player.Win, player.Loses)

	playBlackjack(player)
	//testDatabaseConnection(dbconnection)
	updateBank(filename, player)
	fmt.Printf("Ihr neues Guthaben: %d\n", player.Bank)
	defer dbconnection.Close()
}
