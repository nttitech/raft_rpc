package main

import(
	"my-raft/raft"
	"log"
	"sync"
	"os"
	"strconv"
	"time"
)

func makePeerIds(myID int,serversNum int) []int{
	var peerId []int
	for i:=0;i<serversNum;i++{
		if i != myID{
			peerId = append(peerId,i)
		}
	}
	return peerId
}

func makeServer(id int,serverNum int) *raft.Server{
	var server *raft.Server
	peerIds:=makePeerIds(id,serverNum)
	server= raft.NewServer(id,peerIds)
	return server
}

func run(server *raft.Server){
	// n :=len(servers)
	// for i:=0;i<n;i++{
	// 	servers[i].RunServer()
	// }
	server.RunServer()
	time.Sleep(2 * time.Second)
	var wg sync.WaitGroup

	for _,peerId:= range server.PeerIds{
		addr := ":600" + strconv.Itoa(peerId)
		err :=server.ConnectToPeer(peerId,addr)
		if err != nil{
			log.Printf("connect fail")
		}
		log.Printf("connect success from %d to %d",server.Id,peerId)
	}

	server.CMStart()
	wg.Wait()
	select{}
}

func main(){
	id,_:= strconv.Atoi(os.Args[1])
	serverNum,_ := strconv.Atoi(os.Args[2])
	server := makeServer(id,serverNum)
	run(server)
}