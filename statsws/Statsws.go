// Statsws
package statsws

import (
  "fmt"
  "net/http"
  "sync"
  "github.com/gorilla/websocket"
)

type Statsws struct {
  collector   *Collector              // collector for retrieving metrics data
  connections map[string]*Connection  // map to store connected clients
  messages    *RingBuffer             // buffer to store composed message data
  mutex       sync.Mutex              // mutex for access to connections map
  register    chan *Connection        // notification of newly-connected clients
  unregister  chan *Connection        // notification of disconnected clients
  count       int                     // max number of metrics values
  interval    int                     // collector polling interval in seconds
}

func NewStatsws(count int, interval int) *Statsws {
  return &Statsws{
    collector:   NewCollector(interval),
    connections: make(map[string]*Connection),
    messages:    NewRingBuffer(count),
    register:    make(chan *Connection),
    unregister:  make(chan *Connection),
    count:       count,
    interval:    interval,
  }
}

func (s *Statsws) Start() {
  go s.collector.Start()
  go s.handleClients()
  go s.readMetrics()
}

// Handle an incoming HTTP request
func (s *Statsws) HandleRequest(w http.ResponseWriter, r *http.Request) {
  fmt.Printf("[ %v ]\t%v %v\n", r.RemoteAddr, r.Method, r.URL)
  conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
  if err != nil {
    http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
    return
  }

  c := newConnection(conn, r.RemoteAddr)
  s.register <- c
}

// Handle client connection / disconnection
func (s *Statsws) handleClients() {
  for {
    select {

    // When a new client connects
    case c := <-s.register:
      s.mutex.Lock()
      s.connections[c.socket.RemoteAddr().String()] = c
      s.mutex.Unlock()
      go c.writeInitial(s.messages.Values(), s.count, s.interval)
      go s.readClient(c)

    // When a client disconnects
    case c := <-s.unregister:
      addr := c.socket.RemoteAddr().String()
      if _, ok := s.connections[addr]; ok {
        s.mutex.Lock()
        c.log("Disconnected")
        delete(s.connections, addr)
        s.mutex.Unlock()
        c.socket.Close()
      }
    }
  }
}

// Read metrics from collector
func (s *Statsws) readMetrics() {

  // Collect initial metric value
  lastMetric := <-s.collector.Out

  for {
    // Read data from collector and add to buffer
    metric := <-s.collector.Out

    // Compose message data
    msg := &MsgData{
      Cpu: make([]int, len(metric.Cpu), len(metric.Cpu)),
      Net: make(map[string][]uint64, len(metric.Net)),
      Ts:  metric.Ts,
    }

    // Calculate CPU diff from last value
    for i, lv := range lastMetric.Cpu {
      msg.Cpu[i] = int((metric.Cpu[i].Busy - lv.Busy) / (metric.Cpu[i].All - lv.All) * 100)
    }

    // Calculate percent memory used
    msg.Mem = int((float64(metric.Mem.Used) / float64(metric.Mem.Total)) * 100)

    // Calculate network interface diff from last value
    for mac, ifaceData := range metric.Net {
      msg.Net[mac] = []uint64{
        ifaceData.BytesRxTotal - lastMetric.Net[mac].BytesRxTotal,
        ifaceData.BytesTxTotal - lastMetric.Net[mac].BytesTxTotal,
      }
    }

    // Write message to buffer
    s.messages.Add(msg)

    // Store last metric value
    lastMetric = metric

    // Send message to connected clients
    s.mutex.Lock()
    for _, c := range s.connections {
      go c.writeUpdate(msg)
    }
    s.mutex.Unlock()
  }
}

// Read messages from a client, ignoring anything except disconnection event
func (s *Statsws) readClient(c *Connection) {
  for {
    _, _, err := c.socket.ReadMessage()
    if err != nil {
      if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
        fmt.Printf("error: %v", err)
      }
      break
    }
  }

  // On disconnect, write connection to unregister channel
  s.unregister <- c
}
