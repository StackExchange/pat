package silence

// Cobbled together from bosun.org/cmd/silence/main.go
// and gitlab.stackexchange.com/sre/patcher/cmd/client/actions/actions.go

// TODO(tlim):  This should be part of Bosun.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strings"
	"time"
)

type SilenceRequest struct {
	User    string `json:"user"`
	Start   string `json:"start"`
	End     string `json:"end"`
	Tags    string `json:"tags"`
	Alert   string `json:"alert"`
	Message string `json:"message"`
	Confirm string `json:"confirm"`
	Forget  string `json:"forget"`
}

// EasySilence inserts a Bosun silence with some reasonable defaults.
func EasySilence(alert, duration, message string, hosts []string) (string, error) {
	if len(hosts) == 0 {
		// Add our own hostname (the shortname).
		// (If you want the long name, file a PR.)
		h, err := os.Hostname()
		if err == nil {
			h = strings.SplitN(h, ".", 2)[0]
			hosts = append(hosts, h)
		}
	}

	now := time.Now().UTC()
	d, err := time.ParseDuration(duration)
	if err != nil {
		log.Fatal(err)
	}
	end := now.Add(d)

	s := &SilenceRequest{
		Start:   now.Format("2006-01-02 15:04:05 MST"),
		End:     end.Format("2006-01-02 15:04:05 MST"),
		Tags:    "host=" + strings.Join(hosts, "|"),
		Alert:   alert,
		Message: message,
		Confirm: "confirm",
	}
	return SetSilence("", s)
}

// SetSilence inserts a Bosun silence, filling in some reasonable defaults
// if needed. Returns a summary string of what was done.
func SetSilence(bosunhost string, s *SilenceRequest) (string, error) {

	if bosunhost == "" {
		bosunhost = "bosun"
	}

	if s.User == "" {
		u, err := user.Current()
		if err != nil {
			return "ERROR: Not running on an OS that supports usernames", err
		}
		username := u.Username
		userParts := strings.Split(username, "\\")
		if len(userParts) > 1 {
			username = userParts[1]
		}
		sudo := os.Getenv("SUDO_USER")
		if sudo != "" {
			username = sudo
		}
		s.User = username
	}

	if s.Start == "" {
		s.Start = time.Now().UTC().Format("2006-01-02 15:04:05 MST")
	}

	b, err := json.Marshal(s)
	if err != nil {
		return fmt.Sprintf("Marshal failed: %#v", s), err
	}
	u := url.URL{
		Scheme: "https",
		Host:   bosunhost,
		Path:   "/api/silence/set",
	}
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "POST ERROR", err
	}
	c, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "ERROR ReadAll", err
	}
	if resp.StatusCode != 200 {
		return "STATUS", fmt.Errorf("%s", c)
	}
	if s.Alert == "" {
		s.Alert = "None"
	}
	if s.Tags == "" {
		s.Tags = "None"
	}
	if s.Message == "" {
		s.Message = "None"
	}
	return fmt.Sprintf("Created silence: Start: %s, End: %s, Tags: %s, Alert: %s, Message: %s\n", s.Start, s.End, s.Tags, s.Alert, s.Message), nil
}
