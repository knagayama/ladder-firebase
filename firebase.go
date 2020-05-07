package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"cloud.google.com/go/firestore"

	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Round represents a round in the tournament.
type Round int

func (r Round) String() string {
	return strconv.Itoa(int(r))
}

// Division represents a division in the tournament.
type Division int

// This iota represents a list of divisions.
const (
	X Division = iota
	SPlusUpper
	SPlusLower
	SUpper
	SMiddle
	SLower
	APlusUpper
	APlusMiddle
	APlusLower
	AUpper
	AMiddle
	ALower
	AMinusUpper
	AMinusMiddle
	AMinusLower
	BPlusUpper
	BPlusMiddle
	BPlusLower
	BUpper
	BMiddle
	BLower
	BMinusUpper
	BMinusMiddle
	BMinusLower
	CPlusUpper
	CPlusMiddle
	CPlusLower
	CUpper
	CMiddle
	CLower
	CMinusUpper
	CMinusMiddle
	CMinusLower
)

func (div Division) String() string {
	switch div {
	case X:
		return "X"
	case SPlusUpper:
		return "S+ Upper"
	case SPlusLower:
		return "S+ Lower"
	case SUpper:
		return "S Upper"
	case SMiddle:
		return "S Middle"
	case SLower:
		return "S Lower"
	case APlusUpper:
		return "A+ Upper"
	case APlusMiddle:
		return "A+ Middle"
	case APlusLower:
		return "A+ Lower"
	case AUpper:
		return "A Upper"
	case AMiddle:
		return "A Middle"
	case ALower:
		return "A Lower"
	case AMinusUpper:
		return "A- Upper"
	case AMinusMiddle:
		return "A- Middle"
	case AMinusLower:
		return "A- Lower"
	case BPlusUpper:
		return "B+ Upper"
	case BPlusMiddle:
		return "B+ Middle"
	case BPlusLower:
		return "B+ Lower"
	case BUpper:
		return "B Upper"
	case BMiddle:
		return "B Middle"
	case BLower:
		return "B Lower"
	case BMinusUpper:
		return "B- Upper"
	case BMinusMiddle:
		return "B- Middle"
	case BMinusLower:
		return "B- Lower"
	case CPlusUpper:
		return "C+ Upper"
	case CPlusMiddle:
		return "C+ Middle"
	case CPlusLower:
		return "C+ Lower"
	case CUpper:
		return "C Upper"
	case CMiddle:
		return "C Middle"
	case CLower:
		return "C Lower"
	case CMinusUpper:
		return "C- Upper"
	case CMinusMiddle:
		return "C- Middle"
	case CMinusLower:
		return "C- Lower"
	default:
		return "Unknown"
	}
}

// Challenge holds data for a challenge.
type Challenge struct {
	Round           Round    `firestore:"Round"`
	Code            int      `firestore:"Code"`
	Challenger      string   `firestore:"Challenger"`
	ChallengerRank  int      `firestore:"ChallengerRank"`
	ChallengerScore int      `firestore:"ChallengerScore"`
	Defender        string   `firestore:"Defender"`
	DefenderRank    int      `firestore:"DefenderRank"`
	DefenderScore   int      `firestore:"DefenderScore"`
	Division        Division `firestore:"Division"`
}

// InitTeams loads a local csv file given by path and uploads it to Firestore.
func InitTeams(ctx context.Context, teams *firestore.CollectionRef, path string) error {
	// Load team CSV file.
	csvfile, err := os.Open(path)
	if err != nil {
		return err
	}

	teamsDoc := make([]map[string]string, 0)

	r := csv.NewReader(csvfile)
	localTeams, err := r.ReadAll()
	if err != nil {
		return err
	}
	for _, team := range localTeams {
		doc := make(map[string]string)
		for column, name := range team {
			switch column {
			case 0:
				doc["name"] = name
			case 1:
				doc["player1"] = name
			case 2:
				doc["player2"] = name
			case 3:
				doc["player3"] = name
			case 4:
				doc["player4"] = name
			case 5:
				if name != "" {
					doc["player5"] = name
				}
			}
		}
		teamsDoc = append(teamsDoc, doc)
	}

	for _, doc := range teamsDoc {
		name := doc["name"]
		_, err := teams.Doc(name).Set(ctx, doc)
		if err != nil {
			return err
		}
	}
	return nil
}

