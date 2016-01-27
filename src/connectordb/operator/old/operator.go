/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package operator

import "connectordb/operator/interfaces"

// Operator defines extension functions which work with any BaseOperator, adding
// the ability to query things by path without requiring permission specialization
type Operator interfaces.Operator
