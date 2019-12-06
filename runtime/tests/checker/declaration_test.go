package checker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dapperlabs/flow-go/language/runtime/common"
	"github.com/dapperlabs/flow-go/language/runtime/sema"
	. "github.com/dapperlabs/flow-go/language/runtime/tests/utils"
)

func TestCheckConstantAndVariableDeclarations(t *testing.T) {

	checker, err := ParseAndCheck(t, `
        let x = 1
        var y = 1
    `)

	require.NoError(t, err)

	assert.IsType(t,
		&sema.IntType{},
		checker.GlobalValues["x"].Type,
	)

	assert.IsType(t,
		&sema.IntType{},
		checker.GlobalValues["y"].Type,
	)
}

func TestCheckInvalidGlobalConstantRedeclaration(t *testing.T) {

	_, err := ParseAndCheck(t, `
        fun x() {}

        let y = true
        let y = false
    `)

	errs := ExpectCheckerErrors(t, err, 1)

	assert.IsType(t, &sema.RedeclarationError{}, errs[0])
}

func TestCheckInvalidGlobalFunctionRedeclaration(t *testing.T) {

	_, err := ParseAndCheck(t, `
        let x = true

        fun y() {}
        fun y() {}
    `)

	errs := ExpectCheckerErrors(t, err, 1)

	assert.IsType(t, &sema.RedeclarationError{}, errs[0])
}

func TestCheckInvalidLocalRedeclaration(t *testing.T) {

	_, err := ParseAndCheck(t, `
        fun test() {
            let x = true
            let x = false
        }
    `)

	errs := ExpectCheckerErrors(t, err, 1)

	assert.IsType(t, &sema.RedeclarationError{}, errs[0])
}

func TestCheckInvalidLocalFunctionRedeclaration(t *testing.T) {

	_, err := ParseAndCheck(t, `
        fun test() {
            let x = true

            fun y() {}
            fun y() {}
        }
    `)

	errs := ExpectCheckerErrors(t, err, 1)

	assert.IsType(t, &sema.RedeclarationError{}, errs[0])
}

func TestCheckInvalidUnknownDeclaration(t *testing.T) {

	_, err := ParseAndCheck(t, `
       fun test() {
           return x
       }
    `)

	errs := ExpectCheckerErrors(t, err, 2)

	assert.IsType(t, &sema.NotDeclaredError{}, errs[0])
	assert.IsType(t, &sema.InvalidReturnValueError{}, errs[1])
}

func TestCheckInvalidUnknownDeclarationInGlobal(t *testing.T) {

	_, err := ParseAndCheck(t, `
       let x = y
    `)

	errs := ExpectCheckerErrors(t, err, 1)

	assert.IsType(t, &sema.NotDeclaredError{}, errs[0])
}

func TestCheckInvalidUnknownDeclarationInGlobalAndUnknownType(t *testing.T) {

	_, err := ParseAndCheck(t, `
       let x: X = y
    `)

	errs := ExpectCheckerErrors(t, err, 2)

	assert.IsType(t, &sema.NotDeclaredError{}, errs[0])
	assert.Equal(t,
		"y",
		errs[0].(*sema.NotDeclaredError).Name,
	)
	assert.Equal(t,
		common.DeclarationKindVariable,
		errs[0].(*sema.NotDeclaredError).ExpectedKind,
	)

	assert.IsType(t, &sema.NotDeclaredError{}, errs[1])
	assert.Equal(t,
		"X",
		errs[1].(*sema.NotDeclaredError).Name,
	)
	assert.Equal(t,
		common.DeclarationKindType,
		errs[1].(*sema.NotDeclaredError).ExpectedKind,
	)
}

func TestCheckInvalidUnknownDeclarationCallInGlobal(t *testing.T) {

	_, err := ParseAndCheck(t, `
       let x = y()
    `)

	errs := ExpectCheckerErrors(t, err, 1)

	assert.IsType(t, &sema.NotDeclaredError{}, errs[0])
}

