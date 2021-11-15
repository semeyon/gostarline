package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/rivo/tview"
)

const RU_TIME_FORMAT = "15:04:05"

type RawEvent struct {
	Type      int `json:"type"`
	GroupId   int `json:"groupId"`
	Timestamp int `json:"timestamp"`
}

type EventType struct {
	Code    int    `json:"code"`
	Desc    string `json:"desc"`
	GroupId int    `json:"group_id"`
}

type eventDescriptions struct {
	Events     []EventType `json:"eventDescriptions"`
	Code       int         `json:"code"`
	CodeString string      `json:"codestring"`
}

type Event struct {
	Code      int
	Desc      string
	Timestamp int64
}

type Balance struct {
	Currency string `json:"currency"`
	Key      string `json:"key"`
	Ts       int    `json:"ts"`
	Operator string `json:"operator"`
	State    int    `json:"state"`
	Value    int    `json:"value"`
}

type Position struct {
	S      int     `json:"s"`
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	IsMove bool    `json:"is_move"`
	Dir    int     `json:"dir"`
	R      int     `json:"r"`
	Ts     int     `json:"ts"`
	SatQty int     `json:"sat_qty"`
}

type OBD struct {
	Ts          int `json:"ts"`
	FuelLitres  int `json:"fuel_litres"`
	Mileage     int `json:"mileage"`
	FuelPercent int `json:"fuel_percent"`
}

type InnerEvent struct {
	Type      int `json:"type"`
	Timestamp int `json:"timestamp"`
}

type Common struct {
	RegDate   int     `json:"reg_date"`
	Etemp     int     `json:"etemp"`
	GsmLvl    float32 `json:"gsm_lvl"`
	GpsLvl    float32 `json:"gps_lvl"`
	Ts        int     `json:"ts"`
	MayakTemp int     `json:"mayak_temp"`
	Battery   float32 `json:"battery"`
	CTemp     int     `json:"ctemp"`
}

type InnerAlarmState struct {
	AddH   bool `json:"add_h"`
	AddL   bool `json:"add_l"`
	Door   bool `json:"door"`
	Hbrake bool `json:"hbrake"`
	Hijack bool `json:"hijack"`
	Hood   bool `json:"hood"`
	Ign    bool `json:"ign"`
	Pbrake bool `json:"pbrake"`
	Shockh bool `json:"shock_h"`
	Shockl bool `json:"shock_l"`
	Tilt   bool `json:"tilt"`
	Trunk  bool `json:"trunk"`
	Ts     int  `json:"ts"`
}

type State struct {
	AddSensBpass bool `json:"add_sens_bpass"`
	Alarm        bool `json:"alarm"`
	Arm          bool `json:"arm"`
	ArmAuthWait  bool `json:"arm_auth_wait"`
	ArmMovingPb  bool `json:"arm_moving_pb"`
	Door         bool `json:"door"`
	Hbrake       bool `json:"hbrake"`
	Hfree        bool `json:"hfree"`
	Hijack       bool `json:"hijack"`
	Hood         bool `json:"hood"`
	Ign          bool `json:"ign"`
	Neutral      bool `json:"neutral"`
	Out          bool `json:"out"`
	Pbrake       bool `json:"pbrake"`
	RStart       bool `json:"r_start"`
	RStartTimer  int  `json:"r_start_timer"`
	Run          bool `json:"run"`
	ShockBpass   bool `json:"shock_bpass"`
	TiltBpass    bool `json:"tilt_bpass"`
	Trunk        bool `json:"trunk"`
	Valet        bool `json:"valet"`
	Webasto      bool `json:"webasto"`
	WebastoTimer int  `json:"webasto_timer"`
	Ts           int  `json:"ts"`
}

type InnerData struct {
	Common     Common          `json:"common"`
	Event      InnerEvent      `json:"event"`
	AlarmState InnerAlarmState `json:"alarm_state"`
	OBD        OBD             `json:"obd"`
	Position   Position        `json:"position"`
	// State           State           `json:"state"`
	Balance         []Balance `json:"balance"`
	Telephone       string    `json:"telephone"`
	FirmwareVersion string    `json:"firmware_version"`
	Status          int       `json:"status"`
	UaUrl           string    `json:"ua_url"`
	Sn              string    `json:"sn"`
	Type            string    `json:"type"`
	Alias           string    `json:"alias"`
	DeviceId        string    `json:"device_id"`
	ActivityTs      int64     `json:"activity_ts"`
	Typename        string    `json:"typename"`
}

type Data struct {
	Data       InnerData `json:"data"`
	Code       int       `json:"code"`
	CodeString string    `json:"codestring"`
}

type EventsContainer struct {
	Events     []RawEvent `json:"events"`
	Code       int        `json:"code"`
	CodeString string     `json:"codestring"`
}

type EventRequestParams struct {
	Start int64 `json:"period_start"`
	End   int64 `json:"period_end"`
}

func getMovingState(state bool) string {
	if state {
		return "Moving"
	} else {
		return "Stopped"
	}
}

