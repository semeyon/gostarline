package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type RawEvent struct {
	Type      int `json:"type"`
	groupId   int `json:"groupId"`
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

// Get predefined events from the starline server
func getEvents() []EventType {
	var eventDescriptions eventDescriptions

	resp, err := http.Get("https://developer.starline.ru/json/v3/library/events")
	resp.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	log.Println(resp.Status)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(bodyBytes, &eventDescriptions)
	return eventDescriptions.Events
}

// curl "https://developer.starline.ru/json/v2/device/38406090/events" --cookie 'slnet=3FE11515FA267DE4BE81594B0A8C60A5' -d '{"period_start": 1635818531, "period_end": 1635926531}'
// func getRawEvent(deviceId string, token string, start int, end int) RawEvent {
// 	var rawEvent RawEvent
// 	cookie := &http.Cookie{
// 		Name:  "slnet",
// 		Value: token,
// 		// Expires: 3600,
// 	}
// 	url := "https://developer.starline.ru/json/v3/device/" + deviceId + "/events"
// 	log.Println(url)
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	req.AddCookie(cookie)
// 	client := http.Client{}
// 	resp, err := client.Do(req)

// 	defer resp.Body.Close()
// 	log.Println(resp.Status)
// 	bodyBytes, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	json.Unmarshal(bodyBytes, &rawEvent)
// 	return rawEvent
// }

func getData(deviceId string, token string) Data {
	var data Data
	cookie := &http.Cookie{
		Name:  "slnet",
		Value: token,
		// Expires: 3600,
	}
	url := "https://developer.starline.ru/json/v3/device/" + deviceId + "/data"
	log.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.AddCookie(cookie)
	client := http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()
	log.Println(resp.Status)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(bodyBytes, &data)
	return data
}

func main() {
	log.Println("GoStarline")
	slnetToken := "3FE11515FA267DE4BE81594B0A8C60A5"
	// user_id := "1827506"
	device_id := "38406090"
	events := getEvents()
	log.Println(events)
	data := getData(device_id, slnetToken)
	log.Println(data.Data.Balance)

	// Get events from the starline server
}
