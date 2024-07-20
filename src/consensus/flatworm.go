package consensus

import (
	aba "flatworm/src/aba/waterbear"
	"flatworm/src/broadcast/ecrbc"
	"flatworm/src/broadcast/rbc"
	"flatworm/src/config"
	"flatworm/src/logging"
	"flatworm/src/quorum"
	"flatworm/src/utils"
	"fmt"
	"log"
	"time"
)

func StartFlatWorm(data []byte) {
	if rbcType == RBC {
		rbc_time.Set(rbc.GetEpoch(), utils.MakeTimestamp())
		if rbc.GetEpoch() == 1 {
			t1 = utils.MakeTimestamp()
		}
		if config.MaliciousNode() && config.MaliciousMode() == 2 {
			intid, err := utils.Int64ToInt(id)
			if err != nil {
				log.Fatal("Failed transform int64 to int", err)
			}
			if intid < quorum.FSize() {
				log.Printf("I'm a malicious node %v, don't propose RBC!", id)
			} else {
				rbc.StartRBC(rbc.GetRBCInstanceID(iid), data)
			}

		} else {
			rbc.StartRBC(rbc.GetRBCInstanceID(iid), data)
		}
		go MonitorFlatWormRBCStatus(rbc.GetEpoch())
	} else if rbcType == ECRBC {
		rbc_time.Set(ecrbc.GetEpoch(), utils.MakeTimestamp())
		if ecrbc.GetEpoch() == 1 {
			t1 = utils.MakeTimestamp()
		}
		if config.MaliciousNode() && config.MaliciousMode() == 2 {
			intid, err := utils.Int64ToInt(id)
			if err != nil {
				log.Fatal("Failed transform int64 to int", err)
			}
			if intid < quorum.FSize() {
				log.Printf("I'm a malicious node %v, don't propose RBC!", id)
			} else {
				ecrbc.StartECRBC(ecrbc.GetRBCInstanceID(iid), data)
			}
		} else {
			ecrbc.StartECRBC(ecrbc.GetRBCInstanceID(iid), data)
		}

		go MonitorFlatWormECRBCStatus(ecrbc.GetEpoch())
	}
}

func MonitorFlatWormRBCStatus(e int) {
	for {
		if rbc.GetEpoch() > e {
			return
		}
		rbc_deliver := 0
		for i := 0; i < n; i++ {
			instanceid := rbc.GetRBCInstanceID(members[i])
			if rbc.QueryStatus(instanceid) {
				astatus.Insert(instanceid, true)
				rbc_deliver++
			}
		}
		if rbc_deliver >= quorum.QuorumSize() {
			rbc.IncrementEpoch()
			curStatus.Set(READY)
			return
		}
		time.Sleep(time.Duration(sleepTimerValue) * time.Millisecond)
	}
}

func MonitorFlatWormECRBCStatus(e int) {
	for {
		if ecrbc.GetEpoch() > e {
			return
		}
		rbc_deliver := 0
		for i := 0; i < n; i++ {
			instanceid := ecrbc.GetRBCInstanceID(members[i])
			if ecrbc.QueryStatus(instanceid) {
				astatus.Insert(instanceid, true)
				rbc_deliver++
			}
		}
		if rbc_deliver >= quorum.QuorumSize() {
			ecrbc.IncrementEpoch()
			curStatus.Set(READY)
			return
		}
		time.Sleep(time.Duration(sleepTimerValue) * time.Millisecond)
	}
}

func MonitorStartABA() {
	for {
		var e int
		switch rbcType {
		case RBC:
			e = rbc.GetEpoch()
		case ECRBC:
			e = ecrbc.GetEpoch()
		}
		if epoch.Get() < e && otherlock.Get() != 1 {
			otherlock.Set(1)
			for i := 0; i < n; i++ {
				instanceid := GetInstanceID(members[i])
				if astatus.GetStatus(instanceid) {
					go StartFWABA(instanceid, 1)
				} else {
					go StartFWABA(instanceid, 0)
				}
			}
		}
		time.Sleep(time.Duration(sleepTimerValue) * time.Millisecond)
	}
}

