package consensus

import (
	"encoding/json"
	"flatworm/src/logging"
	"flatworm/src/message"
	"flatworm/src/utils"
	"fmt"
	"log"
	"time"
)

var verbose bool //verbose level
var id int64     //id of server
var iid int      //id in type int, start a RBC using it to instanceid
var errs error
var queue Queue         // cached client requests
var queueHead QueueHead // hash of the request that is in the fist place of the queue
var sleepTimerValue int // sleeptimer for the while loop that continues to monitor the queue or the request status
var consensus ConsensusType
var rbcType RbcType
var n int
var members []int
var t1 int64
var baseinstance int
var rbc_time utils.IntInt64Map

var batchSize int
var requestSize int

func ExitEpoch() {
	t2 := utils.MakeTimestamp()
	txcount := 0
	//如果tx相同则导致output只有一个rep
	for _, v := range output.SetList() {
		rawops := message.DeserializeRawOPS(v)
		txcount += len(rawops.OPS)
	}
	lat, exist := rbc_time.Get(epoch.Get())
	if !exist {
		lat = t1
	}
	p := fmt.Sprintf("*****epoch %v: output size %v, tx count %d, latency %v ms, epoch latency %d ms, tps %d", epoch.Get(), outputSize.Get(), txcount*outputSize.Get(), t2-t1, t2-lat, int64(txcount*outputSize.Get()*1000)/(t2-t1))
	log.Printf(p)
	logging.PrintLog(true, logging.EvaluationLog, p)
}

func CaptureRBCLat() {
	t3 := utils.MakeTimestamp()
	if (t3 - t1) == 0 {
		log.Printf("Latancy is zero!")
		return
	}
	log.Printf("*****RBC phase ends with %v ms", t3-t1)

}

func CaptureLastRBCLat() {
	t3 := utils.MakeTimestamp()
	if (t3 - t1) == 0 {
		log.Printf("Latancy is zero!")
		return
	}
	log.Printf("*****Final RBC phase ends with %v ms", t3-t1)

}

func RequestMonitor() {
	for {
		if curStatus.Get() == READY && !queue.IsEmpty() {
			curStatus.Set(PROCESSING)

			batch := queue.GrabWtihMaxLenAndClear()
			rops := message.RawOPS{
				OPS: batch,
			}
			data, err := rops.Serialize()
			if err != nil {
				continue
			}
			StartProcessing(data)
		} else {
			time.Sleep(time.Duration(sleepTimerValue) * time.Millisecond)
		}
	}
}

func HandleRequest(request []byte, hash string) {
	//log.Printf("Handling request")
	//rawMessage := message.DeserializeMessageWithSignature(request)
	//m := message.DeserializeClientRequest(rawMessage.Msg)

	/*if !cryptolib.VerifySig(m.ID, rawMessage.Msg, rawMessage.Sig) {
		log.Printf("[Authentication Error] The signature of client request has not been verified.")
		return
	}*/
	//log.Printf("Receive len %v op %v\n",len(request),m.OP)
	batchSize = 1
	requestSize = len(request)
	queue.Append(request)
}

func HandleBatchRequest(requests []byte) {
	requestArr := DeserializeRequests(requests)
	//var hashes []string
	Len := len(requestArr)
	log.Printf("Handling batch requests with len %v\n", Len)
	//for i:=0;i<Len;i++{
	//	hashes = append(hashes,string(cryptolib.GenHash(requestArr[i])))
	//}
	//for i:=0;i<Len;i++{
	//	HandleRequest(requestArr[i],hashes[i])
	//}
	/*for i:=0;i<Len;i++{
		rawMessage := message.DeserializeMessageWithSignature(requestArr[i])
		m := message.DeserializeClientRequest(rawMessage.Msg)

		if !cryptolib.VerifySig(m.ID, rawMessage.Msg, rawMessage.Sig) {
			log.Printf("[Authentication Error] The signature of client logout request has not been verified.")
			return
		}
	}*/
	batchSize = Len
	requestSize = len(requestArr[0])
	queue.AppendBatch(requestArr)
}

func DeserializeRequests(input []byte) [][]byte {
	var requestArr [][]byte
	json.Unmarshal(input, &requestArr)
	return requestArr
}
