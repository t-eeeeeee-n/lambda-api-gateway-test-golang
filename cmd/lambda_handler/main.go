package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gorilla/mux"
)

// APIGatewayLambdaHandler Lambda に適用する HTTP ハンドラー
func APIGatewayLambdaHandler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Received request:", req)

	// Gorilla Mux ルーターを作成
	r := mux.NewRouter()

	// ルートハンドラーを登録
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{"message": "Hello from /test"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	// リクエストを Mux で処理
	reqPath := req.Path
	httpReq, _ := http.NewRequest(req.HTTPMethod, reqPath, nil)
	rw := &ResponseWriter{Headers: map[string]string{}, StatusCode: 200}
	r.ServeHTTP(rw, httpReq)

	// Lambda のレスポンスを構成
	response := events.APIGatewayProxyResponse{
		StatusCode: rw.StatusCode,
		Headers:    rw.Headers,
		Body:       rw.Body,
	}

	log.Println("Response:", response)
	return response, nil
}

// ResponseWriter カスタムレスポンスライター
type ResponseWriter struct {
	StatusCode int
	Headers    map[string]string
	Body       string
}

func (rw *ResponseWriter) Header() http.Header {
	return http.Header{}
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	rw.Body = string(b)
	return len(b), nil
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
}

func main() {
	// 環境変数で Lambda 環境かどうかチェック
	if _, isLambda := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); isLambda {
		lambda.Start(APIGatewayLambdaHandler)
	} else {
		// ローカル開発用サーバー
		r := mux.NewRouter()
		r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			response := map[string]string{"message": "Hello from /test"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}).Methods("GET")

		log.Println("Starting local server on :8080")
		log.Fatal(http.ListenAndServe(":8080", r))
	}
}
