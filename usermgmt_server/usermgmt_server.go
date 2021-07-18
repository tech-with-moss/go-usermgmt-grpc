package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"

	pb "github.com/tech-with-moss/go-usermgmt-grpc/usermgmt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	port = ":50051"
)

func NewUserManagementServer() *UserManagementServer {
	return &UserManagementServer{
		users_list: &pb.UsersList{},
	}
}

type UserManagementServer struct {
	pb.UnimplementedUserManagementServer
	users_list *pb.UsersList
}

func (server *UserManagementServer) Run() error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserManagementServer(s, server)
	log.Printf("server listening at %v", lis.Addr())
	return s.Serve(lis)
}

func (server *UserManagementServer) CreateNewUser(ctx context.Context, in *pb.NewUser) (*pb.User, error) {
	log.Printf("Received: %v", in.GetName())

	readBytes, err := ioutil.ReadFile("../users.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("%s: File not found.  Creating new file.\n", "../users.json")
		} else {
			log.Fatalln("Error reading file:", err)
		}
	}

	var users_list *pb.UsersList = &pb.UsersList{}

	if err := protojson.Unmarshal(readBytes, users_list); err != nil {
		log.Fatalf("Failed to parse user list: %v", err)
	}

	var user_id = int32(rand.Intn(100))
	created_user := &pb.User{Name: in.GetName(), Age: in.GetAge(), Id: user_id}
	users_list.Users = append(users_list.Users, created_user)
	jsonBytes, err := protojson.Marshal(users_list)
	if err != nil {
		log.Fatalf("JSON Marshaling failed: %v", err)
	}
	if err := ioutil.WriteFile("../users.json", jsonBytes, 0664); err != nil {
		log.Fatalf("Failed write to file: %v", err)
	}
	return created_user, nil
}

func (server *UserManagementServer) GetUsers(ctx context.Context, in *pb.GetUsersParams) (*pb.UsersList, error) {
	jsonBytes, err := ioutil.ReadFile("../users.json")
	if err != nil {
		log.Fatalf("Failed read from file: %v", err)
	}
	var users_list *pb.UsersList = &pb.UsersList{}
	if err := protojson.Unmarshal(jsonBytes, users_list); err != nil {
		log.Fatalf("Unmarshaling failed: %v", err)
	}

	return users_list, nil
}

func main() {
	var user_mgmt_server *UserManagementServer = NewUserManagementServer()
	if err := user_mgmt_server.Run(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
