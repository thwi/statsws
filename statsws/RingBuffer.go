// RingBuffer
// A ring buffer for storage of message data
package statsws

type RingBuffer struct {
  buf  []*MsgData
  head int
  size int
  tail int
}

func NewRingBuffer(size int) *RingBuffer {
  return &RingBuffer{
    buf:  make([]*MsgData, size),
    head: -1,
    size: size,
    tail:  0,
  }
}

// Add a value to the buffer
func (rb *RingBuffer) Add(v *MsgData) {
  rb.set(rb.head + 1, v)
  old := rb.head
  rb.head = rb.mod(rb.head + 1)
  if old != -1 && rb.head == rb.tail {
    rb.tail = rb.mod(rb.tail + 1)
  }
}

// Return a slice of all values in buffer
func (rb *RingBuffer) Values() []*MsgData {
  if rb.head == -1 { return []*MsgData{} }
  arr := make([]*MsgData, 0, rb.size)
  for i := 0; i < rb.size; i++ {
    idx := rb.mod(i + rb.tail)
    arr = append(arr, rb.get(idx))
    if idx == rb.head { break }
  }
  return arr
}

// Get a value at a given unmodified index
func (rb *RingBuffer) get(p int) *MsgData {
  return rb.buf[rb.mod(p)]
}

// Set a value at a given unmodified index and return modified index of value
func (rb *RingBuffer) set(p int, v *MsgData) {
  rb.buf[rb.mod(p)] = v
}

// Return modified index of an unmodified index
func (rb *RingBuffer) mod(p int) int {
  return p % len(rb.buf)
}
