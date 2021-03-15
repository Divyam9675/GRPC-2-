package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"Product/productpb"
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

	c := productpb.NewProductServiceClient(cc)

	// create Blog
	fmt.Println("Creating the product")
	blog := &productpb.Product{
		DevelopedBy: "Divyam",
		Title:    "My First Product",
		Content:  "Content of the first Product",
		Budget: "400000",
	}
	createBlogRes, err := c.CreateBlog(context.Background(), &productpb.CreateProductRequest{Product: blog})
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
	fmt.Printf("Product has been created: %v", createBlogRes)
	blogID := createBlogRes.GetProduct().GetId()

	// read Blog
    fmt.Println("Reading the blog")

    _, err2 := c.ReadBlog(context.Background(), &productpb.ReadProductRequest{ProductId: "6046faaf5e8e5e2a5ccdf568"})
    if err2 != nil {
	fmt.Printf("Error happened while reading: %v \n", err2)
    }

    readBlogReq := &productpb.ReadProductRequest{ProductId: blogID}
    readBlogRes, readBlogErr := c.ReadBlog(context.Background(), readBlogReq)
    if readBlogErr != nil {
	fmt.Printf("Error happened while reading: %v \n", readBlogErr)
    }

    fmt.Printf("Blog was read: %v \n", readBlogRes)
   

   newBlog := &productpb.Product{
	Id:       blogID,
	DevelopedBy: "Kalai Seenan",
	Title:    "My Second Project (edited)",
	Content:  "Content of the Second project, with some awesome additions!",
}
updateRes, updateErr := c.UpdateBlog(context.Background(), &productpb.UpdateProductRequest{Product: newBlog})
if updateErr != nil {
	fmt.Printf("Error happened while updating: %v \n", updateErr)
}
fmt.Printf("Blog was updated: %v\n", updateRes)

// delete Blog
deleteRes, deleteErr := c.DeleteBlog(context.Background(), &productpb.DeleteProductRequest{ProductId: blogID})

if deleteErr != nil {
	fmt.Printf("Error happened while deleting: %v \n", deleteErr)
}
fmt.Printf("Blog was deleted: %v \n", deleteRes)

// list Blogs

stream, err := c.ListBlog(context.Background(), &productpb.ListProductRequest{})
if err != nil {
	log.Fatalf("error while calling ListBlog RPC: %v", err)
}
for {
	res, err := stream.Recv()
	if err == io.EOF {
		break
	}
	if err != nil {
		log.Fatalf("Something happened: %v", err)
	}
	fmt.Println(res.GetProduct())
}
}