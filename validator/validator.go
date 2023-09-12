package validator

import (
	"GoRelCli/schema_model"
	"errors"
	"fmt"
	"slices"
	"strings"
)

func cleanupNames(schema *schema_model.GoRelSchema) (newEnumNames []string, newModelNames []string) {
	enumMap := make(map[string]string)

	for index, enum := range schema.Enums {
		oldEnumName := enum.Name
		newEnumName := enum.Name

		newEnumName = strings.Replace(newEnumName, "\"", "\\\"", -1)
		newEnumName = strings.Replace(newEnumName, "'", "\\'", -1)

		schema.Enums[index].Name = newEnumName

		enumMap[oldEnumName] = newEnumName
		newEnumNames = append(newEnumNames, schema.Enums[index].Name)
	}

	modelMap := make(map[string]string)

	//TODO: format reference and relational fields
	for index, model := range schema.Models {
		for indexP, property := range model.Properties {
			newEnumName := enumMap[property.Type]
			if newEnumName != "" {
				schema.Models[index].Properties[indexP].Name = newEnumName
			}
		}

		oldModelName := model.Name
		newModelName := model.Name

		newModelName = strings.Replace(newModelName, "\"", "\\\"", -1)
		newModelName = strings.Replace(newModelName, "'", "\\'", -1)
		schema.Models[index].Name = newModelName

		modelMap[oldModelName] = newModelName
		modelMap[oldModelName+"[]"] = newModelName + "[]"
		newModelNames = append(newModelNames, schema.Models[index].Name)
	}

	for index, model := range schema.Models {
		for indexP, property := range model.Properties {
			newType := modelMap[property.Type]
			if newType != "" {
				schema.Models[index].Properties[indexP].Type = newType
			}
		}
	}

	return
}

func cleanupEnumValues(schema *schema_model.GoRelSchema) {
	for index, enum := range schema.Enums {
		for indexV, value := range enum.Values {
			schema.Enums[index].Values[indexV] = strings.Replace(value, "\"", "\\\"", -1)
			schema.Enums[index].Values[indexV] = strings.Replace(value, "\\'", "\\'", -1)
		}
	}
}

func validateType(property schema_model.Property, enumNames []string, modelNames []string, modelName string) error {
	referenceNames := make([]string, len(modelNames))

	for index, modelName := range modelNames {
		referenceNames[index] = fmt.Sprintf("%s[]", modelName)
	}

	if isEnum := slices.Contains(enumNames, property.Type); isEnum {
		return nil
	}

	if isRelationalField := property.RelationField != "" && property.ReferenceField != ""; isRelationalField {
		return nil
	}

	if isReferenceField := slices.Contains(referenceNames, property.Type); isReferenceField {
		return nil
	}

	postgresType, isValid := property.GetPostgresType()

	if !isValid {
		return errors.New(fmt.Sprintf("%s type in %s property (%s model) is not valid", postgresType, property.Name, modelName))
	}

	return nil
}

// TODO: Some enums can have empty values, but only if they are not used. Create function that will scan for those enums and delete them.

func validateEnums(schema schema_model.GoRelSchema) error {
	for _, enum := range schema.Enums {
		isNameEmpty := enum.Name == ""
		hasLessThanTwoValues := len(enum.Values) < 2

		if isNameEmpty {
			return errors.New("enum name is empty")
		}

		if hasLessThanTwoValues {
			return errors.New(fmt.Sprintf("enum with name %s has less than 2 values (%v values)", enum.Name, len(enum.Values)))
		}

		for _, value := range enum.Values {
			if value == "" {
				return errors.New(fmt.Sprintf("enum with name %s has values with empty values", enum.Name))
			}
		}

	}

	return nil
}

func validateModels(schema schema_model.GoRelSchema, enumNames []string, modelNames []string) error {
	if len(schema.Models) == 0 {
		return errors.New("no models provided")
	}

	for _, model := range schema.Models {
		isNameEmpty := model.Name == ""
		hasLessThanTwoProperties := len(model.Properties) < 2
		idFieldCount := 0

		if isNameEmpty {
			return errors.New("model name is empty")
		}

		if hasLessThanTwoProperties {
			return errors.New(fmt.Sprintf("model with name %s has less than 2 properties (%v properties)", model.Name, len(model.Properties)))
		}

		for _, property := range model.Properties {
			if property.Name == "" {
				return errors.New(fmt.Sprintf("model with name %s has property with empty name", model.Name))
			}

			if property.Id {
				idFieldCount++
				isEnumType := slices.Contains(enumNames, property.Type)

				if isEnumType {
					return errors.New(fmt.Sprintf("model with name %s has id property (%s) with enum type", model.Name, property.Name))
				}

				typeAsPropertyType := schema_model.PropertyType(property.Type)
				isOptionalType := typeAsPropertyType == schema_model.StringNullable || typeAsPropertyType == schema_model.IntNullable || typeAsPropertyType == schema_model.BooleanNullable || typeAsPropertyType == schema_model.FloatNullable || typeAsPropertyType == schema_model.DateTimeNullable
				isArrayType := typeAsPropertyType == schema_model.StringArr || typeAsPropertyType == schema_model.IntArr || typeAsPropertyType == schema_model.BooleanArr || typeAsPropertyType == schema_model.FloatArr || typeAsPropertyType == schema_model.DateTimeArr

				if isOptionalType {
					return errors.New(fmt.Sprintf("model with name %s has id property (%s) with optional type", model.Name, property.Name))
				}

				if isArrayType {
					return errors.New(fmt.Sprintf("model with name %s has id property (%s) with array type", model.Name, property.Name))
				}
			}

			//TODO: Table can possibly have 2 id fields. Add support for that.
			if idFieldCount > 1 {
				return errors.New(fmt.Sprintf("model with name %s has more than one id field", model.Name))
			}

			if err := validateType(property, enumNames, modelNames, model.Name); err != nil {
				return err
			}

			if property.Default != "" {
				if _, err := property.ValidateDefaultValue(); err != nil {
					return err
				}
			}
		}

		if idFieldCount == 0 {
			return errors.New(fmt.Sprintf("model with name %s does not have id field", model.Name))
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

func validateRelations(schema schema_model.GoRelSchema, modelNames []string) error {
	for _, model := range schema.Models {
		for _, property := range model.Properties {
			if property.ReferenceField != "" && property.RelationField != "" {
				if !referenceFieldExists(schema, property.Type, model.Name) {
					return errors.New(fmt.Sprintf("relations should be created for both models %s and %s", property.Type, model.Name))
				}
			}
		}
	}
	return nil
}

func ValidateSchema(schema *schema_model.GoRelSchema) (enumNames []string, modelNames []string, err error) {
	enumNames, modelNames = cleanupNames(schema)
	cleanupEnumValues(schema)
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