func getStandartTimeFormat(ts int64) string {
	return time.Unix(ts, 0).Format(RU_TIME_FORMAT)
}

// Get predefined events from the starline server
func getEvents() []EventType {
	log.Println("Request event types by https://developer.starline.ru/json/v3/library/events")
	var eventDescriptions eventDescriptions

	resp, err := http.Get("https://developer.starline.ru/json/v3/library/events")
	resp.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	// log.Println(resp.Status)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(bodyBytes, &eventDescriptions)
	log.Printf("Number of event types: %v", len(eventDescriptions.Events))
	return eventDescriptions.Events
}

func getRawEvent(deviceId string, token string, start int64, end int64) EventsContainer {
	var eventsContainer EventsContainer
	params := &EventRequestParams{Start: start, End: end}
	bParams, err := json.Marshal(params)
	if err != nil {
		log.Fatal(err)
	}
	cookie := &http.Cookie{
		Name:  "slnet",
		Value: token,
	}
	url := "https://developer.starline.ru/json/v2/device/" + deviceId + "/events"
	// log.Printf("Request raw events %s, -d %s", url, string(bParams))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bParams))
	if err != nil {
		log.Fatal(err)
	}
	req.AddCookie(cookie)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	// log.Println(resp.Status)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(bodyBytes, &eventsContainer)
	// log.Println(string(bodyBytes))
	// log.Printf("Number of raw event: %v", len(eventsContainer.Events))
	return eventsContainer
}

func getData(deviceId string, token string) Data {
	var data Data
	cookie := &http.Cookie{
		Name:  "slnet",
		Value: token,
	}
	url := "https://developer.starline.ru/json/v3/device/" + deviceId + "/data"
	// log.Printf("Request device data %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.AddCookie(cookie)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	// log.Println(resp.Status)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(bodyBytes, &data)
	return data
}

func mapEvents(eventTypes []EventType, rawEvents []RawEvent) []Event {
	var eventsMapped []Event
	var eventMap = make(map[int]string)
	for _, eventType := range eventTypes {
		eventMap[eventType.Code] = eventType.Desc
	}
	for _, rawEvent := range rawEvents {
		event := Event{
			Code:      rawEvent.Type,
			Desc:      eventMap[rawEvent.Type],
			Timestamp: int64(rawEvent.Timestamp),
		}
		eventsMapped = append(eventsMapped, event)
	}
	return eventsMapped
}

func getEventById(eventTypes []EventType, eventId int) string {
	for _, eventType := range eventTypes {
		if eventType.Code == eventId {
			return eventType.Desc
		}
	}
	return "Unknown event"
}

