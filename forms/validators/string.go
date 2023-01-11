package validators

import "github.com/har4s/ohmygo/validation"

// Compare checks whether the validated value matches another string.
//
// Example:
//
//	validation.Field(&form.PasswordConfirm, validation.By(validators.Compare(form.Password)))
func Compare(valueToCompare string) validation.RuleFunc {
	return func(value any) error {
		v, _ := value.(string)

		if v != valueToCompare {
			return validation.NewError("validation_values_mismatch", "Values don't match.")
		}

		return nil
	}
}
