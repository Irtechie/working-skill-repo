package graphrouting

// ProviderRequest is the bounded input for optional graph/evidence providers.
// Providers may enrich the packet, but they do not own the packet contract.
type ProviderRequest struct {
	PacketID   string
	Repository Repository
	Seed       Symbol
	MaxEdges   int
	MaxBytes   int
}

type ProviderLookup struct {
	Packet        Packet
	Provider      string
	Available     bool
	Authoritative bool
	Issues        []string
}

type ExactSymbolProvider interface {
	Name() string
	ResolveExactSymbols(request ProviderRequest) (ProviderLookup, error)
}

func normalizeProviderRequest(request ProviderRequest) ProviderRequest {
	if request.PacketID == "" {
		request.PacketID = "exact-symbol-provider"
	}
	if request.MaxEdges <= 0 {
		request.MaxEdges = 100
	}
	if request.MaxBytes <= 0 {
		request.MaxBytes = 64000
	}
	return request
}

func fallbackProviderLookup(request ProviderRequest, provider string, reason string, issues ...string) ProviderLookup {
	request = normalizeProviderRequest(request)
	packet := Packet{
		SchemaVersion: 1,
		PacketID:      request.PacketID,
		Repository:    request.Repository,
		Seeds: Seeds{
			Symbols: []Symbol{request.Seed},
		},
		Fallback:    Fallback{Mode: "file-native", Reason: reason},
		Budget:      Budget{MaxEdges: request.MaxEdges, MaxBytes: request.MaxBytes},
		Limitations: []string{reason},
		GeneratedBy: provider,
	}
	return ProviderLookup{
		Packet:        packet,
		Provider:      provider,
		Available:     false,
		Authoritative: false,
		Issues:        issues,
	}
}
