package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// DocsHandler gerencia requisições de documentação
type DocsHandler struct{}

// NewDocsHandler cria um novo handler de documentação
func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

// ServeHTML gerencia GET /docs - serve a página HTML da documentação
func (h *DocsHandler) ServeHTML(c echo.Context) error {
	html := `<!doctype html>
<html>
  <head>
    <title>Control API Docs</title>
    <link rel="icon" type="image/png" href="https://storage.deliveryvip.com.br/v-EIxwJDl5033-PzQEHUKDdFsInfAbijuPJHj5l9P0c/s:512:512/Z3M6Ly9kZWxpdmVy/eXZpcC9qdHp5dmtx/d3R0dXdtZzY3bXZu/Mzl6MTB3dmRm" />
    <meta charset="utf-8" />
    <meta
      name="viewport"
      content="width=device-width, initial-scale=1" />
  </head>

  <body>
    <div id="app"></div>

    <!-- Load the Script -->
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>

    <!-- Initialize the Scalar API Reference -->
    <script>
      Scalar.createApiReference('#app', {
        // The URL of the OpenAPI/Swagger document
        url: 'docs/openapi.yml',
      })
    </script>
  </body>
</html>`

	return c.HTML(http.StatusOK, html)
}

// ServeOpenAPI gerencia GET /docs/openapi.yml
func (h *DocsHandler) ServeOpenAPI(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/x-yaml")
	return c.File("docs/openapi.yml")
}
