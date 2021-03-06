package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

type assetImpl struct {
	schema.AssetAliases
}

// ID implements response to request for 'id' field.
func (*assetImpl) ID(p graphql.ResolveParams) (interface{}, error) {
	return globalid.AssetTranslator.EncodeToString(p.Source), nil
}

// Namespace implements response to request for 'namespace' field.
func (*assetImpl) Namespace(p graphql.ResolveParams) (interface{}, error) {
	return p.Source, nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*assetImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*types.Asset)
	return ok
}
