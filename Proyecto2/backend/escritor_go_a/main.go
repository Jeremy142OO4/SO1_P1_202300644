package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "escritor_go_a/gen/escritor_go_a/pb"
)

type server struct {
	pb.UnimplementedProductSaleServiceServer
	writer *kafka.Writer
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func categoriaToStr(c pb.Categoria) string {
	switch c {
	case pb.Categoria_ELECTRONICA:
		return "ELECTRONICA"
	case pb.Categoria_ROPA:
		return "ROPA"
	case pb.Categoria_HOGAR:
		return "HOGAR"
	case pb.Categoria_BELLEZA:
		return "BELLEZA"
	default:
		return "CATEGORIA_UNSPECIFIED"
	}
}
func (s *server) ProcesarVenta(ctx context.Context, req *pb.ProductSaleRequest) (*pb.ProductSaleResponse, error) {
	payload := map[string]any{
		"categoria":        categoriaToStr(req.Categoria),
		"producto_id":      req.ProductoId,
		"precio":           req.Precio,
		"cantidad_vendida": req.CantidadVendida,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return &pb.ProductSaleResponse{Ok: false, Mensaje: "error json"}, nil
	}

	err = s.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(req.ProductoId),
		Value: b,
		Time:  time.Now(),
	})
	if err != nil {
		log.Printf("Kafka write error: %v", err)
		return &pb.ProductSaleResponse{Ok: false, Mensaje: "error kafka"}, nil
	}

	log.Printf("Writer A -> Kafka: %s", string(b))
	return &pb.ProductSaleResponse{Ok: true, Mensaje: "enviado a kafka"}, nil
}

func main() {
	broker := getenv("KAFKA_BROKER", "kafka-service:29092")
	topic := getenv("KAFKA_TOPIC", "ventas")
	port := getenv("GRPC_PORT", "50051")

	w := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 50 * time.Millisecond,
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterProductSaleServiceServer(s, &server{writer: w})
	reflection.Register(s)

	log.Printf("Writer A gRPC :%s -> Kafka %s topic=%s", port, broker, topic)
	log.Fatal(s.Serve(lis))
}
