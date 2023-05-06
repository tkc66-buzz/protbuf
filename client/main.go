package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"protobuf/pb"

	"google.golang.org/grpc"
)

func main() {
	// 本来ならばSSL通信を行うべきだが,ローカルでのみ動かすため,Insecureを指定する.
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to listen: %v¥n", err)
	}
	defer conn.Close()

	client := pb.NewFileServiceClient(conn)
	// callListFiles(client)
	callDownload(client)
}

func callListFiles(client pb.FileServiceClient) {
	res, err := client.ListFiles(context.Background(), &pb.ListFilesRequest{})
	if err != nil {
		log.Fatalf("failed to invoke ListFiles: %v¥n", err)
	}
	fmt.Println(res.GetFilenames())
}

func callDownload(client pb.FileServiceClient) {
	stream, err := client.Download(context.Background(), &pb.DownloadRequest{Filename: "name.txt"})
	if err != nil {
		log.Fatalf("failed to invoke Download: %v¥n", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("failed to receive chunk data: %v¥n", err)
		}
		fmt.Println(res.GetData())
	}
}
