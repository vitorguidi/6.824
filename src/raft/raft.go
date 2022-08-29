package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"math/rand"
	//	"bytes"
	"sync"
	"sync/atomic"
	"time"

	//	"6.824/labgob"
	"6.824/labrpc"
)

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

const (
	FOLLOWER = iota
	LEADER
	CANDIDATE
)

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	//volatile state on everybody
	role        int
	commitIndex int
	lastApplied int

	//Persistent state for servers
	currentTerm int
	votedFor    int
	log         []LogEntry

	//volatile state leaders
	nextIndex  []int
	matchIndex []int

	//Channels
	chanApply     chan ApplyMsg
	chanHeartbeat chan bool
}

type LogEntry struct{}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	// Your code here (2A).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.currentTerm, rf.role == LEADER
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

//
// A service wants to switch to snapshot.  Only do so if Raft hasn't
// have more recent info since it communicate the snapshot on applyCh.
//
func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {

	// Your code here (2D).

	return true
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

}

//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int
	VoteGranted bool
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A,ckets to/from the ol 2B).

	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.currentTerm > args.Term {
		reply.VoteGranted = false
		reply.Term = rf.currentTerm
		Dprintf(dWarn, "S%d (term %d) got vote requested by S%d (term %d), not voting cuz candidate on lower term",
			rf.me, rf.currentTerm, args.CandidateId, args.Term)
		return
	}

	if rf.currentTerm < args.Term {
		Dprintf(dWarn, "S%d (term %d) got vote requested by S%d (term %d), converting to follower",
			rf.me, rf.currentTerm, args.CandidateId, args.Term)
		rf.currentTerm = args.Term
		rf.votedFor = -1
		rf.role = FOLLOWER
	}

	if rf.votedFor != -1 {
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		Dprintf(dWarn, "S%d (term %d) already voted for S%d, refusing vote for S%d",
			rf.me, rf.currentTerm, rf.votedFor, args.CandidateId)
		return
	}

	rf.votedFor = args.CandidateId
	reply.VoteGranted = true
	reply.Term = rf.currentTerm

	Dprintf(dWarn, "S%d granting vote for S%d on term %d", rf.me, args.CandidateId, rf.currentTerm)

}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {

	rf.mu.Lock()
	defer rf.mu.Unlock()

	reply.Term = rf.currentTerm
	reply.Success = false

	if args.Term < rf.currentTerm {
		Dprintf(dWarn, "S%d is on term %d, rejecting heartbeat from S%d at term %d", rf.me, rf.currentTerm, args.LeaderId, args.Term)
		return
	}

	if args.Term > rf.currentTerm {
		Dprintf(dWarn, "S%d is leader on term %d, stepping down for S%d at term %d", rf.me, rf.currentTerm, args.LeaderId, args.Term)
		rf.role = FOLLOWER
		rf.currentTerm = args.Term
		rf.votedFor = -1
		rf.chanHeartbeat <- true
		return
	}

	Dprintf(dWarn, "S%d received valid heartbeat from S%d at term %d", rf.me, args.LeaderId, args.Term)
	reply.Success = true
	rf.chanHeartbeat <- true

}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).

	return index, term, isLeader
}

//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

// The ticker go routine starts a new election if this peer hasn't received
// heartsbeats recently.
func (rf *Raft) ticker() {
	for rf.killed() == false {

		// Your code here to check if a leader election should
		// be started and to randomize sleeping time using
		// time.Sleep().
		delay := time.Duration(300+rand.Intn(150)) * time.Millisecond
		select {
		case <-rf.chanHeartbeat:
		case <-time.After(delay):
			rf.mu.Lock()
			if rf.role == LEADER {
				rf.mu.Unlock()
				continue
			}
			rf.votedFor = rf.me
			voteCount := 1
			rf.currentTerm++
			curTerm := rf.currentTerm
			rf.role = CANDIDATE
			Dprintf(dWarn, "S%d starting election at term %d", rf.me, rf.currentTerm)
			rf.mu.Unlock()
			chanVote := make(chan bool, 100000)

			for i := 0; i < len(rf.peers); i++ {
				if i == rf.me {
					continue
				}
				go func(peer int) {
					args := RequestVoteArgs{
						Term:        curTerm,
						CandidateId: rf.me,
					}
					reply := RequestVoteReply{}
					ok := rf.sendRequestVote(peer, &args, &reply)
					if ok && reply.VoteGranted {
						Dprintf(dWarn, "S%d got vote from S%d on term %d", rf.me, peer, args.Term)
						chanVote <- true
					} else {
						Dprintf(dWarn, "S%d did not get vote from S%d on term %d", rf.me, peer, args.Term)
						chanVote <- false
					}
				}(i)
			}
			cnt := 1
			for {
				ack := <-chanVote
				cnt++
				if ack {
					voteCount++
				}
				if cnt == len(rf.peers) || voteCount > len(rf.peers)/2 {
					break
				}
			}
			rf.mu.Lock()
			Dprintf(dWarn, "S%d counting votes on term %d: %d", rf.me, curTerm, voteCount)
			if rf.currentTerm == curTerm && voteCount > len(rf.peers)/2 {
				rf.role = LEADER
			}
			rf.mu.Unlock()
		}

	}
}

func (rf *Raft) leaderLoop() {
	for {
		rf.mu.Lock()
		curTerm := rf.currentTerm
		role := rf.role
		rf.mu.Unlock()
		if role == LEADER {
			for i := 0; i < len(rf.peers); i++ {
				go func(peer int) {
					args := AppendEntriesArgs{
						Term:     curTerm,
						LeaderId: rf.me,
					}
					reply := AppendEntriesReply{}
					rf.sendAppendEntries(peer, &args, &reply)
				}(i)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).

	rf.role = FOLLOWER
	rf.currentTerm = 0
	rf.votedFor = -1

	rf.chanApply = applyCh
	rf.chanHeartbeat = make(chan bool, 100000)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	Dprintf(dWarn, "S%d was just created", me)
	go rf.ticker()
	go rf.leaderLoop()

	return rf
}
