package modules

import (
	"context"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/auth"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/chat"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/documents"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/health"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/modules/users"
	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/rag-engine/ai"
	"github.com/go-chi/chi/v5"
)

func Register(r chi.Router, app *app.Application) {

	userModule := users.New(app)
	getUser := func(ctx context.Context, id string) (any, error) {
		return userModule.Service.GetUser(ctx, id)
	}

	aiModule := ai.New(app)
	healthModule := health.New(app)
	documentsModule := documents.New(app, getUser)
	authModule := auth.New(app, userModule.Service,getUser)
	chatModule := chat.New(app, aiModule.Service, userModule.Service, getUser)

	modules := []Module{
		healthModule,
		userModule,
		chatModule,
		documentsModule,
		authModule,
	}

	for _, m := range modules {
		m.RegisterRoutes(r)
	}
}
