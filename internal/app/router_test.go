package app

import (
	"net/http"
	"testing"

	aihttp "github.com/ksamwang/PersonalContentPlatform/internal/ai/transport"
	assethttp "github.com/ksamwang/PersonalContentPlatform/internal/asset/transport"
	collectionhttp "github.com/ksamwang/PersonalContentPlatform/internal/collection/transport"
	contenthttp "github.com/ksamwang/PersonalContentPlatform/internal/content/transport"
	identityhttp "github.com/ksamwang/PersonalContentPlatform/internal/identity/transport"
	inboxhttp "github.com/ksamwang/PersonalContentPlatform/internal/inbox/transport"
	integrationhttp "github.com/ksamwang/PersonalContentPlatform/internal/integration/transport"
	knowledgehttp "github.com/ksamwang/PersonalContentPlatform/internal/knowledge/transport"
	publicationhttp "github.com/ksamwang/PersonalContentPlatform/internal/publication/transport"
	retrievalhttp "github.com/ksamwang/PersonalContentPlatform/internal/retrieval/transport"
	settingshttp "github.com/ksamwang/PersonalContentPlatform/internal/settings/transport"
)

func TestRouterRegistersAllModuleRoutes(t *testing.T) {
	passthrough := func(next http.Handler) http.Handler { return next }
	application := &App{
		identity:    identityhttp.NewHTTP(nil, nil, "test_session", 0, false),
		content:     contenthttp.NewHTTP(nil, passthrough),
		public:      publicationhttp.NewHTTP(nil),
		asset:       assethttp.NewHTTP(nil, passthrough),
		knowledge:   knowledgehttp.NewHTTP(nil, passthrough),
		ai:          aihttp.NewHTTP(nil, passthrough),
		integration: integrationhttp.NewHTTP(nil, nil, passthrough),
		settings:    settingshttp.NewHTTP(nil, passthrough),
		inbox:       inboxhttp.NewHTTP(nil, passthrough),
		collection:  collectionhttp.NewHTTP(nil, passthrough),
		retrieval:   retrievalhttp.NewHTTP(nil, passthrough),
	}
	if router := application.Router(); router == nil {
		t.Fatal("expected router")
	}
}
