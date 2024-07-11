package main

import (
	"fmt"
	"strings"

	"github.com/huandu/xstrings"
)

func sqlDataTypeToGoDataType(column Column) string {
	goDataType := ""
	switch column.DataType {
	case "uuid":
		goDataType = "string"
	case "time":
		goDataType = "time.Time"
	case "string":
		goDataType = "string"
	case "number":
		goDataType = "int"
		if column.DataTypeLen >= 0 {
			goDataType = fmt.Sprintf("int%d", column.DataTypeLen)
		}
	}

	if column.RefTable != nil {
		goDataType = column.RefTable.Model
	}

	nullIndicator := ""
	if column.IsNull {
		nullIndicator = "*"
	}

	return fmt.Sprintf("%s%s", nullIndicator, goDataType)
}

func sqlDataTypeToGoDataTypeForInput(column Column) string {
	goDataType := ""
	switch column.DataType {
	case "uuid":
		goDataType = "string"
	case "time":
		goDataType = "time.Time"
	case "string":
		goDataType = "string"
	case "number":
		goDataType = "int"
		if column.DataTypeLen >= 0 {
			goDataType = fmt.Sprintf("int%d", column.DataTypeLen)
		}
	}

	nullIndicator := ""
	if column.IsNull {
		nullIndicator = "*"
	}

	return fmt.Sprintf("%s%s", nullIndicator, goDataType)
}

func kebabCase(str string) string {
	return xstrings.ToKebabCase(str)
}

func createColumn(column Column) string {

	var dataType = ""
	switch column.DataType {
	case "uuid":
		dataType = "uuid"
	case "time":
		dataType = "TIMESTAMP WITH TIME ZONE"
	case "string":
		dataType = fmt.Sprintf("varchar(%d)", column.DataTypeLen)
	case "number":
		dataType = "integer"
	}

	var isNull = "NULL"
	if !column.IsNull {
		isNull = "NOT NULL"
	}

	var ref = ""
	if column.RefTable != nil {
		ref = fmt.Sprintf("REFERENCES %s(id)", column.RefTable.Name)
	}

	return fmt.Sprintf("\"%s\" %s %s %s", column.Name, dataType, isNull, ref)
}

func left(str string, len int) string {
	return str[0:len]
}

func notNull(n interface{}) bool {
	return n != nil
}

func isNull(n interface{}) bool {
	return n == nil
}

func sub(n1 int, n2 int) int {
	return n1 - n2
}

func camelCase(str string) string {
	return xstrings.ToCamelCase(str)
}

func pascalCase(str string) string {
	return xstrings.ToPascalCase(str)
}

func createColumnForStruct(column Column) string {
	goDataType := sqlDataTypeToGoDataType(column)
	return fmt.Sprintf("%s\t%s\t`json:\"%s\" db:\"%s\"`", xstrings.ToPascalCase(column.Name), goDataType, xstrings.ToCamelCase(column.Name), column.Name)
}

func createColumnForStructInputDTO(column Column) string {
	//Name              string                  `json:"name" validate:"required"`
	goDataType := sqlDataTypeToGoDataType(column)

	validate := ""
	if !column.IsNull {
		validate = " validate:\"required\""
	}

	return fmt.Sprintf("%s\t%s\t`json:\"%s\"%s`", xstrings.ToPascalCase(column.Name), goDataType, xstrings.ToCamelCase(column.Name), validate)
}

func generateSelectForRefColumn(columns []Column) string {
	//"uCreatedBy"."id" AS "createdBy.id",
	//"uCreatedBy"."email" AS "createdBy.email",

	selectColumns := []string{}
	for _, column := range columns {
		if column.RefTable != nil {

			for _, selCol := range column.RefTable.SelectColumns {
				sel := fmt.Sprintf("%s%s.\"%s\" AS \"%s.%s\"", column.RefTable.Name[0:1], xstrings.ToPascalCase(column.Name), selCol, xstrings.ToCamelCase(column.Name), selCol)
				selectColumns = append(selectColumns, sel)
			}
		}
	}
	selectColumnsWithComma := strings.Join(selectColumns, ",\n")
	return selectColumnsWithComma
}

func joinInSelect(cg CodeGen) string {
	/*
		LEFT JOIN "users" "uPerson" ON "uPerson"."person_id" = "p"."id"
		INNER JOIN "users" "uCreatedBy" ON "uCreatedBy"."id" = "p"."created_by"
		LEFT JOIN "users" "uUpdatedBy" ON "uUpdatedBy"."id" = "p"."updated_by"
	*/
	columns := cg.Columns

	joins := []string{}
	for _, column := range columns {
		if column.RefTable != nil {
			joinType := "INNER JOIN"
			if column.IsNull {
				joinType = "LEFT JOIN"
			}

			alias := fmt.Sprintf("%s%s", column.RefTable.Name[0:1], xstrings.ToPascalCase(column.Name))

			join := fmt.Sprintf("%s \"%s\" %s ON %s.id = %s.\"%s\"", joinType, column.RefTable.Name, alias, alias, cg.AliasTableNameInSelect, column.Name)

			joins = append(joins, join)
		}
	}

	return strings.Join(joins, "\n")
}
