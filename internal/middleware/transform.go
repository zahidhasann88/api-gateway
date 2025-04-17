package middleware

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zahidhasann88/api-gateway/internal/config"
	"github.com/zahidhasann88/api-gateway/pkg/logger"
)

// TransformationMiddleware applies request/response transformations
func TransformationMiddleware(cfg *config.Config, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract service name from path
		path := c.Request.URL.Path
		serviceName := "unknown"

		// Assuming format: /api/{service}/...
		if strings.HasPrefix(path, "/api/") {
			parts := strings.Split(path[5:], "/")
			if len(parts) > 0 {
				serviceName = parts[0]
			}
		}

		// Get service config
		serviceConfig, exists := cfg.Services[serviceName]
		if !exists || serviceConfig.Transformations == nil {
			c.Next()
			return
		}

		// Apply request transformations if enabled
		if serviceConfig.Transformations.Request != nil {
			applyRequestTransformations(c, serviceConfig.Transformations.Request, log)
		}

		// Create a response buffer
		responseBuffer := &responseBuffer{
			ResponseWriter: c.Writer,
			buffer:         new(bytes.Buffer),
		}
		c.Writer = responseBuffer

		// Process the request
		c.Next()

		// Apply response transformations if enabled
		if serviceConfig.Transformations.Response != nil {
			applyResponseTransformations(c, responseBuffer, serviceConfig.Transformations.Response, log)
		} else {
			// Write the response as-is if no transformations
			responseBuffer.ResponseWriter.Write(responseBuffer.buffer.Bytes())
		}
	}
}

// responseBuffer captures the response body
type responseBuffer struct {
	gin.ResponseWriter
	buffer *bytes.Buffer
	status int
}

func (r *responseBuffer) Write(b []byte) (int, error) {
	return r.buffer.Write(b)
}

func (r *responseBuffer) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// applyRequestTransformations modifies the incoming request
func applyRequestTransformations(c *gin.Context, config *config.TransformConfig, log logger.Logger) {
	// Only transform JSON bodies
	contentType := c.GetHeader("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return
	}

	// Read the request body
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body", "error", err)
		return
	}

	// Parse the JSON
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Error("Failed to parse request JSON", "error", err)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return
	}

	// Apply field mappings
	for fromField, toField := range config.FieldMapping {
		if value, exists := data[fromField]; exists {
			data[toField] = value
			if fromField != toField {
				delete(data, fromField)
			}
		}
	}

	// Add header fields to body
	for headerName, fieldName := range config.HeaderToBody {
		if headerValue := c.GetHeader(headerName); headerValue != "" {
			data[fieldName] = headerValue
		}
	}

	// Convert back to JSON
	newBody, err := json.Marshal(data)
	if err != nil {
		log.Error("Failed to marshal transformed request", "error", err)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return
	}

	// Replace the request body
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(newBody))
	c.Request.ContentLength = int64(len(newBody))
}

// applyResponseTransformations modifies the outgoing response
func applyResponseTransformations(c *gin.Context, responseBuffer *responseBuffer, config *config.TransformConfig, log logger.Logger) {
	// Check if we have a JSON response
	body := responseBuffer.buffer.Bytes()

	// Try to parse as JSON
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		// Not JSON or invalid, write as-is
		responseBuffer.ResponseWriter.Write(body)
		return
	}

	// Apply field mappings
	for fromField, toField := range config.FieldMapping {
		if value, exists := data[fromField]; exists {
			data[toField] = value
			if fromField != toField {
				delete(data, fromField)
			}
		}
	}

	// Add fields to headers
	for fieldName, headerName := range config.BodyToHeader {
		if value, exists := data[fieldName]; exists {
			strValue := ""
			switch v := value.(type) {
			case string:
				strValue = v
			default:
				if jsonValue, err := json.Marshal(v); err == nil {
					strValue = string(jsonValue)
				}
			}

			if strValue != "" {
				responseBuffer.ResponseWriter.Header().Set(headerName, strValue)
			}
		}
	}

	// Convert back to JSON
	newBody, err := json.Marshal(data)
	if err != nil {
		log.Error("Failed to marshal transformed response", "error", err)
		responseBuffer.ResponseWriter.Write(body)
		return
	}

	// Write the transformed response
	responseBuffer.ResponseWriter.Write(newBody)
}
