package server

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func HandleStaticResources(c *gin.Context) {
	originalURL := "https://pm.pstatic.net/resources" + strings.Replace(c.Request.URL.Path, "/pmpstaticresources", "", 1)

	fmt.Println("Fetching static resource:", originalURL)

	resp, err := http.Get(originalURL)
	if err != nil {
		log.Printf("Failed to fetch static resource: %v", err)
		c.AbortWithStatus(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read static resource: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Data(http.StatusOK, resp.Header.Get("Content-Type"), bodyBytes)
}

func ReverseProxyHandler(target string) gin.HandlerFunc {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}

	return func(c *gin.Context) {
		fmt.Println("Request URL:", c.Request.URL.Path)
		if strings.HasPrefix(c.Request.URL.Path, "/pmpstaticresources") {
			fmt.Println("Handling static resources")
			HandleStaticResources(c)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.Host = targetURL.Host
		}

		proxy.ModifyResponse = func(resp *http.Response) error {
			if resp.StatusCode == http.StatusMovedPermanently {
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

			modifiedBodyString := strings.ReplaceAll(bodyString, "https://pm.pstatic.net/resources", "http://localhost:8080/pmpstaticresources")
			modifiedBodyString = strings.ReplaceAll(modifiedBodyString, "NAVER", "NotAVER")

			modifiedBodyBytes := []byte(modifiedBodyString)
			resp.Body = io.NopCloser(bytes.NewBuffer(modifiedBodyBytes))
			resp.ContentLength = int64(len(modifiedBodyBytes))
			resp.Header.Set("Content-Length", string(len(modifiedBodyBytes)))

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
	r := gin.Default()

	r.GET("/*path", ReverseProxyHandler("https://www.naver.com"))

	return r
}
