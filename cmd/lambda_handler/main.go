package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gorilla/mux"
)

// APIGatewayLambdaHandler - API Gateway からのリクエストを処理
func APIGatewayLambdaHandler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// 🔥 API Gateway のリクエスト詳細をログに出力
	logRequestDetails(req)

	// Gorilla Mux ルーターを作成
	r := mux.NewRouter().StrictSlash(true)

	// ルートハンドラーを登録 (エンドポイントごとに処理を分ける)
	r.HandleFunc("/", rootHandler).Methods("GET")
	r.HandleFunc("/test", testHandler).Methods("GET")
	r.HandleFunc("/user", userHandler).Methods("GET", "POST")
	r.HandleFunc("/order", orderHandler).Methods("GET", "POST")

	// `req.Path` からパスを取得
	reqPath := normalizePath(req.Path)
	httpMethod := req.HTTPMethod

	// リクエストを Mux で処理
	body := ioutil.NopCloser(strings.NewReader(req.Body))
	httpReq, err := http.NewRequest(httpMethod, reqPath, body)
	if err != nil {
		log.Println("Error creating request:", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Internal Server Error"}, nil
	}

	// ヘッダーをコピー
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Mux に渡すカスタムレスポンスライター
	rw := &ResponseWriter{Headers: map[string]string{}, StatusCode: 404}
	r.ServeHTTP(rw, httpReq)

	// Lambda のレスポンスを構成
	response := events.APIGatewayProxyResponse{
		StatusCode: rw.StatusCode,
		Headers:    rw.Headers,
		Body:       rw.Body,
	}

	// 🔥 レスポンスの詳細をログに出力
	log.Printf("Response: %+v\n", response)

	return response, nil
}

// 🔹 ルート ("/") のハンドラー
func rootHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"message": "Welcome to the root endpoint"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 🔹 `/test` のハンドラー
func testHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"message": "Hello from /test"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 🔹 `/user` のハンドラー (GET & POST)
func userHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		response := map[string]string{"message": "User created"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		response := map[string]string{"message": "User endpoint"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// 🔹 `/order` のハンドラー (GET & POST)
func orderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		response := map[string]string{"message": "Order created"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		response := map[string]string{"message": "Order endpoint"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// 🔥 API Gateway のリクエスト詳細をログに出力
func logRequestDetails(req events.APIGatewayProxyRequest) {
	logData := map[string]interface{}{
		"HTTPMethod":  req.HTTPMethod,
		"Path":        req.Path,
		"Headers":     req.Headers,
		"QueryParams": req.QueryStringParameters,
		"PathParams":  req.PathParameters,
		"RequestID":   req.RequestContext.RequestID,
		"Stage":       req.RequestContext.Stage,
		"Domain":      req.RequestContext.DomainName,
		"Body":        req.Body,
	}
	jsonData, err := json.MarshalIndent(logData, "", "  ")
	if err != nil {
		log.Println("Error marshaling request data:", err)
	} else {
		log.Println("🔥 Received API Gateway Event:\n", string(jsonData))
	}
}

func normalizePath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return "/" + strings.Join(parts[2:], "/")
	}
	return path
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
	if _, isLambda := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); isLambda {
		lambda.Start(APIGatewayLambdaHandler)
	} else {
		r := mux.NewRouter().StrictSlash(true)
		r.HandleFunc("/", rootHandler).Methods("GET")
		r.HandleFunc("/test", testHandler).Methods("GET")
		r.HandleFunc("/user", userHandler).Methods("GET", "POST")
		r.HandleFunc("/order", orderHandler).Methods("GET", "POST")

		log.Println("Starting local server on :8080")
		log.Fatal(http.ListenAndServe(":8080", r))
	}
}
