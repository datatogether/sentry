package main

import (
	"github.com/qri-io/archive"
	"time"
)

func StartCron(d time.Duration) (stop func()) {
	t := time.NewTicker(d)
	go func() {
		for {
			select {
			case <-t.C:
				if err := CalcBasePrimerStats(); err != nil {
					logger.Println(err)
				}
			}
		}
	}()

	return t.Stop
}

func CalcBasePrimerStats() error {
	logger.Println("[INFO] starting base primer stat calculation")
	ps, err := archive.BasePrimers(appDB, 100, 0)
	if err != nil {
		return err
	}
	for _, p := range ps {
		if err := p.CalcStats(appDB); err != nil {
			return err
		}
	}
	logger.Println("[INFO] base primer stat calculation finished.")
	return nil
}
