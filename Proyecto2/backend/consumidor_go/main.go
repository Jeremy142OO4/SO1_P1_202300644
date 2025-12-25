package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type Venta struct {
	Categoria       string  `json:"categoria"`
	ProductoID      string  `json:"producto_id"`
	Precio          float64 `json:"precio"`
	CantidadVendida int     `json:"cantidad_vendida"`
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func main() {
	// Kafka
	broker := getenv("KAFKA_BROKER", "kafka-service:29092")
	topic := getenv("KAFKA_TOPIC", "ventas")
	groupID := getenv("KAFKA_GROUP_ID", "grupo-consumidor-ventas")

	// Valkey
	valkeyAddr := getenv("VALKEY_ADDR", "valkey-service:6379")

	log.Printf("Kafka broker=%s topic=%s group=%s", broker, topic, groupID)
	log.Printf("Valkey addr=%s", valkeyAddr)

	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: valkeyAddr,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("No conecta a Valkey: %v", err)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{broker},
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1e3,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})
	defer reader.Close()

	log.Println("Consumidor listo. Esperando mensajes...")

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error leyendo Kafka: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		var v Venta
		if err := json.Unmarshal(msg.Value, &v); err != nil {
			log.Printf("JSON inválido: %v | raw=%s", err, string(msg.Value))
			continue
		}

		// Normaliza categoría (por si viene Electronica/Ropa/Hogar/Belleza)
		catKey := "categoria:" + v.Categoria + ":reportes"

		pipe := rdb.Pipeline()
		pipe.Incr(ctx, "total_reportes")
		pipe.Incr(ctx, catKey)
		_, err = pipe.Exec(ctx)
		if err != nil {
			log.Printf("Error escribiendo en Valkey: %v", err)
			continue
		}

		log.Printf("OK venta categoria=%s producto=%s cantidad=%d precio=%.2f", v.Categoria, v.ProductoID, v.CantidadVendida, v.Precio)
	}
}
