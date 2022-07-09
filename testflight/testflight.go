package testflight

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func ParseTestflightApp(appUrl string) (name string, status int, err error) {

	parseUrl, err := url.Parse(appUrl)
	if err != nil {
		err = errors.Wrap(err, "Error ParseTestflightApp parse appUrl")
		return
	}

	if parseUrl.Host != "testflight.apple.com" {
		err = errors.Errorf("Error ParseTestflightApp parseUrl.Host:%v", parseUrl.Host)
		return
	}

	client := http.Client{}

	req, err := http.NewRequest("GET", appUrl, nil)

	if err != nil {
		err = errors.Wrap(err, "Error ParseTestflightApp http.NewRequest")
		return
	}

	resp, err := client.Do(req)

	if err != nil {
		err = errors.Wrap(err, "Error ParseTestflightApp client.Do")
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New("Error resp.StatusCode != 200")
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		err = errors.New("Error goquery.NewDocumentFromReader")
		return
	}

	head := doc.Find("head")

	title := head.Find("title").Text()

	regExp := regexp.MustCompile("Join the (.*) beta - TestFlight - Apple")
	matchSlice := regExp.FindStringSubmatch(title)
	if len(matchSlice) < 1 {
		status = 2
	} else {
		statusText := doc.Find(".beta-status span").Text()
		if strings.Contains(statusText, "full") {
			status = 3
		} else {
			status = 1
		}

		name = matchSlice[1]
	}

	return
}
