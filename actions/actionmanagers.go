package actions

import (
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/state-parser/actions/action"
)

var ingestionMap map[string]chan action.Payload

type actionManager struct {
	Name       string
	Action     Action
	ActionChan chan action.Payload
}

func init() {
	ingestionMap = make(map[string]chan action.Payload)
}

func StartActionManagers() {
	var actionList []string
	for name := range Actions {
		actionList = append(actionList, name)
	}

	log.L.Infof("Starting action scheduler. Executing action types: %v", actionList)

	// build each of the individual action managers
	for name, act := range Actions {
		ingestionMap[name] = make(chan action.Payload, 2000) // TODO make this size configurable?

		manager := &actionManager{
			Name:       name,
			Action:     act,
			ActionChan: ingestionMap[name],
		}
		go manager.start()
	}
}

// Execute queues a slice of actions to be executed.
func Execute(actions []action.Payload) {
	for i := range actions {
		if _, ok := ingestionMap[actions[i].Type]; ok {
			ingestionMap[actions[i].Type] <- actions[i]
		}
	}
}

func (a *actionManager) start() {
	// TODO scale number of action managers as size of payload chan increases?
	for act := range a.ActionChan {
		go func(action action.Payload) {
			result := a.Action.Execute(action)
			if result.Error != nil {
				log.L.Warnf("failed to execute %s action: %s", result.Payload.Type, result.Error.String())
			}
		}(act)
	}
}