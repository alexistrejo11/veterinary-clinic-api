package medical

import (
	"context"
	"testing"

	"clinic-vet-api/internal/shared/page"
)

type vaccineCatalogStub struct {
	item  VaccineCatalog
	saved []VaccineCatalog
}

func (s *vaccineCatalogStub) FindAll(context.Context, page.Pagination) (page.Page[VaccineCatalog], error) {
	return page.NewPage([]VaccineCatalog{s.item}, 1, page.PaginationRequest{Page: 1, PageSize: 10}), nil
}
func (s *vaccineCatalogStub) FindByID(_ context.Context, id VaccineCatalogID) (VaccineCatalog, error) {
	if s.item.ID.Value() == id.Value() {
		return s.item, nil
	}
	return VaccineCatalog{}, context.Canceled
}
func (s *vaccineCatalogStub) FindBySpecies(context.Context, string, page.Pagination) (page.Page[VaccineCatalog], error) {
	return page.Page[VaccineCatalog]{}, nil
}
func (s *vaccineCatalogStub) Save(_ context.Context, v VaccineCatalog) (VaccineCatalog, error) {
	if v.ID.Value() == 0 {
		v.ID = NewVaccineCatalogID(1)
	}
	s.item = v
	s.saved = append(s.saved, v)
	return v, nil
}
func (s *vaccineCatalogStub) DeleteByID(context.Context, VaccineCatalogID) error { return nil }

func TestVaccineCatalogServiceCreatesActiveVaccine(t *testing.T) {
	repo := &vaccineCatalogStub{}
	svc := NewVaccineCatalogApplicationService(repo)

	vaccine, err := svc.CreateVaccine(context.Background(), CreateVaccineCatalogCommand{Name: "Rabies", TotalDoses: 1})
	if err != nil {
		t.Fatalf("CreateVaccine() error = %v", err)
	}
	if !vaccine.IsActive || vaccine.ID.Value() != 1 {
		t.Fatalf("CreateVaccine() = %#v, want active persisted vaccine", vaccine)
	}
}

func TestVaccineCatalogServiceDeactivatesInsteadOfDeleting(t *testing.T) {
	repo := &vaccineCatalogStub{item: VaccineCatalog{ID: NewVaccineCatalogID(4), Name: "Rabies", TotalDoses: 1, IsActive: true}}
	svc := NewVaccineCatalogApplicationService(repo)

	if err := svc.DeactivateVaccine(context.Background(), NewVaccineCatalogID(4)); err != nil {
		t.Fatalf("DeactivateVaccine() error = %v", err)
	}
	if repo.item.IsActive {
		t.Fatal("DeactivateVaccine() left vaccine active")
	}
}