func main() {
	log.Println("GoStarline starting up")
	slnetToken := flag.String("token", "", "slnet token")
	device_id := flag.String("device_id", "38406090", "device id")
	flag.Parse()

	log.Printf("Token: %s and device id: %s will be used ", *slnetToken, *device_id)
	eventTypes := getEvents()
	data := getData(*device_id, *slnetToken)

	now := time.Now()
	startdTs := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	endTs := startdTs.Add(24 * time.Hour)
	startdTsUnix := startdTs.Unix()
	endTsUnix := endTs.Unix()

	app := tview.NewApplication()

	list := tview.NewList()
	list.ShowSecondaryText(false).
		SetBorder(true).
		SetTitle("Events Today (?)")

	list.SetHighlightFullLine(false)
	list.SetWrapAround(false)

	rawEvents := getRawEvent(*device_id, *slnetToken, startdTsUnix, endTsUnix)
	events := mapEvents(eventTypes, rawEvents.Events)
	list.SetTitle(fmt.Sprintf("Events Today (%d)@%s", len(events), now.Format(time.RFC3339)))
	// count = count + 1
	list.AddItem(fmt.Sprintf("%d", rawEvents.Code)+" | "+rawEvents.CodeString, "", '0', nil)
	for _, event := range events {
		tm := time.Unix(event.Timestamp, 0)
		list.AddItem(tm.Format(time.RFC3339)+" > "+event.Desc, "", '0', nil)
	}

	rawEvents2 := getRawEvent(*device_id, *slnetToken, startdTsUnix-48*3600, endTsUnix-24*3600)
	events2 := mapEvents(eventTypes, rawEvents2.Events)

	list2 := tview.NewList()
	// list2.SetBorder(true).SetTitle("Events Yesterday")
	list2.SetBorder(true).SetTitle(fmt.Sprintf("Events Yesterday (%d)", len(events2)))
	list2.ShowSecondaryText(false)
	for _, event := range events2 {
		tm := time.Unix(event.Timestamp, 0)
		list2.AddItem(tm.Format(time.RFC3339)+" > "+event.Desc, "", '0', nil)
	}

	rawEvents3 := getRawEvent(*device_id, *slnetToken, startdTsUnix-72*3600, endTsUnix-48*3600)
	events3 := mapEvents(eventTypes, rawEvents3.Events)

	list3 := tview.NewList()
	list3.SetBorder(true).SetTitle(fmt.Sprintf("Events 48 hours ago (%d)", len(events3)))
	list3.ShowSecondaryText(false)
	for _, event := range events3 {
		tm := time.Unix(event.Timestamp, 0)
		list3.AddItem(tm.Format(time.RFC3339)+" > "+event.Desc, "", '0', nil)
	}

	textView := tview.NewTextView()
	textView.SetBorder(true).SetTitle("Data")
	textView.SetDynamicColors(true).SetRegions(true)
	drawData := data.Data

	fmt.Fprintf(textView, "[blue]%s%s %s %s\n", drawData.Typename, drawData.Type, drawData.Alias, drawData.FirmwareVersion)
	fmt.Fprintf(textView, "[bold]Request status: [white]%d %s @%s\n", data.Code, data.CodeString, getStandartTimeFormat(drawData.ActivityTs))
	fmt.Fprintf(textView, "[bold]Position: [white]%f, %f %s @%s\n", drawData.Position.X, drawData.Position.Y, getMovingState(drawData.Position.IsMove), getStandartTimeFormat(int64(drawData.Position.Ts)))
	fmt.Fprintf(textView, "[bold]ODB: [white]%d litres, %d km @%s\n", drawData.OBD.FuelLitres, drawData.OBD.Mileage, getStandartTimeFormat(int64(drawData.OBD.Ts)))
	fmt.Fprintf(textView, "[bold]Common: [white]Auto: %d°C, Engine: %d°C, %fV, GPS:%f, GSM:%f @%s\n", drawData.Common.CTemp, drawData.Common.Etemp, drawData.Common.Battery, drawData.Common.GpsLvl, drawData.Common.GsmLvl, getStandartTimeFormat(int64(drawData.Common.Ts)))
	fmt.Fprintf(textView, "[bold]State: [white]%s @%s\n", getEventById(eventTypes, drawData.Event.Type), getStandartTimeFormat(int64(drawData.Event.Timestamp)))

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(list, 0, 3, false).
			AddItem(list2, 0, 3, false).
			AddItem(list3, 0, 3, false), 0, 5, false)

	ticker := time.NewTicker(60 * time.Second)
	quit := make(chan struct{})
	// count := 0
	go func() {
		for {
			select {
			case <-ticker.C:
				// TODO: could not update if it works 24+ hours
				rawEvents := getRawEvent(*device_id, *slnetToken, startdTsUnix, endTsUnix)
				events := mapEvents(eventTypes, rawEvents.Events)
				app.QueueUpdateDraw(func() {
					list.Clear()
					list.SetTitle(fmt.Sprintf("Events Today (%d)@%s", len(events), now.Format(time.RFC3339)))
					// count = count + 1
					list.AddItem(fmt.Sprintf("%d", rawEvents.Code)+" | "+rawEvents.CodeString, "", '0', nil)
					for _, event := range events {
						tm := time.Unix(event.Timestamp, 0)
						list.AddItem(tm.Format(time.RFC3339)+" > "+event.Desc, "", '0', nil)
					}
				})

				app.QueueUpdateDraw(func() {
					textView.Clear()
					data := getData(*device_id, *slnetToken)
					drawData = data.Data
					fmt.Fprintf(textView, "[red]%s\n", time.Now())
					fmt.Fprintf(textView, "[blue]%s%s %s %s\n", drawData.Typename, drawData.Type, drawData.Alias, drawData.FirmwareVersion)
					fmt.Fprintf(textView, "[bold]Request status: [white]%d %s @%s\n", data.Code, data.CodeString, getStandartTimeFormat(drawData.ActivityTs))
					fmt.Fprintf(textView, "[bold]Position: [white]%f %f %s @%s\n", drawData.Position.X, drawData.Position.Y, getMovingState(drawData.Position.IsMove), getStandartTimeFormat(int64(drawData.Position.Ts)))
					fmt.Fprintf(textView, "[bold]ODB: [white]%d litres %d km @%s\n", drawData.OBD.FuelLitres, drawData.OBD.Mileage, getStandartTimeFormat(int64(drawData.OBD.Ts)))
					fmt.Fprintf(textView, "[bold]Common: [white]Auto: %d°C Engine: %d°C %fV GPS:%f GSM:%f, @%s\n", drawData.Common.CTemp, drawData.Common.Etemp, drawData.Common.Battery, drawData.Common.GpsLvl, drawData.Common.GsmLvl, getStandartTimeFormat(int64(drawData.Common.Ts)))
					fmt.Fprintf(textView, "[bold]State: [white]%s @%s\n", getEventById(eventTypes, drawData.Event.Type), getStandartTimeFormat(int64(drawData.Event.Timestamp)))
				})
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	// go func() {
	// 	app.QueueUpdateDraw(func() {
	// 		time.Sleep(10 * time.Second)
	// 		for _, event := range events {
	// 			tm := time.Unix(event.Timestamp, 0)
	// 			list.AddItem(tm.Format(time.RFC3339)+" > "+event.Desc, "", '0', nil)
	// 		}

	// 	})

	// }()

	if err := app.SetRoot(flex, true).SetFocus(flex).Run(); err != nil {
		panic(err)
	}

	// Get events from the starline server
}
