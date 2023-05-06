package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"protobuf/pb"
	"time"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedFileServiceServer
}

func (s *server) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	fmt.Println("ListFiles was invoked")
	dir := "/Users/takeshiwatanabe/EureWorks/udemy/protobuf/storage"
	paths, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var filenames []string
	for _, path := range paths {
		if !path.IsDir() {
			filenames = append(filenames, path.Name())
		}
	}
	res := &pb.ListFilesResponse{
		Filenames: filenames,
	}
	return res, nil
}

func (s *server) Download(req *pb.DownloadRequest, stream pb.FileService_DownloadServer) error {
	fmt.Println("Download was invoked")
	filename := req.GetFilename()
	path := "/Users/takeshiwatanabe/EureWorks/udemy/protobuf/storage/" + filename

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	buf := make([]byte, 5)
	for {
		n, err := file.Read(buf)
		if err != nil {
			return err
		}
		if n == 0 || err == io.EOF {
			break
		}
		res := &pb.DownloadResponse{Data: buf[:n]}
		sendErr := stream.Send(res)
		if sendErr != nil {
			return sendErr
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v¥n", err)
	}

	s := grpc.NewServer()
	pb.RegisterFileServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v¥n", err)
	}

}
