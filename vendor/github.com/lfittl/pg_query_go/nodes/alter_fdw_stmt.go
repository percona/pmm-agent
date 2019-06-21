// Auto-generated from postgres/src/include/nodes/parsenodes.h - DO NOT EDIT

package pg_query

import "encoding/json"

/* ----------------------
 *		Create/Alter FOREIGN DATA WRAPPER Statements
 * ----------------------
 */
type AlterFdwStmt struct {
	Fdwname     *string `json:"fdwname"`      /* foreign-data wrapper name */
	FuncOptions List    `json:"func_options"` /* HANDLER/VALIDATOR options */
	Options     List    `json:"options"`      /* generic options to FDW */
}

func (node AlterFdwStmt) MarshalJSON() ([]byte, error) {
	type AlterFdwStmtMarshalAlias AlterFdwStmt
	return json.Marshal(map[string]interface{}{
		"AlterFdwStmt": (*AlterFdwStmtMarshalAlias)(&node),
	})
}

func (node *AlterFdwStmt) UnmarshalJSON(input []byte) (err error) {
	var fields map[string]json.RawMessage

	err = json.Unmarshal(input, &fields)
	if err != nil {
		return
	}

	if fields["fdwname"] != nil {
		err = json.Unmarshal(fields["fdwname"], &node.Fdwname)
		if err != nil {
			return
		}
	}

	if fields["func_options"] != nil {
		node.FuncOptions.Items, err = UnmarshalNodeArrayJSON(fields["func_options"])
		if err != nil {
			return
		}
	}

	if fields["options"] != nil {
		node.Options.Items, err = UnmarshalNodeArrayJSON(fields["options"])
		if err != nil {
			return
		}
	}

	return
}
