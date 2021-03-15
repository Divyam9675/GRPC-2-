package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"gopkg.in/mgo.v2/bson"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"Product/productpb"
	"google.golang.org/grpc"
)

//var collection *mongo.Collection

type server struct {
	productpb.UnimplementedProductServiceServer
}


func (*server) CreateBlog(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.CreateProductResponse, error) {
	fmt.Println("Create blog request")
	blog := req.GetProduct()

	data := productItem{
		DevelopedBy: blog.GetDevelopedBy(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
		Budget:   blog.GetBudget(),
	}

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot convert to OID"),
		)
	}

	return &productpb.CreateProductResponse{
		Product: &productpb.Product{
			Id:       oid.Hex(),
			DevelopedBy: blog.GetDevelopedBy(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
			Budget: blog.GetBudget(),
			
		},
	}, nil

}





func (*server) ReadBlog(ctx context.Context, req *productpb.ReadProductRequest) (*productpb.ReadProductResponse, error) {
	fmt.Println("Read blog request")

	blogID := req.GetProductId()
	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}

	// create an empty struct
	data := &productItem{}
	filter := bson.M{"_id": oid}

	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID: %v", err),
		)
	}

	return &productpb.ReadProductResponse{
		Product: dataToBlogPb(data),
	}, nil
}

func dataToBlogPb(data *productItem) *productpb.Product {
		return &productpb.Product{
			Id:       data.ID.Hex(),
			DevelopedBy: data.DevelopedBy,
			Content:  data.Content,
			Title:    data.Title,
			Budget: data.Budget,
		}
	}


	func (*server) UpdateBlog(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.UpdateProductResponse, error) {
		fmt.Println("Update blog request")
		blog := req.GetProduct()
		oid, err := primitive.ObjectIDFromHex(blog.GetId())
		if err != nil {
			return nil, status.Errorf(
				codes.InvalidArgument,
				fmt.Sprintf("Cannot parse ID"),
			)
		}
	
		// create an empty struct
		data := &productItem{}
		filter := bson.M{"_id": oid}
	
		res := collection.FindOne(context.Background(), filter)
		if err := res.Decode(data); err != nil {
			return nil, status.Errorf(
				codes.NotFound,
				fmt.Sprintf("Cannot find blog with specified ID: %v", err),
			)
		}
	
		// we update our internal struct
		data.DevelopedBy = blog.GetDevelopedBy()
		data.Content = blog.GetContent()
		data.Title = blog.GetTitle()
		data.Budget = blog.GetBudget()
	
		_, updateErr := collection.ReplaceOne(context.Background(), filter, data)
		if updateErr != nil {
			return nil, status.Errorf(
				codes.Internal,
				fmt.Sprintf("Cannot update object in MongoDB: %v", updateErr),
			)
		}
	
		return &productpb.UpdateProductResponse{
			Product: dataToBlogPb(data),
		}, nil
	
	}
	
	func (*server) DeleteBlog(ctx context.Context, req *productpb.DeleteProductRequest) (*productpb.DeleteProductResponse, error) {
		fmt.Println("Delete blog request")
		oid, err := primitive.ObjectIDFromHex(req.GetProductId())
		if err != nil {
			return nil, status.Errorf(
				codes.InvalidArgument,
				fmt.Sprintf("Cannot parse ID"),
			)
		}
	
		filter := bson.M{"_id": oid}
	
		res, err := collection.DeleteOne(context.Background(), filter)
	
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				fmt.Sprintf("Cannot delete object in MongoDB: %v", err),
			)
		}
	
		if res.DeletedCount == 0 {
			return nil, status.Errorf(
				codes.NotFound,
				fmt.Sprintf("Cannot find blog in MongoDB: %v", err),
			)
		}
	
		return &productpb.DeleteProductResponse{ProductId: req.GetProductId()}, nil
	}
	
	func (*server) ListBlog(req *productpb.ListProductRequest, stream productpb.ProductService_ListBlogServer) error {
		fmt.Println("List blog request")
	
		cur, err := collection.Find(context.Background(), primitive.D{{}})
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Unknown internal error: %v", err),
			)
		}
		defer cur.Close(context.Background())
		for cur.Next(context.Background()) {
			data := &productItem{}
			err := cur.Decode(data)
			if err != nil {
				return status.Errorf(
					codes.Internal,
					fmt.Sprintf("Error while decoding data from MongoDB: %v", err),
				)
	
			}
			stream.Send(&productpb.ListProductResponse{Product: dataToBlogPb(data)})
		}
		if err := cur.Err(); err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Unknown internal error: %v", err),
			)
		}
		return nil
	}



var collection *mongo.Collection

type productItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	DevelopedBy string          `bson:"developed_by"`
	Title string                `bson:"title"`
	Content  string             `bson:"content"`
	Budget    string            `bson:"budget"`
}

func main(){
   
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Connecting to MongoDB")
	// connect to MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Product Service Started")
	collection = client.Database("mydb").Collection("product")

	// Firstly we need to write a listener

	lis, err := net.Listen("tcp","0.0.0.0:50051")

	if err!= nil{
		log.Fatal("Failed to listen: %v", err)
	}

	// creating a new server
	
	s:= grpc.NewServer()

	// register services on a server 

	productpb.RegisterProductServiceServer(s, &server{}) 

	//Serve accepts incoming connections on the listener lis, creating a new ServerTransport and service goroutine for each.
	//The service goroutines read gRPC requests and then call the registered handlers to reply to them


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

	// First we close the connection with MongoDB:
	fmt.Println("Closing MongoDB Connection")
    	// client.Disconnect(context.TODO())	
	if err := client.Disconnect(context.TODO()); err != nil {
        	log.Fatalf("Error on disconnection with MongoDB : %v", err)
    	}
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