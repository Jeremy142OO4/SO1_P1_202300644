use actix_web::{post, web, App, HttpResponse, HttpServer, Responder};
use serde::{Deserialize, Serialize};

#[derive(Debug, Deserialize, Serialize)]
struct VentaIn {
    categoria: String,
    producto: String,
    precio: f64,
    cantidad_vendida: i32,
}

#[derive(Debug, Serialize)]
struct ApiResp {
    ok: bool,
    mensaje: String,
}

#[post("/venta")]
async fn venta(
    body: web::Json<VentaIn>,
    client: web::Data<reqwest::Client>,
    go_url: web::Data<String>,
) -> impl Responder {

    let url = format!("{}/venta", go_url.get_ref());

    let resp = client
        .post(url)
        .json(&*body)
        .send()
        .await;

    match resp {
        Ok(r) => {
            let status = r.status();
            let text = r.text().await.unwrap_or_default();
            if status.is_success() {
                HttpResponse::Ok().json(ApiResp { ok: true, mensaje: text })
            } else {
                HttpResponse::BadRequest().json(ApiResp { ok: false, mensaje: text })
            }
        }
        Err(e) => HttpResponse::InternalServerError().json(ApiResp {
            ok: false,
            mensaje: format!("error enviando a go-deploy1: {e}"),
        }),
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let bind = std::env::var("BIND").unwrap_or_else(|_| "0.0.0.0:8080".to_string());
    let go_url = std::env::var("GO_API_URL").unwrap_or_else(|_| "http://go-deploy1-svc:8080".to_string());

    let client = reqwest::Client::new();

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(client.clone()))
            .app_data(web::Data::new(go_url.clone()))
            .service(venta)
    })
    .bind(bind)?
    .run()
    .await
}