// InitRanking initializes the ranking based on user input.
func InitRanking(ctx context.Context, teams *firestore.CollectionRef, ranking *firestore.DocumentRef) error {
	// Read from the teams document.
	var rank string

	iter := teams.Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		data := doc.Data()
		fmt.Println("Enter the ranking for: ", data["name"])
		fmt.Scanln(&rank)
		_, err = ranking.Set(ctx, map[string]interface{}{
			rank: data["name"],
		}, firestore.MergeAll)
		if err != nil {
			log.Printf("An error has occurred: %s", err)
		}
	}
	return nil
}

// InputScores uploads challenge scores based on user input.
func InputScores(ctx context.Context, challenges firestore.Query) {
	// Read challengesng for the current round and input the scores.
	iter := challenges.Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		var challenge Challenge
		if err = doc.DataTo(&challenge); err != nil {
			log.Println(err)
		}

		fmt.Printf("%d-%d Div %s: %s (rank %d) vs %s (rank %d)\n", challenge.Round, challenge.Code,
			challenge.Division.String(), challenge.Challenger, challenge.ChallengerRank,
			challenge.Defender, challenge.DefenderRank)
		fmt.Printf("Input score for challenger %s: ", challenge.Challenger)
		var cs, ds int
		fmt.Scanf("%d", &cs)
		fmt.Printf("Input score for defender %s: ", challenge.Defender)
		fmt.Scanf("%d", &ds)
		challenge.ChallengerScore = cs
		challenge.DefenderScore = ds
		_, err = doc.Ref.Set(ctx, challenge)
		if err != nil {
			log.Printf("Error occurred writing to Firestore: %s", err)
		}
		fmt.Println("Written to firebase.")
	}
}

// GenerateRanking generates ranking for the current round based on the last challenge scores.
func GenerateRanking(ctx context.Context, ranking *firestore.DocumentRef, challenges firestore.Query) {
	// Read the scores from last challenge scores.

	// Identify the winner and the loser for each division.

	// Swap the ranking for loser of the upper division and the winner of the lower division.
}

// CreateChallenges generate challenges based on the current team ranking and uploads it to Firestore.
func CreateChallenges(ctx context.Context, tournament *firestore.DocumentRef) {
	// Get ranking from current round.
	teams := make(map[int]string)
	tournamentdsnap, err := tournament.Get(ctx)
	if err != nil {
		log.Fatalln("Error reading tournament data from Firestore:", err)
	}
	tournamentdata := tournamentdsnap.Data()
	var currentRound Round
	t := tournamentdata["currentRound"].(int64)
	currentRound = Round(t)
	rankingdsnap, err := tournament.Collection("ranking").Doc(currentRound.String()).Get(ctx)
	if err != nil {
		log.Fatalln("Error reading ranking from Firestore: ", err)
	}
	rankingdata := rankingdsnap.Data()

	for key, value := range rankingdata {
		i, err := strconv.Atoi(key)
		if err != nil {
			log.Printf("An error has occurred converting a to i: %s", err)
		}
		teams[i] = value.(string)
	}

	// Figure out divisions. Assign map[string][]int which maps divisions to ranks.
	// Division have the followings names:
	// X, S+ Upper, S+ Lower, S Upper, S Middle, S Lower, A+ Upper, A+ Middle, A+ Lower,
	// A Upper, A Middle, A Lower, B Upper, B Middle, B Lower.

	divisionToTeam := make(map[Division][]int)

	switch len(teams) % 3 {
	// If len(teams) % 3 == 0, then all divisions have 3 teams.
	// If len(teams) % 3 == 2, then last division has 2 teams.
	case 0, 2:
		for i, j := 1, X; i < len(teams)+1; i++ {
			divisionToTeam[j] = append(divisionToTeam[j], i)
			if i%3 == 0 {
				j++
			}
		}
	// If len(teams) % 3 == 1, then the top division and the last division have 2 teams.
	case 1:
		for i, j := 1, X; i < len(teams)+1; i++ {
			divisionToTeam[j] = append(divisionToTeam[j], i)
			if i%3 == 2 {
				j++
			}
		}
	}

	challenges := make(map[string]Challenge)

	// Generate a challenge for each team within the division.
	for division, divTeam := range divisionToTeam {
		for key, teamRank := range divTeam {
			fmt.Println(division, key, teamRank, teams[teamRank])
			switch key {
			case 0:
				var challenge Challenge
				challenge.Division = division
				challenge.Round = currentRound
				challenge.Defender = teams[teamRank]
				challenge.DefenderRank = teamRank
				challenge.Challenger = teams[teamRank+1]
				challenge.ChallengerRank = teamRank + 1
				challenge.Code = teamRank
				challenges[strconv.Itoa(challenge.Code)] = challenge

				challenge.Division = division
				challenge.Round = currentRound
				challenge.Defender = teams[teamRank]
				challenge.DefenderRank = teamRank
				challenge.Challenger = teams[teamRank+2]
				challenge.ChallengerRank = teamRank + 2
				challenge.Code = teamRank + 1
				challenges[strconv.Itoa(challenge.Code)] = challenge

			case 1:
				var challenge Challenge
				challenge.Division = division
				challenge.Round = currentRound
				challenge.Defender = teams[teamRank]
				challenge.DefenderRank = teamRank
				challenge.Challenger = teams[teamRank+1]
				challenge.ChallengerRank = teamRank + 1
				challenge.Code = teamRank + 1
				challenges[strconv.Itoa(challenge.Code)] = challenge
			}
		}
	}

	for i := 1; i < len(teams)+1; i++ {
		challenge := challenges[strconv.Itoa(i)]
		code := i
		fmt.Printf("Spladder#4 Div %s [%d-%d] %d位 %s vs %d位 %s \n", challenge.Division.String(), challenge.Round, code,
			challenge.ChallengerRank, challenge.Challenger, challenge.DefenderRank, challenge.Defender)
	}

	// Populate to Firestore.
	for matchCode, challenge := range challenges {
		key := challenge.Round.String() + "-" + matchCode
		_, err = tournament.Collection("challenges").Doc(key).Set(ctx, challenge)
		if err != nil {
			log.Printf("An error has occurred: %s", err)
		}
	}
}

