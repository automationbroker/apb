package jsval_test

import "github.com/lestrrat/go-jsval"

var V0 *jsval.JSVal
var M *jsval.ConstraintMap
var R0 jsval.Constraint
var R1 jsval.Constraint
var R2 jsval.Constraint
var R3 jsval.Constraint
var R4 jsval.Constraint
var R5 jsval.Constraint

func init() {
	M = &jsval.ConstraintMap{}
	R0 = jsval.Object().
		AdditionalProperties(
			jsval.EmptyConstraint,
		).
		AddProp(
			"$schema",
			jsval.String().Format("uri"),
		).
		AddProp(
			"additionalItems",
			jsval.Any().
				Add(
					jsval.Boolean(),
				).
				Add(
					jsval.Reference(M).RefersTo("#"),
				),
		).
		AddProp(
			"additionalProperties",
			jsval.Any().
				Add(
					jsval.Boolean(),
				).
				Add(
					jsval.Reference(M).RefersTo("#"),
				),
		).
		AddProp(
			"allOf",
			jsval.Reference(M).RefersTo("#/definitions/schemaArray"),
		).
		AddProp(
			"anyOf",
			jsval.Reference(M).RefersTo("#/definitions/schemaArray"),
		).
		AddProp(
			"default",
			jsval.EmptyConstraint,
		).
		AddProp(
			"definitions",
			jsval.Object().
				AdditionalProperties(
					jsval.Reference(M).RefersTo("#"),
				),
		).
		AddProp(
			"dependencies",
			jsval.Object().
				AdditionalProperties(
					jsval.Any().
						Add(
							jsval.Reference(M).RefersTo("#"),
						).
						Add(
							jsval.Reference(M).RefersTo("#/definitions/stringArray"),
						),
				),
		).
		AddProp(
			"description",
			jsval.String(),
		).
		AddProp(
			"enum",
			jsval.Array().
				AdditionalItems(
					jsval.EmptyConstraint,
				).
				MinItems(1).
				UniqueItems(true),
		).
		AddProp(
			"exclusiveMaximum",
			jsval.Boolean().Default(false),
		).
		AddProp(
			"exclusiveMinimum",
			jsval.Boolean().Default(false),
		).
		AddProp(
			"id",
			jsval.String().Format("uri"),
		).
		AddProp(
			"items",
			jsval.Any().
				Add(
					jsval.Reference(M).RefersTo("#"),
				).
				Add(
					jsval.Reference(M).RefersTo("#/definitions/schemaArray"),
				),
		).
		AddProp(
			"maxItems",
			jsval.Reference(M).RefersTo("#/definitions/positiveInteger"),
		).
		AddProp(
			"maxLength",
			jsval.Reference(M).RefersTo("#/definitions/positiveInteger"),
		).
		AddProp(
			"maxProperties",
			jsval.Reference(M).RefersTo("#/definitions/positiveInteger"),
		).
		AddProp(
			"maximum",
			jsval.Number(),
		).
		AddProp(
			"minItems",
			jsval.Reference(M).RefersTo("#/definitions/positiveIntegerDefault0"),
		).
		AddProp(
			"minLength",
			jsval.Reference(M).RefersTo("#/definitions/positiveIntegerDefault0"),
		).
		AddProp(
			"minProperties",
			jsval.Reference(M).RefersTo("#/definitions/positiveIntegerDefault0"),
		).
		AddProp(
			"minimum",
			jsval.Number(),
		).
		AddProp(
			"multipleOf",
			jsval.Number().Minimum(0.000000).ExclusiveMinimum(true),
		).
		AddProp(
			"not",
			jsval.Reference(M).RefersTo("#"),
		).
		AddProp(
			"oneOf",
			jsval.Reference(M).RefersTo("#/definitions/schemaArray"),
		).
		AddProp(
			"pattern",
			jsval.String().Format("regex"),
		).
		AddProp(
			"patternProperties",
			jsval.Object().
				AdditionalProperties(
					jsval.Reference(M).RefersTo("#"),
				),
		).
		AddProp(
			"properties",
			jsval.Object().
				AdditionalProperties(
					jsval.Reference(M).RefersTo("#"),
				),
		).
		AddProp(
			"required",
			jsval.Reference(M).RefersTo("#/definitions/stringArray"),
		).
		AddProp(
			"title",
			jsval.String(),
		).
		AddProp(
			"type",
			jsval.Any().
				Add(
					jsval.Reference(M).RefersTo("#/definitions/simpleTypes"),
				).
				Add(
					jsval.Array().
						Items(
							jsval.Reference(M).RefersTo("#/definitions/simpleTypes"),
						).
						AdditionalItems(
							jsval.EmptyConstraint,
						).
						MinItems(1).
						UniqueItems(true),
				),
		).
		AddProp(
			"uniqueItems",
			jsval.Boolean().Default(false),
		).
		PropDependency("exclusiveMaximum", "maximum").
		PropDependency("exclusiveMinimum", "minimum")
	R1 = jsval.Integer().Minimum(0)
	R2 = jsval.All().
		Add(
			jsval.Reference(M).RefersTo("#/definitions/positiveInteger"),
		).
		Add(
			jsval.EmptyConstraint,
		)
	R3 = jsval.Array().
		Items(
			jsval.Reference(M).RefersTo("#"),
		).
		AdditionalItems(
			jsval.EmptyConstraint,
		).
		MinItems(1)
	R4 = jsval.String().Enum("array", "boolean", "integer", "null", "number", "object", "string")
	R5 = jsval.Array().
		Items(
			jsval.String(),
		).
		AdditionalItems(
			jsval.EmptyConstraint,
		).
		MinItems(1).
		UniqueItems(true)
	M.SetReference("#", R0)
	M.SetReference("#/definitions/positiveInteger", R1)
	M.SetReference("#/definitions/positiveIntegerDefault0", R2)
	M.SetReference("#/definitions/schemaArray", R3)
	M.SetReference("#/definitions/simpleTypes", R4)
	M.SetReference("#/definitions/stringArray", R5)
	V0 = jsval.New().
		SetName("V0").
		SetConstraintMap(M).
		SetRoot(R0)
}