package announce

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/slack-go/slack"
	"google.golang.org/api/iterator"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
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

var (
	client *firestore.Client
)

func init() {
	ctx := context.Background()
	var err error
	client, err = firestore.NewClient(ctx, "splathon-ladder")
	if err != nil {
		log.Fatal(err)
	}
}

func SendMatchesToSlack(ctx context.Context, m PubSubMessage) error {
	// Initialise Slack.
	// api := slack.New("xoxb-16607544897-1147123322855-DIDH9Uh0NZ0uzO3QKh58HQA1")

	tournament := client.Collection("tournaments").Doc("spladder5")

	tdoc, err := tournament.Get(ctx)
	if err != nil {
		return err
	}

	currentRound, err := tdoc.DataAt("currentRound")
	if err != nil {
		return err
	}

	iter := tournament.Collection("challenges").Where("Round", "==", currentRound).OrderBy(
		"Date", firestore.Asc).Documents(ctx)
	message := "@channel 本日のお品書きはこちら！\n"
	matches := 0
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		data := doc.Data()
		if err != nil {
			return err
		}
		var date time.Time
		if val, ok := data["Date"]; ok {
			date = val.(time.Time)
		}
		diff := date.Local().Sub(time.Now()).Hours() / 24
		if diff > 1 {
			break
		}
		if diff >= 0 && diff <= 1 {
			var division Division
			division = Division(int(data["Division"].(int64)))
			jst := time.FixedZone("Asia/Tokyo", 9*60*60)
			strt := date.In(jst).Format("2006-01-02 15:04") + " "
			text := fmt.Sprintf("[%d-%d] Div %s: %s (%d位) vs %s (%d位)\n", currentRound, data["Code"],
				division.String(), data["Challenger"], data["ChallengerRank"], data["Defender"],
				data["DefenderRank"])
			message += strt
			message += text
			matches++
		}
	}
	if matches > 0 {
		log.Printf(message)
		url := "https://hooks.slack.com/services/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"
		var whm slack.WebhookMessage
		whm.Text = message
		whm.Username = "ladder_bot_for_fireba"
		err := slack.PostWebhook(url, &whm)
		if err != nil {
			return err
		}
	}
	client.Close()
	return nil
}
