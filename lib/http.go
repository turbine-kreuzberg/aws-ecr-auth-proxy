package lib

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

func fetchDockerHubPrefix(ctx context.Context, svc *ecr.Client) (string, error) {
	result, err := svc.DescribePullThroughCacheRules(ctx, &ecr.DescribePullThroughCacheRulesInput{})
	if err != nil {
		return "", fmt.Errorf("failed to describe pull-through cache rules: %v", err)
	}

	for _, rule := range result.PullThroughCacheRules {
		if *rule.UpstreamRegistryUrl == "registry-1.docker.io" {
			return *rule.EcrRepositoryPrefix, nil
		}
	}

	return "", fmt.Errorf("no Docker Hub pull-through cache rule found")
}

func RunHttpServer(ctx context.Context, port int) error {
	// Create an ECR client
	svc, err := ecrClient()
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %v", err)
	}

	// Fetch Docker Hub prefix
	dockerHubPrefix, err := fetchDockerHubPrefix(ctx, svc)
	if err != nil {
		return fmt.Errorf("failed to fetch Docker Hub prefix: %v", err)
	}

	// Setup HTTP server
	http.HandleFunc("/", handler(ctx, svc, dockerHubPrefix))

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	log.Printf("Starting server on %s...", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		return fmt.Errorf("could not start server: %v", err)
	}

	return nil
}

// getECRAuthToken retrieves a fresh token for ECR pull-through cache authorization.
func getECRAuthToken(ctx context.Context, svc *ecr.Client) (string, string, error) {
	input := &ecr.GetAuthorizationTokenInput{}

	result, err := svc.GetAuthorizationToken(ctx, input)
	if err != nil {
		return "", "", fmt.Errorf("failed to get ECR authorization token: %w", err)
	}

	if len(result.AuthorizationData) == 0 {
		return "", "", fmt.Errorf("no authorization data found")
	}

	token := *result.AuthorizationData[0].AuthorizationToken
	proxyEndpoint := *result.AuthorizationData[0].ProxyEndpoint
	return token, proxyEndpoint, nil
}

// newProxy creates a new ReverseProxy that forwards requests to the ECR domain.
func newProxy(ecrURL *url.URL, authToken string, dockerHubPrefix string) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(ecrURL)
	originalDirector := proxy.Director

	proxy.Director = func(req *http.Request) {
		log.Printf("Original request path: %s", req.URL.Path)
		if strings.HasPrefix(req.URL.Path, "/v2/") {
			req.URL.Path = "/" + dockerHubPrefix + req.URL.Path
		}
		log.Printf("Modified request path: %s", req.URL.Path)

		originalDirector(req)
		req.Host = ecrURL.Host
		req.Header.Set("Authorization", "Basic "+authToken)
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		log.Printf("response for %s is %d - %s", resp.Request.URL, resp.StatusCode, resp.Status)

		return nil
	}

	return proxy
}

// handler forwards incoming requests to the ECR endpoint with proper authorization headers.
func handler(ctx context.Context, svc *ecr.Client, dockerHubPrefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Retrieve a fresh token and the ECR domain.
		authToken, proxyEndpoint, err := getECRAuthToken(ctx, svc)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get ECR token: %v", err), http.StatusInternalServerError)
			return
		}

		// Parse the ECR domain URL.
		ecrURL, err := url.Parse(proxyEndpoint)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to parse ECR domain: %v", err), http.StatusInternalServerError)
			return
		}

		// Create a new ReverseProxy and forward the request.
		proxy := newProxy(ecrURL, authToken, dockerHubPrefix)
		proxy.ServeHTTP(w, r)
	}
}
