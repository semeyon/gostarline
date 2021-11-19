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
const USER_UNFO_URL = "ttps://developer.starline.ru/json/v1/user/%s/user_info/"
const BASE_EVENTS_TYPES_URL = "https://developer.starline.ru/json/v3/library/events"
const DEVICES_EVENTS_URL = "https://developer.starline.ru/json/v2/device/%s/events"
const DEVICE_DATA_URL = "https://developer.starline.ru/json/v3/device/%s/data"
const COOKIE_NAME = "slnet"
const POST = "POST"
const GET = "GET"

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
	GroupId   int
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
	Common          Common          `json:"common"`
	Event           InnerEvent      `json:"event"`
	AlarmState      InnerAlarmState `json:"alarm_state"`
	OBD             OBD             `json:"obd"`
	Position        Position        `json:"position"`
	State           State           `json:"state"`
	Balance         []Balance       `json:"balance"`
	Telephone       string          `json:"telephone"`
	FirmwareVersion string          `json:"firmware_version"`
	Status          int             `json:"status"`
	UaUrl           string          `json:"ua_url"`
	Sn              string          `json:"sn"`
	Type            string          `json:"type"`
	Alias           string          `json:"alias"`
	DeviceId        string          `json:"device_id"`
	ActivityTs      int64           `json:"activity_ts"`
	Typename        string          `json:"typename"`
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

func getEvents() []EventType {
	log.Println("Request event types by " + BASE_EVENTS_TYPES_URL)
	var eventDescriptions eventDescriptions
	resp, err := http.Get(BASE_EVENTS_TYPES_URL)
	resp.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
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
		Name:  COOKIE_NAME,
		Value: token,
	}
	url := fmt.Sprintf(DEVICES_EVENTS_URL, deviceId)
	req, err := http.NewRequest(POST, url, bytes.NewBuffer(bParams))
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
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(bodyBytes, &eventsContainer)
	return eventsContainer
}

func getData(deviceId string, token string) Data {
	var data Data
	cookie := &http.Cookie{
		Name:  "slnet",
		Value: token,
	}
	url := fmt.Sprintf(DEVICE_DATA_URL, deviceId)
	req, err := http.NewRequest(GET, url, nil)
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
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(bodyBytes, &data)
	return data
}

