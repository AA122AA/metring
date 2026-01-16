package handler

import (
	"context"
	"net/http"

	"github.com/AA122AA/metring/internal/server/database"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type PingHandler struct {
	db *database.Database
	lg *zap.Logger
}

func NewPingHandler(ctx context.Context, db *database.Database) *PingHandler {
	return &PingHandler{
		db: db,
		lg: zctx.From(ctx).Named("Ping handler"),
	}
}

func (p *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if p.db == nil {
		http.Error(w, "Can not ping Database", http.StatusInternalServerError)
		p.lg.Info("Database is not connected")
		return
	}

	err := p.db.Ping(r.Context())
	if err != nil {
		http.Error(w, "Can not ping Database", http.StatusInternalServerError)
		p.lg.Error("Can not ping Database", zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusOK)
}
