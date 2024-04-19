package crow_test

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"pjm.dev/crow"
)

type User struct {
	gorm.Model
	Name string
}

// User implements the schema.Tabler interface
func (u User) TableName() string {
	return "users"
}

func TestCrow(t *testing.T) {
	// open an empty in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	// create crow
	c := crow.New(db)

	// create our seed with two users: Foo and Bar
	seed := crow.Database{User{}: {User{Name: "Foo"}, User{Name: "Bar"}}}

	// seed crow
	err = c.Seed(seed)
	if err != nil {
		t.Fatal(err)
	}

	// delete users named Foo
	if err := db.Where("name = ?", "Foo").Delete(&User{}).Error; err != nil {
		t.Fatal(err)
	}

	// dump crow
	dump, err := c.Dump()
	if err != nil {
		t.Fatal(err)
	}

	// perform some assertions on the dump
	users, ok := dump[User{}]
	if !ok {
		t.Fatal("dump does not contain users")
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].(User).Name != "Bar" {
		t.Fatalf("expected user name Bar, got %s", users[0].(User).Name)
	}
}
