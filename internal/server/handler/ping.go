package handler

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type PingHandler struct {
	db *sql.DB
	lg *zap.Logger
}

func NewPingHandler(ctx context.Context, db *sql.DB) *PingHandler {
	return &PingHandler{
		db: db,
		lg: zctx.From(ctx).Named("Ping handler"),
	}
}

func (p *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	err := p.db.Ping()
	if err != nil {
		http.Error(w, "Can not ping Database", http.StatusInternalServerError)
		p.lg.Error("Can not ping Database", zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusOK)
}
