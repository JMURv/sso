package utils

import (
	"fmt"
	"strings"
)

func BuildFilterQuery(filters map[string]any) (string, []any) {
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
		case "roles":
			if roles, ok := value.([]string); ok {
				placeholders := make([]string, len(roles))

				for i, role := range roles {
					args = append(args, role)
					placeholders[i] = fmt.Sprintf("$%d", len(args))
				}
				conds = append(
					conds, fmt.Sprintf(
						`
                    (SELECT COUNT(DISTINCT r.name) 
                     FROM user_roles ur 
                     JOIN roles r ON ur.role_id = r.id 
                     WHERE ur.user_id = u.id 
                       AND r.name IN (%s)
                    ) = %d`,
						strings.Join(placeholders, ", "),
						len(roles),
					),
				)
			}
		}
	}

	if len(conds) > 0 {
		q.WriteString(" WHERE ")
		q.WriteString(strings.Join(conds, " AND "))
	}

	return q.String(), args
}
