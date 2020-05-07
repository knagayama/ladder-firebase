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

// Challenge holds data for a challenge.
type Challenge struct {
	Round           int
	Code            int
	Challenger      string
	ChallengerRank  int
	ChallengerScore int
	Defender        string
	DefenderRank    int
	DefenderScore   int
	Division        string
}

// InitTeams loads a local csv file given by path and uploads it to Firestore.
func InitTeams(ctx context.Context, client *firestore.Client, path string) {
	// Load team CSV file.
	csvfile, err := os.Open(path)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	teamsDoc := make([]map[string]string, 0)

	r := csv.NewReader(csvfile)
	teams, err := r.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}
	for _, team := range teams {
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
		_, err := client.Collection("teams").Doc(name).Set(ctx, doc)
		if err != nil {
			log.Printf("Error writing to teams collection: %s", err)
		}
	}
}

// InitRanking initializes the ranking based on user input.
func InitRanking(ctx context.Context, client *firestore.Client, currentRound int) error {
	// Read from the teams document.
	var rank string

	iter := client.Collection("teams").Documents(ctx)
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
		_, err = client.Collection("ranking").Doc(strconv.Itoa(currentRound)).Set(ctx, map[string]interface{}{
			rank: data["name"],
		}, firestore.MergeAll)
		if err != nil {
			log.Printf("An error has occurred: %s", err)
		}
	}
	return nil
}

// InputScores uploads challenge scores based on user input.
func InputScores(ctx context.Context, client *firestore.Client, currentRound int) {
	// Read challenges for the current round and input the scores.
	dsnap, err := client.Collection("challenges").Doc(strconv.Itoa(currentRound)).Get(ctx)
	if err != nil {
		log.Fatalln("Error reading challenges from Firestore: ", err)
	}

	data := dsnap.Data()

	for i := 1; i <= len(data); i++ {
		challenge := data[strconv.Itoa(i)].(map[string]interface{})
		fmt.Printf("%d-%d Div %s: %s (rank %d) vs %s (rank %d)\n", challenge["Round"], challenge["Code"],
			challenge["Division"], challenge["Challenger"], challenge["ChallengerRank"],
			challenge["Defender"], challenge["DefenderRank"])
		fmt.Printf("Input score for challenger %s: ", challenge["Challenger"])
		var cs, ds int
		fmt.Scanf("%d", &cs)
		fmt.Printf("Input score for defender %s: ", challenge["Defender"])
		fmt.Scanf("%d", &ds)
		challenge["ChallengerScore"] = cs
		challenge["DefenderScore"] = ds
		_, err = client.Collection("challenges").Doc(strconv.Itoa(currentRound)).Set(ctx, map[string]interface{}{
			strconv.Itoa(i): challenge,
		}, firestore.MergeAll)
		if err != nil {
			log.Printf("Error occurred writing to Firestore: %s", err)
		}
		fmt.Println("Written to firebase.")
	}
}

// GenerateRanking generates ranking for the current round based on the last challenge scores.
func GenerateRanking(ctx context.Context, client *firestore.Client, currentRound int) {
	// Read the scores from last challenge scores.

	// Identify the winner and the loser for each division.

	// Swap the ranking for loser of the upper division and the winner of the lower division.
}

// CreateChallenges generate challenges based on the current team ranking and uploads it to Firestore.
func CreateChallenges(ctx context.Context, client *firestore.Client, currentRound int) {
	// Get ranking from current round.
	teams := make(map[int]string)
	dsnap, err := client.Collection("ranking").Doc(strconv.Itoa(currentRound)).Get(ctx)
	if err != nil {
		log.Fatalln("Error reading ranking from Firestore: ", err)
	}

	data := dsnap.Data()

	for key, value := range data {
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

	divisions := []string{"X", "S+ Upper", "S+ Lower", "S Upper", "S Middle", "S Lower",
		"A+ Upper", "A+ Middle", "A+ Lower", "A Upper", "A Middle", "A Lower",
		"B+ Upper", "B+ Middle", "B+ Lower", "B Upper", "B Middle", "B Lower",
		"C+ Upper", "C+ Middle", "C+ Lower", "C Upper", "C Middle", "C Lower"}
	divisionToTeam := make(map[string][]int)

	switch len(teams) % 3 {
	// If len(teams) % 3 == 0, then all divisions have 3 teams.
	// If len(teams) % 3 == 2, then last division has 2 teams.
	case 0, 2:
		for i, j := 1, 0; i < len(teams)+1; i++ {
			divisionToTeam[divisions[j]] = append(divisionToTeam[divisions[j]], i)
			if i%3 == 0 {
				j++
			}
		}
	// If len(teams) % 3 == 1, then the top division and the last division have 2 teams.
	case 1:
		for i, j := 1, 0; i < len(teams)+1; i++ {
			divisionToTeam[divisions[j]] = append(divisionToTeam[divisions[j]], i)
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
		fmt.Println("Spladder#4 Division", challenge.Division, " [", currentRound, "-", code, "]", challenge.ChallengerRank, "位", challenge.Challenger, "vs", challenge.DefenderRank, "位", challenge.Defender)
	}

	// Populate to Firestore.
	for matchCode, challenge := range challenges {
		_, err = client.Collection("challenges").Doc(strconv.Itoa(currentRound)).Set(ctx, map[string]interface{}{
			matchCode: challenge,
		}, firestore.MergeAll)
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

	// Get current raound.
	var currentRound int
	fmt.Println("Enter the current round: ")
	fmt.Scanln(&currentRound)

	var s string
	fmt.Println("Init? y/n")
	fmt.Scanln(&s)

	if s == "y" {
		// If this is a new tournament, then initialize the teams and the ranking.
		fmt.Println("Init teams? y/n")
		fmt.Scanln(&s)
		if s == "y" {
			InitTeams(ctx, client, "spladder-teams.csv")
		}

		// Manually seed the initial ranking.
		fmt.Println("Init ranking? y/n")
		fmt.Scanln(&s)
		if s == "y" {
			err := InitRanking(ctx, client, currentRound)
			if err != nil {
				fmt.Println("Error initialising ranking: ", err)
			}
		}

		// Generate challenges for the first round.
		fmt.Println("Create challenges for initial round? y/n")
		fmt.Scanln(&s)
		if s == "y" {
			CreateChallenges(ctx, client, currentRound)
		}
	}

	fmt.Println("Input scores for the current round? y/n")
	fmt.Scanln(&s)
	if s == "y" {
		InputScores(ctx, client, currentRound)
	}

	fmt.Println("Generate new ranking based on the previous round scores? y/n")
	fmt.Scanln(&s)
	if s == "y" {
		GenerateRanking(ctx, client, currentRound)
	}

	fmt.Println("Create challenges for current round? y/n")
	fmt.Scanln(&s)
	if s == "y" {
		CreateChallenges(ctx, client, currentRound)
	}
}
