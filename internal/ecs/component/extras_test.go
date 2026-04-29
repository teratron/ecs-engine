package component

import (
	"reflect"
	"testing"
)

func TestRegisterByTypeIdempotent(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	id1 := r.RegisterByType(reflect.TypeOf(Position{}))
	id2 := r.RegisterByType(reflect.TypeOf(Position{}))
	if id1 != id2 {
		t.Fatalf("RegisterByType must be idempotent; got %d and %d", id1, id2)
	}
	if !id1.IsValid() {
		t.Fatalf("RegisterByType must return a valid ID; got %d", id1)
	}
}

func TestRegisterByTypeNilPanics(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	defer func() {
		if recover() == nil {
			t.Fatal("RegisterByType(nil) must panic")
		}
	}()
	r.RegisterByType(nil)
}

func TestRegisterByTypeMatchesGenericRegisterType(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	idGeneric := RegisterType[Position](r)
	idReflect := r.RegisterByType(reflect.TypeOf(Position{}))
	if idGeneric != idReflect {
		t.Fatalf("RegisterType[T] and RegisterByType(reflect.Type) must agree; got %d vs %d",
			idGeneric, idReflect)
	}
}

// Helpers for the table extras tests.
func newPosVelTable(t *testing.T) (*Table, ID, ID) {
	t.Helper()
	r := NewRegistry()
	posID := RegisterType[Position](r)
	velID := RegisterType[Velocity](r)
	specs := []ColumnSpec{
		ColumnSpecFromInfo(r.Info(posID)),
		ColumnSpecFromInfo(r.Info(velID)),
	}
	return NewTable(specs, 0), posID, velID
}

func TestTableHasColumn(t *testing.T) {
	t.Parallel()

	tbl, posID, _ := newPosVelTable(t)
	if !tbl.HasColumn(posID) {
		t.Fatal("HasColumn must return true for registered column")
	}
	if tbl.HasColumn(ID(99)) {
		t.Fatal("HasColumn must return false for unknown ID")
	}
}

func TestTableCellPtrByID(t *testing.T) {
	t.Parallel()

	tbl, posID, velID := newPosVelTable(t)
	row := tbl.AddRow(map[ID]any{
		posID: Position{X: 1, Y: 2, Z: 3},
		velID: Velocity{DX: 7, DY: 8, DZ: 9},
	})

	ptr, ok := tbl.CellPtrByID(posID, row)
	if !ok {
		t.Fatal("CellPtrByID must succeed for registered column")
	}
	if ptr == nil {
		t.Fatal("CellPtrByID must return non-nil for non-zero-size column")
	}
	got := *(*Position)(ptr)
	if got.X != 1 || got.Y != 2 || got.Z != 3 {
		t.Fatalf("Position via CellPtrByID = %+v, want {1,2,3}", got)
	}
}

func TestTableCellPtrByIDUnknown(t *testing.T) {
	t.Parallel()

	tbl, posID, _ := newPosVelTable(t)
	tbl.AddRow(map[ID]any{posID: Position{}})
	if _, ok := tbl.CellPtrByID(ID(99), 0); ok {
		t.Fatal("CellPtrByID with unknown ID must return ok=false")
	}
}

func TestTableCellPtrByIDOutOfRangePanics(t *testing.T) {
	t.Parallel()

	tbl, posID, _ := newPosVelTable(t)
	tbl.AddRow(map[ID]any{posID: Position{}})
	defer func() {
		if recover() == nil {
			t.Fatal("CellPtrByID with out-of-range row must panic")
		}
	}()
	tbl.CellPtrByID(posID, 99)
}

func TestTableSetCellByID(t *testing.T) {
	t.Parallel()

	tbl, posID, _ := newPosVelTable(t)
	row := tbl.AddRow(map[ID]any{posID: Position{X: 0}})
	tbl.SetCellByID(posID, row, Position{X: 42, Y: 7, Z: 1})

	ptr, _ := tbl.CellPtrByID(posID, row)
	got := *(*Position)(ptr)
	if got.X != 42 || got.Y != 7 || got.Z != 1 {
		t.Fatalf("SetCellByID round-trip = %+v, want {42,7,1}", got)
	}
}

func TestTableSetCellByIDUnknownPanics(t *testing.T) {
	t.Parallel()

	tbl, posID, _ := newPosVelTable(t)
	tbl.AddRow(map[ID]any{posID: Position{}})
	defer func() {
		if recover() == nil {
			t.Fatal("SetCellByID with unknown ID must panic")
		}
	}()
	tbl.SetCellByID(ID(99), 0, Position{})
}

func TestTableRowValues(t *testing.T) {
	t.Parallel()

	tbl, posID, velID := newPosVelTable(t)
	row := tbl.AddRow(map[ID]any{
		posID: Position{X: 5, Y: 6, Z: 7},
		velID: Velocity{DX: 1, DY: 2, DZ: 3},
	})

	values := tbl.RowValues(row)
	if len(values) != 2 {
		t.Fatalf("RowValues len = %d, want 2", len(values))
	}
	if v, ok := values[posID].(Position); !ok || v.X != 5 || v.Y != 6 || v.Z != 7 {
		t.Fatalf("RowValues[posID] = %+v, want Position{5,6,7}", values[posID])
	}
	if v, ok := values[velID].(Velocity); !ok || v.DX != 1 || v.DY != 2 || v.DZ != 3 {
		t.Fatalf("RowValues[velID] = %+v, want Velocity{1,2,3}", values[velID])
	}
}

func TestTableRowValuesOutOfRangePanics(t *testing.T) {
	t.Parallel()

	tbl, posID, _ := newPosVelTable(t)
	tbl.AddRow(map[ID]any{posID: Position{}})
	defer func() {
		if recover() == nil {
			t.Fatal("RowValues with out-of-range row must panic")
		}
	}()
	tbl.RowValues(99)
}

// Tag (zero-size) round-trip for CellPtrByID and RowValues.
type ZeroTag struct{}

func TestTableCellPtrByIDZeroSize(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	tagID := RegisterType[ZeroTag](r)
	posID := RegisterType[Position](r)
	specs := []ColumnSpec{
		ColumnSpecFromInfo(r.Info(tagID)),
		ColumnSpecFromInfo(r.Info(posID)),
	}
	tbl := NewTable(specs, 0)
	row := tbl.AddRow(map[ID]any{tagID: ZeroTag{}, posID: Position{}})

	ptr, ok := tbl.CellPtrByID(tagID, row)
	if !ok {
		t.Fatal("zero-size column must report ok=true")
	}
	if ptr != nil {
		t.Fatalf("zero-size column pointer must be nil; got %p", ptr)
	}

	values := tbl.RowValues(row)
	if _, ok := values[tagID].(ZeroTag); !ok {
		t.Fatalf("RowValues[tagID] = %+v, want ZeroTag{}", values[tagID])
	}
}
