package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"log"
	"net"
	"strings"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
)

// empty struct because this isn't a fancy example
type AuthorizationServer struct{}

// inject a header that can be used for future rate limiting
func (a *AuthorizationServer) Check(ctx context.Context, req *auth.CheckRequest) (*auth.CheckResponse, error) {
	authHeader, ok := req.Attributes.Request.Http.Headers["authorization"]
	var splitToken []string
	if ok {
		splitToken = strings.Split(authHeader, "Bearer ")
	}
	log.Printf("checking bearer token")
	if len(splitToken) == 2 {
		token := splitToken[1]
		sha := sha256.New()
		sha.Write([]byte(token))
		tokenSha := base64.StdEncoding.EncodeToString(sha.Sum(nil))

		// valid tokens have exactly 3 characters. #secure.
		// Normally this is where you'd go check with the system that knows if it's a valid token.

		if len(token) == 3 {
			log.Printf("bearer token is good")
			return &auth.CheckResponse{
				Status: &status.Status{
					Code: int32(code.Code_OK),
				},
				HttpResponse: &auth.CheckResponse_OkResponse{
					OkResponse: &auth.OkHttpResponse{
						Headers: []*core.HeaderValueOption{
							{
								Header: &core.HeaderValue{
									Key:   "x-ext-auth-ratelimit",
									Value: tokenSha,
								},
							},
						},
					},
				},
			}, nil
		}
	}
	log.Printf("bearer token is not good")
	return &auth.CheckResponse{
		Status: &status.Status{
			Code: int32(code.Code_UNAUTHENTICATED),
		},
		HttpResponse: &auth.CheckResponse_DeniedResponse{
			DeniedResponse: &auth.DeniedHttpResponse{
				Status: &envoy_type.HttpStatus{
					Code: envoy_type.StatusCode_Unauthorized,
				},
				Body: "Need an Authorization Header with a 3 character bearer token! #secure",
			},
		},
	}, nil
}

func main() {
	// create a TCP listener on port 4000
	lis, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("listening on %s", lis.Addr())

	grpcServer := grpc.NewServer()
	authServer := &AuthorizationServer{}
	auth.RegisterAuthorizationServer(grpcServer, authServer)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
