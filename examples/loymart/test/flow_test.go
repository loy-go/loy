package test

import (
	"context"
	"testing"
	"time"

	"loymart/internal/notification/domain"
	"loymart/internal/notification/service"
	"loymart/internal/onuserregistered/listener"
	"loymart/internal/platform/tenant"
	"loymart/internal/rbac/policy"
	"loymart/internal/sendnotification/job"
	userDomain "loymart/internal/user/domain"
	"loymart/internal/userregistered/event"
)

type mockNotificationRepo struct {
	items []*domain.Notification
}

func (m *mockNotificationRepo) FindByID(ctx context.Context, id int64) (*domain.Notification, error) {
	for _, it := range m.items {
		if it.ID == id {
			return it, nil
		}
	}
	return nil, nil
}

func (m *mockNotificationRepo) FindAll(ctx context.Context, limit, offset int) ([]*domain.Notification, error) {
	return m.items, nil
}

func (m *mockNotificationRepo) Create(ctx context.Context, entity *domain.Notification) error {
	entity.ID = int64(len(m.items) + 1)
	m.items = append(m.items, entity)
	return nil
}

func (m *mockNotificationRepo) Update(ctx context.Context, entity *domain.Notification) error {
	return nil
}

func (m *mockNotificationRepo) Delete(ctx context.Context, id int64) error {
	return nil
}

func TestLoymart_SynchronousFlow(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc, err := service.NewService(repo)
	if err != nil {
		t.Fatalf("initializing service: %v", err)
	}

	ctx := tenant.WithTenantID(context.Background(), "org-123")
	tenantID, ok := tenant.GetTenantID(ctx)
	if !ok || tenantID != "org-123" {
		t.Fatalf("expected tenant org-123, got: %s", tenantID)
	}

	notif := &domain.Notification{
		Channel:   "email",
		Recipient: "user@example.com",
		Body:      "Welcome to Loymart!",
	}

	if err := svc.Create(ctx, notif); err != nil {
		t.Fatalf("creating notification: %v", err)
	}

	items, err := svc.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("listing notifications: %v", err)
	}
	if len(items) != 1 || items[0].Recipient != "user@example.com" {
		t.Errorf("expected 1 notification with recipient, got: %v", items)
	}
}

func TestLoymart_AsynchronousQueueFlow(t *testing.T) {
	ctx := context.Background()

	// 1. Trigger user registration event
	evt := event.UserRegisteredCreatedEvent{
		ID:        42,
		Timestamp: time.Now(),
	}

	lst := listener.NewOnUserRegisteredCreatedListener(nil)
	if err := lst.Handle(ctx, evt); err != nil {
		t.Fatalf("listener failed: %v", err)
	}

	// 2. Create and dispatch background Asynq notification job
	task, err := job.NewSendNotificationTask(evt.ID)
	if err != nil {
		t.Fatalf("creating notification task: %v", err)
	}

	processor := job.NewSendNotificationProcessor()
	if err := processor.ProcessTask(ctx, task); err != nil {
		t.Fatalf("processing background notification task: %v", err)
	}
}

func TestLoymart_RBACAuthorizationPolicy(t *testing.T) {
	p := policy.NewRbacPolicy()

	admin := &userDomain.User{ID: 1, Role: "admin"}
	member := &userDomain.User{ID: 2, Role: "member"}

	adminCtx := context.WithValue(context.Background(), policy.RoleContextKey, "admin")
	memberCtx := context.WithValue(context.Background(), policy.RoleContextKey, "member")

	// Admin can edit member
	if !p.CanEdit(adminCtx, admin.ID, member) {
		t.Errorf("expected admin to have edit permission")
	}

	// Member can edit self
	if !p.CanEdit(memberCtx, member.ID, member) {
		t.Errorf("expected member to edit self")
	}

	// Member cannot edit admin
	if p.CanEdit(memberCtx, member.ID, admin) {
		t.Errorf("did not expect member to edit admin")
	}

	// Only admin can delete
	if p.CanDelete(memberCtx, member.ID, member) {
		t.Errorf("did not expect member to delete")
	}
	if !p.CanDelete(adminCtx, admin.ID, member) {
		t.Errorf("expected admin to delete")
	}
}
