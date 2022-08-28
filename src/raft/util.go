package raft

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Debugging
var Debug = false

var cnt = 0
var mu sync.Mutex
var debugStart time.Time

//func DPrintf(format string, a ...interface{}) (n int, err error) {
//	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
//
//	if Debug {
//		log.Printf(format, a...)
//	}
//	return
//}

type logTopic string

const (
	dClient  logTopic = "CLNT"
	dCommit  logTopic = "CMIT"
	dDrop    logTopic = "DROP"
	dError   logTopic = "ERRO"
	dInfo    logTopic = "INFO"
	dLeader  logTopic = "LEAD"
	dLog     logTopic = "LOG1"
	dLog2    logTopic = "LOG2"
	dPersist logTopic = "PERS"
	dSnap    logTopic = "SNAP"
	dTerm    logTopic = "TERM"
	dTest    logTopic = "TEST"
	dTimer   logTopic = "TIMR"
	dTrace   logTopic = "TRCE"
	dVote    logTopic = "VOTE"
	dWarn    logTopic = "WARN"
)

func init() {
	Debug = os.Getenv("Debug") != ""
	debugStart = time.Now()
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
}

func Dprintf(topic logTopic, format string, a ...interface{}) {
	if Debug {
		time := time.Since(debugStart).Microseconds()
		//time /= 100
		prefix := fmt.Sprintf("%012d %v ", time, string(topic))
		format = prefix + format
		log.Printf(format, a...)
	}
}
