package ringbuffer

import (
	"sync"
	"sync/atomic"
)

// RingBuffer is a lock-free multi-producer, multi-consumer ring buffer,
// but now includes a sync.Cond for optional blocking Dequeue.
type RingBuffer[T any] struct {
	// ring is the actual ring of items. We store pointers to T (rather than T) to avoid unintended copying
	ring []*node[T]

	// capacity is the size of the ring, which must be a power of two for proper masking
	capacity uint64
	// mask is capacity-1, so (index & mask) is index % capacity when capacity is power-of-two
	mask uint64

	// Enqueue fields:
	// producers read and increment enqueuePos to reserve slots in the ring
	enqueuePos uint64

	// Dequeue fields:
	// consumers read and increment dequeuePos to claim items
	dequeuePos uint64

	// condMu is the mutex used by cond. We keep it separate from the lock-free data
	// to avoid interfering with concurrent producers
	// need this for blocking Dequeues (so we don't spin lock)
	condMu sync.Mutex
	cond   *sync.Cond

	// closed is set by Close (under condMu) so blocked consumers can exit
	closed bool
}

// node holds a single ring buffer slot plus a sequence number used to synchronize producers/consumers
type node[T any] struct {
	// seq indicates which "turn" (index in the linear sense) this slot can be safely written or read
	seq    uint64
	value  T
	hasVal bool
}

// maxRingCapacity caps the ring size at an allocatable power of two.
const maxRingCapacity = 1 << 30

// NewRingBuffer creates a new ring buffer with the requested capacity.
// The capacity is rounded up to the next power of two (and capped at
// maxRingCapacity), since the index masking requires a power-of-two ring size.
func NewRingBuffer[T any](capacity uint64) *RingBuffer[T] {
	if capacity == 0 {
		capacity = 1
	}
	if capacity > maxRingCapacity {
		capacity = maxRingCapacity
	}
	if capacity&(capacity-1) != 0 {
		p := uint64(1)
		for p < capacity {
			p <<= 1
		}
		capacity = p
	}
	rb := &RingBuffer[T]{}
	rb.capacity = capacity
	rb.mask = capacity - 1
	rb.ring = make([]*node[T], capacity)
	for i := uint64(0); i < capacity; i++ {
		rb.ring[i] = &node[T]{seq: i}
	}

	// Set up the sync.Cond on top of the condMu
	rb.cond = sync.NewCond(&rb.condMu)

	return rb
}

// Enqueue tries to put val into the ring buffer
// Returns true if successful, false if the buffer is full at this moment
func (rb *RingBuffer[T]) Enqueue(val T) bool {
	var n *node[T]
	var pos uint64

	for {
		// Get the current enqueuePos. We'll attempt to reserve a slot at that position
		pos = atomic.LoadUint64(&rb.enqueuePos)
		n = rb.ring[pos&rb.mask]

		// Read the sequence value of this slot. The slot is ready for us if n.seq == pos (in the linear sense)
		seq := atomic.LoadUint64(&n.seq)
		diff := int64(seq) - int64(pos)
		if diff == 0 {
			// The slot is ready for the current pos. Attempt to reserve it by
			// incrementing enqueuePos. If successful, we own the slot
			if atomic.CompareAndSwapUint64(&rb.enqueuePos, pos, pos+1) {
				break
			}
			// Otherwise, some other producer raced us. Retry
		} else if diff < 0 {
			// This means seq < pos, so the consumer has not moved far enough to free this slot yet
			// => buffer is full right now
			return false
		} else {
			// diff > 0 means the slot is still owned by a previous position; return false or spin
			return false
		}
	}

	// We own slot n now. Write our value and update seq so consumers see it
	n.value = val
	n.hasVal = true
	// Next sequence for this slot is pos+1, so that a consumer can claim it
	atomic.StoreUint64(&n.seq, pos+1)

	// Wake any blocking dequeuer. Signal unconditionally: gating this on a
	// Size()==1 check races with concurrent producers/consumers and can lose
	// wakeups, leaving a consumer blocked while items are available.
	rb.condMu.Lock()
	rb.cond.Signal()
	rb.condMu.Unlock()

	return true
}

