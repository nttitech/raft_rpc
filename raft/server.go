package raft 

import(
	"log"
	"sync"
	"net"
	"net/rpc"
	"strconv"
)

type Server struct{
	mu sync.Mutex
	Id int
	PeerIds []int
	peerServers map[int]*rpc.Client
	rpcServer *rpc.Server
	listener net.Listener
	ConsensusModule *RaftState
}

func NewServer(id int,peerIds []int) *Server{
	s := new(Server)

	s.Id = id
	s.PeerIds = peerIds
	s.peerServers = make(map[int]*rpc.Client)

	return s
}

func (s *Server) RunServer(){
	var ConsensusModule *RaftState

	ConsensusModule = &RaftState{
		id:s.Id,
		peerIds:s.PeerIds,
		currentTerm:0,
		votedFor:-1,
		role:Follower,
		server:s,
	}
	s.ConsensusModule =ConsensusModule
	s.rpcServer = rpc.NewServer()
	s.rpcServer.RegisterName("ConsensusModule",s.ConsensusModule)

	var err error
	var addr string
	addr = ":600" + strconv.Itoa(s.Id)
	s.listener, err = net.Listen("tcp",addr)
	if err != nil{
		log.Fatal(err)
	}
	//fmt.Printf("make new server")
	go func(){
			for{
			conn, err:=s.listener.Accept()
			if err != nil{
				log.Fatal(err)
			}

			go s.rpcServer.ServeConn(conn)
			//fmt.Printf("complete connection")
		}
	}()	

}

func (s *Server) ShutDown(){
	for id := range s.peerServers {
		if s.peerServers[id] != nil {
			s.peerServers[id].Close()
			s.peerServers[id] = nil
		}
	}
	s.listener.Close()
}

func (s *Server) GetListenAddr() net.Addr{
	return s.listener.Addr()
}

func (s *Server) ConnectToPeer(peerId int,addr string) error{
	if s.peerServers[peerId] == nil{
		client,err :=rpc.Dial("tcp",addr)
		if err != nil{
			return err
		}
		s.peerServers[peerId] = client
		s.ConsensusModule.peerIds = append(s.ConsensusModule.peerIds,peerId)
		//fmt.Printf("Dial")
	}
	return nil
}

func (s * Server) DisconnectPeer(peerId int) error{
	if s.peerServers[peerId] != nil {
		err := s.peerServers[peerId].Close()
		s.peerServers[peerId] = nil
		return err
	}
	return nil
}

func (s *Server) Call(id int,serviceMethod string,args interface{},reply interface{}) error{
	peer := s.peerServers[id]

	if peer == nil{
		log.Printf("peer does not exist")
		return nil
	}else{
		return peer.Call(serviceMethod, args, reply)
	}
}

func(s *Server)CMStart(){
	s.ConsensusModule.becomeFollower(0)
}