func MonitorFlatWormABAStatus() {
	for {
		var e int
		switch rbcType {
		case RBC:
			e = rbc.GetEpoch()
		case ECRBC:
			e = ecrbc.GetEpoch()
		}
		if epoch.Get() > e {
			continue
		}
		for i := 0; i < n; i++ {
			instanceid := GetInstanceID(members[i])
			status := aba.QueryStatus(instanceid)
			if !fstatus.GetStatus(instanceid) && status {
				fstatus.Insert(instanceid, true)
				go UpdateFWOutput(instanceid)
			}
		}
		time.Sleep(time.Duration(sleepTimerValue) * time.Millisecond)
	}
}

func UpdateFWOutput(instanceid int) {
	p := fmt.Sprintf("[Consensus] Update Output for instance %v in epoch %v", instanceid, epoch.Get())
	logging.PrintLog(true, logging.NormalLog, p)
	value := aba.QueryValue(instanceid)

	if value == 0 {
		outputCount.Increment()
	} else {
		outputSize.Increment()
		outputCount.Increment()
		go UpdateWBOutputSet(instanceid)
	}
	if outputCount.Get() == n {
		ExitEpoch()
		output.Init()
		outputCount.Init()
		outputSize.Init()
		epoch.Increment()
		otherlock.Init()
		t1 = utils.MakeTimestamp()
		return
	}
}

func UpdateWBOutputSet(instanceid int) {
	for {
		var v []byte
		switch rbcType {
		case RBC:
			v = rbc.QueryReq(instanceid)
		case ECRBC:
			v = ecrbc.QueryReq(instanceid)
		}
		if v != nil {
			output.AddItem(v)
			break
		} else {
			time.Sleep(time.Duration(sleepTimerValue) * time.Millisecond)
		}
	}
}

func StartFWABA(instanceid int, input int) {
	if bstatus.GetStatus(instanceid) {
		return
	}
	bstatus.Insert(instanceid, true)
	//log.Printf("[%v] Starting ABA from zero with input %v in epoch %v", instanceid, input,epoch.Get())
	if config.MaliciousNode() {
		switch config.MaliciousMode() {
		case 0:
			intid, err := utils.Int64ToInt(id)
			if err != nil {
				log.Fatal("Failed transform int64 to int", err)
			}
			if intid < quorum.FSize() {
				//log.Printf("I'm a malicious node %v, start ABA %v with %v!",id,instanceid,0)
				p := fmt.Sprintf("[%v] I'm a malicious node %v, start ABA with %v!", id, instanceid, 0)
				logging.PrintLog(verbose, logging.NormalLog, p)
				input = 0
			}
		case 1:
			intid, err := utils.Int64ToInt(id)
			if err != nil {
				log.Fatal("Failed transform int64 to int", err)
			}
			if intid > 2*quorum.FSize() {
				//log.Printf("I'm a malicious node %v, start ABA %v with %v!",id,instanceid,0)
				p := fmt.Sprintf("[%v] I'm a malicious node %v, start ABA with %v!", id, instanceid, 0)
				logging.PrintLog(verbose, logging.NormalLog, p)
				input = 0
			}
		case 3:
			intid, err := utils.Int64ToInt(id)
			if err != nil {
				log.Fatal("Failed transform int64 to int", err)
			}
			if intid < quorum.FSize() {
				//log.Printf("I'm a malicious node %v, start ABA %v with %v!",id,instanceid,input ^ 1)
				p := fmt.Sprintf("[%v] I'm a malicious node %v, start ABA with %v!", id, instanceid, input^1)
				logging.PrintLog(verbose, logging.NormalLog, p)
				if input != 2 {
					input = input ^ 1
				}

			}

		}

	}
	switch consensus {
	case FlatWorm:
		aba.StartABAFromRoundZero(instanceid, input)
	default:
		log.Fatalf("This script only supports WaterBear and biased WaterBear")
	}
}

func InitFlatWormBFT(ct bool) {
	InitStatus(n)
	//aba.SetEpoch(epoch.Get())
	if rbcType == RBC {
		rbc.SetEpoch(epoch.Get())
	} else if rbcType == ECRBC {
		rbc.SetEpoch(epoch.Get())
		ecrbc.SetEpoch(epoch.Get())
	}

	aba.InitCoinType(ct)
}