// Dequeue tries to remove and return one item from the ring buffer
// Returns (val, true) if successful, or (zeroValue, false) if empty
func (rb *RingBuffer[T]) Dequeue() (T, bool) {
	var n *node[T]
	var pos uint64
	var zero T

	for {
		// Get the current dequeuePos. We'll attempt to claim a slot at that position
		pos = atomic.LoadUint64(&rb.dequeuePos)
		n = rb.ring[pos&rb.mask]

		// We expect n.seq to be pos+1 if the item is ready for consumption
		seq := atomic.LoadUint64(&n.seq)
		diff := int64(seq) - int64(pos+1)
		if diff == 0 {
			// The slot has data for this position. Attempt to claim it by
			// incrementing dequeuePos. If successful, we can safely read
			if atomic.CompareAndSwapUint64(&rb.dequeuePos, pos, pos+1) {
				break
			}
			// Otherwise, some other consumer raced us. Retry
		} else if diff < 0 {
			// This means seq < pos+1 => no new data available
			return zero, false
		} else {
			// seq > pos+1 => producer is still writing or we are behind
			return zero, false
		}
	}

	// We own the slot's data now
	val := n.value
	n.value = zero
	n.hasVal = false

	// Mark this slot as free for future producers by setting n.seq to pos + rb.mask + 1
	// That is effectively the next "turn" in the linear sense when the producer can re-use this slot
	atomic.StoreUint64(&n.seq, pos+rb.mask+1)
	return val, true
}

// DequeueBlocking blocks (using sync.Cond) until it can dequeue an item.
// Returns (val, true) if an item is successfully dequeued, or (zeroValue, false)
// if the ring is closed or in some unexpected state
func (rb *RingBuffer[T]) DequeueBlocking() (T, bool) {
	for {
		rb.condMu.Lock()
		// Wait while size is 0, i.e. no items available
		for rb.Size() == 0 {
			if rb.closed {
				rb.condMu.Unlock()
				var zero T
				return zero, false
			}
			rb.cond.Wait()
		}
		rb.condMu.Unlock()

		// Attempt the normal Dequeue; another consumer may have raced us to
		// the item, in which case go back to waiting instead of returning a
		// spurious (zero, false)
		if val, ok := rb.Dequeue(); ok {
			return val, true
		}
	}
}

// Close marks the ring buffer as closed and wakes all blocked consumers.
// Items still in the ring can be drained with Dequeue/DequeueBlocking;
// DequeueBlocking returns (zero, false) once the ring is closed and empty.
func (rb *RingBuffer[T]) Close() {
	rb.condMu.Lock()
	rb.closed = true
	rb.cond.Broadcast()
	rb.condMu.Unlock()
}

// Size returns the number of items currently in the ring buffer
// Note that this is only an approximate snapshot because producers/consumers
// may be concurrently modifying enqueuePos/dequeuePos
func (rb *RingBuffer[T]) Size() uint64 {
	ep := atomic.LoadUint64(&rb.enqueuePos)
	dp := atomic.LoadUint64(&rb.dequeuePos)
	return ep - dp
}

// Capacity returns the total capacity of the ring buffer
func (rb *RingBuffer[T]) Capacity() uint64 {
	return rb.capacity
}

// Example usage (with blocking Dequeue):
//
// func main() {
//     rb := NewRingBuffer[string](1024)
//
//     // Producer (example)
//     go func() {
//         for i := 0; i < 10000; i++ {
//             for !rb.Enqueue(fmt.Sprintf("Log line %d", i)) {
//                 // If we fail to enqueue (buffer is full),
//                 // we might block, expand capacity, or drop the message.
//             }
//         }
//     }()
//
//     // Consumer using blocking Dequeue
//     go func() {
//         for {
//             val, ok := rb.DequeueBlocking()
//             if !ok {
//                 // handle unexpected closure (if we ever decide to stop the ring buffer)
//                 return
//             }
//             // process "val" here
//         }
//     }()
//
//     // ...
// }
