// Package orm provides the generic ORM layer for Kora.
// All documents are represented as doctype.Document (map[string]any).
package orm

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/yourorg/kora/doctype"
)

// generateName creates a unique document name based on the DocType.
// Format: {PREFIX}-{NNNN} where PREFIX is derived from the DocType name.
func generateName(dt *doctype.DocType, existingCount int) string {
	prefix := derivePrefix(dt.Name)
	return fmt.Sprintf("%s-%04d", prefix, existingCount+1)
}

func derivePrefix(name string) string {
	// For multi-word names, take the first letter of each word.
	// For single-word names, take the first 4 letters.
	// Examples: "Customer" → "CUST", "Work Order" → "WO", "Work Order Item" → "WOI"
	parts := strings.Fields(name)
	if len(parts) == 1 {
		s := strings.ToUpper(name)
		if len(s) > 4 {
			s = s[:4]
		}
		return s
	}
	var result strings.Builder
	for _, p := range parts {
		if len(p) > 0 {
			c := p[0]
			if c >= 'a' && c <= 'z' {
				c = c - 32
			}
			result.WriteByte(c)
		}
	}
	return result.String()
}

// TxManager provides transactional operations on documents.
type TxManager struct {
	DB       *sql.DB
	Registry *doctype.Registry
}

// Insert creates a new document in the database.
func (tx *TxManager) Insert(dt *doctype.DocType, doc *doctype.Document, owner string) error {
	if !doc.IsNew {
		return fmt.Errorf("cannot insert an existing document")
	}

	// Check unique constraints before inserting.
	if err := tx.checkUniqueConstraints(dt, doc, ""); err != nil {
		return err
	}

	if doc.Name == "" {
		var count int
		err := tx.DB.QueryRow(
			fmt.Sprintf("SELECT COUNT(*) FROM %s", dt.TableName()),
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("counting existing rows: %w", err)
		}
		doc.Name = generateName(dt, count)
	}

	now := time.Now()
	doc.DocStatus = 0

	dataFields := dt.DataFields()
	var columns []string
	var placeholders []string
	var values []any

	columns = append(columns, "name", "owner", "creation", "modified", "modified_by", "doc_status", "idx")
	placeholders = append(placeholders, "?", "?", "?", "?", "?", "?", "?")
	values = append(values, doc.Name, owner, now, now, owner, doc.DocStatus, 0)

	for _, f := range dataFields {
		if f.Fieldtype == "Table" {
			continue
		}
		columns = append(columns, f.Fieldname)
		placeholders = append(placeholders, "?")

		val := doc.Get(f.Fieldname)
		if val == nil && f.Default != "" {
			val = convertDefault(f.Default, f.Fieldtype)
		}
		values = append(values, val)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		dt.TableName(),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := tx.DB.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("inserting document: %w", err)
	}

	for _, f := range dt.TableFields() {
		children := doc.GetTable(f.Fieldname)
		if children == nil {
			continue
		}
		childDT := tx.Registry.Get(f.Options)
		if childDT == nil {
			return fmt.Errorf("child doctype %q not found", f.Options)
		}

		for i, child := range children {
			child.DocType = childDT.Name
			if err := tx.insertChild(dt, f.Fieldname, childDT, child, doc.Name, i); err != nil {
				return fmt.Errorf("inserting child row %d in %s: %w", i, f.Fieldname, err)
			}
		}
	}

	doc.IsNew = false
	return nil
}

