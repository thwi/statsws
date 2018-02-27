// Connection
// Wrapper for reading/writing to a websocket connection
package statsws

import (
  "fmt"
  "sync"
  "github.com/gorilla/websocket"
)

const INITIAL = "initial"
const UPDATE  = "update"

type Connection struct {
  Address string
  mutex   sync.Mutex
  socket  *websocket.Conn
}

type MsgData struct {
  Cpu []int               `json:"cpu"`
  Mem int                 `json:"mem"`
  Net map[string][]uint64 `json:"net"`
  Ts  int64               `json:"ts"`
}

type InitialMessage struct {
  Type     string     `json:"type"`
  Data     []*MsgData `json:"data"`
  Count    int        `json:"count"`
  Interval int        `json:"interval"`
}

type UpdateMessage struct {
  Type string   `json:"type"`
  Data *MsgData `json:"data"`
}

func newConnection(conn *websocket.Conn, addr string) *Connection {
  return &Connection{
    Address: addr,
    socket:  conn,
  }
}

func (c *Connection) log(msg string) {
  fmt.Printf("[ %+v ]\t%+v\n", c.Address, msg)
}

func (c *Connection) writeInitial(data []*MsgData, count int, interval int) {
  c.write(&InitialMessage{
    Type:     INITIAL,
    Data:     data,
    Count:    count,
    Interval: interval,
  })
}

func (c *Connection) writeUpdate(data *MsgData) {
  c.write(&UpdateMessage{
    Type: UPDATE,
    Data: data,
  })
}

// Write a message to websocket, using mutex to ensure single concurrent writer
func (c *Connection) write(msg interface{}) {
  c.mutex.Lock()
  c.log(fmt.Sprintf("Sending message: %+v", msg))
  c.socket.WriteJSON(msg)
  c.mutex.Unlock()
}
