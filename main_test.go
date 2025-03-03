package main

import (
	"fmt"
	"reflect"
	"testing"
)

const unixTimeString = "1610768692"
const parsedUnixTime = "Sat, 16 Jan 2021 03:44:52 UTC"

var mockConfig = config{
	BuildNumber:                  "1",
	AppTitle:                     "Some app title",
	AppURL:                       "https://www.github.com/username/repo",
	BuildURL:                     "https://www.bitrise.io/some/app/build",
	BuildTime:                    unixTimeString,
	CommitMessageBody:            "Some commit message body",
	GitBranch:                    "master",
	Workflow:                     "master_branch",
	WebhookURL:                   "https://microsoft.com/some/webhook",
	CardTitle:                    "The heading for the card",
	SuccessThemeColor:            "FFFFFF",
	FailedThemeColor:             "000000",
	SectionTitle:                 "Git author name",
	SectionSubtitle:              "Commit message",
	SectionText:                  "Commit message body",
	SectionHeaderImage:           "",
	SectionImage:                 "https://www.example.com/image.png",
	SectionImageDescription:      "A description of the image",
	EnablePrimarySectionMarkdown: "no",
	EnableBuildFactsMarkdown:     "no",
	EnableDefaultActions:         "yes",
	EnableDebug:                  "no",
	RepoURL:                      "https://www.github.com/username/repo",
	Actions: `[
		{
			"text": "Some text",
			"targets": [
				{
					"uri": "www.google.com", 
					"os": "android"
				},
				{
					"uri": "www.google.com", 
					"os": "iOS"
				},
				{
					"uri": "www.google.com", 
					"os": "windows"
				}
			]
		}
	]`,
}

func TestGetValueForBuildStatus(t *testing.T) {
	const success = "success"
	const fail = "fail"

	successValue := getValueForBuildStatus(success, fail, true)
	failValue := getValueForBuildStatus(success, fail, false)

	if successValue != success {
		t.Errorf("Test failed: expected %v but input was %v", success, successValue)
	}
	if failValue != fail {
		t.Errorf("Test failed: expected %v but input was %v", fail, failValue)
	}
}

func TestOptionalUserValue(t *testing.T) {
	const defaultValue = "default value"
	const userValue = "user value"

	fallbackToDefault := optionalUserValue(defaultValue, "")
	userCustomValue := optionalUserValue(defaultValue, userValue)
	if fallbackToDefault != defaultValue {
		t.Errorf("Test failed: expected %v but input was %v", defaultValue, fallbackToDefault)
	}
	if userCustomValue != userValue {
		t.Errorf("Test failed: expected %v but input was %v", userValue, userCustomValue)
	}
}

func TestParseTimeString(t *testing.T) {

	var tests = []struct {
		input    config
		expected string
	}{
		// successful UTC
		{
			mockConfig,
			parsedUnixTime,
		},
		// successful local
		{
			config{
				BuildTime: unixTimeString,
				Timezone:  "Australia/Sydney",
			},
			"Sat, 16 Jan 2021 14:44:52 AEDT",
		},
		// invalid timezone, returns UTC
		{
			config{
				BuildTime: unixTimeString,
				Timezone:  "Bermuda Triangle",
			},
			parsedUnixTime,
		},
		// invalid `buildTime``
		{
			config{
				BuildTime: "unixTimeString",
			},
			string("Couldn't parse build time"),
		},
	}

	for _, test := range tests {
		if output := parseTimeString(test.input); output != test.expected {
			t.Errorf("Test failed: output was %v, expected %v", output, test.expected)
		}
	}
}

func TestBuildPrimarySection(t *testing.T) {
	var defaultValuesConfig = config{
		SectionTitle:                 "Some author",
		SectionSubtitle:              "A commit message",
		SectionText:                  "The commits message body",
		EnablePrimarySectionMarkdown: "no",
	}
	var tests = []struct {
		input    config
		expected AdaptiveCardBody
	}{
		{
			defaultValuesConfig,
			AdaptiveCardBody{
				Type: "Container",
				Items: []AdaptiveCardBody{
					{
						Type: "TextBlock",
						Text: defaultValuesConfig.SectionTitle,
						Weight: "Bolder",
					},
					{
						Type: "TextBlock",
						Text: defaultValuesConfig.SectionSubtitle,
						IsSubtle: true,
					},
					{
						Type: "TextBlock",
						Text: defaultValuesConfig.SectionText,
						Wrap: true,
						IsSubtle: true,
					},
				},
			},
		},
	}

	for _, test := range tests {
		if output := buildPrimarySection(test.input); !reflect.DeepEqual(output, test.expected) {
			t.Errorf("Test failed: config input was %v, expected %v", test.input, test.expected)
		}
	}
}

