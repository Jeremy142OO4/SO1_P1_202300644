IMAGES=("img_bajo" "img_cpu" "img_ram")

echo "Generando 10 contenedores aleatorios..."

for i in {1..10}
do
    # Elegir imagen aleatoria
    RAND_INDEX=$((RANDOM % 3))
    IMAGE=${IMAGES[$RAND_INDEX]}

    # Nombre único para el contenedor
    CONTAINER_NAME="contenedor_${i}_${RANDOM}"

    docker run -d --name "$CONTAINER_NAME" "$IMAGE"

    echo "[$i] Contenedor creado: $CONTAINER_NAME usando $IMAGE"
done

echo "Proceso finalizado."
