package docs

import "testing"

func TestSwaggerInfoRegistration(t *testing.T) {
	t.Parallel()
	// Apenas verifica se SwaggerInfo está inicializado corretamente
	if SwaggerInfo == nil || SwaggerInfo.SwaggerTemplate == "" {
		t.Error("SwaggerInfo não inicializado corretamente")
	}
}
