package main

import (
	"context"
	"log"
	"math/rand"
	"net"

	pb "github.com/tech-with-moss/go-usermgmt-grpc/usermgmt"
	"google.golang.org/grpc"
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
	var user_id = int32(rand.Intn(100))
	created_user := &pb.User{Name: in.GetName(), Age: in.GetAge(), Id: user_id}
	server.users_list.Users = append(server.users_list.Users, created_user)
	return created_user, nil
}

func (server *UserManagementServer) GetUsers(ctx context.Context, in *pb.GetUsersParams) (*pb.UsersList, error) {
	return server.users_list, nil
}

func main() {
	var user_mgmt_server *UserManagementServer = NewUserManagementServer()
	if err := user_mgmt_server.Run(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
