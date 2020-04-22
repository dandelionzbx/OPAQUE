package main

import (
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"net"
)
func Round1(x []byte,r *big.Int)([]byte){
	//generate alpha = H(x)+rBig*G
	//1. convert H(x) into a point on the curve
	H := sha256.Sum256(x)
	hX, hY := elliptic.P256().ScalarBaseMult(H[:])
	//fmt.Println(elliptic.P256().IsOnCurve(hX,hY))
	rText, err := r.MarshalJSON()
	if err != nil{
		fmt.Println("big random number marshal failed:",err)
	}
	tempX,tempY := elliptic.P256().Params().ScalarBaseMult(rText)
	alphaX,alphaY := elliptic.P256().Add(tempX,tempY,hX,hY)
	fmt.Println("alpha_x = ",alphaX)
	fmt.Println("alpha_y = ",alphaY)
	alpha := elliptic.Marshal(elliptic.P256(),alphaX,alphaY)
	return alpha
}

func Round3(x []byte,r *big.Int,v []byte, beta []byte)([]byte){
	r = r.Sub(elliptic.P256().Params().N,r) // -r = N -r
	rText,_ := r.MarshalJSON()
	vX,vY := elliptic.Unmarshal(elliptic.P256(),v)
	betaX,betaY := elliptic.Unmarshal(elliptic.P256(),beta)
	tempX, tempY := elliptic.P256().ScalarMult(vX,vY,rText)
	x1,y1 := elliptic.P256().Add(betaX,betaY,tempX,tempY)
	out := elliptic.Marshal(elliptic.P256(),x1,y1)
	s1 := append(x,v...)
	s2 := append(s1,out...)
	PRF := sha256.Sum256(s2)
	fmt.Println("The output of the OPRF is :",PRF[:])
	return PRF[:]
}
func main(){
	//1.根据地址找到连接
	conn, err := net.Dial("tcp","127.0.0.1:20000")
	if err != nil{
		fmt.Println("连接服务端失败：err",err)
		return
	}
	defer conn.Close()
	//2.向server端发送消息
	//2.1. generate a random number in range [0,N-1]
	q := elliptic.P256().Params().N
	rBig, err := rand.Int(rand.Reader, q)
	if err != nil{
		fmt.Println("Generate random number is invalid")
	}
	x := []byte{123}
	//2.2. compute alpha
	alpha := Round1(x,rBig)
	//2.3. transfer alpha to server
	_, err = conn.Write(alpha)
	if err != nil{
		fmt.Println("发送消息失败，err：",err)
		return
	}

	//3.客户端从服务端那里收取消息
	var buf [4096]byte
	len, err := conn.Read(buf[:])
	if err != nil{
		fmt.Println("从服务器端接收数据失败：",err)
		return
	}
	fmt.Println("从服务器接收的数据长度为：",len)
	//3.1. 解析v, beta
	vText := buf[:len/2]
	betaText := buf[len/2:len]
	//3.2. 计算最终的prf的值
	Round3(x,rBig,vText,betaText)
}
