package raft

import(
	"time"
	"fmt"
)

type RequestVoteArgs struct{
	Term int
	CandidateId int
	LastLogIndex int
	LastLogTerm int
}

type RequestVoteReply struct{
	Term int
	VoteGranted bool
}

type LogEntry struct{
	Term int
	Command string
}

type AppendEntryArgs struct{
	Term     int
	LeaderId int

	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}
type AppendEntryReply struct{
	Term int
	Success bool
}

func (r *RaftState) RequestVote(args RequestVoteArgs,reply *RequestVoteReply) error{
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.crash {
		return fmt.Errorf("server is temporarily unavailable")
	}

	if args.Term > r.currentTerm{
		r.becomeFollower(args.Term)
	}
	//r.dlog("is received requestVote,args.Term:%d,r.currentTerm:%d,r.votedfor:%d,args.CandidateId:%d",args.Term,r.currentTerm,r.votedFor,args.CandidateId)
	if args.Term == r.currentTerm && (r.votedFor == -1 || r.votedFor == args.CandidateId ){
		reply.Term = r.currentTerm
		reply.VoteGranted = true
		r.votedFor = args.CandidateId
		r.dlog("vote for %d",r.votedFor)
		return  nil
	} else{
		reply.Term = r.currentTerm
		reply.VoteGranted = false
	}
	return nil
}

func (r *RaftState) AppendEntry(args AppendEntryArgs,reply *AppendEntryReply) error{
	r.mu.Lock()
	defer r.mu.Unlock()
	// if r.crash {
	// 	time.Sleep(10000*time.Millisecond)
	// }
	if args.Entries != nil{
		if r.checkConsistensy(args){
			consistentLog := r.log[:args.PrevLogIndex + 1]
			r.log = append(consistentLog,args.Entries...)
			reply.Success = true
			r.dlog("has log entry:%v",r.log)
		}else{
			reply.Success = false
			r.dlog("has log entry:%v",r.log)
		}
		return nil
	}

	if args.Term > r.currentTerm{
		r.becomeFollower(args.Term)
		return nil
	}

	if args.Term < r.currentTerm{
		reply.Term = r.currentTerm
		return nil
	}

	if args.Term > r.currentTerm{
		r.becomeFollower(args.Term)
	}
	if args.Term == r.currentTerm{
		if r.role != Follower{
			r.becomeFollower(args.Term)
		}

		r.electionResetEvent = time.Now()
		r.votedFor = -1
	}
	return nil
}