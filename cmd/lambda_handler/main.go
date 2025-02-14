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
	// 🔥 API Gateway のリクエストを詳細にログ出力
	logRequestDetails(req)

	// Gorilla Mux ルーターを作成
	r := mux.NewRouter().StrictSlash(true)

	// ルートハンドラーを登録
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{"message": "Hello from /test"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	// `req.Path` を `/test` のみに変換
	reqPath := normalizePath(req.Path)

	// リクエストを Mux で処理
	body := ioutil.NopCloser(strings.NewReader(req.Body))
	httpReq, err := http.NewRequest(req.HTTPMethod, reqPath, body)
	if err != nil {
		log.Println("Error creating request:", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Internal Server Error"}, nil
	}

	// ヘッダーをコピー
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Mux に渡すカスタムレスポンスライター
	rw := &ResponseWriter{Headers: map[string]string{}, StatusCode: 404} // 🔥 デフォルト `404`
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

// normalizePath - API Gateway の `/{proxy}` 形式を `/test` に変換
func normalizePath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		return "/" + strings.Join(parts[1:], "/") // `/{proxy}` → `/test`
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

// 🔥 API Gateway のリクエスト内容をログ出力
func logRequestDetails(req events.APIGatewayProxyRequest) {
	// ログ用データ構造体
	logData := map[string]interface{}{
		"HTTPMethod":     req.HTTPMethod,
		"Path":           req.Path,
		"Headers":        req.Headers,
		"QueryParams":    req.QueryStringParameters,
		"PathParams":     req.PathParameters,
		"RequestContext": req.RequestContext,
		"Body":           req.Body,
	}

	// JSON に変換して出力
	jsonData, err := json.MarshalIndent(logData, "", "  ")
	if err != nil {
		log.Println("Error marshaling request data:", err)
	} else {
		log.Println("🔥 Received API Gateway Event:\n", string(jsonData))
	}
}

func main() {
	// 環境変数で Lambda 環境かどうかチェック
	if _, isLambda := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); isLambda {
		lambda.Start(APIGatewayLambdaHandler)
	} else {
		// ローカル開発用サーバー
		r := mux.NewRouter().StrictSlash(true)
		r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			response := map[string]string{"message": "Hello from /test"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}).Methods("GET")

		log.Println("Starting local server on :8080")
		log.Fatal(http.ListenAndServe(":8080", r))
	}
}
