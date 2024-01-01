package main;

import (
	"log"
	"fmt"
  "os"
  "time"
  "math"
  "strconv"
  "strings"
  "os/signal"
  "net/url"
  "encoding/json"

  "github.com/gorilla/websocket"
)

const MYID = "7acd3a1a6cc84244c19385f2"

var ownerx float64 = 0
var ownery float64 = 0

var token = os.Getenv("MPPTOKEN")

func HandleChatMessage(con *websocket.Conn, msg map[string]interface{}) {
  user := msg["p"].(map[string]interface{})
  log.Printf("Wow, user %s sent %s", user["name"].(string), msg["a"].(string));
  if strings.ToLower(msg["a"].(string)) == "охуеть" {
    SendArray(con, []map[string]interface{} {{
      "m": "a",
      "message": "охуеть!",
      "reply_to": msg["id"].(string),
    }})
  }
}

func HandleMouseMovement(con *websocket.Conn, msg map[string]interface{}) {
  var err error
  if msg["id"].(string) == MYID {
    ownerx, err = strconv.ParseFloat(msg["x"].(string), 64)
    if err != nil { log.Fatal(err) }
    ownery, err = strconv.ParseFloat(msg["y"].(string), 64)
    if err != nil { log.Fatal(err) }
  }
}

func MoveMouse(con *websocket.Conn, x0 float64, y0 float64, t int64) error {
  mag := 9.0/16.0
  amp := 5.0
  var speedinv float64 = 550
  nx := x0 + mag*amp*math.Cos(-float64(t)/speedinv)
  ny := y0 + amp*math.Sin(-float64(t)/speedinv)
  return SendArray(con, []map[string]interface{} {{
    "m": "m",
    "x": nx,
    "y": ny,
  }})
}

func HandleMessage(con *websocket.Conn, msg string) {
  var parsed []interface{}
  err := json.Unmarshal([]byte(msg), &parsed);

  if err != nil {
    log.Fatal(err);
  }

  for _, elem := range parsed {
    casted := elem.(map[string]interface{})
    switch casted["m"].(string) {
    case "a":
      HandleChatMessage(con, casted);
    case "m":
      HandleMouseMovement(con, casted);
    default:
      log.Printf("Unknown event: %s", casted["m"].(string));
    }
  }
}

func SendArray(con *websocket.Conn, array []map[string]interface{}) error {
  jsonStr, err := json.Marshal(array);

  if err != nil {
    return err
  }

  err = con.WriteMessage(websocket.TextMessage, []byte(string(jsonStr)))
  if err != nil {
    return err
  }

  return nil
}

func main() {
  log.Println("Started");
  
  // messageOut := make(chan string)
  interrupt := make(chan os.Signal, 1)
  signal.Notify(interrupt, os.Interrupt)

  message := make(chan string)

  u := url.URL{Scheme: "wss", Host: "www.mppclone.com", Path: "",}
  con, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
  defer con.Close()

  if err != nil {
    fmt.Printf("((\n")
		log.Fatal(err)
	}

  done := make(chan struct{})

  // receive
  go func() {
    defer close(done)
    for {
      _, msg, err := con.ReadMessage()
      if err != nil {
        log.Println("read:", err)
        return
      }
      log.Printf("recv: %s", string(msg[:]));
      message <- string(msg[:]);
    }
  }()

  // SendHi(con, token)
  SendArray(con, []map[string]interface{} {{
    "m": "hi",
    "token": token,
  }});
  // GoToRoom(con, "Room625346841249")
  SendArray(con, []map[string]interface{} {{
    "m": "ch",
    "_id": "qwe",
  }});

  // ebanemsya
  SendArray(con, []map[string]interface{} {{
    "m": "userset",
    "set": map[string]string {
      "name": "GOBOT1",
      "color": "#32a852",
    },
  }});

  // writing to the chat
  SendArray(con, []map[string]interface{} {{
    "m": "a",
    "message": "ГОВНО С ДЫМОМ",
  }});

	ticker := time.NewTicker(20*time.Second)
	defer ticker.Stop()

	mouseMover := time.NewTicker(50*time.Millisecond)
	defer mouseMover.Stop()

  // wating
  for {
    select {
    case <-done:
      return

    case msg := <-message:
      HandleMessage(con, msg)

    case _ = <-ticker.C:
      log.Printf("a 20 seconds passed")
      SendArray(con, []map[string]interface{} {{
        "m": "t",
        "e": time.Now().UnixNano()/1000000,
      }});

    case _ = <-mouseMover.C:
      MoveMouse(con, ownerx, ownery, time.Now().UnixNano()/1000000);

    case <-interrupt:
      log.Printf("Interrupted");
      			err := con.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
      if err != nil {
        log.Println("write close:", err)
        return
      }
      select {
      case <-done:
      case <-time.After(time.Second):
      }
      return
    }
  }
}
