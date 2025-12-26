package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"

	pb "go_deploy1/gen/ventas/pb"
)

type VentaIn struct {
	Categoria       string  `json:"categoria"`
	Producto        string  `json:"producto"`
	Precio          float64 `json:"precio"`
	CantidadVendida int32   `json:"cantidad_vendida"`
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func catToEnum(cat string) pb.Categoria {
	c := strings.ToUpper(strings.TrimSpace(cat))
	switch c {
	case "ELECTRONICA":
		return pb.Categoria_ELECTRONICA
	case "ROPA":
		return pb.Categoria_ROPA
	case "HOGAR":
		return pb.Categoria_HOGAR
	case "BELLEZA":
		return pb.Categoria_BELLEZA
	default:
		return pb.Categoria_CATEGORIA_UNSPECIFIED
	}
}

type server struct {
	writerA pb.ProductSaleServiceClient
	writerB pb.ProductSaleServiceClient
}

func (s *server) pickWriter() pb.ProductSaleServiceClient {
	// balanceo simple (random)
	if rand.Intn(2) == 0 {
		return s.writerA
	}
	return s.writerB
}

func (s *server) handleVenta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var in VentaIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("json invalido"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	req := &pb.ProductSaleRequest{
		Categoria:       catToEnum(in.Categoria),
		ProductoId:      in.Producto,
		Precio:          in.Precio,
		CantidadVendida: in.CantidadVendida,
	}

	resp, err := s.pickWriter().ProcesarVenta(ctx, req)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("error grpc: " + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	httpAddr := getenv("HTTP_ADDR", ":8080")
	writerAAddr := getenv("WRITER_A_ADDR", "writer-go-a-svc:50051")
	writerBAddr := getenv("WRITER_B_ADDR", "writer-go-b-svc:50051")

	connA, err := grpc.Dial(writerAAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer connA.Close()

	connB, err := grpc.Dial(writerBAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer connB.Close()

	s := &server{
		writerA: pb.NewProductSaleServiceClient(connA),
		writerB: pb.NewProductSaleServiceClient(connB),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/venta", s.handleVenta)

	log.Printf("go-deploy1 REST %s -> grpc A=%s B=%s", httpAddr, writerAAddr, writerBAddr)
	log.Fatal(http.ListenAndServe(httpAddr, mux))
}
