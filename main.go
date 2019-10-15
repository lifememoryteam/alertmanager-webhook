package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/ak1ra24/alermanager-webhook/githubapi"
	"gopkg.in/yaml.v2"
)

var (
	githubinfo GithubInfo
)

type AlerManager struct {
	Receiver string  `json:"receiver"`
	Status   string  `json:"status"`
	Alerts   []Alert `json:"alerts"`
	// GroupLabels GroupLabel `json:"groupLabels"`
}

type Alert struct {
	Status string `json:"status"`
	Labels struct {
		AlertName string `json:"alertname"`
		Instance  string `json:"instance"`
		Job       string `json:"job"`
		Name      string `json:"name"`
	} `json:"labels"`
	Annotations struct {
		Description string `json:"description"`
		Summary     string `json:"summary"`
	} `json:"annotations"`
	StartsAt    string `json:"startsAt"`
	URL         string `json:"generatorURL"`
	Fingerprint string `json:"fingerprint"`
}

type AlertIssue struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type GithubInfo struct {
	Github struct {
		Token      string `yaml:"token"`
		Repository struct {
			Owner string `yaml:"owner"`
			Name  string `yaml:"name"`
		} `yaml:"repository"`
	} `yaml:"github"`
}

func ReadYaml(filename string) (GithubInfo, error) {

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return GithubInfo{}, err
	}

	var info GithubInfo
	err = yaml.Unmarshal(buf, &info)
	if err != nil {
		return GithubInfo{}, err
	}

	return info, nil
}

func ParseAlertManagerJson() error {

	bytes, err := ioutil.ReadFile("./NodeInstanceDown_Resolve.json")
	if err != nil {
		return err
	}

	var alertManager AlerManager

	if err := json.Unmarshal(bytes, &alertManager); err != nil {
		return err
	}

	for _, alert := range alertManager.Alerts {
		// alert title
		alertIssue := fmt.Sprintf("[%s] %s", alert.Labels.AlertName, alert.Labels.Instance)
		fmt.Println(alertIssue)
		// alert description
		alertIssueDescription := fmt.Sprintf("`%s`\n\n```\n%s\n```\n", alert.Annotations.Summary, alert.Annotations.Description)
		fmt.Println(alertIssueDescription)
	}

	return nil
}

func DumpJsonHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := make([]byte, length)
	length, err = r.Body.Read(body)
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// var jsonBody map[string]interface{}
	var alertManager AlerManager
	err = json.Unmarshal(body[:length], &alertManager)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Printf("%v\n", alertManager)
	for _, alert := range alertManager.Alerts {
		// alert title
		alertIssueTitle := fmt.Sprintf("[%s] %s (%s)", alert.Labels.AlertName, alert.Labels.Instance, alert.Fingerprint)
		// alert description
		alertIssueDescription := fmt.Sprintf("### %s [%s] \n\n```\n%s\n```\n", alert.Annotations.Summary, alert.StartsAt, alert.Annotations.Description)

		alertIssue := AlertIssue{Title: alertIssueTitle, Description: alertIssueDescription, Status: alert.Status}

		res, err := json.Marshal(alertIssue)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		labels := []string{alertIssue.Status}

		if err := IssueHandle(alertIssue.Title, alertIssue.Description, labels); err != nil {
			log.Fatal(err)
		}

		log.Printf("AlertIssue: %v, Labels: %s\n", alertIssue, labels)

		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
		w.WriteHeader(http.StatusOK)
	}
}

func IssueHandle(title, body string, labels []string) error {

	fmt.Println(githubinfo.Github.Token)
	g := githubapi.NewClient(githubinfo.Github.Repository.Owner, githubinfo.Github.Repository.Name, githubinfo.Github.Token)

	log.Printf("Title: %s, Body: %s, Labels: %s\n", title, body, labels)
	issueNum, dup, err := g.DuplicateIssueTitle(title)
	if err != nil {
		return err
	}

	if !dup {
		log.Printf("Create Issue\n")
		if err := g.CreateIssue(title, body, labels); err != nil {
			return err
		}
	} else {
		log.Printf("Comment\n")
		log.Printf("Issue Number: %d, Duplicate Title: %t\n", issueNum, dup)
		if err := g.CreateIssueComment(issueNum, body); err != nil {
			return err
		}

		if err := g.ReplaceLabel(issueNum, labels); err != nil {
			return err
		}
	}

	return nil
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func main() {
	info, err := ReadYaml("webhook.yaml")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%v\n", info)

	githubinfo.Github.Token = info.Github.Token
	githubinfo.Github.Repository.Owner = info.Github.Repository.Owner
	githubinfo.Github.Repository.Name = info.Github.Repository.Name

	http.HandleFunc("/api/health", HealthHandler)
	http.HandleFunc("/api/webhook", DumpJsonHandler)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