func mapEvents(eventTypes []EventType, rawEvents []RawEvent) []Event {
	var eventsMapped []Event
	var eventMap = make(map[int]EventType)
	for _, eventType := range eventTypes {
		eventMap[eventType.Code] = eventType
	}
	for _, rawEvent := range rawEvents {
		event := Event{
			Code:      rawEvent.Type,
			Desc:      eventMap[rawEvent.Type].Desc,
			GroupId:   eventMap[rawEvent.Type].GroupId,
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

func initNewTextView() *tview.TextView {
	list := tview.NewTextView()
	list.SetBorder(true)
	list.SetDynamicColors(true).SetRegions(true)
	return list
}

func initNewDataTextView(drawData InnerData) *tview.TextView {
	title := fmt.Sprintf("[[blue]%s%s %s %s[white]]", drawData.Typename, drawData.Type, drawData.Alias, drawData.FirmwareVersion)
	textView := tview.NewTextView()
	textView.SetBorder(true).SetTitle(title)
	textView.SetDynamicColors(true).SetRegions(true)
	return textView
}

func getCurrency(b Balance) string {
	if (b.Currency == "") || (b.Currency == "RUB") {
		return "₽"
	} else {
		return b.Currency
	}
}

func drawDataTextView(textView *tview.TextView, data Data, eventTypes []EventType) {
	var balances []string
	textView.Clear()
	drawData := data.Data
	balancesArr := drawData.Balance
	for _, balance := range balancesArr {
		bTime := getStandartTimeFormat(int64(balance.Ts))
		bCur := getCurrency(balance)
		balances = append(balances, fmt.Sprintf("%s: %s%d @%s", balance.Key, bCur, balance.Value, bTime))
	}

	currentState := getEventById(eventTypes, drawData.Event.Type)
	movingState := getMovingState(drawData.Position.IsMove)
	requestTime := getStandartTimeFormat(drawData.ActivityTs)
	positionTime := getStandartTimeFormat(int64(drawData.Position.Ts))
	odbTime := getStandartTimeFormat(int64(drawData.OBD.Ts))
	commonTime := getStandartTimeFormat(int64(drawData.Common.Ts))
	stateTime := getStandartTimeFormat(int64(drawData.Event.Timestamp))
	fmt.Fprintf(textView, "[bold]Request status: [white]%d %s @%s\n", data.Code, data.CodeString, requestTime)
	fmt.Fprintf(textView, "[bold]Position: [white]%f, %f %s @%s\n", drawData.Position.Y, drawData.Position.X, movingState, positionTime)
	fmt.Fprintf(textView, "[bold]ODB: [white]%d litres, %d km @%s\n", drawData.OBD.FuelLitres, drawData.OBD.Mileage, odbTime)
	fmt.Fprintf(textView, "[bold]Common: [white]Auto: %d°C, Engine: %d°C, %.2fV, GPS:%.1f, GSM:%.1f @%s\n", drawData.Common.CTemp, drawData.Common.Etemp, drawData.Common.Battery, drawData.Common.GpsLvl, drawData.Common.GsmLvl, commonTime)
	fmt.Fprintf(textView, "[bold]Billing: [white]%s @%s\n", currentState, stateTime)
	fmt.Fprintf(textView, "[bold]State: [white]%s\n", balances)
}

func setColorOnEvenGroupId(groupId int) string {
	if (groupId == 0) || (groupId == 2) {
		return "[red]"
	} else if (groupId == 3) || (groupId == 4) {
		return "[yellow]"
	}
	return "[white]"
}

func drawListView(list *tview.TextView, titleFormat string, rawEvents EventsContainer, events []Event) {
	list.Clear()
	now := time.Now()
	list.SetTitle(fmt.Sprintf(titleFormat+" (%d)@%s", len(events), now.Format(time.RFC3339)))
	fmt.Fprintf(list, "[blue]%d %s\n", rawEvents.Code, rawEvents.CodeString)
	for _, event := range events {
		tm := time.Unix(event.Timestamp, 0)
		// TODO: Move to a function
		fmt.Fprintf(list, "%s%s > %s\n[white]", setColorOnEvenGroupId(event.GroupId), tm.Format(RU_TIME_FORMAT), event.Desc)
	}
}

func prepareStartEndDate() (int64, int64) {
	now := time.Now()
	startdTs := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	endTs := startdTs.Add(24 * time.Hour)
	startdTsUnix := startdTs.Unix()
	endTsUnix := endTs.Unix()
	return startdTsUnix, endTsUnix
}

func main() {
	log.Println("GoStarline starting up")
	slnetToken := flag.String("token", "", "slnet token")
	device_id := flag.String("device_id", "38406090", "device id")
	flag.Parse()
	log.Printf("Token: %s and device id: %s will be used ", *slnetToken, *device_id)

	startdTsUnix, endTsUnix := prepareStartEndDate()

	app := tview.NewApplication()

	eventTypes := getEvents()
	data := getData(*device_id, *slnetToken)
	textView := initNewDataTextView(data.Data)
	drawDataTextView(textView, data, eventTypes)

	list := initNewTextView()
	rawEvents := getRawEvent(*device_id, *slnetToken, startdTsUnix, endTsUnix)
	events := mapEvents(eventTypes, rawEvents.Events)
	drawListView(list, "Events Today", rawEvents, events)

	list2 := initNewTextView()
	rawEvents2 := getRawEvent(*device_id, *slnetToken, startdTsUnix-48*3600, endTsUnix-24*3600)
	events2 := mapEvents(eventTypes, rawEvents2.Events)
	drawListView(list2, "Events Yesturday", rawEvents2, events2)

	list3 := initNewTextView()
	rawEvents3 := getRawEvent(*device_id, *slnetToken, startdTsUnix-72*3600, endTsUnix-48*3600)
	events3 := mapEvents(eventTypes, rawEvents3.Events)
	drawListView(list3, "Events 48 hours ago", rawEvents3, events3)

	list4 := initNewTextView()
	rawEvents4 := getRawEvent(*device_id, *slnetToken, startdTsUnix-96*3600, endTsUnix-72*3600)
	events4 := mapEvents(eventTypes, rawEvents4.Events)
	drawListView(list4, "Events 72 hours ago", rawEvents4, events4)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(list, 0, 3, false).
			AddItem(list2, 0, 3, false).
			AddItem(list3, 0, 3, false).
			AddItem(list4, 0, 3, false), 0, 5, false)

	ticker := time.NewTicker(60 * time.Second)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				rawEvents := getRawEvent(*device_id, *slnetToken, startdTsUnix, endTsUnix)
				events := mapEvents(eventTypes, rawEvents.Events)
				app.QueueUpdateDraw(func() {
					drawListView(list, "Events Today", rawEvents, events)
				})
				data := getData(*device_id, *slnetToken)
				app.QueueUpdateDraw(func() {
					drawDataTextView(textView, data, eventTypes)
				})
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	if err := app.SetRoot(flex, true).SetFocus(flex).Run(); err != nil {
		panic(err)
	}

}
