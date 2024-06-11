package raft
import(
	"fmt"
	"log"
	"sync"
	"time"
	"math/rand"
)
type RaftState struct{
	mu sync.Mutex
	id int
	peerIds []int

	currentTerm int
	votedFor int
	//log []LogEntry

	role Role
	electionResetEvent time.Time

	server *Server

}

func (r *RaftState) dlog(format string, args ...interface{}) {
	format = fmt.Sprintf("[%d] ", r.id) + format
	log.Printf(format, args...)
}
func (r *RaftState) electionTimeout() time.Duration {
	return time.Duration(500+rand.Intn(500)) * time.Millisecond
}

func (r *RaftState) runElectionTimer(){
	timeoutDuration := r.electionTimeout()
	r.mu.Lock()
	termStarted := r.currentTerm
	r.mu.Unlock()
	//r.dlog("election timerstarted (%v), term=%d", timeoutDuration, termStarted)

	ticker:= time.NewTicker(30 * time.Millisecond)
	defer ticker.Stop()
	for{
		<-ticker.C

		r.mu.Lock()
		if r.role != Candidate && r.role != Follower{
			//r.dlog("in election timerstate=%s, bailing out", r.role)
			r.mu.Unlock()
			return
		}

		if termStarted != r.currentTerm {
			//r.dlog("in election timerterm changed from %d to %d, bailing out", termStarted, r.currentTerm)
			r.mu.Unlock()
			return
		}

		if elapsed := time.Since(r.electionResetEvent); elapsed >= timeoutDuration {
			r.dlog("dosen't catch a heartbeat")
			r.StartElection()
			r.mu.Unlock()
			return
		}
		r.mu.Unlock()
	}
}

func (r *RaftState) StartElection(){
	r.role = Candidate
	r.currentTerm += 1
	r.electionResetEvent = time.Now()
	r.votedFor= r.id
	r.dlog("becomes Candidate (currentTerm=%d);", r.currentTerm)

	votesReceived := 1
	var quorum = (len(r.peerIds)+1)/2 + 1
	for _,peerId := range r.peerIds{
		go func(peerId int){
		args := RequestVoteArgs{
			Term: r.currentTerm,
			CandidateId: r.id,
		}
		//r.dlog("peerId is %d",peerId)
		var reply RequestVoteReply
			err := r.server.Call(peerId,"ConsensusModule.RequestVote",args,&reply)
			if err != nil{
				r.dlog("RequestVote fail")
			}

			if reply.Term > r.currentTerm{
				r.becomeFollower(reply.Term)
				return
			}

			if reply.VoteGranted && reply.Term == r.currentTerm{
				r.mu.Lock()
				votesReceived++
				r.dlog("quorum is:%d votesReceived is:%d r.role is %s",quorum,votesReceived,r.role )
				if votesReceived >= quorum && r.role == Candidate{
					r.becomeLeader()
					return
				}
				r.mu.Unlock()
			}
		}(peerId)
	}

	go r.runElectionTimer()
}

func (r *RaftState) LeaderSendHeartbeats(){
	for _,peerId := range r.peerIds{
		args := AppendEntryArgs{
			Term: r.currentTerm,
		}
		go func(peerId int){
			var reply AppendEntryReply
				err:=r.server.Call(peerId,"ConsensusModule.AppendEntry",args,&reply)
				if err !=nil {
					r.dlog("AppendEntry fail")
				}
				if reply.Term > r.currentTerm{
					r.becomeFollower(reply.Term)
					return
				}
			}(peerId)
		}
}

