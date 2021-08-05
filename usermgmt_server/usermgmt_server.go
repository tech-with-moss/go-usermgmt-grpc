package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4"
	pb "github.com/tech-with-moss/go-usermgmt-grpc/usermgmt"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

func NewUserManagementServer() *UserManagementServer {
	return &UserManagementServer{}
}

type UserManagementServer struct {
	conn                *pgx.Conn
	first_user_creation bool
	pb.UnimplementedUserManagementServer
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

//When user is added, read full userlist from file into
//userlist struct, then append new user and write new userlist back to file
func (server *UserManagementServer) CreateNewUser(ctx context.Context, in *pb.NewUser) (*pb.User, error) {

	if server.first_user_creation == true {
		createSql := `
		create table users(
		  id integer,
		  unique (id),
		  name text,
		  age int
		);
	  `
		_, err := server.conn.Exec(context.Background(), createSql)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Table creation failed: %v\n", err)
			os.Exit(1)
		}

		server.first_user_creation = false
	}

	log.Printf("Received: %v", in.GetName())

	//var users_list *pb.UsersList = &pb.UsersList{}
	var user_id = int32(rand.Intn(1000))
	created_user := &pb.User{Name: in.GetName(), Age: in.GetAge(), Id: user_id}
	tx, err := server.conn.Begin(context.Background())
	if err != nil {
		log.Fatalf("conn.Begin failed: %v", err)
	}

	_, err = tx.Exec(context.Background(), "insert into users values ($1,$2,$3)",
		created_user.Id, created_user.Name, created_user.Age)
	if err != nil {
		log.Fatalf("tx.Exec failed: %v", err)
	}
	return created_user, nil

}

func (server *UserManagementServer) GetUsers(ctx context.Context, in *pb.GetUsersParams) (*pb.UsersList, error) {

	var users_list *pb.UsersList = &pb.UsersList{}

	return users_list, nil
}

func main() {
	database_url := "postgres://postgres:mysecretpassword@localhost:5432/postgres"
	var user_mgmt_server *UserManagementServer = NewUserManagementServer()
	conn, err := pgx.Connect(context.Background(), database_url)
	if err != nil {
		log.Fatalf("Unable to establish connection: %v", err)
	}
	defer conn.Close(context.Background())
	user_mgmt_server.conn = &conn
	user_mgmt_server.first_user_creation = false
	if err := user_mgmt_server.Run(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
