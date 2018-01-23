package base

//AlertFactory corresponds to a struct that is run to generate alerts.
type AlertFactory interface {
	Run(loggingLevel int) (int, []Alert, error)
}

//Alerts are the types that get generated. For now there's only Slack alerts, the intention is to add at least E-Mail alerts and MOM alerts
type Alert struct {
	AlertType string //The type of alert, see constants.go to see the available values
	Content   []byte //The content of the alert to send
}

type SlackAlert struct {
	Markdown bool   `json:"mrkdwn"`
	Text     string `json:"text'`
}