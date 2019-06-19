// Auto-generated - DO NOT EDIT

package pg_query

import "strconv"

func (node CaseTestExpr) Fingerprint(ctx FingerprintContext, parentNode Node, parentFieldName string) {
	ctx.WriteString("CaseTestExpr")

	if node.Collation != 0 {
		ctx.WriteString("collation")
		ctx.WriteString(strconv.Itoa(int(node.Collation)))
	}

	if node.TypeId != 0 {
		ctx.WriteString("typeId")
		ctx.WriteString(strconv.Itoa(int(node.TypeId)))
	}

	if node.TypeMod != 0 {
		ctx.WriteString("typeMod")
		ctx.WriteString(strconv.Itoa(int(node.TypeMod)))
	}

	if node.Xpr != nil {
		subCtx := FingerprintSubContext{}
		node.Xpr.Fingerprint(&subCtx, node, "Xpr")

		if len(subCtx.parts) > 0 {
			ctx.WriteString("xpr")
			for _, part := range subCtx.parts {
				ctx.WriteString(part)
			}
		}
	}
}
