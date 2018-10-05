package eventbased

import (
	"os"
	"strings"
	"time"

	"github.com/byuoitav/common/events"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	v2 "github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/byuoitav/state-parser/actions/action"
	"github.com/byuoitav/state-parser/state/cache"
	"github.com/fatih/color"
)

var (
	// APIForward is the url to forward events to
	APIForward = os.Getenv("ELASTIC_API_EVENTS")

	// SecondAPIForward is a second location to forward events to
	SecondAPIForward = os.Getenv("ELASTIC_API_EVENTS_TWO")

	// HeartbeatForward is the url to forward heartbeats to
	HeartbeatForward = os.Getenv("ELASTIC_HEARTBEAT_EVENTS")

	// DMPSEventsForward is the url to forward DMPS events to
	DMPSEventsForward = os.Getenv("ELASTIC_DMPS_EVENTS")

	// DMPSHeartbeatForward is the url to forward DMPS events to
	DMPSHeartbeatForward = os.Getenv("ELASTIC_DMPS_HEARTBEATS")
)

func init() {
	if len(APIForward) == 0 || len(HeartbeatForward) == 0 {
		log.L.Fatalf("$ELASTIC_API_EVENTS and $ELASTIC_HEARTBEAT_EVENTS must be set.")
	}
	log.L.Infof("\n\nForwarding URLs:\n\tAPI Forward:\t\t%v\n\tSecond API Forward\t\t%v\n\tHeartbeat Forward:\t%v\n", APIForward, SecondAPIForward, HeartbeatForward)
}

// SimpleForwardingJob is exported to add it as a job.
type SimpleForwardingJob struct {
}

// Run fowards events to an elk timeseries index.
func (*SimpleForwardingJob) Run(context interface{}, actionWrite chan action.Payload) {

	var err *nerr.E
	//	cache.GetCache(cache.DEFAULT)
	switch v := context.(type) {
	case *elkreporting.ElkEvent:
		//translate
		_, err = cache.GetCache(cache.DEFAULT).StoreAndForwardEvent(TranslateEvent(*v))
	case elkreporting.ElkEvent:
		//translate
		_, err = cache.GetCache(cache.DEFAULT).StoreAndForwardEvent(TranslateEvent(v))
	case v2.Event:
		_, err = cache.GetCache(cache.DEFAULT).StoreAndForwardEvent(v)
	case *v2.Event:
		_, err = cache.GetCache(cache.DEFAULT).StoreAndForwardEvent(*v)
	default:
	}

	if err != nil {
		log.L.Warnf("Problem storing event: %v", err.Error())
	}

	return
}

func TranslateEvent(e elkreporting.ElkEvent) v2.Event {
	time, err := time.Parse(time.RFC3339, e.Timestamp)
	if err != nil {
		log.L.Warnf("Couldn't parse time %v: %v", e.Timestamp, err.Error())
	}

	//check to see if the room id is already good
	if !strings.Contains(e.Room, "-") {
		//we need to build it
		e.Room = e.Building + "-" + e.Room
	}

	if !strings.Contains(e.Event.Event.Device, "-") {
		//we need to build it
		e.Event.Event.Device = e.Room + "-" + e.Event.Event.Device
	}

	toReturn := v2.Event{
		GeneratingSystem: e.Hostname,
		Timestamp:        time,
		TargetDevice: v2.BasicDeviceInfo{
			DeviceID: e.Event.Event.Device,
			BasicRoomInfo: v2.BasicRoomInfo{
				RoomID:     e.Room,
				BuildingID: e.Building,
			}},
		AffectedRoom: v2.BasicRoomInfo{
			BuildingID: e.Building,
			RoomID:     e.Room,
		},
		Key:   e.Event.Event.EventInfoKey,
		Value: e.Event.Event.EventInfoValue,
		User:  e.Event.Event.Requestor,
	}

	//generate tags
	toReturn.EventTags = GetTagsFromOldEvent(e)

	return toReturn
}

func GetTagsFromOldEvent(e elkreporting.ElkEvent) []string {

	log.L.Debugf(color.HiBlueString("%v", e.Event.Event.EventCause))
	log.L.Debugf(color.HiBlueString("%v", e.Event.Event.Type))

	var toReturn []string
	switch e.Event.Event.Type {
	case events.CORESTATE:
		toReturn = append(toReturn, v2.CoreState)
	case events.HEARTBEAT:
		toReturn = append(toReturn, v2.Heartbeat)
	case events.DETAILSTATE:
		toReturn = append(toReturn, v2.DetailState)
	default:
		toReturn = append(toReturn, e.Event.Event.Type.String())
	}

	switch e.Event.Event.EventCause {
	case events.USERINPUT:
		toReturn = append(toReturn, v2.UserGenerated)
	case events.AUTOGENERATED:
		toReturn = append(toReturn, v2.AutoGenerated)
	default:
		toReturn = append(toReturn, e.Event.Event.EventCause.String())
	}

	return toReturn
}
