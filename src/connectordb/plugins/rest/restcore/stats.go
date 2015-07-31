package restcore

import (
	"connectordb/streamdb/operator/authoperator"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	//The following globals are atomically incremented/decremnted to give statistics
	StatsAuthFails = uint32(0)
	StatsQueries   = uint32(0)
	StatsInserts   = uint32(0)
	StatsErrors    = uint32(0)
	StatsPanics    = uint32(0)

	StatsTimePeriod = 1.0 * time.Minute

	QueryTimers     = map[string]*QueryTimer{}
	QueryTimePeriod = 24 * time.Hour
)

//QueryTimer holds timing statistics for a specific query
type QueryTimer struct {
	sync.Mutex

	//TimeSum is the total sum of times that the given query ran
	TimeSum float64

	//TimeVarSum is the sum of squares of times that the given query ran
	TimeVarSum float64

	//NumQueries is the number of queries that were handled in the given time period
	NumQueries int32
}

//Clear resets the QueryTimer to reload data
func (qt *QueryTimer) Clear() {
	qt.Lock()
	qt.TimeSum = 0.0
	qt.TimeVarSum = 0.0
	qt.NumQueries = 0
	qt.Unlock()
}

//Add adds a new duration
func (qt *QueryTimer) Add(t time.Duration) {
	qt.Lock()
	qt.NumQueries++
	tdiff := float64(t.Nanoseconds()) * 1e-9
	qt.TimeSum += tdiff
	qt.TimeVarSum += tdiff * tdiff
	qt.Unlock()
}

//GetClear gets the internal variance, and then clears the values
func (qt *QueryTimer) GetClear() (num int32, mean float64, variance float64) {
	qt.Lock()
	defer qt.Unlock()
	if qt.NumQueries != 0 {
		num = qt.NumQueries
		mean = qt.TimeSum / float64(num)
		variance = qt.TimeVarSum / float64(num)
	}
	qt.TimeSum = 0.0
	qt.TimeVarSum = 0.0
	qt.NumQueries = 0
	return num, mean, variance

}

//Get gets the timer values for the given query
func (qt *QueryTimer) Get() (num int32, mean float64, variance float64) {
	qt.Lock()
	defer qt.Unlock()
	if qt.NumQueries != 0 {
		num = qt.NumQueries
		mean = qt.TimeSum / float64(num)
		variance = qt.TimeVarSum / float64(num)
	}
	return num, mean, variance
}

func toDuration(t float64) time.Duration {
	return time.Duration(int64(t * 1e9))
}

//RunQueryTimers periodically gets and prints the query average runtime and variance
func RunQueryTimers() {
	for {
		time.Sleep(QueryTimePeriod)
		s := fmt.Sprintf("Statistics for the past %v:\n", QueryTimePeriod)
		for qname := range QueryTimers {
			num, mean, variance := QueryTimers[qname].GetClear()
			s += fmt.Sprintf("- %s: num=%d mean=%v sd=%v\n", qname, num, toDuration(mean), toDuration(math.Sqrt(variance-mean*mean)))
		}
		log.Info(s)
	}
}

//StatsAddFail adds an authentication failure to the statistics
func StatsAddFail(err error) {
	if err == authoperator.ErrPermissions {
		atomic.AddUint32(&StatsAuthFails, 1)
	}
}

//RunStats periodically displays query amounts and relevant data. It does not display anything
//if there was no action within a time period.
func RunStats() {
	for {
		time.Sleep(StatsTimePeriod)
		q := atomic.SwapUint32(&StatsQueries, 0)
		a := atomic.SwapUint32(&StatsAuthFails, 0)
		i := atomic.SwapUint32(&StatsInserts, 0)
		e := atomic.SwapUint32(&StatsErrors, 0)
		p := atomic.LoadUint32(&StatsPanics)

		//Only display stat view if there was something going on
		if q > 0 {
			logger := log.WithFields(log.Fields{"queries": q, "authfails": a, "inserts": i, "errors": e, "panics": p})
			if p > 0 {
				logger.Warnf("%.2f queries/s", float64(q)/StatsTimePeriod.Seconds())
			} else {
				logger.Infof("%.2f queries/s", float64(q)/StatsTimePeriod.Seconds())
			}
		}

	}
}
