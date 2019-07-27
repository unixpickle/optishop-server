package db

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/unixpickle/optishop-server/optishop"
)

// runGenericTests runs a series of tests on an initially
// empty database.
func runGenericTests(t *testing.T, db DB) {
	t.Run("Users", func(t *testing.T) {
		uid1, err := db.CreateUser("bob", "pass", nil)
		if err != nil {
			t.Fatal(err)
		}
		if name, err := db.Username(uid1); err != nil {
			t.Fatal(err)
		} else if name != "bob" {
			t.Error("unexpected username:", name)
		}
		if _, err := db.CreateUser("bob", "pass", nil); err == nil {
			t.Error("account creation should have failed (with same password)")
		}
		if _, err := db.CreateUser("bob", "aoeu", nil); err == nil {
			t.Error("account creation should have failed (with different password)")
		}
		uid2, err := db.CreateUser("joe", "ssap", nil)
		if err != nil {
			t.Fatal(err)
		}
		if name, err := db.Username(uid2); err != nil {
			t.Fatal(err)
		} else if name != "joe" {
			t.Error("unexpected username:", name)
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
		if err := db.Chpass(uid1, "pass", "pass1"); err != nil {
			t.Error(err)
		}
		if err := db.Chpass(uid1, "pass", "pass1"); err == nil {
			t.Error("chpass should have failed")
		}
		if uid, err := db.Login("bob", "pass1"); err != nil || uid != uid1 {
			t.Error("login failed:", err)
		}
		if _, err := db.Login("bob", "pass"); err == nil {
			t.Error("login should fail with old password")
		}

		// Make sure changing the password didn't affect
		// the other user.
		if uid, err := db.Login("joe", "ssap"); err != nil || uid != uid2 {
			t.Error("login failed:", err)
		}
	})

	t.Run("Metadata", func(t *testing.T) {
		uid1, err := db.CreateUser("metaTester", "pass", map[string]string{"secret": "hi"})
		if err != nil {
			t.Fatal(err)
		}
		uid2, err := db.CreateUser("metaTester2", "pass", nil)
		if err != nil {
			t.Fatal(err)
		}
		if data, err := db.UserMetadata(uid1, "secret"); err != nil {
			t.Error(err)
		} else if data != "hi" {
			t.Error("unexpected data:", data)
		}
		if _, err := db.UserMetadata(uid2, "secret"); err == nil {
			t.Error("expected error when reading metadata")
		}

		db.SetUserMetadata(uid1, "secret", "hey")
		if data, err := db.UserMetadata(uid1, "secret"); err != nil {
			t.Error(err)
		} else if data != "hey" {
			t.Error("unexpected data:", data)
		}

		db.SetUserMetadata(uid2, "secret", "hello")
		if data, err := db.UserMetadata(uid2, "secret"); err != nil {
			t.Error(err)
		} else if data != "hello" {
			t.Error("unexpected data:", data)
		}
	})

	t.Run("Stores", func(t *testing.T) {
		user, err := db.CreateUser("storeTester", "pass", nil)
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

		if store, err := db.Store(user, newID2); err != nil {
			t.Error(err)
		} else if store.ID != newID2 {
			t.Error("unexpected result")
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
		if _, err := db.Store(user, newID1); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Lists", func(t *testing.T) {
		user, err := db.CreateUser("listTester", "pass", nil)
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

	t.Run("Permute", func(t *testing.T) {
		user, err := db.CreateUser("permuteTester", "pass", nil)
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

		ids := make([]ListEntryID, 10)
		idToInfo := map[ListEntryID]*ListEntryInfo{}
		for i := range ids {
			item := &ListEntryInfo{
				InventoryProductData: []byte(fmt.Sprintf("product %d", i)),
				Zone: &optishop.Zone{
					Name:     fmt.Sprintf("A%d", i),
					Location: optishop.Point{X: float64(i), Y: float64(i)},
				},
				Floor: i,
			}
			id, err := db.AddListEntry(user, store, item)
			if err != nil {
				t.Fatal(err)
			}
			ids[i] = id
			idToInfo[id] = item
		}

		checkContents := func() {
			entries, err := db.ListEntries(user, store)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 10 {
				t.Fatalf("invalid count: %d", len(entries))
			}
			for i, entry := range entries {
				if entry.ID != ids[i] || !listEntriesEqual(entry.Info, idToInfo[entry.ID]) {
					t.Fatal("invalid entry or ID")
				}
			}
		}

		checkContents()
		for i := 0; i < 10; i++ {
			rand.Shuffle(len(ids), func(i, j int) {
				ids[i], ids[j] = ids[j], ids[i]
			})
			if err := db.PermuteListEntries(user, store, ids); err != nil {
				t.Fatal(err)
			}
			checkContents()
		}

		errorPerms := [][]ListEntryID{
			append([]ListEntryID{ids[0]}, ids...),
			ids[:len(ids)-1],
			append([]ListEntryID{ids[0]}, ids[:len(ids)-1]...),
			append([]ListEntryID{"notarealid1231231"}, ids[:len(ids)-1]...),
			append([]ListEntryID{"notarealid1231231"}, ids[:len(ids)]...),
		}
		for i, perm := range errorPerms {
			if err := db.PermuteListEntries(user, store, perm); err == nil {
				t.Fatalf("case %d should have failed", i)
			}
			checkContents()
		}
	})
}

func listEntriesEqual(l1, l2 *ListEntryInfo) bool {
	return bytes.Equal(l1.InventoryProductData, l2.InventoryProductData) &&
		reflect.DeepEqual(l1.Zone, l2.Zone) &&
		l1.Floor == l2.Floor
}
