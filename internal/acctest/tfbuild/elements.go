package tfbuild

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type FileElement func(*hclwrite.File)

func File(children ...FileElement) *hclwrite.File {
	hclFile := hclwrite.NewEmptyFile()

	for _, child := range children {
		child(hclFile)
	}

	return hclFile
}

type BlockElement func(*hclwrite.Block)

func Block(typeName string, labels []string, children ...BlockElement) FileElement {
	return func(file *hclwrite.File) {
		block := file.Body().AppendNewBlock(typeName, labels)

		for _, child := range children {
			child(block)
		}
	}
}

func Resource(resourceType string, resourceName string, children ...BlockElement) FileElement {
	return Block("resource", []string{resourceType, resourceName}, children...)
}

func Data(resourceType string, resourceName string, children ...BlockElement) FileElement {
	return Block("data", []string{resourceType, resourceName}, children...)
}

func Provider(providerType string, children ...BlockElement) FileElement {
	return Block("provider", []string{providerType}, children...)
}

func InnerBlock(typeName string, children ...BlockElement) BlockElement {
	return func(block *hclwrite.Block) {
		innerBlock := block.Body().AppendNewBlock(typeName, []string{})

		for _, child := range children {
			child(innerBlock)
		}
	}
}

func DependsOn(deps ...hcl.Traversal) BlockElement {
	var tokens hclwrite.Tokens

	tokens = append(tokens,
		&hclwrite.Token{
			Type:         hclsyntax.TokenOBrack,
			Bytes:        []byte{'['},
			SpacesBefore: 0,
		},
	)

	depsLen := len(deps)
	for i, dep := range deps {
		depExpr := hclwrite.NewExpressionAbsTraversal(dep)
		tokens = depExpr.BuildTokens(tokens)
		if i < depsLen-1 {
			tokens = append(tokens, &hclwrite.Token{
				Type:         hclsyntax.TokenComma,
				Bytes:        []byte{','},
				SpacesBefore: 0,
			})
		}
	}

	tokens = append(tokens,
		&hclwrite.Token{
			Type:         hclsyntax.TokenCBrack,
			Bytes:        []byte{']'},
			SpacesBefore: 0,
		},
	)

	return Attribute("depends_on", hclwrite.NewExpressionRaw(tokens))
}

func Attribute(name string, expr *hclwrite.Expression) BlockElement {
	return func(block *hclwrite.Block) {
		var tokens hclwrite.Tokens
		tokens = expr.BuildTokens(tokens)
		block.Body().SetAttributeRaw(name, tokens)
	}
}

func AttributeInt(name string, val int64) BlockElement {
	return func(block *hclwrite.Block) {
		block.Body().SetAttributeValue(name, cty.NumberIntVal(val))
	}
}

func AttributeBool(name string, val bool) BlockElement {
	return func(block *hclwrite.Block) {
		block.Body().SetAttributeValue(name, cty.BoolVal(val))
	}
}

func AttributeString(name string, val string) BlockElement {
	return func(block *hclwrite.Block) {
		block.Body().SetAttributeValue(name, cty.StringVal(val))
	}
}

func AttributeTraversal(name string, ref hcl.Traversal) BlockElement {
	return func(block *hclwrite.Block) {
		block.Body().SetAttributeTraversal(name, ref)
	}
}

func TraversalResource(resourceType string, resourceName string) hcl.Traversal {
	return hcl.Traversal{
		hcl.TraverseRoot{
			Name: resourceType,
		},
		hcl.TraverseAttr{
			Name: resourceName,
		},
	}
}

func TraversalResourceAttribute(resourceType string, resourceName string, attr string) hcl.Traversal {
	return hcl.Traversal{
		hcl.TraverseRoot{
			Name: resourceType,
		},
		hcl.TraverseAttr{
			Name: resourceName,
		},
		hcl.TraverseAttr{
			Name: attr,
		},
	}
}

func String(val string) *hclwrite.Expression {
	return hclwrite.NewExpressionLiteral(cty.StringVal(val))
}

func StringList(vals ...string) *hclwrite.Expression {
	var ctyVals []cty.Value
	for _, val := range vals {
		ctyVals = append(ctyVals, cty.StringVal(val))
	}
	return hclwrite.NewExpressionLiteral(cty.ListVal(ctyVals))
}

func List(elements ...*hclwrite.Expression) *hclwrite.Expression {
	var tokens hclwrite.Tokens

	tokens = append(tokens, &hclwrite.Token{
		Type:         hclsyntax.TokenOBrack,
		Bytes:        []byte{'['},
		SpacesBefore: 0,
	})

	for _, e := range elements {
		tokens = e.BuildTokens(tokens)
	}

	tokens = append(tokens, &hclwrite.Token{
		Type:         hclsyntax.TokenCBrack,
		Bytes:        []byte{']'},
		SpacesBefore: 0,
	})

	return hclwrite.NewExpressionRaw(tokens)
}

func Identifier(name string) *hclwrite.Expression {
	return hclwrite.NewExpressionRaw(hclwrite.Tokens{
		{
			Type:         hclsyntax.TokenIdent,
			Bytes:        []byte(name),
			SpacesBefore: 0,
		},
	})
}

func FunctionCall(name string, attrExpressions ...*hclwrite.Expression) *hclwrite.Expression {
	var tokens hclwrite.Tokens

	tokens = append(tokens,
		&hclwrite.Token{
			Type:         hclsyntax.TokenIdent,
			Bytes:        []byte(name),
			SpacesBefore: 0,
		},
		&hclwrite.Token{
			Type:         hclsyntax.TokenOParen,
			Bytes:        []byte{'('},
			SpacesBefore: 0,
		},
	)

	attrExpressionsLen := len(attrExpressions)
	for i, attrExpression := range attrExpressions {
		tokens = attrExpression.BuildTokens(tokens)
		if i < attrExpressionsLen-1 {
			tokens = append(tokens, &hclwrite.Token{
				Type:         hclsyntax.TokenComma,
				Bytes:        []byte{','},
				SpacesBefore: 0,
			})
		}
	}

	tokens = append(tokens,
		&hclwrite.Token{
			Type:         hclsyntax.TokenCParen,
			Bytes:        []byte{')'},
			SpacesBefore: 0,
		},
	)

	return hclwrite.NewExpressionRaw(tokens)
}
