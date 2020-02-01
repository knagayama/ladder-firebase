package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"

	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Loads csv file given by path and uploads it to Firestore.
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
			log.Printf("Error: %s", err)
		}
	}
}

func InitRanking(ctx context.Context, client *firestore.Client) error {
	// Read from the teams document.
	var currentRound string
	fmt.Println("Enter the current round: ")
	fmt.Scanln(&currentRound)
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
		_, err = client.Collection("ranking").Doc(currentRound).Set(ctx, map[string]interface{}{
			rank: data["name"],
		}, firestore.MergeAll)
		if err != nil {
			log.Printf("An error has occurred: %s", err)
		}
	}
	return nil
}

func CreateChallenges(ctx context.Context, client *firestore.Client) {
	// Do nothing.
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

	var s string
	fmt.Println("Init Teams? y/n")
	fmt.Scanln(&s)
	if s == "y" {
		InitTeams(ctx, client, "spladder-teams.csv")
	}

	fmt.Println("Init Ranking? y/n")
	fmt.Scanln(&s)
	if s == "y" {
		err := InitRanking(ctx, client)
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println("Create Challenges? y/n")
	fmt.Scanln(&s)
	if s == "y" {
		CreateChallenges(ctx, client)
	}
}
