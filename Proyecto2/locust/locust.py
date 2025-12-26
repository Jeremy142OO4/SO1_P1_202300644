from locust import HttpUser, task, between
import random, string

CATS = ["ELECTRONICA", "ROPA", "HOGAR", "BELLEZA"]

class ApiUser(HttpUser):
    wait_time = between(0.1, 0.5)

    @task
    def post_venta(self):
        payload = {
            "categoria": random.choice(CATS),
            "producto": "PRD-" + "".join(random.choices(string.ascii_uppercase, k=6)),
            "precio": round(random.uniform(1, 200), 2),
            "cantidad_vendida": random.randint(1, 10),
        }
        self.client.post("/venta", json=payload)