func TestCheckInvalidRedeclarations(t *testing.T) {

	_, err := ParseAndCheck(t, `
      fun test(a: Int, a: Int) {
        let x = 1
        let x = 2
      }
    `)

	errs := ExpectCheckerErrors(t, err, 2)

	assert.IsType(t, &sema.RedeclarationError{}, errs[0])
	assert.IsType(t, &sema.RedeclarationError{}, errs[1])
}

func TestCheckInvalidConstantValue(t *testing.T) {

	_, err := ParseAndCheck(t, `
      let x: Bool = 1
    `)

	errs := ExpectCheckerErrors(t, err, 1)

	assert.IsType(t, &sema.TypeMismatchError{}, errs[0])
}

func TestCheckInvalidReference(t *testing.T) {

	_, err := ParseAndCheck(t, `
      fun test() {
          testX
      }
    `)

	errs := ExpectCheckerErrors(t, err, 1)

	assert.IsType(t, &sema.NotDeclaredError{}, errs[0])
}

func TestCheckInvalidVariableDeclarationSecondValueNotDeclared(t *testing.T) {

	_, err := ParseAndCheck(t, `
       var y = 2
       let z = y = x
   `)

	errs := ExpectCheckerErrors(t, err, 2)

	assert.IsType(t, &sema.NonResourceTypeError{}, errs[0])
	assert.IsType(t, &sema.NotDeclaredError{}, errs[1])
}

func TestCheckInvalidVariableDeclarationSecondValueCopyTransfers(t *testing.T) {

	_, err := ParseAndCheck(t, `
       var x = 1
       var y = 2
       let z = y = x
   `)

	errs := ExpectCheckerErrors(t, err, 1)

	assert.IsType(t, &sema.NonResourceTypeError{}, errs[0])
}

func TestCheckInvalidVariableDeclarationSecondValueNotTarget(t *testing.T) {

	_, err := ParseAndCheck(t, `
      resource X {}
      fun f() {}

      let x <- create X()
      let z = f() <- x
  `)

	errs := ExpectCheckerErrors(t, err, 1)

	assert.IsType(t, &sema.InvalidAssignmentTargetError{}, errs[0])
}

func TestCheckInvalidVariableDeclarationSecondValueCopyTransferSecond(t *testing.T) {

	_, err := ParseAndCheck(t, `
     resource R {}

     let x <- create R()
     var y <- create R()
     let z <- y = x
   `)

	errs := ExpectCheckerErrors(t, err, 1)

	assert.IsType(t, &sema.IncorrectTransferOperationError{}, errs[0])
}

func TestCheckInvalidVariableDeclarationSecondValueCopyTransferFirst(t *testing.T) {

	_, err := ParseAndCheck(t, `
     resource R {}

     let x <- create R()
     var y <- create R()
     let z = y <- x
   `)

	errs := ExpectCheckerErrors(t, err, 1)

	assert.IsType(t, &sema.IncorrectTransferOperationError{}, errs[0])
}

func TestCheckInvalidVariableDeclarationSecondValueConstant(t *testing.T) {

	_, err := ParseAndCheck(t, `
     resource R {}

     let x <- create R()
     let y <- create R()
     let z <- y <- x
   `)

	errs := ExpectCheckerErrors(t, err, 1)

	assert.IsType(t, &sema.AssignmentToConstantError{}, errs[0])
}

func TestCheckInvalidVariableDeclarationSecondValueTypeMismatch(t *testing.T) {

	_, err := ParseAndCheck(t, `
     resource X {}
     resource Y {}

     let x <- create X()
     var y <- create Y()
     let z <- y <- x
   `)

	errs := ExpectCheckerErrors(t, err, 1)

	assert.IsType(t, &sema.TypeMismatchError{}, errs[0])
}

