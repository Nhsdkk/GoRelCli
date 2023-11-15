package validator

import (
	"GoRelCli/models/error_model/validation_error"
	"GoRelCli/models/schema_model"
	"GoRelCli/utils/schema_parser"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// CharactersEnd ASCII character index for end of letters
const CharactersEnd = 127

// CharactersStart ASCII character index for start of letters
const CharactersStart = 33

// CapitalEnd ASCII character index for end of capital letters
const CapitalEnd = 122

// CapitalStart ASCII character index for start of capital letters
const CapitalStart = 97

// LowercaseEnd ASCII character index for end of lowercase letters
const LowercaseEnd = 90

// LowercaseStart ASCII character index for start of lowercase letters
const LowercaseStart = 65

// UNDERSCORE ASCII character index for underscore character
const UNDERSCORE = 95

// MINUS ASCII character index for minus character
const MINUS = 45

// QuestionMark ASCII character index for questionMark
const QuestionMark = 63

// SquareBracketOpening ASCII character index for opening square bracket
const SquareBracketOpening = 91

// SquareBracketClosing ASCII character index for closing square bracket
const SquareBracketClosing = 93

func checkNameForSpecialCharacter(name string) bool {
	for _, charIndex := range name {
		if charIndex == UNDERSCORE || charIndex == MINUS {
			continue
		}

		if (charIndex >= CapitalStart && charIndex <= CapitalEnd) || (charIndex >= LowercaseStart && charIndex <= LowercaseEnd) {
			continue
		}

		return true
	}
	return false
}

//func replaceSpecialCharacters(str string) string {
//	str = strings.Replace(str, "\"", "\\\"", -1)
//	str = strings.Replace(str, "'", "\\'", -1)
//	return str
//}

func cleanupString(name string, mapper func(r rune) bool) string {
	mapperFunction := func(r rune) rune {
		if mapper(r) {
			return r
		}
		return -1
	}

	return strings.Map(mapperFunction, name)

}

func CleanupNames(schema *schema_model.GoRelSchema) {
	nameMapper := func(r rune) bool {
		allowedSpecialCharacters := r == UNDERSCORE || r == MINUS
		chars := (r >= CapitalStart && r <= CapitalEnd) || (r >= LowercaseStart && r <= LowercaseEnd)
		allowedRange := r >= CharactersStart && r <= CharactersEnd
		if allowedRange && !(allowedSpecialCharacters || chars) {
			return false
		}
		return true
	}

	typeMapper := func(r rune) bool {
		allowedSpecialCharacters := r == UNDERSCORE || r == MINUS || r == QuestionMark || r == SquareBracketOpening || r == SquareBracketClosing
		chars := (r >= CapitalStart && r <= CapitalEnd) || (r >= LowercaseStart && r <= LowercaseEnd)
		allowedRange := r >= CharactersStart && r <= CharactersEnd
		if allowedRange && !(allowedSpecialCharacters || chars) {
			return false
		}
		return true
	}

	for index, enum := range schema.Enums {
		if checkNameForSpecialCharacter(enum.Name) {
			schema.Enums[index].Name = cleanupString(enum.Name, nameMapper)
		}
		for valueIndex, value := range enum.Values {
			if checkNameForSpecialCharacter(value) {
				schema.Enums[index].Values[valueIndex] = cleanupString(value, nameMapper)
			}
		}
	}
	for index, model := range schema.Models {
		if checkNameForSpecialCharacter(model.Name) {
			schema.Models[index].Name = cleanupString(model.Name, nameMapper)
		}
		for propertyIndex, property := range model.Properties {
			if checkNameForSpecialCharacter(property.Name) {
				schema.Models[index].Properties[propertyIndex].Name = cleanupString(property.Name, nameMapper)
			}
			if checkNameForSpecialCharacter(property.Type) {
				schema.Models[index].Properties[propertyIndex].Type = cleanupString(property.Type, typeMapper)
			}
			if checkNameForSpecialCharacter(property.RelationField) {
				schema.Models[index].Properties[propertyIndex].RelationField = cleanupString(property.RelationField, nameMapper)
			}
			if checkNameForSpecialCharacter(property.ReferenceField) {
				schema.Models[index].Properties[propertyIndex].ReferenceField = cleanupString(property.ReferenceField, nameMapper)
			}
		}
	}
}

//func cleanupNames(schema *schema_model.GoRelSchema) (newEnumNames []string, newModelNames []string) {
//	enumMap := make(map[string]string)
//
//	for index, enum := range schema.Enums {
//		oldEnumName := enum.Name
//		newEnumName := enum.Name
//
//		newEnumName = replaceSpecialCharacters(newEnumName)
//
//		schema.Enums[index].Name = newEnumName
//
//		enumMap[oldEnumName] = newEnumName
//		newEnumNames = append(newEnumNames, schema.Enums[index].Name)
//	}
//
//	modelMap := make(map[string]string)
//
//	//TODO: format reference and relational fields
//	for index, model := range schema.Models {
//		for indexP, property := range model.Properties {
//			newEnumName := enumMap[property.Type]
//			if newEnumName != "" {
//				schema.Models[index].Properties[indexP].Name = newEnumName
//			}
//		}
//
//		oldModelName := model.Name
//		newModelName := model.Name
//
//		newModelName = replaceSpecialCharacters(newModelName)
//		schema.Models[index].Name = newModelName
//
//		modelMap[oldModelName] = newModelName
//		modelMap[oldModelName+"[]"] = newModelName + "[]"
//		newModelNames = append(newModelNames, schema.Models[index].Name)
//	}
//
//	for index, model := range schema.Models {
//		for indexP, property := range model.Properties {
//			newType := modelMap[property.Type]
//			if newType != "" {
//				schema.Models[index].Properties[indexP].Type = newType
//			}
//		}
//	}
//
//	return
//}
//
//func cleanupEnumValues(schema *schema_model.GoRelSchema) {
//	for index, enum := range schema.Enums {
//		for indexV, value := range enum.Values {
//			schema.Enums[index].Values[indexV] = replaceSpecialCharacters(value)
//		}
//	}
//}

func validateType(property schema_model.Property, enumNames []string, modelNames []string) (isValid bool) {
	referenceNames := make([]string, len(modelNames))

	for index, modelName := range modelNames {
		referenceNames[index] = fmt.Sprintf("%s[]", modelName)
	}

	if isEnum := slices.Contains(enumNames, property.Type); isEnum {
		return true
	}

	if isRelationalField := property.RelationField != "" && property.ReferenceField != ""; isRelationalField {
		return true
	}

	if isReferenceField := slices.Contains(referenceNames, property.Type); isReferenceField {
		return true
	}

	if _, isValid := property.GetPostgresType(); !isValid {
		return false
	}

	return true
}

// TODO: Some enums can have empty values, but only if they are not used. Create function that will scan for those enums and delete them.
func validateEnums(schema schema_model.GoRelSchema) *validation_error.ValidationError {
	for _, enum := range schema.Enums {
		isNameEmpty := enum.Name == ""
		hasLessThanTwoValues := len(enum.Values) < 2

		if isNameEmpty {
			return &validation_error.ValidationError{
				Position: validation_error.EnumValidationError,
				Text:     "There is an enum, which name is empty",
			}
		}

		if checkNameForSpecialCharacter(enum.Name) {
			return &validation_error.ValidationError{
				Position: validation_error.EnumValidationError,
				Text:     fmt.Sprintf("Enum with name %s has special characters in it. Try using \"gorel clean\" to delete these characters.", enum.Name),
			}
		}

		if hasLessThanTwoValues {
			return &validation_error.ValidationError{
				Position: validation_error.EnumValidationError,
				Text:     fmt.Sprintf("Enum with name %s has less than 2 values (%v values)", enum.Name, len(enum.Values)),
			}
		}

		for _, value := range enum.Values {
			if value == "" {
				return &validation_error.ValidationError{
					Position: validation_error.EnumValidationError,
					Text:     fmt.Sprintf("Enum with name %s has empty values", enum.Name),
				}
			}
			if checkNameForSpecialCharacter(value) {
				return &validation_error.ValidationError{
					Position: validation_error.EnumValidationError,
					Text:     fmt.Sprintf("Value %s of enum with name %s has special characters in it", value, enum.Name),
				}
			}
		}

	}

	return nil
}

func validateModels(schema schema_model.GoRelSchema, enumNames []string, modelNames []string) *validation_error.ValidationError {
	if len(schema.Models) == 0 {
		return &validation_error.ValidationError{
			Position: validation_error.ModelValidationError,
			Text:     "No models provided",
		}
	}

	for _, model := range schema.Models {
		isNameEmpty := model.Name == ""
		hasLessThanTwoProperties := len(model.Properties) < 2
		idFieldCount := 0

		if isNameEmpty {
			return &validation_error.ValidationError{
				Position: validation_error.ModelValidationError,
				Text:     "There is a model, which name is empty",
			}
		}

		if checkNameForSpecialCharacter(model.Name) {
			return &validation_error.ValidationError{
				Position: validation_error.ModelValidationError,
				Text:     fmt.Sprintf("Model with name \"%s\" has special characters in it. Try using \"gorel clean\" to delete these characters. ", model.Name),
			}
		}

		if hasLessThanTwoProperties {
			return &validation_error.ValidationError{
				Position: validation_error.ModelValidationError,
				Text:     fmt.Sprintf("Model with name %s has less than 2 properties (%v properties)", model.Name, len(model.Properties)),
			}
		}

		for _, property := range model.Properties {
			if property.Name == "" {
				return &validation_error.ValidationError{
					Position: validation_error.ModelValidationError,
					Text:     fmt.Sprintf("Model with name %s has property with empty name", model.Name),
				}
			}

			if checkNameForSpecialCharacter(property.Name) {
				return &validation_error.ValidationError{
					Position: validation_error.ModelValidationError,
					Text:     fmt.Sprintf("Property %s of model with name %s has special characters in it", property.Name, model.Name),
				}
			}

			if property.Id {
				idFieldCount++
				isEnumType := slices.Contains(enumNames, property.Type)

				if isEnumType {
					return &validation_error.ValidationError{
						Position: validation_error.ModelValidationError,
						Text:     fmt.Sprintf("Model with name %s has id property (%s) with enum type", model.Name, property.Name),
					}
				}

				typeAsPropertyType := schema_model.PropertyType(property.Type)
				isOptionalType := typeAsPropertyType == schema_model.StringNullable || typeAsPropertyType == schema_model.IntNullable || typeAsPropertyType == schema_model.BooleanNullable || typeAsPropertyType == schema_model.FloatNullable || typeAsPropertyType == schema_model.DateTimeNullable
				isArrayType := typeAsPropertyType == schema_model.StringArr || typeAsPropertyType == schema_model.IntArr || typeAsPropertyType == schema_model.BooleanArr || typeAsPropertyType == schema_model.FloatArr || typeAsPropertyType == schema_model.DateTimeArr

				if isOptionalType {
					return &validation_error.ValidationError{
						Position: validation_error.ModelValidationError,
						Text:     fmt.Sprintf("Model with name %s has id property (%s) with optional type", model.Name, property.Name),
					}
				}

				if isArrayType {
					return &validation_error.ValidationError{
						Position: validation_error.ModelValidationError,
						Text:     fmt.Sprintf("Model with name %s has id property (%s) with array type", model.Name, property.Name),
					}
				}
			}

			//TODO: Table can possibly have 2 id fields. Add support for that.
			if idFieldCount > 1 {
				return &validation_error.ValidationError{
					Position: validation_error.ModelValidationError,
					Text:     fmt.Sprintf("model with name %s has more than one id field", model.Name),
				}
			}

			if isValid := validateType(property, enumNames, modelNames); !isValid {
				return &validation_error.ValidationError{
					Position: validation_error.ModelValidationError,
					Text:     fmt.Sprintf("%s type in %s property (%s model) is not valid", property.Type, property.Name, model.Name),
				}
			}

			if property.Default != "" {
				if _, err := property.ValidateDefaultValue(); err != nil {
					return err
				}
			}
		}

		if idFieldCount == 0 {
			return &validation_error.ValidationError{
				Position: validation_error.ModelValidationError,
				Text:     fmt.Sprintf("model with name %s does not have id field", model.Name),
			}
		}
	}

	return nil
}

func referenceFieldExists(schema schema_model.GoRelSchema, referenceModelName string, relationModelName string) bool {
	modelIndex := -1

	for index, model := range schema.Models {
		if model.Name == referenceModelName {
			modelIndex = index
			break
		}
	}

	if modelIndex == -1 {
		return false
	}

	model := schema.Models[modelIndex]

	for _, property := range model.Properties {
		if property.Type == relationModelName || property.Type == fmt.Sprintf("%s[]", relationModelName) {
			return true
		}
	}

	return false
}

func validateRelations(schema schema_model.GoRelSchema, modelNames []string) *validation_error.ValidationError {
	for _, model := range schema.Models {
		for _, property := range model.Properties {
			if property.ReferenceField != "" && property.RelationField != "" {
				if !referenceFieldExists(schema, property.Type, model.Name) {
					return &validation_error.ValidationError{
						Position: validation_error.RelationValidationError,
						Text:     fmt.Sprintf("relations should be created for both models %s and %s", property.Type, model.Name),
					}
				}
			}
		}
	}
	return nil
}

func ValidateSchema(schema *schema_model.GoRelSchema) (enumNames []string, modelNames []string, err error) {
	enumNames, modelNames = schema_parser.IndexSchema(*schema)
	if err := validateEnums(*schema); err != nil {
		return nil, nil, errors.New(fmt.Sprintf("error while validating enums:\n%s", err))
	}
	if err := validateModels(*schema, enumNames, modelNames); err != nil {
		return nil, nil, errors.New(fmt.Sprintf("error while validating models:\n%s", err))
	}
	if err := validateRelations(*schema, modelNames); err != nil {
		return nil, nil, errors.New(fmt.Sprintf("error while validating relations:\n%s", err))
	}

	return enumNames, modelNames, nil
}
