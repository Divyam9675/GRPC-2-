package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "os"
    "os/signal"

  

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/reflection"
    "google.golang.org/grpc/status"

    "GRPC/grpcpb"

    "database/sql"
    _ "github.com/go-sql-driver/mysql"
   
    "google.golang.org/grpc"
)

//var collection *mongo.Collection
//var db *sql.DB
type server struct {
    db *sql.DB
	grpcpb.UnimplementedBlogServiceServer
}

type blogItem struct {
    
    AuthorID int64 `db:"author_id" json:"author_id"`
    Content  string `db:"content" json:"content"`
    Title     string    `db:"title" json:"title"`
   
}

// connect returns SQL database connection from the pool
func (s *server) connect(ctx context.Context) (*sql.Conn, error) {
    c, err := s.db.Conn(ctx)
    if err != nil {
        return nil, status.Error(codes.Unknown, "failed to connect to database-> "+err.Error())
    }
    return c, nil
}
func NewServer(db *sql.DB) grpcpb.BlogServiceServer {
    return &server{db: db}
}

func (s *server) CreateBlog(ctx context.Context, req *grpcpb.CreateBlogRequest) (*grpcpb.CreateBlogResponse, error) {
    fmt.Println("Create blog request")

    // get SQL connection from pool
    c, err := s.connect(ctx)
    if err != nil {
        return nil, err
    }
    defer c.Close()

    //PARSE THE DATA
    blog := req.GetBlog()

    //MAP IT BLOG ITEM
    data := blogItem{
        AuthorID: blog.GetAuthorId(),
        Title:    blog.GetTitle(),
        Content:  blog.GetContent(),
    }

    //INSERT DATA THROUGH Mysql
    // insert ToDo entity data
    _, err = c.ExecContext(ctx, "INSERT INTO blog(`author_id`, `content`, `title`) VALUES(?, ?, ?)",
        data.AuthorID, data.Content,data.Title)
    if err != nil {
        return nil, status.Error(codes.Unknown, "failed to insert into Blog-> "+err.Error())
    }
   

    return &grpcpb.CreateBlogResponse{
        Blog: &grpcpb.Blog{
            //Id:       oid.Hex(),
            AuthorId: blog.GetAuthorId(),
            Title:    blog.GetTitle(),
            Content:  blog.GetContent(),
        },
    }, nil

}


func dataToBlogPb(data *blogItem) *grpcpb.Blog {
    return &grpcpb.Blog{
        //Id:       data.ID.Hex(),
        AuthorId: data.AuthorID,
        Content:  data.Content,
        Title:    data.Title,
    }
}





 


func main() {
    // if we crash the go code, we get the file name and line number
    log.SetFlags(log.LstdFlags | log.Lshortfile)

   

    // // FOR MYSQL 
     db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/blog_pb")
    if err != nil {
        // fmt.Print(err.Error())
        fmt.Println("Error creating DB:", err)
        fmt.Println("To verify, db is:", db)
    }
    defer db.Close()
    fmt.Println("Successfully  Connected to MYSQl")
    
    seri := NewServer(db)

    // //END MYSQL
    
    lis, err := net.Listen("tcp", "0.0.0.0:50051")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    opts := []grpc.ServerOption{}
    s := grpc.NewServer(opts...)
    grpcpb.RegisterBlogServiceServer(s, seri)
    // Register reflection service on gRPC server.
    reflection.Register(s)

    go func() {
        fmt.Println("Starting Server...")
        if err := s.Serve(lis); err != nil {
            log.Fatalf("failed to serve: %v", err)
        }
    }()

    // Wait for Control C to exit
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, os.Interrupt)

    // Block until a signal is received
    <-ch

    
    
    // Second step : closing the listener
    fmt.Println("Closing the listener")
    if err := lis.Close(); err != nil {
        log.Fatalf("Error on closing the listener : %v", err)
    }
    // Finally, we stop the server
    fmt.Println("Stopping the server")
    s.Stop()
    fmt.Println("End of Program")
}