package raft

import(
	"time"
)
type Role int
const(
	Follower Role = iota
	Candidate
	Leader
) 

func (r *RaftState) becomeLeader(){
	r.role = Leader
	for _,peerId := range r.peerIds{
		r.nextIndex[peerId] = len(r.log)
		r.matchIndex[peerId] = -1
	}
	r.dlog("is leader")
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		startTime := time.Now()

		for {
			r.LeaderSendHeartbeats()
			<-ticker.C
			elapsed := time.Since(startTime)
			if elapsed > 10000*time.Millisecond && r.currentTerm == 1{
				r.dlog("sleep")
				r.crash = true
				time.Sleep(time.Millisecond * 10000)
				r.dlog("active")
				r.crash = false
				startTime = time.Now()
			}

			// r.mu.Lock()
			// defer r.mu.Unlock()
			if r.role != Leader {
				return
			}

		}
	}()
	// time.Sleep(time.Millisecond * 200)
	// r.becomeFollower((r.currentTerm))
}

func (r *RaftState) becomeFollower(currentTerm int){
	r.currentTerm = currentTerm
	r.votedFor = -1
	r.role = Follower
	r.electionResetEvent = time.Now()
	for{
		if !r.crash{
			break
		}
	}
	r.dlog("is follower at Term %d",r.currentTerm)
	go r.runElectionTimer()
}
