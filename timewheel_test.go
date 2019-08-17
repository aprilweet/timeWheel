package timeWheel

import (
	"testing"
	"time"
)

func Test(t *testing.T) {
	t.Logf("Limit: %d", Limit)

	for l := 0; l < 3; l++ {
		for i := 0; i <= Limit; i++ {
			a := callArg{
				t:     t,
				stamp: time.Now().Add(time.Duration(i) * time.Second),
			}
			if Add(i, call, a) {
				if i <= 0 || i >= Limit {
					t.Errorf("latency(%d) is not in (%d, %d), but succeed", i, 0, Limit)
					t.Fail()
				}
			} else {
				if i > 0 && i < Limit {
					t.Errorf("latency(%d) is in (%d, %d), but failed", i, 0, Limit)
					t.Fail()
				}
			}
		}
		time.Sleep(time.Second * time.Duration(l))
	}

	time.Sleep(time.Second * time.Duration(Limit))
}

type callArg struct {
	t     *testing.T
	stamp time.Time
}

func call(args interface{}) {
	a := args.(callArg)
	t := a.t
	d := a.stamp.Sub(time.Now())
	if d > (time.Second*time.Duration(-1)) && d < (time.Second*time.Duration(1)) {
		t.Logf("deviation: %v", d)
	} else {
		t.Errorf("deviation(%v) too much", d)
		t.Fail()
	}
}
