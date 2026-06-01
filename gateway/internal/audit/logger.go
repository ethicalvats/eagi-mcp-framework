package audit

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type AuditEntry struct {
	Timestamp string                 `json:"timestamp"`
	Domain    string                 `json:"domain"`
	Tool      string                 `json:"tool,omitempty"`
	UserID    string                 `json:"user_id"`
	Role      string                 `json:"role"`
	Status    int                    `json:"status"`
	Duration  int64                  `json:"duration_ms"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type Logger struct {
	logger *log.Logger
	mu     sync.Mutex
}

func NewLogger(outputFile string) (*Logger, error) {
	var out *os.File
	var err error

	if outputFile == "stdout" || outputFile == "" {
		out = os.Stdout
	} else {
		out, err = os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
	}

	return &Logger{
		logger: log.New(out, "", 0),
	}, nil
}

func (l *Logger) Log(entry AuditEntry) {
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("[Audit Error] Failed to marshal entry: %v", err)
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Println(string(data))
}
