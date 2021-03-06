package main

import (
    "context"
    "fmt"
    "GRPC/grpcpb"
    
    "log"
   
    "google.golang.org/grpc"
)


func main() {

    fmt.Println("Blog Client")

    opts := grpc.WithInsecure()

    cc, err := grpc.Dial("localhost:50051", opts)
    if err != nil {
        log.Fatalf("could not connect: %v", err)
    }
    defer cc.Close()

    c := grpcpb.NewBlogServiceClient(cc)

    // create Blog
    fmt.Println("Creating the blog")
    blog := &grpcpb.Blog{
        AuthorId: 1,
        Title:    "My First Blog",
        Content:  "Content of the first blog",
    }
    createBlogRes, err := c.CreateBlog(context.Background(), &grpcpb.CreateBlogRequest{Blog: blog})
    if err != nil {
        log.Fatalf("Unexpected error: %v", err)
    }
    fmt.Printf("Blog has been created: %v", createBlogRes)
  
  
}