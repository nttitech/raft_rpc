package main

import(
	"log"
	"os"
	"net/rpc"
)

func main(){
	leaderId := os.Args[1]
	leaderAddr := ":600" + leaderId
	client,err := rpc.Dial("tcp",leaderAddr)
	if err != nil{
		log.Printf("connect fail")
	}
	command := os.Args[2]
	// data := []byte(command)
	// _,err = conn.Write(data)
	// if err != nil{
	// 	log.Printf("can not write")
	// }
	var reply string
	err =client.Call("ConsensusModule.ReceiveCommand",&command,&reply)
	if err !=nil {
		log.Printf("can not send command")
	}
}