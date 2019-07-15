package db

import (
	"testing"

	"github.com/unixpickle/optishop-server/optishop"
)

// runGenericTests runs a series of tests on an initially
// empty database.
func runGenericTests(t *testing.T, db DB) {
	t.Run("Users", func(t *testing.T) {
		uid1, err := db.CreateUser("bob", "pass")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.CreateUser("bob", "pass"); err == nil {
			t.Error("account creation should have failed (with same password)")
		}
		if _, err := db.CreateUser("bob", "aoeu"); err == nil {
			t.Error("account creation should have failed (with different password)")
		}
		uid2, err := db.CreateUser("joe", "ssap")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Login("bob", "aoeu"); err == nil {
			t.Error("login should have failed with incorrect password")
		}
		if uid, err := db.Login("bob", "pass"); err != nil || uid != uid1 {
			t.Errorf("incorrect login result: %v, %v", uid, err)
		}
		if _, err := db.Login("joe", "pass"); err == nil {
			t.Error("login should have failed with incorrect password")
		}
		if uid, err := db.Login("joe", "ssap"); err != nil || uid != uid2 {
			t.Errorf("incorrect login result: %v, %v", uid, err)
		}
	})

	t.Run("Stores", func(t *testing.T) {
		user, err := db.CreateUser("storeTester", "pass")
		if err != nil {
			t.Fatal(err)
		}

		stores, err := db.Stores(user)
		if err != nil {
			t.Fatal(err)
		}
		if len(stores) != 0 {
			t.Error("there are already stores in this user")
		}

		newID1, err := db.AddStore(user, &StoreInfo{
			SourceName: "target",
			StoreName:  "tribeca",
			StoreData:  []byte("hello"),
		})
		if err != nil {
			t.Fatal(err)
		}
		newID2, err := db.AddStore(user, &StoreInfo{
			SourceName: "walmart",
			StoreName:  "mart",
			StoreData:  []byte("goodbye"),
		})
		if err != nil {
			t.Fatal(err)
		}

		stores, err = db.Stores(user)
		if err != nil {
			t.Fatal(err)
		}
		if len(stores) != 2 {
			t.Fatal("incorrect number of stores (expected 2):", len(stores))
		}
		if stores[0].ID != newID1 || stores[0].Info.SourceName != "target" ||
			stores[0].Info.StoreName != "tribeca" || string(stores[0].Info.StoreData) != "hello" {
			t.Error("incorrect fields for first store")
		}
		if stores[1].ID != newID2 || stores[1].Info.SourceName != "walmart" ||
			stores[1].Info.StoreName != "mart" || string(stores[1].Info.StoreData) != "goodbye" {
			t.Error("incorrect fields for second store")
		}

		if err := db.RemoveStore(user, newID1); err != nil {
			t.Fatal(err)
		}
		if err := db.RemoveStore(user, newID1); err == nil {
			t.Error("expected error on redundant removal")
		}
		stores, err = db.Stores(user)
		if err != nil {
			t.Fatal(err)
		}
		if len(stores) != 1 {
			t.Fatal("incorrect number of stores (expected 1):", len(stores))
		}
		if stores[0].ID != newID2 || stores[0].Info.SourceName != "walmart" ||
			stores[0].Info.StoreName != "mart" || string(stores[0].Info.StoreData) != "goodbye" {
			t.Error("incorrect fields for first store")
		}
	})

	t.Run("Lists", func(t *testing.T) {
		user, err := db.CreateUser("listTester", "pass")
		if err != nil {
			t.Fatal(err)
		}

		store, err := db.AddStore(user, &StoreInfo{
			SourceName: "target",
			StoreName:  "tribeca",
			StoreData:  []byte("hello"),
		})
		if err != nil {
			t.Fatal(err)
		}

		list, err := db.ListEntries(user, store)
		if err != nil {
			t.Fatal(err)
		} else if len(list) != 0 {
			t.Fatal("expected zero entries but got:", len(list))
		}

		newID1, err := db.AddListEntry(user, store, &ListEntryInfo{
			InventoryProductData: []byte("hello"),
			Zone:                 &optishop.Zone{Name: "hi"},
		})
		if err != nil {
			t.Fatal(err)
		}

		newID2, err := db.AddListEntry(user, store, &ListEntryInfo{
			InventoryProductData: []byte("goodbye"),
			Zone:                 &optishop.Zone{Name: "bye"},
		})
		if err != nil {
			t.Fatal(err)
		}

		list, err = db.ListEntries(user, store)
		if err != nil {
			t.Fatal(err)
		}
		if len(list) != 2 {
			t.Fatal("expected two entries but got:", len(list))
		}

		if list[0].ID != newID1 || string(list[0].Info.InventoryProductData) != "hello" ||
			list[0].Info.Zone.Name != "hi" {
			t.Error("incorrect fields in first entry")
		}
		if list[1].ID != newID2 || string(list[1].Info.InventoryProductData) != "goodbye" ||
			list[1].Info.Zone.Name != "bye" {
			t.Error("incorrect fields in second entry")
		}

		if err := db.RemoveListEntry(user, store, newID1); err != nil {
			t.Fatal(err)
		}

		list, err = db.ListEntries(user, store)
		if err != nil {
			t.Fatal(err)
		}
		if len(list) != 1 {
			t.Fatal("expected one entry but got:", len(list))
		}
		if list[0].ID != newID2 || string(list[0].Info.InventoryProductData) != "goodbye" ||
			list[0].Info.Zone.Name != "bye" {
			t.Error("incorrect fields in first entry")
		}
	})
}
