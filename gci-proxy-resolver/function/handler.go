package function

import (
	"fmt"
	"github.com/gci-proxy-resolver/model"
)

// Handle a serverless request
func Handle(req model.Request) (model.Response, error) {
	resp := []byte(fmt.Sprintf("Hello, Go. You said: %s", string(req.Body)))
	return model.Response{Body: resp}, nil
}
