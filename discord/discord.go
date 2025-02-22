package discord

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/dimfu/kaido/config"
)

func Prompt(webhookUrl ...string) error {
	var s string
	cfg := config.GetConfig()

	if len(webhookUrl) == 0 {
		r := bufio.NewReader(os.Stdin)
		for {
			fmt.Fprint(os.Stderr, "Enter your discord webhook url: ")
			s, _ = r.ReadString('\n')
			if s != "" {
				break
			}
		}
	} else {
		s = webhookUrl[0]
	}

	val := strings.TrimSpace(s)
	if len(val) == 0 {
		return errors.New("webhook cannot be empty")
	}

	u, err := url.Parse(val)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("webhook must be a valid url")
	}

	client := &http.Client{}
	request, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.New("Discord webhook url is not valid")
	}

	cfg.DiscordWebhookURL = u.String()

	if err := cfg.Save(); err != nil {
		return err
	}

	return nil
}
