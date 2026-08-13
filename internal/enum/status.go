package enum

// Status represents the processing lifecycle persisted with an order.
type Status string

const (
	// Criado marks an order that was accepted but not yet processed.
	Criado Status = "CRIADO"
	// Processando marks an order currently handled by the worker.
	Processando Status = "PROCESSANDO"
	// Processado marks an order whose asynchronous work completed.
	Processado Status = "PROCESSADO"
)
