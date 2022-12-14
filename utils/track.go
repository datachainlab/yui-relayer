package utils

import (
	"time"

	"github.com/rs/zerolog/log"
)

// e.g. Caller: set code at first line in function
// ```
// defer utils.Track(time.Now(), "parseFile()", map[string]string{"data1": "500", "data2": "800"})
// ```
func Track(start time.Time, funcName string, kvs map[string]string) {
	elapsed := time.Since(start)
	baseLog := log.Info().Str("func", funcName).Str("elapsed", elapsed.String())
	for k, v := range kvs {
		baseLog = baseLog.Str(k, v)
	}
	baseLog.Send()
}

func TrackLog(elapsed time.Duration, funcName string, kvs map[string]string) {
	baseLog := log.Info().Str("func", funcName).Str("elapsed", elapsed.String())
	for k, v := range kvs {
		baseLog = baseLog.Str(k, v)
	}
	baseLog.Send()
}