func TestCheckInvalidVariableDeclarationSecondValueUseAfterInvalidation(t *testing.T) {

	_, err := ParseAndCheck(t, `
     resource R {}

     let x <- create R()
     var y <- create R()
     let z <- y <- x

     let r <- x
   `)

	errs := ExpectCheckerErrors(t, err, 1)

	assert.IsType(t, &sema.ResourceUseAfterInvalidationError{}, errs[0])
}

func TestCheckVariableDeclarationSecondValue(t *testing.T) {

	checker, err := ParseAndCheck(t, `
     resource R {}

     let x <- create R()
     var y <- create R()
     let z <- y <- x

     let r <- y
   `)

	require.NoError(t, err)

	assert.IsType(t,
		&sema.CompositeType{},
		checker.GlobalValues["x"].Type,
	)

	assert.IsType(t,
		&sema.CompositeType{},
		checker.GlobalValues["y"].Type,
	)

	assert.IsType(t,
		&sema.CompositeType{},
		checker.GlobalValues["z"].Type,
	)

	assert.IsType(t,
		&sema.CompositeType{},
		checker.GlobalValues["r"].Type,
	)
}

func TestCheckVariableDeclarationSecondValueDictionary(t *testing.T) {

	checker, err := ParseAndCheck(t, `
     resource R {}

     let x <- create R()
     var ys <- {"r": <-create R()}
     // NOTE: nested move is valid here
     let z <- ys["r"] <- x

     // NOTE: nested move is invalid here
     let r <- ys.remove(key: "r")
   `)

	require.Nil(t, err)

	assert.IsType(t,
		&sema.CompositeType{},
		checker.GlobalValues["x"].Type,
	)

	assert.IsType(t,
		&sema.DictionaryType{},
		checker.GlobalValues["ys"].Type,
	)

	assert.IsType(t,
		&sema.OptionalType{},
		checker.GlobalValues["z"].Type,
	)

	assert.IsType(t,
		&sema.OptionalType{},
		checker.GlobalValues["r"].Type,
	)
}

func TestCheckVariableDeclarationSecondValueNil(t *testing.T) {

	_, err := ParseAndCheck(t, `
     resource R {}

     fun test() {
         var x: <-R? <- create R()
         let y <- x <- nil
         destroy x
         destroy y
     }
   `)

	require.NoError(t, err)
}

func TestCheckTopLevelContractRestriction(t *testing.T) {

	_, err := ParseAndCheckWithOptions(t,
		`
          contract C {}
        `,
		ParseAndCheckOptions{
			Options: []sema.Option{
				sema.WithValidTopLevelDeclarations(
					[]common.DeclarationKind{
						common.DeclarationKindContract,
						common.DeclarationKindImport,
					},
				),
				sema.WithAccessCheckMode(sema.AccessCheckModeNotSpecifiedUnrestricted),
			},
		},
	)

	require.NoError(t, err)
}

func TestCheckInvalidTopLevelContractRestriction(t *testing.T) {

	tests := map[string]string{
		"resource":           `resource Test {}`,
		"struct":             `struct Test {}`,
		"resource interface": `resource interface Test {}`,
		"struct interface":   `struct interface Test {}`,
		"event":              `event Test()`,
		"function":           `fun test() {}`,
		"transaction":        `transaction { execute {} }`,
		"constant":           `var x = 1`,
		"variable":           `let x = 1`,
	}

	for name, code := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := ParseAndCheckWithOptions(t,
				code,
				ParseAndCheckOptions{
					Options: []sema.Option{
						sema.WithValidTopLevelDeclarations(
							[]common.DeclarationKind{
								common.DeclarationKindContractInterface,
								common.DeclarationKindContract,
								common.DeclarationKindImport,
							},
						),
						sema.WithAccessCheckMode(sema.AccessCheckModeNotSpecifiedUnrestricted),
					},
				},
			)

			errs := ExpectCheckerErrors(t, err, 1)

			assert.IsType(t, &sema.InvalidTopLevelDeclarationError{}, errs[0])
		})
	}
}
