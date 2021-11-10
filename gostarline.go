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

type InnerData struct {
	Common          Common     `json:"common"`
	Event           InnerEvent `json:"event"`
	OBD             OBD        `json:"obd"`
	Position        Position   `json:"position"`
	Balance         []Balance  `json:"balance"`
	Telephone       string     `json:"telephone"`
	FirmwareVersion string     `json:"firmware_version"`
	Status          int        `json:"status"`
	UaUrl           string     `json:"ua_url"`
	Sn              string     `json:"sn"`
	Type            string     `json:"type"`
	Alias           string     `json:"alias"`
	DeviceId        string     `json:"device_id"`
	ActivityTs      int        `json:"activity_ts"`
	Typename        string     `json:"typename"`
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
	log.Printf("Request raw events %s, -d %s", url, string(bParams))
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
	log.Println(resp.Status)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(bodyBytes, &eventsContainer)
	log.Println(string(bodyBytes))
	log.Printf("Number of raw event: %v", len(eventsContainer.Events))
	return eventsContainer
}

func getData(deviceId string, token string) Data {
	var data Data
	cookie := &http.Cookie{
		Name:  "slnet",
		Value: token,
	}
	url := "https://developer.starline.ru/json/v3/device/" + deviceId + "/data"
	log.Printf("Request device data %s", url)
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

func main() {
	log.Println("GoStarline starting up")
	slnetToken := flag.String("token", "", "slnet token")
	device_id := flag.String("device_id", "38406090", "device id")
	flag.Parse()

	log.Printf("Token: %s and device id: %s will be used ", *slnetToken, *device_id)
	eventTypes := getEvents()
	data := getData(*device_id, *slnetToken)
	log.Println(data)
	apochNow := time.Now().Unix()
	rawEvents := getRawEvent(*device_id, *slnetToken, apochNow-24*3600, apochNow)
	events := mapEvents(eventTypes, rawEvents.Events)

	log.Println("---------")
	log.Println(events)

	app := tview.NewApplication()

	list := tview.NewList()
	list.ShowSecondaryText(false).
		SetBorder(true).
		SetTitle("Events Today")

	rawEvents2 := getRawEvent(*device_id, *slnetToken, apochNow-48*3600, apochNow-24*3600)
	events2 := mapEvents(eventTypes, rawEvents2.Events)

	list2 := tview.NewList()
	list2.SetBorder(true).SetTitle("Events Yesterday")
	list2.ShowSecondaryText(false)
	for _, event := range events2 {
		tm := time.Unix(event.Timestamp, 0)
		list2.AddItem(tm.Format(time.RFC3339)+" > "+event.Desc, "", '0', nil)
	}

	rawEvents3 := getRawEvent(*device_id, *slnetToken, apochNow-72*3600, apochNow-48*3600)
	events3 := mapEvents(eventTypes, rawEvents3.Events)

	list3 := tview.NewList()
	list3.SetBorder(true).SetTitle("Events 48 hours ago")
	list3.ShowSecondaryText(false)
	for _, event := range events3 {
		tm := time.Unix(event.Timestamp, 0)
		list3.AddItem(tm.Format(time.RFC3339)+" > "+event.Desc, "", '0', nil)
	}

	textView := tview.NewTextView()
	textView.SetBorder(true).SetTitle("Data")
	// textView.SetText(data)
	fmt.Fprintf(textView, "%d ", data)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(list, 0, 3, false).
			AddItem(list2, 0, 3, false).
			AddItem(list3, 0, 3, false), 0, 5, false)

	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})
	counter := 0
	go func() {
		for {
			select {
			case <-ticker.C:
				app.QueueUpdateDraw(func() {
					list.AddItem(fmt.Sprint(counter), "", '0', nil)
				})
				counter = counter + 1
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
