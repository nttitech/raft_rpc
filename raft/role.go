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
	r.dlog("is leader")
	go func() {
		ticker := time.NewTicker(160 * time.Millisecond)
		defer ticker.Stop()
		startTime := time.Now()

		for {
			r.LeaderSendHeartbeats()
			<-ticker.C
			elapsed := time.Since(startTime)
			if elapsed > 500*time.Millisecond{
				r.dlog("sleep")
				time.Sleep(time.Millisecond * 1001)
				r.dlog("active")
			}

			r.mu.Lock()
			if r.role != Leader {
				r.mu.Unlock()
				return
			}
			r.mu.Unlock()
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
	r.dlog("is follower at Term %d",r.currentTerm)
	go r.runElectionTimer()
}
