package proxy

import (
	"fmt"
	"sync"
)

type Trips struct {
	trips []Trip
	mu    sync.RWMutex
}

func NewTripsRecorder() *Trips {
	return &Trips{
		trips: make([]Trip, 0),
	}
}

func (t *Trips) Add(trip Trip) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.trips = append(t.trips, trip)
	fmt.Printf("%+v\n", trip)
}

func (t *Trips) Remove(index int) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if index < 0 || index >= len(t.trips) {
		return false
	}

	t.trips = append(t.trips[:index], t.trips[index+1:]...)
	return true
}

func (t *Trips) Get(index int) (Trip, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if index < 0 || index >= len(t.trips) {
		return Trip{}, false
	}

	return t.trips[index], true
}

func (t *Trips) GetAll() []Trip {
	t.mu.RLock()
	defer t.mu.RUnlock()

	trips := make([]Trip, len(t.trips))
	copy(trips, t.trips)
	return trips
}

func (t *Trips) Len() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return len(t.trips)
}
