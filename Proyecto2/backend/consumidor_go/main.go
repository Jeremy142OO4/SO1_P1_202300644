package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type Venta struct {
	Categoria       string  `json:"categoria"`
	ProductoID      string  `json:"producto_id"`
	Producto        string  `json:"producto"`
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
	broker := getenv("KAFKA_BROKER", "kafka-service:29092")
	topic := getenv("KAFKA_TOPIC", "ventas")
	groupID := getenv("KAFKA_GROUP_ID", "grupo-consumidor-ventas")
	valkeyAddr := getenv("VALKEY_ADDR", "valkey-primary:6379")

	log.Printf("Kafka broker=%s topic=%s group=%s", broker, topic, groupID)
	log.Printf("Valkey addr=%s", valkeyAddr)

	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{Addr: valkeyAddr})
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

	_ = rdb.SetNX(ctx, "max_precio", "0", 0).Err()
	_ = rdb.SetNX(ctx, "min_precio", "999999999", 0).Err()

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

		producto := strings.TrimSpace(v.ProductoID)
		if producto == "" {
			producto = strings.TrimSpace(v.Producto)
		}
		if producto == "" {
			producto = "UNKNOWN"
		}

		cat := strings.ToUpper(strings.TrimSpace(v.Categoria))
		if cat == "" {
			cat = "CATEGORIA_UNSPECIFIED"
		}

		now := time.Now().Unix()

		pipe := rdb.Pipeline()

		// conteos
		pipe.Incr(ctx, "total_reportes")
		pipe.HIncrBy(ctx, "reportes_por_categoria", cat, 1)

		pipe.ZIncrBy(ctx, "z_ventas_por_producto", float64(v.CantidadVendida), producto)
		pipe.ZIncrBy(ctx, "z_ventas_por_producto:"+cat, float64(v.CantidadVendida), producto)

		// sumas para promedio por categoria
		pipe.HIncrByFloat(ctx, "sum_precio_por_categoria", cat, v.Precio)

		pipe.ZAdd(ctx, "ts_precio:"+cat+":"+producto, redis.Z{
			Score:  float64(now),
			Member: fmt.Sprintf("%.2f", v.Precio),
		})

		_, err = pipe.Exec(ctx)
		if err != nil {
			log.Printf("Error escribiendo en Valkey: %v", err)
			continue
		}

		// Recalcular promedio
		sumStr, _ := rdb.HGet(ctx, "sum_precio_por_categoria", cat).Result()
		cntStr, _ := rdb.HGet(ctx, "reportes_por_categoria", cat).Result()

		sumV, _ := strconv.ParseFloat(sumStr, 64)
		cntV, _ := strconv.ParseFloat(cntStr, 64)
		if cntV > 0 {
			avg := sumV / cntV
			_ = rdb.HSet(ctx, "avg_precio_por_categoria", cat, fmt.Sprintf("%.2f", avg)).Err()
		}

		// max/min
		curMaxStr, _ := rdb.Get(ctx, "max_precio").Result()
		curMax, _ := strconv.ParseFloat(curMaxStr, 64)
		if v.Precio >= curMax {
			_ = rdb.Set(ctx, "max_precio", fmt.Sprintf("%.2f", v.Precio), 0).Err()
			_ = rdb.Set(ctx, "max_precio_producto", producto, 0).Err()
		}

		curMinStr, _ := rdb.Get(ctx, "min_precio").Result()
		curMin, _ := strconv.ParseFloat(curMinStr, 64)
		if v.Precio <= curMin {
			_ = rdb.Set(ctx, "min_precio", fmt.Sprintf("%.2f", v.Precio), 0).Err()
			_ = rdb.Set(ctx, "min_precio_producto", producto, 0).Err()
		}

		log.Printf("OK venta categoria=%s producto=%s cantidad=%d precio=%.2f", cat, producto, v.CantidadVendida, v.Precio)
	}
}
