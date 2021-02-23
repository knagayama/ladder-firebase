package announce

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/slack-go/slack"
)

type FirestoreEvent struct {
	OldValue   FirestoreValue `json:"oldValue"`
	Value      FirestoreValue `json:"value"`
	UpdateMask struct {
		Fieldpaths []string `json:"fieldPaths"`
	} `json:"updateMask`
}

type FirestoreValue struct {
	CreateTime time.Time   `json:"createTime"`
	Fields     interface{} `json:"fields"`
	Name       string      `json:"name"`
	UpdateTime time.Time   `json:"updateTime"`
}

func SendURLToSlack(ctx context.Context, e FirestoreEvent) error {
	f := e.Value.Fields.(map[string]interface{})
	g := f["StreamURL"].(map[string]interface{})
	newURL := g["stringValue"]
	k := e.OldValue.Fields.(map[string]interface{})
	g = k["StreamURL"].(map[string]interface{})
	oldURL := g["stringValue"]

	if newURL != nil && newURL != "" && newURL != oldURL {
		g = f["Round"].(map[string]interface{})
		currentRound := g["integerValue"]
		g = f["Code"].(map[string]interface{})
		code := g["integerValue"]
		g = f["Challenger"].(map[string]interface{})
		challenger := g["stringValue"]
		g = f["ChallengerRank"].(map[string]interface{})
		crank := g["integerValue"]
		g = f["Defender"].(map[string]interface{})
		defender := g["stringValue"]
		g = f["DefenderRank"].(map[string]interface{})
		drank := g["integerValue"]
		g = f["Streamer"].(map[string]interface{})
		streamer := g["stringValue"]
		message := fmt.Sprintf(
			"[配信URL] %+v さんによる [%+v-%+v] %+v (%+v位) vs %+v (%+v位) の配信があがったぞ！クリッククリックぅ→ %s\n",
			streamer, currentRound, code, challenger, crank, defender, drank, newURL)
		log.Printf(message)
		slackURL := "https://hooks.slack.com/services/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"
		var whm slack.WebhookMessage
		whm.Text = message
		whm.Username = "ladder_bot_for_fireba"
		err := slack.PostWebhook(slackURL, &whm)
		if err != nil {
			return err
		}
	}
	return nil
}