func TestBuildFactsSection(t *testing.T) {

	var tests = []struct {
		input          config
		isBuildSuccess bool
		expected       AdaptiveCardBody
	}{
		// Successful build
		{
			mockConfig,
			true,
			AdaptiveCardBody{
				Type: "FactSet",
				Facts: []AdaptiveCardFact{
					{
						Title: "Build Status",
						Value: fmt.Sprintf(`<span style="color:#%s">Success</span>`, mockConfig.SuccessThemeColor),
					},
					{
						Title: "Build Number",
						Value: mockConfig.BuildNumber,
					},
					{
						Title: "Git Branch",
						Value: mockConfig.GitBranch,
					},
					{
						Title: "Build Triggered",
						Value: parsedUnixTime,
					},
					{
						Title: "Workflow",
						Value: mockConfig.Workflow,
					},
				},
			},
		},
		// Failed build
		{
			mockConfig,
			false,
				AdaptiveCardBody{
				Type: "FactSet",
				Facts: []AdaptiveCardFact{
					{
						Title: "Build Status",
						Value: fmt.Sprintf(`<span style="color:#%s">Fail</span>`, mockConfig.FailedThemeColor),
					},
					{
						Title: "Build Number",
						Value: mockConfig.BuildNumber,
					},
					{
						Title: "Git Branch",
						Value: mockConfig.GitBranch,
					},
					{
						Title: "Build Triggered",
						Value: parsedUnixTime,
					},
					{
						Title: "Workflow",
						Value: mockConfig.Workflow,
					},
				},
			},
		},
	}
	for _, test := range tests {
		if output := buildFactsSection(test.input, test.isBuildSuccess); !reflect.DeepEqual(output, test.expected) {
			t.Errorf("Test failed: config input was %v, expected %v", test.input, test.expected)
		}
	}
}

func TestBuildImagesSection(t *testing.T) {
	var defaultValuesConfig = config{
		SectionImage:            "https://www.example.com/image.png",
		SectionImageDescription: "This is the image description",
	}
	var emptyDescriptionConfig = config{
		SectionImage: "https://www.example.com/image.png",
	}
	var emptyImageConfig = config{}

	var tests = []struct {
		input    config
		expected AdaptiveCardBody
	}{
		{
			defaultValuesConfig,
			AdaptiveCardBody{
				Type: "Container",
				Items: []AdaptiveCardBody{
					{
						Type: "Image",
						URL: defaultValuesConfig.SectionImage,
						AltText: defaultValuesConfig.SectionImageDescription,
					},
				},
			},
		},
		{
			emptyDescriptionConfig,
			AdaptiveCardBody{
				Type: "Container",
				Items: []AdaptiveCardBody{
					{
						Type: "Image",
						URL: emptyDescriptionConfig.SectionImage,
					},
				},
			},
		},
		{
			emptyImageConfig,
			AdaptiveCardBody{},
		},
	}

	for _, test := range tests {
		if output := buildImagesSection(test.input); !reflect.DeepEqual(output, test.expected) {
			t.Errorf("Test failed: config input was %v, expected %v", test.input, test.expected)
		}
	}
}