func main() {
	// Initialize Firebase.
	ctx := context.Background()
	sa := option.WithCredentialsFile("./serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	tournament := client.Collection("tournaments").Doc("spladder4")

	// Get current raound.
	var currentRound Round
	fmt.Println("Enter the current round:")
	fmt.Scanln(&currentRound)

	_, err = tournament.Set(ctx, map[string]interface{}{
		"currentRound": currentRound,
	}, firestore.MergeAll)

	if err != nil {
		log.Fatalln("Error writing to firebase:", err)
	}

	teams := tournament.Collection("teams")
	ranking := tournament.Collection("ranking").Doc(currentRound.String())
	challenges := tournament.Collection("challenges").Where("Round", "==", currentRound).
		OrderBy("Code", firestore.Asc)

	var s string
	fmt.Println("Init? y/n")
	fmt.Scanln(&s)

	if s == "y" {
		// If this is a new tournament, then initialize the teams and the ranking.
		fmt.Println("Init teams? y/n")
		fmt.Scanln(&s)
		if s == "y" {
			err = InitTeams(ctx, teams, "spladder-teams.csv")
			if err != nil {
				log.Fatalln("Error initialising teams:", err)
			}
		}

		// Manually seed the initial ranking.
		fmt.Println("Init ranking? y/n")
		fmt.Scanln(&s)
		if s == "y" {
			err := InitRanking(ctx, teams, ranking)
			if err != nil {
				fmt.Println("Error initialising ranking: ", err)
			}
		}

		// Generate challenges for the first round.
		fmt.Println("Create challenges for initial round? y/n")
		fmt.Scanln(&s)
		if s == "y" {
			CreateChallenges(ctx, tournament)
		}
	}

	fmt.Println("Input scores for the current round? y/n")
	fmt.Scanln(&s)
	if s == "y" {
		InputScores(ctx, challenges)
	}

	fmt.Println("Generate new ranking based on the previous round scores? y/n")
	fmt.Scanln(&s)
	if s == "y" {
		GenerateRanking(ctx, ranking, challenges)
	}

	fmt.Println("Create challenges for current round? y/n")
	fmt.Scanln(&s)
	if s == "y" {
		CreateChallenges(ctx, tournament)
	}
}
