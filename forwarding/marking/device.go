package marking

import (
	"fmt"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/state-parsing/forwarding"
)

//toMark is the list of rooms, There may be one or more of them
//secondaryAlertType is the type of alert marking as (e.g. heartbeat)
//secondaryAlertData is the data to be filled there (e.g. last-heartbeat-received, etc)
func MarkDevicesAsAlerting(toMark []string, secondaryAlertType string, secondaryAlertData map[string]interface{}) {
	//build our general alerting
	alerting := forwarding.StateDistribution{
		Key:   "alerting",
		Value: true,
	}

	secondaryAlertValue := make(map[string]interface{})
	secondaryAlertValue[secondaryAlertType] = secondaryAlertData

	//bulid our specifc alert
	secondaryAlert := forwarding.StateDistribution{
		Key:   "alerts",
		Value: secondaryAlertValue,
	}

	//ship it off to go with the rest
	for i := range toMark {
		log.L.Infof("Marking %s as alerting", toMark[i])
		forwarding.SendToStateBuffer(alerting, toMark[i], "device")
		forwarding.SendToStateBuffer(secondaryAlert, toMark[i], "device")
	}
}

func MarkDevicesAsNotHeartbeatAlerting(deviceIDs []string) {
	secondaryData := make(map[string]map[string]interface{})
	secondaryData["lost-heartbeat"] = make(map[string]interface{})

	secondaryData["lost-heartbeat"]["alerting"] = false
	secondaryData["lost-heartbeat"]["message"] = fmt.Sprintf("Alert cleared at %s", time.Now().Format(time.RFC3339))

	secondaryStatus := forwarding.StateDistribution{
		Key:   "alerts",
		Value: secondaryData,
	}

	alertingStatus := forwarding.StateDistribution{
		Key:   "alerting",
		Value: false,
	}

	for _, id := range deviceIDs {
		log.L.Info("Marking %s as not alerting", id)
		forwarding.SendToStateBuffer(secondaryStatus, id, "device")
		forwarding.SendToStateBuffer(alertingStatus, id, "device")
	}
}
