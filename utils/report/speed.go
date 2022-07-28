package report

import (
	"go.uber.org/atomic"
	"log"
	"time"
)

var (
	Speed      atomic.Uint64
	SpeedTotal atomic.Uint64
	Progress   string
)

func Tick(second uint64) {
	for range time.Tick(time.Duration(second) * time.Second) {
		nowSpeed := Speed.Load()
		SpeedTotal.Add(nowSpeed)
		log.Printf("total:%d  speed: %d/s  %s",
			SpeedTotal.Load(),
			Speed.Load()/second,
			Progress,
		)
		Speed.Sub(nowSpeed)
	}
}
