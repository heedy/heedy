/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package transforms

import "github.com/connectordb/duck"

func pipelineGeneratorCompare(left, right TransformFunc, operator string) TransformFunc {
	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		leftResult := left(te.Copy())
		if !leftResult.CanProcess() {
			return leftResult
		}

		rightResult := right(te.Copy())
		if !rightResult.CanProcess() {
			return rightResult
		}

		var ok bool

		leftData := leftResult.Datapoint.Data
		rightData := rightResult.Datapoint.Data

		switch operator {
		case ">":
			te.Datapoint.Data, ok = duck.Gt(leftData, rightData)
		case ">=":
			te.Datapoint.Data, ok = duck.Gte(leftData, rightData)
		case "<":
			te.Datapoint.Data, ok = duck.Lt(leftData, rightData)
		case "<=":
			te.Datapoint.Data, ok = duck.Lte(leftData, rightData)
		case "!=":
			var eq bool
			eq, ok = duck.Eq(leftData, rightData)
			te.Datapoint.Data = !eq
		case "==":
			te.Datapoint.Data, ok = duck.Eq(leftData, rightData)
		default:
			return te.SetErrorString("comparison: incorrectly initialized! (internal error)")
		}

		if ok != true {
			return te.SetErrorString("comparison: invalid comparison types")
		}

		return te
	}
}
