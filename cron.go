package main

import (
	"github.com/datatogether/core"
	"time"
)

// StartCron spins up a ticker that will run cron jobs (currently just calculating base Primer Stats)
// at a given interval
// TODO - this should move to a que to clear the way for running lots & lots
// of copies of sentry
func StartCron(d time.Duration) (stop func()) {
	t := time.NewTicker(d)
	go func() {
		for {
			select {
			case <-t.C:
				if err := CalcBasePrimerStats(); err != nil {
					log.Debug(err)
				}
			}
		}
	}()

	return t.Stop
}

// CalcBasePrimerStats calculates the tallies for primers that have no parent
// Primer by calculating stats for all of their child primers & working up the chain
// This process is very computationally expensive, and should be run selectively
// TODO - this currently spins up at least 1Gig of ram to do it's work, need to refactor
func CalcBasePrimerStats() error {
	log.Info("[INFO] starting base primer stat calculation")
	ps, err := core.BasePrimers(appDB, 100, 0)
	if err != nil {
		return err
	}
	for _, p := range ps {
		if err := p.CalcStats(appDB); err != nil {
			return err
		}
	}
	log.Info("[INFO] base primer stat calculation finished.")
	return nil
}
