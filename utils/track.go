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
	elapsedMS := elapsed.Milliseconds()
	elapsedMicroS := elapsed.Microseconds()
	//baseLog := log.Info().Str("func", funcName).Str("elapsed", elapsed.String())
	// Note: `μ` of elapsed_μs can not be used in jq command or json
	baseLog := log.Info().Str("func", funcName).Str("elapsed", elapsed.String()).Int64("elapsed_ms", elapsedMS).Int64("elapsed_us", elapsedMicroS)

	for k, v := range kvs {
		baseLog = baseLog.Str(k, v)
	}
	baseLog.Send()
}

func TrackLog(elapsed time.Duration, funcName string, kvs map[string]string) {
	//elapsed := time.Since(start)
	elapsedMS := elapsed.Milliseconds()
	elapsedMicroS := elapsed.Microseconds()
	//baseLog := log.Info().Str("func", funcName).Str("elapsed", elapsed.String())
	baseLog := log.Info().Str("func", funcName).Str("elapsed", elapsed.String()).Int64("elapsed_ms", elapsedMS).Int64("elapsed_us", elapsedMicroS)
	for k, v := range kvs {
		baseLog = baseLog.Str(k, v)
	}
	baseLog.Send()
}
