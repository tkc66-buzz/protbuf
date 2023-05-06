package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"protobuf/pb"
	"time"

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
	// callDownload(client)
	// callUpload(client)
	callUploadAndNotifyProgress(client)
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

func callUpload(client pb.FileServiceClient) {
	filename := "sports.txt"
	path := "/Users/takeshiwatanabe/EureWorks/udemy/protobuf/storage/" + filename

	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("failed to open file: %v¥n", err)
	}
	defer file.Close()

	stream, err := client.Upload(context.Background())
	if err != nil {
		log.Fatalf("failed to invoke Upload: %v¥n", err)
	}

	buf := make([]byte, 5)
	for {
		n, err := file.Read(buf)
		if err != nil {
			log.Fatalf("failed to read file: %v¥n", err)
		}
		if n == 0 || err == io.EOF {
			break
		}
		if sendErr := stream.Send(&pb.UploadRequest{Data: buf[:n]}); sendErr != nil {
			log.Fatalf("failed to send chunk data: %v¥n", sendErr)
		}
		time.Sleep(1 * time.Second)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("failed to receive response: %v¥n", err)
	}
	fmt.Println(res.GetSize())
}

func callUploadAndNotifyProgress(client pb.FileServiceClient) {
	filename := "sports.txt"
	path := "/Users/takeshiwatanabe/EureWorks/udemy/protobuf/storage/" + filename

	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("failed to open file: %v¥n", err)
	}
	defer file.Close()

	stream, err := client.UploadAndNotifyProgress(context.Background())
	if err != nil {
		log.Fatalf("failed to invoke UploadAndNotifyProgress: %v¥n", err)
	}

	// request
	buf := make([]byte, 5)
	go func() {
		for {
			n, err := file.Read(buf)
			if err != nil {
				log.Fatalf("failed to read file: %v¥n", err)
			}
			if n == 0 || err == io.EOF {
				break
			}
			if sendErr := stream.Send(&pb.UploadAndNotifyProgressRequest{Data: buf[:n]}); sendErr != nil {
				log.Fatalf("failed to send chunk data: %v¥n", sendErr)
			}
			time.Sleep(1 * time.Second)
		}
		if err := stream.CloseSend(); err != nil {
			log.Fatalf("failed to close send: %v¥n", err)
		}
	}()

	// response
	ch := make(chan struct{})
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("failed to receive chunk data: %v¥n", err)
			}
			fmt.Println(res.GetMessage())
		}
		close(ch)
	}()
	<-ch
}