func TestNewMessage(t *testing.T) {

	var buildSuccessFacts = AdaptiveCardBody{
		Type: "FactSet",
		Facts: []AdaptiveCardFact{
			{
				Title: "Build Status",
				Value: fmt.Sprintf(`<span style="color:#%s">Success</span>`, mockConfig.SuccessThemeColor),
			},
			{
				Title: "Build Number",
				Value: mockConfig.BuildNumber,
			},
			{
				Title: "Git Branch",
				Value: mockConfig.GitBranch,
			},
			{
				Title: "Build Triggered",
				Value: parsedUnixTime,
			},
			{
				Title: "Workflow",
				Value: mockConfig.Workflow,
			},
		},
	}

	var buildFailedFacts = AdaptiveCardBody{
		Type: "FactSet",
		Facts: []AdaptiveCardFact{
			{
				Title: "Build Status",
				Value: fmt.Sprintf(`<span style="color:#%s">Fail</span>`, mockConfig.FailedThemeColor),
			},
			{
				Title: "Build Number",
				Value: mockConfig.BuildNumber,
			},
			{
				Title: "Git Branch",
				Value: mockConfig.GitBranch,
			},
			{
				Title: "Build Triggered",
				Value: parsedUnixTime,
			},
			{
				Title: "Workflow",
				Value: mockConfig.Workflow,
			},
		},
	}

	var primarySection = AdaptiveCardBody{
		Type: "Container",
		Items: []AdaptiveCardBody{
			{
				Type: "TextBlock",
				Text: mockConfig.SectionTitle,
				Weight: "Bolder",
			},
			{
				Type: "TextBlock",
				Text: mockConfig.SectionSubtitle,
				IsSubtle: true,
			},
			{
				Type: "TextBlock",
				Text: mockConfig.SectionText,
				Wrap: true,
				IsSubtle: true,
			},
		},
	}

	var imagesSection = AdaptiveCardBody{
		Type: "Container",
		Items: []AdaptiveCardBody{
			{
				Type: "Image",
				URL: mockConfig.SectionImage,
				AltText: mockConfig.SectionImageDescription,
			},
		},
	}

	var buildSuccessMessage = AdaptiveCard{
		Type: "AdaptiveCard",
		Version: "1.2",
		Body: []AdaptiveCardBody{
			{
				Type: "TextBlock",
				Text: optionalUserValue(mockConfig.AppTitle, mockConfig.CardTitle),
				Size: "Large",
				Weight: "Bolder",
			},
			{
				Type: "TextBlock",
				Text: getValueForBuildStatus(
					fmt.Sprintf("%s #%s succeeded", mockConfig.AppTitle, mockConfig.BuildNumber),
					fmt.Sprintf("%s #%s failed", mockConfig.AppTitle, mockConfig.BuildNumber),
					true,
				),
				Size: "Medium",
				Weight: "Bolder",
				Color: getValueForBuildStatus("Good", "Attention", true),
			},
			primarySection,
			imagesSection,
			buildSuccessFacts,
		},
		Actions: []AdaptiveCardAction{
			{
				Type: "Action.OpenUrl",
				Title: "Go To Repo",
				URL: mockConfig.RepoURL,
			},
			{
				Type: "Action.OpenUrl",
				Title: "Go To Build",
				URL: mockConfig.BuildURL,
			},
			{
				Type: "Action.OpenUrl",
				Title: "Some text",
				URL: "www.google.com",
			},
		},
	}

	var buildFailedMessage = AdaptiveCard{
		Type: "AdaptiveCard",
		Version: "1.2",
		Body: []AdaptiveCardBody{
			{
				Type: "TextBlock",
				Text: optionalUserValue(mockConfig.AppTitle, mockConfig.CardTitle),
				Size: "Large",
				Weight: "Bolder",
			},
			{
				Type: "TextBlock",
				Text: getValueForBuildStatus(
					fmt.Sprintf("%s #%s succeeded", mockConfig.AppTitle, mockConfig.BuildNumber),
					fmt.Sprintf("%s #%s failed", mockConfig.AppTitle, mockConfig.BuildNumber),
					false,
				),
				Size: "Medium",
				Weight: "Bolder",
				Color: getValueForBuildStatus("Good", "Attention", false),
			},
			primarySection,
			imagesSection,
			buildFailedFacts,
		},
		Actions: []AdaptiveCardAction{
			{
				Type: "Action.OpenUrl",
				Title: "Go To Repo",
				URL: mockConfig.RepoURL,
			},
			{
				Type: "Action.OpenUrl",
				Title: "Go To Build",
				URL: mockConfig.BuildURL,
			},
			{
				Type: "Action.OpenUrl",
				Title: "Some text",
				URL: "www.google.com",
			},
		},
	}

	var tests = []struct {
		input          config
		isBuildSuccess bool
		expected       AdaptiveCard
	}{
		{
			mockConfig,
			true,
			buildSuccessMessage,
		},
		{
			mockConfig,
			false,
			buildFailedMessage,
		},
	}
	for _, test := range tests {
		if output := newMessage(test.input, test.isBuildSuccess); !reflect.DeepEqual(output, test.expected) {
			t.Errorf("Test failed: config input was %v, expected %v", test.input, test.expected)
		}
	}

}