func (tx *TxManager) insertChild(parentDT *doctype.DocType, parentField string, childDT *doctype.DocType, doc *doctype.Document, parentName string, idx int) error {
	if doc.Name == "" {
		// Generate unique child row name including parent reference.
		var count int
		prefix := derivePrefix(childDT.Name)
		tx.DB.QueryRow(
			fmt.Sprintf("SELECT COUNT(*) FROM %s", parentDT.ChildTableName(parentField)),
		).Scan(&count)
		doc.Name = fmt.Sprintf("%s-%s-%04d", prefix, parentName, count+1)
	}

	now := time.Now()

	var columns []string
	var placeholders []string
	var values []any

	columns = append(columns, "name", "owner", "creation", "modified", "modified_by", "doc_status", "idx")
	placeholders = append(placeholders, "?", "?", "?", "?", "?", "?", "?")
	values = append(values, doc.Name, "", now, now, "", 0, idx)

	columns = append(columns, "parent", "parentfield", "parenttype")
	placeholders = append(placeholders, "?", "?", "?")
	values = append(values, parentName, parentField, parentDT.Name)

	for _, f := range childDT.DataFields() {
		if f.Fieldtype == "Table" {
			continue
		}
		columns = append(columns, f.Fieldname)
		placeholders = append(placeholders, "?")
		val := doc.Get(f.Fieldname)
		if val == nil && f.Default != "" {
			val = convertDefault(f.Default, f.Fieldtype)
		}
		values = append(values, val)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		parentDT.ChildTableName(parentField),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := tx.DB.Exec(query, values...)
	return err
}

// Save updates an existing document.
// If owner is non-empty, only updates if the document is owned by that user.
func (tx *TxManager) Save(dt *doctype.DocType, doc *doctype.Document, modifiedBy string, owner string) error {
	if doc.IsNew {
		return fmt.Errorf("cannot save a new document; use Insert instead")
	}

	// Check unique constraints before updating.
	if err := tx.checkUniqueConstraints(dt, doc, doc.Name); err != nil {
		return err
	}

	now := time.Now()
	dataFields := dt.DataFields()

	var setClauses []string
	var values []any

	for _, f := range dataFields {
		if f.Fieldtype == "Table" {
			continue
		}
		if f.ReadOnly {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", f.Fieldname))
		values = append(values, doc.Get(f.Fieldname))
	}

	setClauses = append(setClauses, "modified = ?", "modified_by = ?")
	values = append(values, now, modifiedBy)

	where := "name = ?"
	values = append(values, doc.Name)
	if owner != "" {
		where += " AND owner = ?"
		values = append(values, owner)
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		dt.TableName(),
		strings.Join(setClauses, ", "),
		where,
	)

	result, err := tx.DB.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("updating document: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("document %q not found or access denied", doc.Name)
	}

	for _, f := range dt.TableFields() {
		childTableName := dt.ChildTableName(f.Fieldname)
		if _, err := tx.DB.Exec(
			fmt.Sprintf("DELETE FROM %s WHERE parent = ?", childTableName),
			doc.Name,
		); err != nil {
			return fmt.Errorf("deleting old child rows for %s: %w", f.Fieldname, err)
		}

		children := doc.GetTable(f.Fieldname)
		if children == nil {
			continue
		}
		childDT := tx.Registry.Get(f.Options)
		if childDT == nil {
			return fmt.Errorf("child doctype %q not found", f.Options)
		}

		for i, child := range children {
			child.DocType = childDT.Name
			if err := tx.insertChild(dt, f.Fieldname, childDT, child, doc.Name, i); err != nil {
				return fmt.Errorf("inserting child row %d in %s: %w", i, f.Fieldname, err)
			}
		}
	}

	return nil
}

// GetDoc loads a single document by name, including child table expansion.
// If owner is non-empty, only returns the document if owned by that user.
func (tx *TxManager) GetDoc(dt *doctype.DocType, name string, owner string) (*doctype.Document, error) {
	dataFields := dt.DataFields()

	var cols []string
	for _, f := range dataFields {
		if f.Fieldtype == "Table" {
			continue
		}
		cols = append(cols, f.Fieldname)
	}
	cols = append(cols, "name", "owner", "creation", "modified", "modified_by", "doc_status")

	scanTargets := make([]any, len(cols))
	for i := range cols {
		var v any
		scanTargets[i] = &v
	}

	where := "name = ?"
	args := []any{name}
	if owner != "" {
		where += " AND owner = ?"
		args = append(args, owner)
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s",
		strings.Join(cols, ", "),
		dt.TableName(),
		where,
	)

	row := tx.DB.QueryRow(query, args...)
	if err := row.Scan(scanTargets...); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document %q not found in %s", name, dt.Name)
		}
		return nil, fmt.Errorf("scanning document: %w", err)
	}

	doc := doctype.NewDocument(dt.Name)
	doc.Name = name
	doc.IsNew = false

	for i, col := range cols {
		val := *(scanTargets[i].(*any))
		switch col {
		case "name":
			doc.Name = stringVal(val)
		case "doc_status":
			doc.DocStatus = intVal(val)
		case "owner", "creation", "modified", "modified_by":
			// System columns.
		default:
			doc.Fields[col] = byteSliceToString(val)
		}
	}

	for _, f := range dt.TableFields() {
		childDT := tx.Registry.Get(f.Options)
		if childDT == nil {
			continue
		}
		children, err := tx.getChildRows(dt.ChildTableName(f.Fieldname), childDT)
		if err != nil {
			return nil, fmt.Errorf("loading child table %s: %w", f.Fieldname, err)
		}
		doc.Fields[f.Fieldname] = children
	}

	return doc, nil
}

