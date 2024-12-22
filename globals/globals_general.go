package globals_general

type MensajeResultado struct {
	Status string
}

// Estructura para respuestas específicas de compactación
type CompactacionRespuesta struct {
	Status      string `json:"status"`
	Compactable bool   `json:"compactable"`
}