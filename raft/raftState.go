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
	crash bool

}

func (r *RaftState) dlog(format string, args ...interface{}) {
	format = fmt.Sprintf("[%d] ", r.id) + format
	log.Printf(format, args...)
}
func (r *RaftState) electionTimeout() time.Duration {
	return time.Duration(150+rand.Intn(150)) * time.Millisecond
}

func (r *RaftState) runElectionTimer(){
	timeoutDuration := r.electionTimeout()
	r.mu.Lock()
	termStarted := r.currentTerm
	r.mu.Unlock()
	r.dlog("election timerstarted (%v), term=%d", timeoutDuration, termStarted)

	ticker:= time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for{
		<-ticker.C
		//r.dlog("ロック前")
		r.mu.Lock()
		//r.dlog("ロック後")
		//defer r.mu.Unlock()
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
		elapsed := time.Since(r.electionResetEvent)
		if elapsed >= timeoutDuration {
			r.dlog("dosen't catch a heartbeat")
			//r.mu.Unlock()
			r.StartElection()
			r.mu.Unlock()
			return
		}
		//r.dlog("nanionani,elapsed=(%v),r.electionResetEvent=(%v)",elapsed,r.electionResetEvent)
		r.mu.Unlock()
	}
}

func (r *RaftState) StartElection(){
	r.role = Candidate
	r.currentTerm += 1
	r.electionResetEvent = time.Now()
	r.votedFor= r.id
	electionTerm := r.currentTerm
	r.dlog("becomes Candidate (currentTerm=%d);", r.currentTerm)
	//r.dlog("has peerIDs : %v",r.peerIds )
	//r.dlog("server has peerIDs : %v",r.server.PeerIds )
	votesReceived := 1
	var quorum = (len(r.peerIds)+1)/2 + 1
	for _,peerId := range r.peerIds{
		go func(peerId int){
		args := RequestVoteArgs{
			Term: electionTerm,
			CandidateId: r.id,
		}
		//r.dlog("peerId is %d",peerId)
		var reply RequestVoteReply
		err := r.server.Call(peerId,"ConsensusModule.RequestVote",args,&reply)
			if err != nil{
				//r.dlog("RequestVote fail from %d",peerId)
				return
			}
			r.mu.Lock()
			defer r.mu.Unlock()
			if r.role != Candidate{
				r.dlog("role has changed from Candidate to %v",r.role)
				return
			}
			if reply.Term > r.currentTerm{
				r.becomeFollower(reply.Term)
				return
			}

			if reply.VoteGranted && reply.Term == electionTerm{
				//r.mu.Lock()
				//defer r.mu.Unlock()
				votesReceived++
				//r.dlog("quorum is:%d votesReceived is:%d r.role is %s",quorum,votesReceived,r.role )
				if votesReceived >= quorum && r.role == Candidate{
					//r.dlog("これからリーダーになる")
					r.becomeLeader()
					//r.dlog("もうリーダーになった")
					return
				}
			}
		}(peerId)
	}

	go r.runElectionTimer()
	r.dlog(" Run another election timer, in case this election is not successful at term %d",r.currentTerm)
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
					//r.dlog("AppendEntry fail from %d",peerId)
				}
				if reply.Term > r.currentTerm{
					r.becomeFollower(reply.Term)
					return
				}
			}(peerId)
		}
}