func (tx *TxManager) getChildRows(tableName string, childDT *doctype.DocType) ([]*doctype.Document, error) {
	dataFields := childDT.DataFields()

	var cols []string
	for _, f := range dataFields {
		if f.Fieldtype == "Table" {
			continue
		}
		cols = append(cols, f.Fieldname)
	}
	cols = append(cols, "name", "idx", "parent", "parentfield", "parenttype")

	rows, err := tx.DB.Query(
		fmt.Sprintf("SELECT %s FROM %s ORDER BY idx", strings.Join(cols, ", "), tableName),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var children []*doctype.Document
	for rows.Next() {
		scanTargets := make([]any, len(cols))
		for i := range cols {
			var v any
			scanTargets[i] = &v
		}

		if err := rows.Scan(scanTargets...); err != nil {
			return nil, err
		}

		child := doctype.NewDocument(childDT.Name)
		child.IsNew = false

		for i, col := range cols {
			val := *(scanTargets[i].(*any))
			switch col {
			case "name":
				child.Name = stringVal(val)
			default:
				child.Fields[col] = byteSliceToString(val)
			}
		}

		children = append(children, child)
	}

	return children, rows.Err()
}

// GetList returns a paginated list of documents with optional filtering.
// If owner is non-empty, only returns documents owned by that user.
func (tx *TxManager) GetList(dt *doctype.DocType, filters string, orderBy string, limit, offset int, owner string) ([]*doctype.Document, int, error) {
	where := "1=1"
	var whereArgs []any
	if filters != "" && filters != "[]" {
		fs, err := ParseFilters(filters)
		if err != nil {
			return nil, 0, fmt.Errorf("parsing filters: %w", err)
		}
		where, whereArgs, err = fs.ToSQL()
		if err != nil {
			return nil, 0, fmt.Errorf("building filter SQL: %w", err)
		}
	}
	if owner != "" {
		where += " AND owner = ?"
		whereArgs = append(whereArgs, owner)
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", dt.TableName(), where)
	err := tx.DB.QueryRow(countQuery, whereArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting documents: %w", err)
	}

	dataFields := dt.DataFields()
	var cols []string
	for _, f := range dataFields {
		if f.Fieldtype == "Table" {
			continue
		}
		cols = append(cols, f.Fieldname)
	}
	cols = append(cols, "name", "owner", "creation", "modified", "modified_by", "doc_status")

	if orderBy == "" {
		orderBy = dt.SortField + " " + dt.SortOrder
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s ORDER BY %s LIMIT ? OFFSET ?",
		strings.Join(cols, ", "),
		dt.TableName(),
		where,
		orderBy,
	)

	args := append(whereArgs, limit, offset)

	rows, err := tx.DB.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("querying documents: %w", err)
	}
	defer rows.Close()

	docs := make([]*doctype.Document, 0)
	for rows.Next() {
		scanTargets := make([]any, len(cols))
		for i := range cols {
			var v any
			scanTargets[i] = &v
		}

		if err := rows.Scan(scanTargets...); err != nil {
			return nil, 0, fmt.Errorf("scanning row: %w", err)
		}

		doc := doctype.NewDocument(dt.Name)
		doc.IsNew = false

		for i, col := range cols {
			val := *(scanTargets[i].(*any))
			switch col {
			case "name":
				doc.Name = stringVal(val)
			case "doc_status":
				doc.DocStatus = intVal(val)
			default:
				doc.Fields[col] = byteSliceToString(val)
			}
		}

		docs = append(docs, doc)
	}

	return docs, total, rows.Err()
}

// Delete removes a document by name.
// If owner is non-empty, only deletes if the document is owned by that user.
func (tx *TxManager) Delete(dt *doctype.DocType, name string, owner string) error {
	for _, f := range dt.TableFields() {
		childTable := dt.ChildTableName(f.Fieldname)
		if _, err := tx.DB.Exec(
			fmt.Sprintf("DELETE FROM %s WHERE parent = ?", childTable),
			name,
		); err != nil {
			return fmt.Errorf("deleting child rows for %s: %w", f.Fieldname, err)
		}
	}

	where := "name = ?"
	args := []any{name}
	if owner != "" {
		where += " AND owner = ?"
		args = append(args, owner)
	}

	result, err := tx.DB.Exec(
		fmt.Sprintf("DELETE FROM %s WHERE %s", dt.TableName(), where),
		args...,
	)
	if err != nil {
		return fmt.Errorf("deleting document: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("document %q not found or access denied", name)
	}

	return nil
}

// checkUniqueConstraints verifies that unique fields don't conflict with existing documents.
// Returns a doctype.ValidationError if a duplicate is found, so the frontend can show a field-level error.
func (tx *TxManager) checkUniqueConstraints(dt *doctype.DocType, doc *doctype.Document, excludeName string) error {
	for _, f := range dt.Fields {
		if !f.Unique {
			continue
		}
		val := doc.Get(f.Fieldname)
		if val == nil || val == "" {
			continue // nil/empty values don't trigger unique checks.
		}

		// SELECT name FROM tabX WHERE fieldname = ? AND name != ?
		query := fmt.Sprintf("SELECT name FROM %s WHERE %s = ?", dt.TableName(), f.Fieldname)
		args := []any{val}
		if excludeName != "" {
			query += " AND name != ?"
			args = append(args, excludeName)
		}

		var existingName string
		err := tx.DB.QueryRow(query, args...).Scan(&existingName)
		if err == nil {
			return &doctype.ValidationError{
				Type:    "UniqueConstraint",
				Message: fmt.Sprintf("%s must be unique. Value %q already exists in %s.", f.Label, val, existingName),
				Field:   f.Fieldname,
				DocType: dt.Name,
			}
		}
		// sql.ErrNoRows is expected — no duplicate found.
	}
	return nil
}

// --- Helpers ---

func stringVal(v any) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	}
	return fmt.Sprintf("%v", v)
}

// byteSliceToString converts []byte to string for JSON-safe storage.
// The MySQL driver returns []byte for VARCHAR/TEXT columns. JSON marshaling
// encodes []byte as base64, so we must convert to string first.
func byteSliceToString(v any) any {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}

func intVal(v any) int {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return int(n)
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

func convertDefault(def string, fieldtype string) any {
	switch fieldtype {
	case "Int":
		var n int64
		fmt.Sscanf(def, "%d", &n)
		return n
	case "Float", "Currency", "Percent":
		var f float64
		fmt.Sscanf(def, "%f", &f)
		return f
	case "Check":
		return def == "1" || def == "true"
	default:
		return def
	}
}
