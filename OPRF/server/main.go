package main

import (
	"crypto/elliptic"
	"fmt"
	"math/big"
	"net"
)
func Round2(k []byte, alphaX *big.Int, alphaY *big.Int)([]byte,[]byte){
	//compute v = k * G, where k is owned by server
	vX,vY := elliptic.P256().ScalarMult(elliptic.P256().Params().Gx,elliptic.P256().Params().Gy,k)
	//fmt.Println(vX,vY)
	v := elliptic.Marshal(elliptic.P256(),vX,vY)
	betaX,betaY := elliptic.P256().ScalarMult(alphaX,alphaY,k)
	//fmt.Println(betaX,betaY)
	beta := elliptic.Marshal(elliptic.P256(),betaX,betaY)
	return v,beta
}


//单独处理连接的函数
func process(conn net.Conn){
	defer conn.Close()
	//从连接中接收数据
	var buf [1024]byte
	n, err := conn.Read(buf[:])
	if err != nil{
		fmt.Println("接收客户端发来的消息失败了，err:",err)
		return
	}
	alphax, alphay := elliptic.Unmarshal(elliptic.P256(),buf[:n])
	fmt.Println("接收客户端发来的消息：",alphax,alphay)
	//服务端进行计算，并发出v，beta
	salt := []byte{88,99,100}
	v, beta := Round2(salt,alphax,alphay)
	_, err = conn.Write(append(v,beta...))
	fmt.Println("服务端发出去的消息为：",v,beta)
	length := len(v)+len(beta)
	fmt.Println("服务器发出去的消息长度为：",length)
}

func main(){
	//1.监听端口
	listener, err :=net.Listen("tcp","127.0.0.1:20000")
	if err != nil{
		fmt.Println("listen failed,err:",err)
		return
	}
	defer listener.Close()//程序退出时释放20000这个端口
	//2.接收客户端请求建立链接
	for{
		conn, err := listener.Accept() //如果没有客户端连接就阻塞，一直在等待
		if err != nil{
			fmt.Println("连接失败, err:",err)
			continue
		}
		//3.创建goroutine处理链接
		go process(conn)
	}
}