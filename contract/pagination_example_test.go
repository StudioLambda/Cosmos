package contract_test

import (
	"fmt"

	"github.com/studiolambda/cosmos/contract"
)

func ExamplePaginate() {
	items := []string{"a", "b", "c"}

	page := contract.Paginate(items, 10, 2, 3)

	fmt.Println(page.CurrentPage)
	fmt.Println(page.LastPage)
	fmt.Println(page.PerPage)
	fmt.Println(page.Total)
	fmt.Println(page.Items)
	// Output:
	// 2
	// 4
	// 3
	// 10
	// [a b c]
}

func ExampleCursorPaginate() {
	type User struct {
		ID   int
		Name string
	}

	users := []User{{ID: 10, Name: "Alice"}, {ID: 20, Name: "Bob"}}

	cursor, err := contract.CursorPaginate(users, 2, true, false, func(u User) (string, error) {
		return contract.CursorEncode(u.ID)
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(cursor.PerPage)
	fmt.Println(cursor.NextCursor != "")
	fmt.Println(cursor.PrevCursor)

	id, err := contract.CursorDecode[int](cursor.NextCursor)

	if err != nil {
		panic(err)
	}

	fmt.Println(id)
	// Output:
	// 2
	// true
	//
	// 20
}

func ExampleCursorEncode() {
	encoded, err := contract.CursorEncode(42)

	if err != nil {
		panic(err)
	}

	fmt.Println(encoded)
	// Output: NDI
}

func ExampleCursorDecode() {
	value, err := contract.CursorDecode[int]("NDI")

	if err != nil {
		panic(err)
	}

	fmt.Println(value)
	// Output: 42
}
