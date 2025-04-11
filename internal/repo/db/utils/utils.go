package utils

import (
	"fmt"
	"strings"
)

func BuildFilterQuery(filters map[string]any) (string, []any) {
	idx := 1
	q := &strings.Builder{}
	args := make([]any, 0, len(filters))
	conds := make([]string, 0, len(filters))

	for key, value := range filters {
		switch key {
		case "is_active":
			conds = append(conds, "u.is_active = true")
		case "is_email_verified":
			conds = append(conds, "u.is_email_verified = true")
		case "is_wa":
			conds = append(conds, "u.is_wa = true")
		case "search":
			search := fmt.Sprintf("%%%s%%", value.(string))
			conds = append(conds, "(u.name ILIKE $1 OR u.email ILIKE $1)")
			args = append(args, search)
			idx++
		case "roles":
			if roles, ok := value.([]string); ok && len(roles) > 0 {
				placeholders := make([]string, len(roles))

				for i, role := range roles {
					args = append(args, role)
					placeholders[i] = fmt.Sprintf("$%d", idx+i)
				}
				conds = append(conds, fmt.Sprintf("r.name IN (%s)", strings.Join(placeholders, ",")))
				idx += len(roles)
			}
		}
	}

	if len(conds) > 0 {
		q.WriteString(" WHERE ")
		q.WriteString(strings.Join(conds, " AND "))
	}

	return q.String(), args
}

func GetSort(sort any) string {
	order := "DESC"
	if sort == nil {
		return "u.created_at DESC"
	}

	sortMap := map[string]string{
		"name":       "u.name",
		"email":      "u.email",
		"created_at": "u.created_at",
	}

	sortStr := sort.(string)
	if sortStr[0] == '-' {
		order = "DESC"
		sortStr = sortStr[1:]
	}

	if field, ok := sortMap[sortStr]; ok {
		return fmt.Sprintf("%s %s", field, order)
	}
	return "u.created_at DESC"
}
