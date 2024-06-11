package raft

import(
	"time"
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

type AppendEntryArgs struct{
	Term     int
	LeaderId int

	PrevLogIndex int
	PrevLogTerm  int
	//Entries      []LogEntry
	LeaderCommit int
}
type AppendEntryReply struct{
	Term int
	Success bool
}

func (r *RaftState) RequestVote(args RequestVoteArgs,reply *RequestVoteReply) error{
	r.mu.Lock()
	defer r.mu.Unlock()

	//r.dlog("is received requestVote,args.Term:%d,r.currentTerm:%d,r.votedfor:%d,args.CandidateId:%d",args.Term,r.currentTerm,r.votedFor,args.CandidateId)
	if args.Term > r.currentTerm && (r.votedFor == -1 || r.votedFor == args.CandidateId){
		r.becomeFollower(args.Term)
		reply.Term = r.currentTerm
		reply.VoteGranted = true
		r.votedFor = args.CandidateId
		r.dlog("vote for %d",r.votedFor)
		return  nil
	}
	return nil
}

func (r *RaftState) AppendEntry(args AppendEntryArgs,reply *AppendEntryReply) error{
	if args.Term > r.currentTerm{
		r.becomeFollower(args.Term)
		return nil
	}

	if args.Term < r.currentTerm{
		reply.Term = r.currentTerm
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
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