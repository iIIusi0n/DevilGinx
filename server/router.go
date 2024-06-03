package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func ReverseProxyHandler(target string) gin.HandlerFunc {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}

	return func(c *gin.Context) {
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.Host = targetURL.Host

			if req.Header.Get("Origin") != "" {
				req.Header.Set("Origin", targetURL.String())
			}

			if req.Header.Get("Referer") != "" {
				parsed, _ := url.Parse(req.Header.Get("Referer"))
				parsed.Scheme = targetURL.Scheme
				parsed.Host = targetURL.Host
				req.Header.Set("Referer", parsed.String())
			}
		}

		proxy.ModifyResponse = func(resp *http.Response) error {
			if resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusFound {
				parsed, _ := url.Parse(resp.Header.Get("Location"))
				parsed.Scheme = "https"
				parsed.Host = "localhost:8443"
				resp.Header.Set("Location", parsed.String())
				log.Println("Redirecting to:", resp.Header.Get("Location"))
			}

			if resp.StatusCode != http.StatusOK {
				return nil
			}

			var bodyBytes []byte
			var err error
			contentEncoding := resp.Header.Get("Content-Encoding")

			if contentEncoding == "gzip" {
				gzReader, err := gzip.NewReader(resp.Body)
				if err != nil {
					return err
				}
				bodyBytes, err = io.ReadAll(gzReader)
				if err != nil {
					return err
				}
				gzReader.Close()
				resp.Body.Close()
			} else {
				bodyBytes, err = io.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				resp.Body.Close()
			}

			bodyString := string(bodyBytes)

			modifiedBodyString := strings.ReplaceAll(bodyString, "Sign in", "Do not sign in")

			modifiedBodyBytes := []byte(modifiedBodyString)

			if contentEncoding == "gzip" {
				var buf bytes.Buffer
				gzWriter := gzip.NewWriter(&buf)
				gzWriter.Write(modifiedBodyBytes)
				gzWriter.Close()
				modifiedBodyBytes = buf.Bytes()
			}

			resp.Body = io.NopCloser(bytes.NewBuffer(modifiedBodyBytes))
			resp.ContentLength = int64(len(modifiedBodyBytes))
			resp.Header.Set("Content-Length", strconv.Itoa(len(modifiedBodyBytes)))

			return nil
		}

		proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
			http.Error(rw, "The reverse proxy encountered an error", http.StatusBadGateway)
			log.Printf("Reverse proxy error: %v", err)
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func GetRouter() *gin.Engine {
	targetURL := "https://github.com"

	r := gin.Default()

	r.GET("/*path", ReverseProxyHandler(targetURL))
	r.POST("/*path", ReverseProxyHandler(targetURL))

	return r
}
