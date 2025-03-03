package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/colorstring"
)

var buildSucceeded = os.Getenv("BITRISE_BUILD_STATUS") == "0"

func getValueForBuildStatus(ifSuccess, ifFailed string, buildSucceeded bool) string {
	if buildSucceeded || ifFailed == "" {
		return ifSuccess
	}
	return ifFailed
}

func optionalUserValue(defaultValue, userValue string) string {
	if userValue == "" {
		return defaultValue
	}
	return userValue
}

func valueOptionToBool(userValue string) bool {
	return userValue == "yes"
}

func parseTimeString(cfg config) string {
	var timeAtLoc time.Time
	i, err := strconv.ParseInt(cfg.BuildTime, 10, 64)
	if err != nil {
		return string("Couldn't parse build time")
	}
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		fmt.Println(colorstring.Redf("\n%s", err))
		fmt.Println(colorstring.Cyan("\nExporting time in UTC...\n"))

		timeAtLoc = time.Unix(i, 0).In(time.UTC)
		return timeAtLoc.Format(time.RFC1123)
	}
	timeAtLoc = time.Unix(i, 0).In(loc)

	return timeAtLoc.Format(time.RFC1123)
}

func newMessage(cfg config, buildSuccessful bool) AdaptiveCard {
	message := AdaptiveCard{}
	message.Type = "AdaptiveCard"
	message.Version = "1.2"
	message.Body = []AdaptiveCardBody{
		{
			Type: "TextBlock",
			Text: optionalUserValue(cfg.AppTitle, cfg.CardTitle),
			Size: "Large",
			Weight: "Bolder",
		},
		{
			Type: "TextBlock",
			Text: getValueForBuildStatus(
				fmt.Sprintf("%s #%s succeeded", cfg.AppTitle, cfg.BuildNumber),
				fmt.Sprintf("%s #%s failed", cfg.AppTitle, cfg.BuildNumber),
				buildSuccessful,
			),
			Size: "Medium",
			Weight: "Bolder",
			Color: getValueForBuildStatus("Good", "Attention", buildSuccessful),
		},
		buildPrimarySection(cfg),
		buildImagesSection(cfg),
		buildFactsSection(cfg, buildSuccessful),
	}
	message.Actions = buildActions(cfg, buildSuccessful)

	return message
}

// Builds the primary section of the AdaptiveCard content
func buildPrimarySection(cfg config) AdaptiveCardBody {
	return AdaptiveCardBody{
		Type: "Container",
		Items: []AdaptiveCardBody{
			{
				Type: "TextBlock",
				Text: cfg.SectionTitle,
				Weight: "Bolder",
			},
			{
				Type: "TextBlock",
				Text: cfg.SectionSubtitle,
				IsSubtle: true,
			},
			{
				Type: "TextBlock",
				Text: cfg.SectionText,
				Wrap: true,
				IsSubtle: true,
			},
		},
	}
}

// Builds a Container containing a list of Image
func buildImagesSection(cfg config) AdaptiveCardBody {
	if cfg.SectionImage == "" {
		return AdaptiveCardBody{}
	}
	return AdaptiveCardBody{
		Type: "Container",
		Items: []AdaptiveCardBody{
			{
				Type: "Image",
				URL: cfg.SectionImage,
				AltText: cfg.SectionImageDescription,
			},
		},
	}
}

// Builds a Container containing a list of Fact related to build status
func buildFactsSection(cfg config, buildSuccessful bool) AdaptiveCardBody {
	return AdaptiveCardBody{
		Type: "FactSet",
		Facts: []AdaptiveCardFact{
			{
				Title: "Build Status",
				Value: getValueForBuildStatus(
					fmt.Sprintf(`<span style="color:#%s">Success</span>`, cfg.SuccessThemeColor),
					fmt.Sprintf(`<span style="color:#%s">Fail</span>`, cfg.FailedThemeColor),
					buildSuccessful,
				),
			},
			{
				Title: "Build Number",
				Value: cfg.BuildNumber,
			},
			{
				Title: "Git Branch",
				Value: cfg.GitBranch,
			},
			{
				Title: "Build Triggered",
				Value: parseTimeString(cfg),
			},
			{
				Title: "Workflow",
				Value: cfg.Workflow,
			},
		},
	}
}

func buildActions(cfg config, buildSuccessful bool) []AdaptiveCardAction {
	actions := []AdaptiveCardAction{}
	if valueOptionToBool(cfg.EnableDefaultActions) {
		actions = append(actions, AdaptiveCardAction{
			Type: "Action.OpenUrl",
			Title: "Go To Repo",
			URL: cfg.RepoURL,
		})
		actions = append(actions, AdaptiveCardAction{
			Type: "Action.OpenUrl",
			Title: "Go To Build",
			URL: cfg.BuildURL,
		})
	}
	customActions := parseActions(cfg.Actions)
	for _, action := range customActions {
		actions = append(actions, AdaptiveCardAction{
			Type: "Action.OpenUrl",
			Title: action.Text,
			URL: action.Targets[0].URI,
		})
	}
	return actions
}

// postMessage sends a message to a channel.
func postMessage(webhookURL string, msg AdaptiveCard, debugEnabled bool) error {
	b, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return err
	}
	if debugEnabled {
		log.Print(colorstring.Yellowf("\nRequest to Microsoft Teams:\n%s", b))
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to send the request: %s", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); err == nil {
			err = cerr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("server error: %s, failed to read response: %s", resp.Status, err)
		}
		return fmt.Errorf("server error: %s, response: %s", resp.Status, body)
	}

	return nil
}

func main() {
	var cfg config
	if err := stepconf.Parse(&cfg); err != nil {
		log.Fatalf("Error: %s\n", err)
	}
	stepconf.Print(cfg)

	message := newMessage(cfg, buildSucceeded)
	if err := postMessage(cfg.WebhookURL, message, valueOptionToBool(cfg.EnableDebug)); err != nil {
		log.Fatalf("Error: %s", err)
	}

	fmt.Println(colorstring.Cyan("\nMessage successfully sent!"))
}
