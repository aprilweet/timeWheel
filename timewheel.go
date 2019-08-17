package timeWheel

import (
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	Seconds int = 60
	Minutes int = 60
	Hours   int = 24
	Limit   int = Seconds * Minutes * Hours
)

type (
	Callback func(interface{})

	timer struct {
		latency  int
		callback Callback
		args     interface{}
	}

	slot []*timer
)

var second [Seconds]*slot
var minute [Minutes]*slot
var hour [Hours]*slot

var curS int = 0
var curM int = 0
var curH int = 0

var lock sync.Mutex

func init() {
	go func() {
		log.Printf("%d:%d:%d", curH, curM, curS)
		t := time.NewTicker(time.Second)
		for range t.C {
			lock.Lock()
			curS += 1
			if curS == Seconds {
				curS = 0
				curM += 1
				if curM == Minutes {
					curM = 0
					curH += 1
					if curH == Hours {
						curH = 0
					}
				}
			}
			log.Printf("%d:%d:%d", curH, curM, curS)
			expire()
			lock.Unlock()
		}
	}()
}

func Add(latency int, callback Callback, args interface{}) bool {
	lock.Lock()
	defer lock.Unlock()
	return add(latency, callback, args)
}

func add(latency int, callback Callback, args interface{}) bool {
	if latency >= Limit || latency <= 0 {
		return false
	}

	left := latency

	s := left % Seconds
	left /= Seconds
	m := left % Minutes
	left /= Minutes
	h := left % Hours
	// log.Print(latency, h, m, s)

	var pos **slot

	if h > 0 {
		pos = &hour[(curH+h)%Hours]
		left = Seconds*Minutes*h - Seconds*curM - curS
	} else if m > 0 {
		pos = &minute[(curM+m)%Minutes]
		left = Seconds*m - curS
	} else {
		pos = &second[(curS+s)%Seconds]
		left = s
	}

	if *pos == nil {
		*pos = &slot{}
	}

	t := timer{
		latency:  latency - left,
		callback: callback,
		args:     args,
	}
	// log.Print(t)

	**pos = append(**pos, &t)
	return true
}

func expire() {
	const (
		Second = "second"
		Minute = "minute"
		Hour   = "hour"
	)

	for level, pos := range map[string]**slot{Second: &second[curS], Minute: &minute[curM], Hour: &hour[curH]} {
		if *pos != nil && len(**pos) > 0 {
			for _, t := range **pos {
				if t.latency > 0 {
					if level == Second {
						panic(fmt.Errorf("%s latency: %d", level, t.latency))
					}
					if !add(t.latency, t.callback, t.args) {
						panic(fmt.Errorf("%s re-add failed", level))
					}
				} else if t.latency == 0 {
					t.callback(t.args)
				} else {
					panic(fmt.Errorf("%s latency: %d", level, t.latency))
				}
			}
			*pos = nil
		}
	}
}
